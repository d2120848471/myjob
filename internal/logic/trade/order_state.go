package tradelogic

import (
	"context"
	"strings"
	"time"

	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

// markOrderFailed 在单个 fulfillment 最终失败（或超时）时，累加失败数量并在必要时把订单收敛到最终态。
func (l *TradeOrderLogic) markOrderFailed(ctx context.Context, orderID int64, attemptQuantity int, reason string, now time.Time) error {
	if orderID <= 0 || attemptQuantity <= 0 {
		return nil
	}
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.GetOne(`
SELECT quantity, success_quantity, failed_quantity, status, failure_reason
FROM trade_order
WHERE id = ?
`, orderID)
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
		newFailed := currentFailed + attemptQuantity

		nextStatus := "processing"
		finishedAt := time.Time{}
		failureReason := strings.TrimSpace(record["failure_reason"].String())

		if total > 0 && currentSuccess+newFailed >= total {
			finishedAt = now
			if currentSuccess > 0 && newFailed > 0 {
				nextStatus = "manual_review"
				if failureReason == "" {
					failureReason = "partial_success_need_review"
				}
			} else {
				nextStatus = "failed"
				if failureReason == "" {
					if strings.TrimSpace(reason) == "" {
						failureReason = "all_bindings_failed"
					} else {
						failureReason = strings.TrimSpace(reason)
					}
				}
			}
		}

		_, err = tx.Exec(`
UPDATE trade_order
SET failed_quantity = ?,
    status = ?,
    failure_reason = CASE WHEN ? != '' THEN ? ELSE failure_reason END,
    finished_at = ?,
    updated_at = ?
WHERE id = ?
`, newFailed, nextStatus, failureReason, failureReason, nullableTimeArg(finishedAt), now, orderID)
		if err != nil {
			return apiErr(consts.CodeInternalError, "更新订单失败")
		}
		return nil
	})
}

// tryKickoffNextCreatedAttempt 在订单没有进行中 attempt 时，尝试启动最早的待执行 attempt（用于推进下一个 fulfillment）。
func (l *TradeOrderLogic) tryKickoffNextCreatedAttempt(ctx context.Context, orderID int64, traceID string) error {
	if orderID <= 0 {
		return nil
	}

	record, err := l.core.DB().GetCore().GetOne(ctx, `SELECT status FROM trade_order WHERE id = ?`, orderID)
	if err != nil || record == nil || len(record) == 0 {
		return nil
	}
	status := strings.TrimSpace(record["status"].String())
	if status == "success" || status == "failed" || status == "manual_review" {
		return nil
	}

	inProgress, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM trade_order_attempt
WHERE order_id = ? AND attempt_status IN ('submitted','accepted','waiting_callback','querying','unknown')
`, orderID)
	if err == nil && inProgress.Int() > 0 {
		return nil
	}

	value, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT id
FROM trade_order_attempt
WHERE order_id = ? AND attempt_status = 'created'
ORDER BY fulfillment_no ASC, attempt_no ASC, id ASC
LIMIT 1
`, orderID)
	if err != nil || value == nil || value.IsNil() {
		return nil
	}
	attemptID := value.Int64()
	if attemptID <= 0 {
		return nil
	}
	_ = l.executeCreateAttempt(ctx, attemptID, strings.TrimSpace(traceID))
	return nil
}

