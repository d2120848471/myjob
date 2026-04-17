package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"myjob/internal/app"
	"myjob/internal/consts"
	openauth "myjob/internal/library/openauth"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OpenSignatureGuard 提供开放接口 Header 签名鉴权与防重放中间件。
type OpenSignatureGuard struct{ core *app.Core }

// NewOpenSignatureGuard 创建开放签名鉴权中间件。
func NewOpenSignatureGuard(core *app.Core) *OpenSignatureGuard { return &OpenSignatureGuard{core: core} }

// Require 返回一个中间件：校验 app_key、签名、时间戳、nonce，并将 caller 写入请求上下文。
func (g *OpenSignatureGuard) Require() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		appKey := strings.TrimSpace(r.GetHeader("X-App-Key"))
		timestamp := strings.TrimSpace(r.GetHeader("X-Timestamp"))
		nonce := strings.TrimSpace(r.GetHeader("X-Nonce"))
		signature := strings.TrimSpace(r.GetHeader("X-Signature"))
		if appKey == "" {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "app_key_invalid"))
			return
		}
		if timestamp == "" || nonce == "" || signature == "" {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "signature_invalid"))
			return
		}

		caller, ok, err := g.loadCaller(r.Context(), appKey)
		if err != nil {
			r.SetError(err)
			return
		}
		if !ok || caller.ID <= 0 {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "app_key_invalid"))
			return
		}
		if strings.TrimSpace(caller.Status) != "enabled" {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "caller_disabled"))
			return
		}

		ip := strings.TrimSpace(r.GetClientIp())
		if !ipAllowed(ip, strings.TrimSpace(caller.AllowedIPList)) {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "ip_not_allowed"))
			return
		}

		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || ts <= 0 {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "timestamp_expired"))
			return
		}
		skew := g.core.Config().Trade.OpenTimestampSkewSeconds
		if skew <= 0 {
			skew = 300
		}
		now := g.core.Now().Unix()
		if int64(math.Abs(float64(now-ts))) > int64(skew) {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "timestamp_expired"))
			return
		}

		if err := g.consumeNonce(r.Context(), appKey, nonce); err != nil {
			r.SetError(err)
			return
		}

		body := r.GetBody()
		if ok, verifyErr := verifyOpenV1Signature(
			strings.TrimSpace(caller.AppSecret),
			strings.TrimSpace(r.Method),
			strings.TrimSpace(r.URL.Path),
			timestamp,
			nonce,
			body,
			signature,
		); verifyErr != nil {
			r.SetError(verifyErr)
			return
		} else if !ok {
			r.SetError(gerror.NewCode(consts.CodeUnauthorized, "signature_invalid"))
			return
		}

		ctx := openauth.WithCaller(r.Context(), caller)
		r.SetCtx(ctx)
		r.Middleware.Next()
	}
}

func (g *OpenSignatureGuard) loadCaller(ctx context.Context, appKey string) (entity.OpenCaller, bool, error) {
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return entity.OpenCaller{}, false, nil
	}
	caller := entity.OpenCaller{}
	if err := g.core.DB().GetCore().GetScan(ctx, &caller, `
SELECT *
FROM open_caller
WHERE app_key = ? AND is_deleted = 0
ORDER BY id ASC
LIMIT 1
`, appKey); err != nil {
		return entity.OpenCaller{}, false, gerror.NewCode(consts.CodeInternalError, "caller加载失败")
	}
	if caller.ID <= 0 {
		return entity.OpenCaller{}, false, nil
	}
	return caller, true, nil
}

func (g *OpenSignatureGuard) consumeNonce(ctx context.Context, appKey, nonce string) error {
	nonce = strings.TrimSpace(nonce)
	if nonce == "" {
		return gerror.NewCode(consts.CodeUnauthorized, "nonce_replayed")
	}
	ttl := g.core.Config().Trade.OpenNonceTTLSeconds
	if ttl <= 0 {
		ttl = 600
	}
	seconds := int64(ttl)
	key := "open:nonce:" + strings.TrimSpace(appKey) + ":" + nonce
	value, err := g.core.Redis().GroupString().Set(ctx, key, "1", gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
		NX:        true,
	})
	if err != nil {
		return gerror.NewCode(consts.CodeInternalError, "nonce校验失败")
	}
	if value == nil || value.IsNil() {
		return gerror.NewCode(consts.CodeUnauthorized, "nonce_replayed")
	}
	return nil
}

func ipAllowed(ip string, allowedIPList string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false
	}
	if strings.TrimSpace(allowedIPList) == "" {
		return true
	}
	var allowed []string
	if err := json.Unmarshal([]byte(allowedIPList), &allowed); err != nil {
		return false
	}
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if strings.TrimSpace(item) == ip {
			return true
		}
	}
	return false
}

func verifyOpenV1Signature(appSecret, method, path, timestamp, nonce string, body []byte, providedSignature string) (bool, error) {
	appSecret = strings.TrimSpace(appSecret)
	providedSignature = strings.TrimSpace(providedSignature)
	if appSecret == "" || providedSignature == "" {
		return false, gerror.NewCode(consts.CodeUnauthorized, "signature_invalid")
	}

	providedBytes, err := hex.DecodeString(strings.ToLower(providedSignature))
	if err != nil {
		return false, gerror.NewCode(consts.CodeUnauthorized, "signature_invalid")
	}

	canonical := strings.ToUpper(strings.TrimSpace(method)) + "\n" +
		strings.TrimSpace(path) + "\n" +
		strings.TrimSpace(timestamp) + "\n" +
		strings.TrimSpace(nonce) + "\n" +
		sha256Hex(body)
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(canonical))
	expected := mac.Sum(nil)
	return hmac.Equal(providedBytes, expected), nil
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
