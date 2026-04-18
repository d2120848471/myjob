package tradelogic

import (
	"context"
	"io"
	"net"
	"strings"
	"time"

	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"

	"github.com/gogf/gf/v2/database/gdb"
)

type dueAttemptRow struct {
	ID int64 `db:"id"`
}

// RunQueryJob 执行一次“异步查单”扫描与推进（只处理 next_query_at <= now 的 attempt）。
func (l *TradeOrderLogic) RunQueryJob(ctx context.Context, traceID string) (int, error) {
	now := l.core.Now()
	limit := l.core.Config().Trade.AttemptQueryBatchSize
	if limit <= 0 {
		limit = 100
	}

	rows := make([]dueAttemptRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id
FROM trade_order_attempt
WHERE attempt_status IN ('accepted','waiting_callback','querying','unknown')
  AND next_query_at IS NOT NULL
  AND next_query_at <= ?
ORDER BY next_query_at ASC, id ASC
LIMIT ?
`, now, limit); err != nil {
		return 0, apiErr(consts.CodeInternalError, "扫描查单任务失败")
	}

	processed := 0
	for _, row := range rows {
		if row.ID <= 0 {
			continue
		}
		if err := l.executeQueryAttempt(ctx, row.ID, traceID); err != nil {
			// 单条失败不影响下一条，避免任务整体中断。
			continue
		}
		processed++
	}
	return processed, nil
}

type queryAttemptRow struct {
	AttemptID              int64     `db:"attempt_id"`
	OrderID                int64     `db:"order_id"`
	GoodsID                int64     `db:"goods_id"`
	PlatformAccountID      int64     `db:"platform_account_id"`
	ProviderCode           string    `db:"provider_code"`
	ProviderRequestOrderNo string    `db:"provider_request_order_no"`
	ChannelOrderNo         string    `db:"channel_order_no"`
	AttemptQuantity        int       `db:"attempt_quantity"`
	AttemptStatus          string    `db:"attempt_status"`
	QueryCount             int       `db:"query_count"`
	QueryDeadlineAt        time.Time `db:"query_deadline_at"`
}

func (l *TradeOrderLogic) loadQueryAttemptRow(ctx context.Context, db gdb.DB, attemptID int64) (queryAttemptRow, error) {
	row := queryAttemptRow{}
	if err := db.GetCore().GetScan(ctx, &row, `
SELECT
    a.id AS attempt_id,
    a.order_id,
    o.goods_id,
    a.platform_account_id,
    a.provider_code,
    a.provider_request_order_no,
    a.channel_order_no,
    a.attempt_quantity,
    a.attempt_status,
    a.query_count,
    a.query_deadline_at
FROM trade_order_attempt a
JOIN trade_order o ON o.id = a.order_id
WHERE a.id = ?
`, attemptID); err != nil {
		return queryAttemptRow{}, apiErr(consts.CodeInternalError, "读取attempt失败")
	}
	if row.AttemptID <= 0 {
		return queryAttemptRow{}, apiErr(consts.CodeBadRequest, "attempt不存在")
	}
	return row, nil
}

func (l *TradeOrderLogic) executeQueryAttempt(ctx context.Context, attemptID int64, traceID string) error {
	row, err := l.loadQueryAttemptRow(ctx, l.core.DB(), attemptID)
	if err != nil {
		return err
	}
	switch strings.TrimSpace(row.AttemptStatus) {
	case "accepted", "waiting_callback", "querying", "unknown":
		// ok
	default:
		// attempt 已被回调或其他任务推进到最终态；跳过，避免重复推进主订单数量。
		return nil
	}

	now := l.core.Now()
	// 超过查单截止时间：直接判定 timeout，并进入补单或收敛逻辑。
	if !row.QueryDeadlineAt.IsZero() && !now.Before(row.QueryDeadlineAt) {
		updated, _ := l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = 'timeout',
    error_category = CASE WHEN error_category != '' THEN error_category ELSE 'timeout' END,
    finished_at = ?,
    updated_at = ?
WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
`, now, now, attemptID)
		affected := int64(0)
		if updated != nil {
			affected, _ = updated.RowsAffected()
		}
		if affected <= 0 {
			return nil
		}

		createRow, loadErr := l.loadCreateAttemptRow(ctx, l.core.DB(), attemptID)
		if loadErr == nil {
			if nextID, ok, _ := l.tryReplenishAfterFailedAttempt(ctx, createRow, traceID); ok && nextID > 0 {
				_ = l.executeCreateAttempt(ctx, nextID, traceID)
			} else {
				_ = l.markOrderFailed(ctx, createRow.OrderID, createRow.AttemptQuantity, "timeout", now)
				_ = l.tryKickoffNextCreatedAttempt(ctx, createRow.OrderID, traceID)
			}
		}
		return nil
	}

	orderProvider, ok := l.lookupOrderProvider(strings.TrimSpace(row.ProviderCode))
	if !ok {
		return apiErr(consts.CodeBadRequest, "provider不支持查单")
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

	queryInput := supplierprovider.QueryOrderInput{
		ProviderRequestOrderNo: row.ProviderRequestOrderNo,
		ChannelOrderNo:         strings.TrimSpace(row.ChannelOrderNo),
	}

	candidates := orderProvider.CandidateBaseURLs(accountCfg)
	for _, baseURL := range candidates {
		reqStart := time.Now()
		request, buildErr := orderProvider.BuildQueryOrderRequest(context.Background(), accountCfg, queryInput, baseURL)
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

		// 先写入本次查单请求快照，便于排查。
		updated, err := l.core.DB().Exec(ctx, `
	UPDATE trade_order_attempt
	SET attempt_status = 'querying',
	    request_url = ?,
	    request_method = ?,
    request_headers = ?,
    request_payload = ?,
	    trace_id = ?,
	    last_query_at = ?,
	    updated_at = ?
	WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
	`, request.URL.String(), request.Method, headersSnapshot, bodySnapshot, traceID, now, now, attemptID)
		if err != nil {
			return apiErr(consts.CodeInternalError, "更新attempt查单快照失败")
		}
		affected := int64(0)
		if updated != nil {
			affected, _ = updated.RowsAffected()
		}
		if affected <= 0 {
			// attempt 已在其他路径推进到最终态（如回调成功/失败）；跳过即可。
			return nil
		}

		response, requestErr := l.client.Do(request)
		if requestErr != nil {
			category := "server_error"
			if ne, ok := requestErr.(net.Error); ok && ne.Timeout() {
				category = "timeout"
			}
			updated, err := l.core.DB().Exec(ctx, `
	UPDATE trade_order_attempt
	SET attempt_status = 'querying',
	    error_category = ?,
	    error_message = ?,
	    duration_ms = ?,
	    updated_at = ?
	WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
	`, category, truncateSnapshot(requestErr.Error()), int(time.Since(reqStart).Milliseconds()), l.core.Now(), attemptID)
			if err != nil {
				return apiErr(consts.CodeInternalError, "更新attempt查单错误失败")
			}
			affected := int64(0)
			if updated != nil {
				affected, _ = updated.RowsAffected()
			}
			if affected <= 0 {
				return nil
			}
			continue
		}

		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return apiErr(consts.CodeInternalError, "平台响应读取失败")
		}
		responseSnapshot := snapshotResponseBody(body, account.TokenID, account.SecretKey)
		durationMS := int(time.Since(reqStart).Milliseconds())

		result, parseErr := orderProvider.ParseQueryOrderResponse(response.StatusCode, body)
		if parseErr != nil {
			queryCount := row.QueryCount + 1
			nextQueryAt := l.computeNextQueryAt(now, queryCount)

			if shouldRetrySupplierCandidate(response, body, parseErr) {
				updated, err := l.core.DB().Exec(ctx, `
	UPDATE trade_order_attempt
	SET response_payload = ?,
	    http_status = ?,
	    duration_ms = ?,
	    error_category = 'unknown',
	    error_message = ?,
	    updated_at = ?
	WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
	`, responseSnapshot, response.StatusCode, durationMS, truncateSnapshot(parseErr.Error()), l.core.Now(), attemptID)
				if err != nil {
					return apiErr(consts.CodeInternalError, "更新attempt查单解析错误失败")
				}
				affected := int64(0)
				if updated != nil {
					affected, _ = updated.RowsAffected()
				}
				if affected <= 0 {
					return nil
				}
				continue
			}

			updated, err := l.core.DB().Exec(ctx, `
	UPDATE trade_order_attempt
	SET response_payload = ?,
	    http_status = ?,
	    duration_ms = ?,
	    error_category = 'unknown',
	    error_message = ?,
	    query_count = ?,
	    next_query_at = ?,
	    updated_at = ?
	WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
	`, responseSnapshot, response.StatusCode, durationMS, truncateSnapshot(parseErr.Error()), queryCount, nullableTimeArg(nextQueryAt), l.core.Now(), attemptID)
			if err != nil {
				return apiErr(consts.CodeInternalError, "更新attempt查单解析错误失败")
			}
			affected := int64(0)
			if updated != nil {
				affected, _ = updated.RowsAffected()
			}
			if affected <= 0 {
				return nil
			}
			return nil
		}

		attemptStatus := "querying"
		if result.FinalSuccess {
			attemptStatus = "success"
		} else if result.FinalFailed {
			attemptStatus = "failed"
		}

		nextQueryAt := time.Time{}
		finishedAt := time.Time{}
		queryCount := row.QueryCount
		if attemptStatus == "querying" {
			queryCount++
			nextQueryAt = l.computeNextQueryAt(now, queryCount)
		} else {
			finishedAt = now
		}

		finalRes, _ := l.core.DB().Exec(ctx, `
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
    query_count = ?,
    next_query_at = ?,
    finished_at = ?,
    updated_at = ?
WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
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
			queryCount,
			nullableTimeArg(nextQueryAt),
			nullableTimeArg(finishedAt),
			l.core.Now(),
			attemptID,
		)
		affected = int64(0)
		if finalRes != nil {
			affected, _ = finalRes.RowsAffected()
		}
		if affected <= 0 {
			return nil
		}

		if attemptStatus == "success" {
			_ = l.markOrderSuccess(ctx, row.OrderID, row.AttemptQuantity, result.ChannelOrderNo, now)
			_ = l.tryKickoffNextCreatedAttempt(ctx, row.OrderID, traceID)
			return nil
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
			return nil
		}

		return nil
	}

	// 所有候选域名都失败：推进退避调度，避免每轮扫描都重复打上游。
	queryCount := row.QueryCount + 1
	nextQueryAt := l.computeNextQueryAt(now, queryCount)
	scheduleRes, err := l.core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET query_count = ?,
    next_query_at = ?,
    updated_at = ?
WHERE id = ? AND attempt_status IN ('accepted','waiting_callback','querying','unknown')
`, queryCount, nullableTimeArg(nextQueryAt), l.core.Now(), attemptID)
	if err != nil {
		return apiErr(consts.CodeInternalError, "推进attempt退避调度失败")
	}
	affected := int64(0)
	if scheduleRes != nil {
		affected, _ = scheduleRes.RowsAffected()
	}
	if affected <= 0 {
		return nil
	}
	return nil
}

func (l *TradeOrderLogic) computeNextQueryAt(now time.Time, queryCount int) time.Time {
	base := l.core.Config().Trade.AttemptQueryScanIntervalSeconds
	if base <= 0 {
		base = 30
	}
	maxBackoff := l.core.Config().Trade.AttemptQueryMaxBackoffSeconds
	if maxBackoff <= 0 {
		maxBackoff = 300
	}
	seconds := base
	if queryCount > 0 {
		seconds = base << (queryCount - 1)
	}
	if seconds > maxBackoff {
		seconds = maxBackoff
	}
	return now.Add(time.Duration(seconds) * time.Second)
}
