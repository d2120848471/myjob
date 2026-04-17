package tradelogic

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"strings"
	"time"

	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

type createAttemptRow struct {
	AttemptID              int64  `db:"attempt_id"`
	OrderID                int64  `db:"order_id"`
	OrderNo                string `db:"order_no"`
	GoodsID                int64  `db:"goods_id"`
	GoodsCode              string `db:"goods_code"`
	OrderQuantity          int    `db:"order_quantity"`
	PlatformAccountID      int64  `db:"platform_account_id"`
	ProviderCode           string `db:"provider_code"`
	ProviderRequestOrderNo string `db:"provider_request_order_no"`
	FulfillmentNo          string `db:"fulfillment_no"`
	AttemptQuantity        int    `db:"attempt_quantity"`
	AttemptNo              int    `db:"attempt_no"`
	SupplierGoodsNo        string `db:"supplier_goods_no"`
	PayloadJSON            string `db:"payload_json"`
	SalePrice              string `db:"sale_price"`
}

func (l *TradeOrderLogic) loadCreateAttemptRow(ctx context.Context, db gdb.DB, attemptID int64) (createAttemptRow, error) {
	row := createAttemptRow{}
	if err := db.GetCore().GetScan(ctx, &row, `
SELECT
    a.id AS attempt_id,
    a.order_id,
    o.order_no,
    o.goods_id,
    o.goods_code_snapshot AS goods_code,
    o.quantity AS order_quantity,
    a.platform_account_id,
    a.provider_code,
    a.provider_request_order_no,
    a.fulfillment_no,
    a.attempt_quantity,
    a.attempt_no,
    a.binding_supplier_goods_no_snapshot AS supplier_goods_no,
    o.payload_json,
    o.sale_price
FROM trade_order_attempt a
JOIN trade_order o ON o.id = a.order_id
WHERE a.id = ?
`, attemptID); err != nil {
		return createAttemptRow{}, apiErr(consts.CodeInternalError, "读取attempt失败")
	}
	if row.AttemptID <= 0 {
		return createAttemptRow{}, apiErr(consts.CodeBadRequest, "attempt不存在")
	}
	return row, nil
}

func (l *TradeOrderLogic) loadPlatformAccount(ctx context.Context, db gdb.DB, accountID int64) (entity.SupplierPlatformAccount, error) {
	account := entity.SupplierPlatformAccount{}
	if err := db.GetCore().GetScan(ctx, &account, `SELECT * FROM supplier_platform_account WHERE id = ? AND is_deleted = 0`, accountID); err != nil {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeInternalError, "读取渠道账号失败")
	}
	if account.ID <= 0 {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	return account, nil
}

func decodeJSONMap(value string) map[string]any {
	result := make(map[string]any)
	_ = json.Unmarshal([]byte(value), &result)
	return result
}

func (l *TradeOrderLogic) executeCreateAttempt(ctx context.Context, attemptID int64, traceID string) error {
	row, err := l.loadCreateAttemptRow(ctx, l.core.DB(), attemptID)
	if err != nil {
		return err
	}

	orderProvider, ok := l.lookupOrderProvider(strings.TrimSpace(row.ProviderCode))
	if !ok {
		return apiErr(consts.CodeBadRequest, "provider不支持下单")
	}

	account, err := l.loadPlatformAccount(ctx, l.core.DB(), row.PlatformAccountID)
	if err != nil {
		return err
	}
	extraConfig := decodeJSONMap(strings.TrimSpace(account.ExtraConfig))
	accountCfg := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  extraConfig,
	}

	payload := decodeJSONMap(row.PayloadJSON)
	createInput := supplierprovider.CreateOrderInput{
		ProviderRequestOrderNo: row.ProviderRequestOrderNo,
		SupplierGoodsNo:        strings.TrimSpace(row.SupplierGoodsNo),
		Quantity:               row.AttemptQuantity,
		Payload:                payload,
	}

	now := l.core.Now()
	candidates := orderProvider.CandidateBaseURLs(accountCfg)
	var lastErr error
	for _, baseURL := range candidates {
		reqStart := time.Now()
		request, buildErr := orderProvider.BuildCreateOrderRequest(context.Background(), accountCfg, createInput, baseURL)
		if buildErr != nil {
			return apiErr(consts.CodeInternalError, "平台请求构建失败")
		}

		headersSnapshot, snapshotErr := snapshotHeaders(request.Header, account.TokenID, account.SecretKey)
		if snapshotErr != nil {
			return apiErr(consts.CodeInternalError, "平台请求快照生成失败")
		}
		bodySnapshot, snapshotErr := snapshotRequestBody(request, account.TokenID, account.SecretKey)
		if snapshotErr != nil {
			return apiErr(consts.CodeInternalError, "平台请求快照生成失败")
		}

		_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = 'submitted',
    request_url = ?,
    request_method = ?,
    request_headers = ?,
    request_payload = ?,
    trace_id = ?,
    updated_at = ?
WHERE id = ?
`, request.URL.String(), request.Method, headersSnapshot, bodySnapshot, traceID, now, attemptID)

		response, requestErr := l.client.Do(request)
		if requestErr != nil {
			lastErr = requestErr
			category := "server_error"
			if ne, ok := requestErr.(net.Error); ok && ne.Timeout() {
				category = "timeout"
			}
			_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = 'unknown',
    error_category = ?,
    error_message = ?,
    duration_ms = ?,
    updated_at = ?
WHERE id = ?
`, category, truncateSnapshot(requestErr.Error()), int(time.Since(reqStart).Milliseconds()), l.core.Now(), attemptID)
			continue
		}

		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return apiErr(consts.CodeInternalError, "平台响应读取失败")
		}

		responseSnapshot := snapshotResponseBody(body, account.TokenID, account.SecretKey)
		durationMS := int(time.Since(reqStart).Milliseconds())

		result, parseErr := orderProvider.ParseCreateOrderResponse(response.StatusCode, body)
		if parseErr != nil {
			// HTML 首页等场景继续尝试下一个候选。
			if shouldRetrySupplierCandidate(response, body, parseErr) {
				lastErr = parseErr
				continue
			}
			_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = 'unknown',
    response_payload = ?,
    http_status = ?,
    duration_ms = ?,
    error_category = 'unknown',
    error_message = ?,
    updated_at = ?
WHERE id = ?
`, responseSnapshot, response.StatusCode, durationMS, truncateSnapshot(parseErr.Error()), l.core.Now(), attemptID)
			return nil
		}

		attemptStatus := "unknown"
		if result.Accepted {
			attemptStatus = "accepted"
		} else if result.FinalSuccess {
			attemptStatus = "success"
		} else if result.FinalFailed {
			attemptStatus = "failed"
		} else if result.Uncertain {
			attemptStatus = "unknown"
		}

		nextQueryAt, queryDeadlineAt := l.computeQuerySchedule(ctx, row.GoodsID, now)
		finishedAt := time.Time{}
		if attemptStatus == "success" || attemptStatus == "failed" {
			finishedAt = now
			nextQueryAt = time.Time{}
			queryDeadlineAt = time.Time{}
		}

		_, _ = l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = ?,
    upstream_status = ?,
    channel_order_no = ?,
    response_payload = ?,
    http_status = ?,
    duration_ms = ?,
    error_category = ?,
    error_code = ?,
    error_message = ?,
    next_query_at = ?,
    query_deadline_at = ?,
    finished_at = ?,
    updated_at = ?
WHERE id = ?
`,
			attemptStatus,
			truncateSnapshot(strings.TrimSpace(result.UpstreamStatus)),
			truncateSnapshot(strings.TrimSpace(result.ChannelOrderNo)),
			responseSnapshot,
			response.StatusCode,
			durationMS,
			truncateSnapshot(strings.TrimSpace(result.ErrorCategory)),
			truncateSnapshot(strings.TrimSpace(result.ErrorCode)),
			truncateSnapshot(strings.TrimSpace(result.ErrorMessage)),
			nullableTimeArg(nextQueryAt),
			nullableTimeArg(queryDeadlineAt),
			nullableTimeArg(finishedAt),
			l.core.Now(),
			attemptID,
		)

		// 同步明确成功：直接推进主订单数量与状态。
		if attemptStatus == "success" {
			_ = l.markOrderSuccess(ctx, row.OrderID, row.AttemptQuantity, result.ChannelOrderNo, now)
			_ = l.tryKickoffNextCreatedAttempt(ctx, row.OrderID, traceID)
		}
		// 同步明确失败：若开启智能补单，则切换下一条绑定继续建单。
		if attemptStatus == "failed" {
			if nextID, ok, _ := l.tryReplenishAfterFailedAttempt(ctx, row, traceID); ok && nextID > 0 {
				_ = l.executeCreateAttempt(ctx, nextID, traceID)
			} else {
				reason := strings.TrimSpace(result.ErrorCategory)
				if reason == "" {
					reason = "all_bindings_failed"
				}
				_ = l.markOrderFailed(ctx, row.OrderID, row.AttemptQuantity, reason, now)
				_ = l.tryKickoffNextCreatedAttempt(ctx, row.OrderID, traceID)
			}
		}
		return nil
	}

	if lastErr != nil {
		_ = lastErr
	}
	return nil
}

func (l *TradeOrderLogic) markOrderSuccess(ctx context.Context, orderID int64, attemptQuantity int, channelOrderNo string, now time.Time) error {
	if orderID <= 0 || attemptQuantity <= 0 {
		return nil
	}
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.GetOne(`SELECT quantity, success_quantity, failed_quantity, status, failure_reason FROM trade_order WHERE id = ?`, orderID)
		if err != nil {
			return err
		}
		if record == nil || len(record) == 0 {
			return nil
		}
		status := strings.TrimSpace(record["status"].String())
		if status == "success" || status == "failed" || status == "manual_review" {
			return nil
		}

		total := record["quantity"].Int()
		currentSuccess := record["success_quantity"].Int()
		currentFailed := record["failed_quantity"].Int()
		newSuccess := currentSuccess + attemptQuantity
		failureReason := strings.TrimSpace(record["failure_reason"].String())

		finishedAt := time.Time{}
		nextStatus := "processing"
		if total > 0 {
			if newSuccess >= total {
				nextStatus = "success"
				finishedAt = now
			} else if newSuccess+currentFailed >= total {
				nextStatus = "manual_review"
				finishedAt = now
				if failureReason == "" {
					failureReason = "partial_success_need_review"
				}
			}
		}

		_, err = tx.Exec(`
UPDATE trade_order
SET success_quantity = ?,
    status = ?,
    channel_order_no = CASE WHEN ? != '' THEN ? ELSE channel_order_no END,
    failure_reason = CASE WHEN ? != '' THEN ? ELSE failure_reason END,
    finished_at = ?,
    updated_at = ?
WHERE id = ?
`, newSuccess, nextStatus, strings.TrimSpace(channelOrderNo), strings.TrimSpace(channelOrderNo), failureReason, failureReason, nullableTimeArg(finishedAt), now, orderID)
		return err
	})
}

func (l *TradeOrderLogic) computeQuerySchedule(ctx context.Context, goodsID int64, now time.Time) (time.Time, time.Time) {
	interval := l.core.Config().Trade.AttemptQueryScanIntervalSeconds
	if interval <= 0 {
		interval = 30
	}
	next := now.Add(time.Duration(interval) * time.Second)

	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT attempt_timeout_enabled, attempt_timeout_minutes
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil || record == nil || len(record) == 0 {
		return next, time.Time{}
	}
	if record["attempt_timeout_enabled"].Int() == 0 {
		return next, time.Time{}
	}
	minutes := record["attempt_timeout_minutes"].Int()
	if minutes <= 0 {
		return next, time.Time{}
	}
	return next, now.Add(time.Duration(minutes) * time.Minute)
}
