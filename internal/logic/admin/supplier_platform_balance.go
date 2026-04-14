package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/google/uuid"
)

func (l *SupplierPlatformLogic) RefreshBalance(ctx context.Context, req *adminapi.SupplierPlatformRefreshBalanceReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformRefreshBalanceRes, error) {
	account, err := l.getSupplierPlatform(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "平台账号不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号详情查询失败")
	}
	provider, ok := supplierprovider.Lookup(account.ProviderCode)
	if !ok {
		return nil, apiErr(consts.CodeBadRequest, "平台适配器不存在")
	}
	extraConfig, err := parseExtraConfig(account.ExtraConfig)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台扩展配置解析失败")
	}
	traceID := "spb_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	now := l.core.Now()
	result := refreshExecutionResult{
		Balance:           formatMoney(account.LastBalance.String),
		ConnectStatus:     2,
		ConnectStatusText: connectStatusText(2),
		Message:           "余额查询失败",
		RefreshedAt:       now.Format("2006-01-02 15:04:05"),
		TraceID:           traceID,
		RequestMethod:     "POST",
	}

	providerAccount := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  extraConfig,
	}
	candidates := provider.CandidateBaseURLs(providerAccount)
	// 同一次刷新复用同一个 client/transport，避免循环内反复创建并丢弃 transport
	// 导致废弃连接未清理、CDN 侧频控误判。
	client := l.httpClientForProvider(account.ProviderCode)
	var lastErr error
	for _, baseURL := range candidates {
		reqStart := time.Now()
		// 外部平台请求不要复用 ghttp 入站上下文，避免请求生命周期被框架侧提前结束。
		request, buildErr := provider.BuildRequest(context.Background(), providerAccount, now, baseURL)
		if buildErr != nil {
			return nil, apiErr(consts.CodeInternalError, "平台请求构建失败")
		}
		result.RequestURL = request.URL.String()
		result.RequestMethod = request.Method
		requestSnapshot, snapshotErr := snapshotRequest(request, account)
		if snapshotErr != nil {
			return nil, apiErr(consts.CodeInternalError, "平台请求快照生成失败")
		}
		result.RequestSnapshot = requestSnapshot

		response, requestErr := client.Do(request)
		if requestErr != nil {
			lastErr = requestErr
			result.Message = requestErr.Error()
			result.DurationMS = int(time.Since(reqStart).Milliseconds())
			result.ResponseSnapshot = truncateSnapshot(result.Message)
			continue
		}

		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return nil, apiErr(consts.CodeInternalError, "平台响应读取失败")
		}
		result.HTTPStatus = response.StatusCode
		result.DurationMS = int(time.Since(reqStart).Milliseconds())
		result.ResponseSnapshot = truncateSnapshot(sanitizeSnapshot(string(body), account))

		amount, message, parseErr := provider.ParseBalanceResponse(response.StatusCode, body)
		if parseErr != nil {
			// 部分平台配置域名的 https 入口会返回前端首页，识别后继续降级到同域名 http。
			if shouldRetrySupplierCandidate(request, response, body, parseErr) {
				lastErr = parseErr
				result.Message = parseErr.Error()
				continue
			}
			if strings.TrimSpace(message) == "" {
				message = parseErr.Error()
			}
			result.Message = message
			break
		}

		result.Success = true
		result.Balance = amount.StringFixed(4)
		result.BalanceAmount = amount.StringFixed(4)
		result.ConnectStatus = 1
		result.ConnectStatusText = connectStatusText(1)
		if strings.TrimSpace(message) == "" {
			message = "查询成功"
		}
		result.Message = message
		break
	}

	if !result.Success && lastErr != nil && strings.TrimSpace(result.Message) == "" {
		result.Message = lastErr.Error()
	}
	if strings.TrimSpace(result.Balance) == "" {
		result.Balance = formatMoney(account.LastBalance.String)
	}
	if strings.TrimSpace(result.Message) == "" {
		result.Message = "余额查询失败"
	}

	if err = l.persistRefreshResult(ctx, account, actor, result); err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台余额刷新结果保存失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("刷新第三方对接余额：%d -> %s -> %s", req.ID, account.ProviderName, result.Message), ip)
	return &adminapi.SupplierPlatformRefreshBalanceRes{
		ID:                account.ID,
		Balance:           result.Balance,
		ConnectStatus:     result.ConnectStatus,
		ConnectStatusText: result.ConnectStatusText,
		Message:           result.Message,
		RefreshedAt:       result.RefreshedAt,
		TraceID:           result.TraceID,
	}, nil
}

func shouldRetrySupplierCandidate(request *http.Request, response *http.Response, body []byte, parseErr error) bool {
	if parseErr == nil || request == nil || request.URL == nil || response == nil {
		return false
	}
	// HTTP-first 顺序下，无论 HTTP 还是 HTTPS，收到 HTML（CDN 默认页/前端首页）
	// 而非 JSON API 响应时，都应继续尝试下一个候选地址。
	contentType := strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Type")))
	trimmedBody := strings.ToLower(strings.TrimSpace(string(body)))
	if strings.Contains(contentType, "text/html") {
		return true
	}
	return strings.HasPrefix(trimmedBody, "<!doctype html") || strings.HasPrefix(trimmedBody, "<html")
}

func (l *SupplierPlatformLogic) httpClientForProvider(providerCode string) *http.Client {
	if providerCode != "kakayun" {
		return l.httpClient
	}
	client := *l.httpClient
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := l.httpClient.Transport.(*http.Transport); ok && baseTransport != nil {
		transport = baseTransport.Clone()
	}
	transport.DisableCompression = true
	client.Transport = transport
	return &client
}

type refreshExecutionResult struct {
	Success           bool
	Balance           string
	BalanceAmount     string
	ConnectStatus     int
	ConnectStatusText string
	Message           string
	RefreshedAt       string
	TraceID           string
	RequestURL        string
	RequestMethod     string
	RequestSnapshot   string
	ResponseSnapshot  string
	HTTPStatus        int
	DurationMS        int
}

func (l *SupplierPlatformLogic) persistRefreshResult(ctx context.Context, account entity.SupplierPlatformAccount, actor entity.AdminUser, result refreshExecutionResult) error {
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		now := l.core.Now()
		balanceValue := any(nil)
		if strings.TrimSpace(result.BalanceAmount) != "" {
			balanceValue = result.BalanceAmount
		}
		if _, err := tx.Exec(`
INSERT INTO supplier_platform_balance_log (
    platform_id, operator_id, operator_name, provider_code, request_url, request_method,
    request_snapshot, response_snapshot, http_status, success, balance_amount, message,
    duration_ms, trace_id, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, account.ID, actor.ID, actor.Username, account.ProviderCode, result.RequestURL, result.RequestMethod, result.RequestSnapshot, result.ResponseSnapshot, result.HTTPStatus, boolToInt(result.Success), balanceValue, result.Message, result.DurationMS, result.TraceID, now); err != nil {
			return err
		}

		if result.Success {
			_, err := tx.Exec(`
UPDATE supplier_platform_account
SET last_balance = ?, last_balance_status = 1, last_balance_message = ?, last_balance_at = ?, last_balance_trace_id = ?, updated_at = ?
WHERE id = ?
`, result.BalanceAmount, result.Message, now, result.TraceID, now, account.ID)
			return err
		}
		_, err := tx.Exec(`
UPDATE supplier_platform_account
SET last_balance_status = 2, last_balance_message = ?, last_balance_at = ?, last_balance_trace_id = ?, updated_at = ?
WHERE id = ?
`, result.Message, now, result.TraceID, now, account.ID)
		return err
	})
}

func snapshotRequest(request *http.Request, account entity.SupplierPlatformAccount) (string, error) {
	body := ""
	if request.GetBody != nil {
		reader, err := request.GetBody()
		if err != nil {
			return "", err
		}
		raw, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return "", err
		}
		body = string(raw)
	}
	headers := make(map[string][]string, len(request.Header))
	for key, values := range request.Header {
		copied := make([]string, len(values))
		copy(copied, values)
		headers[key] = copied
	}
	payload := map[string]any{
		"url":     request.URL.String(),
		"method":  request.Method,
		"headers": headers,
		"body":    body,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return truncateSnapshot(sanitizeSnapshot(string(raw), account)), nil
}

func sanitizeSnapshot(value string, account entity.SupplierPlatformAccount) string {
	value = strings.ReplaceAll(value, account.SecretKey, app.MaskSecret(account.SecretKey))
	value = strings.ReplaceAll(value, account.TokenID, app.MaskSecret(account.TokenID))
	return value
}

func truncateSnapshot(value string) string {
	if len(value) <= 4096 {
		return value
	}
	return value[:4096]
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
