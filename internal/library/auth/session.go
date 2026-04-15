package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims 是后台访问令牌的 JWT Claims，包含用户标识与会话版本信息。
type Claims struct {
	UserID       int64 `json:"user_id"`
	GroupID      int64 `json:"group_id"`
	TokenVersion int   `json:"token_version"`
	jwt.RegisteredClaims
}

// IssueToken 签发一个 JWT 访问令牌，并返回 tokenString 与 jti。
//
// jti 同时作为服务端 Session 的主键（用于 Redis 存储与主动失效）。
func IssueToken(secret string, payload modelruntime.SessionPayload, now time.Time, ttl time.Duration) (string, string, error) {
	jti := uuid.NewString()
	expiresAt := now.Add(ttl)
	claims := Claims{
		UserID:       payload.UserID,
		GroupID:      payload.GroupID,
		TokenVersion: payload.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	return tokenString, jti, nil
}

// ParseToken 解析并校验 JWT token（签名/过期/格式）。
func ParseToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// SaveSession 将会话负载保存到 Redis：
// - admin:session:<jti> 保存 JSON payload（带 TTL）
// - admin:user:sessions:<uid> 维护用户当前活跃会话集合（用于批量踢下线）
func SaveSession(ctx context.Context, client *gredis.Redis, payload modelruntime.SessionPayload, ttl time.Duration) error {
	data, _ := json.Marshal(payload)
	seconds := int64(math.Ceil(ttl.Seconds()))
	if _, err := client.GroupString().Set(ctx, SessionKey(payload.JTI), string(data), gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
	}); err != nil {
		return err
	}
	if _, err := client.GroupSet().SAdd(ctx, UserSessionsKey(payload.UserID), payload.JTI); err != nil {
		return err
	}
	_, _ = client.GroupGeneric().Expire(ctx, UserSessionsKey(payload.UserID), seconds)
	return nil
}

// SessionKey 返回单个会话的 Redis key。
func SessionKey(jti string) string {
	return "admin:session:" + jti
}

// UserSessionsKey 返回用户会话集合的 Redis key。
func UserSessionsKey(userID int64) string {
	return fmt.Sprintf("admin:user:sessions:%d", userID)
}

// TempLoginKey 返回短信二次验证临时登录态的 Redis key。
func TempLoginKey(loginToken string) string {
	return "admin:login:tmp:" + loginToken
}

// SMSCodeKey 返回短信验证码缓存的 Redis key。
func SMSCodeKey(userID int64) string {
	return fmt.Sprintf("sms:login:%d", userID)
}

// SMSSendLockKey 返回短信发送频控锁的 Redis key。
func SMSSendLockKey(userID int64) string {
	return fmt.Sprintf("sms:login:send_lock:%d", userID)
}

// PermissionCacheKey 返回用户组权限码列表缓存的 Redis key。
func PermissionCacheKey(groupID int64) string {
	return fmt.Sprintf("admin:perm:group:%d", groupID)
}

// SMSConfigCacheKey 返回短信配置缓存的 Redis key。
func SMSConfigCacheKey() string {
	return "admin:config:sms"
}

// SystemConfigCacheKey 返回系统参数分组缓存的 Redis key。
func SystemConfigCacheKey(group string) string {
	return fmt.Sprintf("admin:config:system:%s", group)
}
