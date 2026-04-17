package tradelogic

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
)

// HandleProviderOrderCallback 处理上游订单回调：验签、幂等写日志、定位 attempt，并推进 attempt 与主订单状态。
//
// 注意：
// - Provider 回调 ACK 往往不是统一 JSON 包装，因此由 controller 直接写回本返回值。
func (l *TradeOrderLogic) HandleProviderOrderCallback(ctx context.Context, providerCode string, headers http.Header, body []byte) ([]byte, string, error) {
	providerCode = strings.TrimSpace(strings.ToLower(providerCode))
	if providerCode == "" {
		return nil, "", apiErr(consts.CodeBadRequest, "provider_code不能为空")
	}

	cbProvider, ok := supplierprovider.LookupCallback(providerCode)
	if !ok {
		return nil, "", apiErr(consts.CodeBadRequest, "provider不支持回调")
	}

	account, err := l.loadPlatformAccountForCallback(ctx, providerCode, body)
	if err != nil {
		return nil, "", err
	}
	accountCfg := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  decodeJSONMap(strings.TrimSpace(account.ExtraConfig)),
	}

	verifyErr := cbProvider.VerifyCallbackSignature(accountCfg, headers, body)
	verifyResult := "ok"
	if verifyErr != nil {
		verifyResult = "failed"
	}

	result, parseErr := cbProvider.ParseCallbackPayload(accountCfg, headers, body)
	if parseErr != nil {
		return nil, "", apiErr(consts.CodeBadRequest, "回调解析失败")
	}
	if result == nil {
		return nil, "", apiErr(consts.CodeBadRequest, "回调解析失败")
	}
	if strings.TrimSpace(result.IdempotencyKey) == "" {
		// 没有稳定幂等键时，用请求体摘要兜底。
		result.IdempotencyKey = fmt.Sprintf("%x", sha256Sum(body))
	}

	ackBody, contentType, ackErr := cbProvider.BuildCallbackAck(supplierprovider.CallbackAckInput{
		ProviderRequestOrderNo: strings.TrimSpace(result.ProviderRequestOrderNo),
		ChannelOrderNo:         strings.TrimSpace(result.ChannelOrderNo),
		FinalSuccess:           result.FinalSuccess,
		FinalFailed:            result.FinalFailed,
	})
	if ackErr != nil {
		return nil, "", apiErr(consts.CodeInternalError, "回调ACK构建失败")
	}

	now := l.core.Now()
	traceID := fmt.Sprintf("callback-%d", now.UnixNano())
	headersSnapshot, _ := snapshotHeaders(headers, account.TokenID, account.SecretKey)
	bodySnapshot := truncateSnapshot(sanitizeSnapshot(string(body), account.TokenID, account.SecretKey))
	ackSnapshot := truncateSnapshot(string(ackBody))

	processResult := "processed"
	if verifyErr != nil {
		processResult = "verify_failed"
	}

	// 写回调日志：幂等由 (provider_code, idempotency_key) 保证。
	inserted, err := l.core.DB().Exec(ctx, `
INSERT INTO provider_callback_log (
    provider_code, platform_account_id, idempotency_key,
    provider_request_order_no, channel_order_no,
    request_headers, request_body,
    verify_result, process_result, ack_body,
    created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		providerCode,
		account.ID,
		strings.TrimSpace(result.IdempotencyKey),
		strings.TrimSpace(result.ProviderRequestOrderNo),
		strings.TrimSpace(result.ChannelOrderNo),
		headersSnapshot,
		bodySnapshot,
		verifyResult,
		processResult,
		ackSnapshot,
		now,
	)
	if err != nil {
		// 幂等冲突：认为已处理过，直接返回 ACK，避免上游重试风暴。
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return ackBody, contentType, nil
		}
		return nil, "", apiErr(consts.CodeInternalError, "回调日志写入失败")
	}
	logID, _ := inserted.LastInsertId()

	if verifyErr != nil {
		return ackBody, contentType, gerror.WrapCode(consts.CodeBadRequest, verifyErr, "回调验签失败")
	}

	attemptID, orderID, attemptQuantity, locateErr := l.locateAttemptForCallback(ctx, result)
	if locateErr != nil {
		_ = l.updateCallbackLogProcessResult(ctx, logID, "attempt_not_found")
		return ackBody, contentType, nil
	}

	if ignored, ignoreErr := l.ignoreLateCallbackIfNeeded(ctx, logID, attemptID, orderID, result, account, now); ignoreErr == nil && ignored {
		return ackBody, contentType, nil
	}

	attemptStatus := ""
	if result.FinalSuccess {
		attemptStatus = "success"
	} else if result.FinalFailed {
		attemptStatus = "failed"
	}

	finishedAt := time.Time{}
	if attemptStatus == "success" || attemptStatus == "failed" {
		finishedAt = now
	}

	_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET callback_payload = ?,
    callback_received_at = ?,
    callback_processed_at = ?,
    attempt_status = CASE WHEN ? != '' THEN ? ELSE attempt_status END,
    upstream_status = ?,
    channel_order_no = CASE WHEN ? != '' THEN ? ELSE channel_order_no END,
    finished_at = ?,
    updated_at = ?
WHERE id = ?
`,
		truncateSnapshot(sanitizeSnapshot(result.RawPayload, account.TokenID, account.SecretKey)),
		now,
		now,
		attemptStatus,
		attemptStatus,
		truncateSnapshot(strings.TrimSpace(result.UpstreamStatus)),
		strings.TrimSpace(result.ChannelOrderNo),
		strings.TrimSpace(result.ChannelOrderNo),
		nullableTimeArg(finishedAt),
		now,
		attemptID,
	)

	if attemptStatus == "success" {
		_ = l.markOrderSuccess(ctx, orderID, attemptQuantity, result.ChannelOrderNo, now)
		_ = l.tryKickoffNextCreatedAttempt(ctx, orderID, traceID)
		return ackBody, contentType, nil
	}
	if attemptStatus == "failed" {
		createRow, loadErr := l.loadCreateAttemptRow(ctx, l.core.DB(), attemptID)
		if loadErr == nil {
			if nextID, ok, _ := l.tryReplenishAfterFailedAttempt(ctx, createRow, traceID); ok && nextID > 0 {
				_ = l.executeCreateAttempt(ctx, nextID, traceID)
			} else {
				_ = l.markOrderFailed(ctx, createRow.OrderID, createRow.AttemptQuantity, "all_bindings_failed", now)
				_ = l.tryKickoffNextCreatedAttempt(ctx, createRow.OrderID, traceID)
			}
		}
	}
	return ackBody, contentType, nil
}

func (l *TradeOrderLogic) updateCallbackLogProcessResult(ctx context.Context, logID int64, processResult string) error {
	if logID <= 0 || strings.TrimSpace(processResult) == "" {
		return nil
	}
	_, err := l.core.DB().Exec(ctx, `UPDATE provider_callback_log SET process_result = ? WHERE id = ?`, strings.TrimSpace(processResult), logID)
	return err
}

func (l *TradeOrderLogic) ignoreLateCallbackIfNeeded(ctx context.Context, logID, attemptID, orderID int64, result *supplierprovider.CallbackResult, account entity.SupplierPlatformAccount, now time.Time) (bool, error) {
	if attemptID <= 0 || orderID <= 0 || result == nil {
		return false, nil
	}
	if !result.FinalSuccess && !result.FinalFailed {
		return false, nil
	}

	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT attempt_status, fulfillment_no
FROM trade_order_attempt
WHERE id = ?
LIMIT 1
`, attemptID)
	if err != nil || record == nil || len(record) == 0 {
		return false, err
	}
	attemptStatus := strings.TrimSpace(record["attempt_status"].String())
	fulfillmentNo := strings.TrimSpace(record["fulfillment_no"].String())
	if fulfillmentNo == "" {
		return false, nil
	}

	// 已经有“成功 attempt”兜底的前提下，timeout/failed attempt 的成功回调视为晚到，不允许覆盖主订单与履约归属。
	if attemptStatus == "timeout" || attemptStatus == "failed" {
		successCount, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM trade_order_attempt
WHERE order_id = ? AND fulfillment_no = ? AND attempt_status = 'success' AND id != ?
`, orderID, fulfillmentNo, attemptID)
		if err == nil && successCount.Int() > 0 {
			processResult := "late_callback_ignored"
			if result.FinalSuccess {
				processResult = "late_success_ignored"
			} else if result.FinalFailed {
				processResult = "late_failed_ignored"
			}
			_ = l.updateCallbackLogProcessResult(ctx, logID, processResult)
			_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET callback_payload = ?,
    callback_received_at = ?,
    callback_processed_at = ?,
    upstream_status = ?,
    channel_order_no = CASE WHEN ? != '' THEN ? ELSE channel_order_no END,
    updated_at = ?
WHERE id = ?
`,
				truncateSnapshot(sanitizeSnapshot(result.RawPayload, account.TokenID, account.SecretKey)),
				now,
				now,
				truncateSnapshot(strings.TrimSpace(result.UpstreamStatus)),
				strings.TrimSpace(result.ChannelOrderNo),
				strings.TrimSpace(result.ChannelOrderNo),
				now,
				attemptID,
			)
			return true, nil
		}
	}
	return false, nil
}

func (l *TradeOrderLogic) loadPlatformAccountForCallback(ctx context.Context, providerCode string, body []byte) (entity.SupplierPlatformAccount, error) {
	locator := extractCallbackAccountLocator(providerCode, body)
	query := `SELECT * FROM supplier_platform_account WHERE provider_code = ? AND is_deleted = 0 ORDER BY id ASC LIMIT 1`
	args := []any{providerCode}
	if strings.TrimSpace(locator) != "" {
		query = `SELECT * FROM supplier_platform_account WHERE provider_code = ? AND token_id = ? AND is_deleted = 0 ORDER BY id ASC LIMIT 1`
		args = []any{providerCode, strings.TrimSpace(locator)}
	}
	account := entity.SupplierPlatformAccount{}
	if err := l.core.DB().GetCore().GetScan(ctx, &account, query, args...); err != nil {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeInternalError, "读取渠道账号失败")
	}
	if account.ID <= 0 {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	return account, nil
}

func extractCallbackAccountLocator(providerCode string, body []byte) string {
	providerCode = strings.TrimSpace(strings.ToLower(providerCode))
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return ""
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	switch providerCode {
	case "youkayun":
		return strings.TrimSpace(fmt.Sprint(payload["userid"]))
	case "xingquanyi":
		return strings.TrimSpace(fmt.Sprint(payload["customer_id"]))
	default:
		return ""
	}
}

func (l *TradeOrderLogic) locateAttemptForCallback(ctx context.Context, result *supplierprovider.CallbackResult) (int64, int64, int, error) {
	providerRequestOrderNo := strings.TrimSpace(result.ProviderRequestOrderNo)
	channelOrderNo := strings.TrimSpace(result.ChannelOrderNo)

	if providerRequestOrderNo != "" {
		record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT id, order_id, attempt_quantity
FROM trade_order_attempt
WHERE provider_request_order_no = ?
LIMIT 1
`, providerRequestOrderNo)
		if err == nil && record != nil && len(record) > 0 {
			return record["id"].Int64(), record["order_id"].Int64(), record["attempt_quantity"].Int(), nil
		}
	}

	if channelOrderNo != "" {
		record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT id, order_id, attempt_quantity
FROM trade_order_attempt
WHERE channel_order_no = ?
ORDER BY id DESC
LIMIT 1
`, channelOrderNo)
		if err == nil && record != nil && len(record) > 0 {
			return record["id"].Int64(), record["order_id"].Int64(), record["attempt_quantity"].Int(), nil
		}
	}
	return 0, 0, 0, fmt.Errorf("attempt not found")
}

func sha256Sum(body []byte) [32]byte {
	return sha256.Sum256(body)
}
