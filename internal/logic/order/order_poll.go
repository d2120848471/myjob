package orderlogic

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"myjob/internal/library/channelpricing"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

type queryPollResult struct {
	Status            string
	AttemptStatus     string
	SupplierOrderNo   string
	SupplierUSOrderNo string
	SupplierStatus    string
	RefundStatus      string
	Receipt           string
	Message           string
	RequestSnapshot   string
	ResponseSnapshot  string
}

// PollDueOnce 扫描一批到期未终态订单，并查询当前云发卡提交记录。
func (l *OrderLogic) PollDueOnce(ctx context.Context) error {
	if err := l.recoverStuckSubmittingOrders(ctx); err != nil {
		return err
	}
	rows := make([]entity.ExternalOrder, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT *
FROM external_order
WHERE status IN (?, ?) AND next_poll_at IS NOT NULL AND next_poll_at <= ?
ORDER BY next_poll_at ASC, id ASC
LIMIT 20
`, OrderStatusProcessing, OrderStatusUnknown, l.core.Now()); err != nil {
		return err
	}
	for _, row := range rows {
		if err := l.pollOrder(ctx, row); err != nil {
			continue
		}
	}
	return nil
}

func (l *OrderLogic) pollOrder(ctx context.Context, order entity.ExternalOrder) error {
	if !order.CurrentAttemptID.Valid || order.CurrentAttemptID.Int64 <= 0 {
		return l.recoverStuckSubmittingOrder(ctx, order)
	}
	attempt, err := l.loadOrderAttempt(ctx, order.CurrentAttemptID.Int64)
	if err != nil {
		return err
	}
	account, err := l.loadPlatformAccount(ctx, attempt.PlatformAccountID)
	if err != nil {
		return err
	}
	provider, ok := supplierprovider.LookupOrder(account.ProviderCode)
	if !ok {
		return l.markOrderFailed(ctx, order.ID, attempt.ID, "平台适配器不存在")
	}
	segments, err := l.loadAttemptSegments(ctx, attempt.ID)
	if err != nil {
		return err
	}
	if len(segments) > 0 {
		return l.pollAttemptSegments(ctx, order, attempt, provider, account, segments)
	}
	result, err := l.executeQueryOrder(provider, account, supplierprovider.QueryOrderInput{
		SupplierOrderNo:   attempt.SupplierOrderNo,
		SupplierUSOrderNo: attempt.SupplierUSOrderNo,
	})
	if err != nil && result.Status == "" {
		return err
	}
	switch result.Status {
	case supplierprovider.SupplierOrderStatusSuccess:
		return l.applyPollSuccess(ctx, order, attempt, result)
	case supplierprovider.SupplierOrderStatusProcessing:
		return l.applyPollProcessing(ctx, order, attempt, result)
	case supplierprovider.SupplierOrderStatusFailed:
		if err := l.applyPollAttemptFailed(ctx, order, attempt, result); err != nil {
			return err
		}
		return l.handleAttemptFailed(ctx, order, attempt, defaultOrderMessage(result.Receipt, result.Message))
	default:
		return l.applyPollUnknown(ctx, order, attempt, result)
	}
}

func (l *OrderLogic) executeQueryOrder(provider supplierprovider.OrderProvider, account entity.SupplierPlatformAccount, input supplierprovider.QueryOrderInput) (queryPollResult, error) {
	providerAccount := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  map[string]any{},
	}
	client := l.httpClientForOrderProvider(account.ProviderCode)
	result := queryPollResult{
		Status:            supplierprovider.SupplierOrderStatusUnknown,
		AttemptStatus:     OrderAttemptStatusUnknown,
		SupplierOrderNo:   input.SupplierOrderNo,
		SupplierUSOrderNo: input.SupplierUSOrderNo,
		Message:           "查单响应无法确认",
	}
	var lastErr error
	for _, baseURL := range provider.CandidateBaseURLs(providerAccount) {
		request, buildErr := provider.BuildQueryOrderRequest(context.Background(), providerAccount, l.core.Now(), baseURL, input)
		if buildErr != nil {
			return queryPollResult{}, buildErr
		}
		requestSnapshot, snapshotErr := snapshotOrderRequest(request, account)
		if snapshotErr != nil {
			return queryPollResult{}, snapshotErr
		}
		result.RequestSnapshot = requestSnapshot
		response, requestErr := client.Do(request)
		if requestErr != nil {
			lastErr = requestErr
			result.Message = requestErr.Error()
			result.ResponseSnapshot = truncateOrderSnapshot(sanitizeOrderSnapshot(requestErr.Error(), account))
			continue
		}
		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return queryPollResult{}, readErr
		}
		result.ResponseSnapshot = truncateOrderSnapshot(sanitizeOrderSnapshot(string(body), account))
		parsed, parseErr := provider.ParseQueryOrderResponse(response.StatusCode, body)
		if parseErr != nil {
			result.SupplierStatus = parsed.SupplierStatus
			result.RefundStatus = parsed.RefundStatus
			result.Message = defaultOrderMessage(parsed.Message, parseErr.Error())
			if errors.Is(parseErr, supplierprovider.ErrSupplierUnknownResponse) {
				return result, nil
			}
			return result, nil
		}
		result.Status = parsed.Status
		result.SupplierOrderNo = defaultOrderMessage(parsed.SupplierOrderNo, input.SupplierOrderNo)
		result.SupplierUSOrderNo = input.SupplierUSOrderNo
		result.SupplierStatus = parsed.SupplierStatus
		result.RefundStatus = parsed.RefundStatus
		result.Receipt = parsed.Receipt
		result.Message = defaultOrderMessage(parsed.Message, parsed.Receipt)
		switch parsed.Status {
		case supplierprovider.SupplierOrderStatusSuccess:
			result.AttemptStatus = OrderAttemptStatusSuccess
		case supplierprovider.SupplierOrderStatusFailed:
			result.AttemptStatus = OrderAttemptStatusFailed
		case supplierprovider.SupplierOrderStatusProcessing:
			result.AttemptStatus = OrderAttemptStatusProcessing
		default:
			result.AttemptStatus = OrderAttemptStatusUnknown
		}
		return result, nil
	}
	if lastErr != nil {
		return result, nil
	}
	return result, nil
}

func (l *OrderLogic) applyPollSuccess(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, result queryPollResult) error {
	now := l.core.Now()
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := updateAttemptFromPoll(ctx, tx, attempt.ID, result, now); err != nil {
			return err
		}
		_, err := tx.Exec(`
UPDATE external_order
SET status = ?, last_receipt = ?, next_poll_at = NULL, last_poll_at = ?, poll_count = poll_count + 1, updated_at = ?
WHERE id = ?
`, OrderStatusSuccess, defaultOrderMessage(result.Receipt, result.Message), now, now, order.ID)
		return err
	})
}

func (l *OrderLogic) applyPollProcessing(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, result queryPollResult) error {
	now := l.core.Now()
	nextPollAt := now.Add(pollIntervalDuration(l.core.Config().OpenOrder.PollIntervalSeconds))
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := updateAttemptFromPoll(ctx, tx, attempt.ID, result, now); err != nil {
			return err
		}
		nextStatus := order.Status
		if strings.TrimSpace(nextStatus) == "" {
			nextStatus = OrderStatusProcessing
		}
		_, err := tx.Exec(`
UPDATE external_order
SET status = ?, last_receipt = ?, next_poll_at = ?, last_poll_at = ?, poll_count = poll_count + 1, updated_at = ?
WHERE id = ?
`, nextStatus, defaultOrderMessage(result.Receipt, result.Message), nextPollAt, now, now, order.ID)
		return err
	})
}

func (l *OrderLogic) applyPollAttemptFailed(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, result queryPollResult) error {
	now := l.core.Now()
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := updateAttemptFromPoll(ctx, tx, attempt.ID, result, now); err != nil {
			return err
		}
		_, err := tx.Exec(`
UPDATE external_order
SET last_receipt = ?, last_poll_at = ?, poll_count = poll_count + 1, updated_at = ?
WHERE id = ?
`, defaultOrderMessage(result.Receipt, result.Message), now, now, order.ID)
		return err
	})
}

func (l *OrderLogic) applyPollUnknown(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, result queryPollResult) error {
	now := l.core.Now()
	nextPollAt := now.Add(pollIntervalDuration(l.core.Config().OpenOrder.PollIntervalSeconds))
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := updateAttemptFromPoll(ctx, tx, attempt.ID, result, now); err != nil {
			return err
		}
		_, err := tx.Exec(`
UPDATE external_order
SET last_poll_error = ?, next_poll_at = ?, last_poll_at = ?, poll_count = poll_count + 1, updated_at = ?
WHERE id = ?
`, result.Message, nextPollAt, now, now, order.ID)
		return err
	})
}

func updateAttemptFromPoll(ctx context.Context, tx gdb.TX, attemptID int64, result queryPollResult, now time.Time) error {
	_, err := tx.Exec(`
UPDATE external_order_attempt
SET supplier_order_no = ?, supplier_us_order_no = ?, supplier_status = ?, refund_status = ?,
    response_snapshot = ?, receipt = ?, status = ?, last_checked_at = ?, updated_at = ?
WHERE id = ?
`, result.SupplierOrderNo, result.SupplierUSOrderNo, result.SupplierStatus, result.RefundStatus,
		result.ResponseSnapshot, result.Receipt, result.AttemptStatus, now, now, attemptID)
	return err
}

func (l *OrderLogic) loadOrderAttempt(ctx context.Context, attemptID int64) (entity.ExternalOrderAttempt, error) {
	attempt := entity.ExternalOrderAttempt{}
	err := l.core.DB().GetCore().GetScan(ctx, &attempt, `SELECT * FROM external_order_attempt WHERE id = ?`, attemptID)
	return attempt, err
}

func (l *OrderLogic) recoverStuckSubmittingOrders(ctx context.Context) error {
	rows := make([]entity.ExternalOrder, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT *
FROM external_order
WHERE status = ?
  AND (current_attempt_id IS NULL OR current_attempt_id = 0)
  AND next_poll_at IS NULL
ORDER BY updated_at ASC, id ASC
LIMIT 20
`, OrderStatusProcessing); err != nil {
		return err
	}
	for _, row := range rows {
		if err := l.recoverStuckSubmittingOrder(ctx, row); err != nil {
			continue
		}
	}
	return nil
}

func (l *OrderLogic) recoverStuckSubmittingOrder(ctx context.Context, order entity.ExternalOrder) error {
	attempted, err := l.loadAttemptedBindingIDs(ctx, order.ID)
	if err != nil {
		return err
	}
	candidates, err := l.loadCandidateChannels(ctx, order.GoodsID, attempted)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return l.markOrderFailed(ctx, order.ID, 0, "提交状态异常且暂无可用云发卡渠道")
	}
	config, err := l.loadReorderConfig(ctx, order.GoodsID)
	if err != nil {
		return err
	}
	candidate := selectCandidate(candidates, attempted, config.OrderStrategy, l.core.Now())
	if candidate.BindingID == 0 {
		return l.markOrderFailed(ctx, order.ID, 0, "提交状态异常且暂无可用云发卡渠道")
	}
	attemptNo := order.AttemptCount + 1
	supplierUSOrderNo := order.OrderNo + "-T" + intToString(attemptNo)
	now := l.core.Now()
	// 恢复后的订单需要在本轮继续查单；MySQL DATETIME 无小数秒时可能把 now 四舍五入到下一秒。
	immediatePollAt := now.Add(-time.Second)
	priceSnapshot, err := channelpricing.OrderSnapshot(candidate.pricingRule(), order.Quantity)
	if err != nil {
		return err
	}
	receipt := "提交状态异常，转入查单确认"

	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Exec(`
UPDATE external_order
SET status = ?, attempt_count = ?, last_receipt = ?, next_poll_at = ?,
    unit_price = ?, order_amount = ?, cost_amount = ?, profit_amount = ?, updated_at = ?
WHERE id = ?
  AND status = ?
  AND (current_attempt_id IS NULL OR current_attempt_id = 0)
  AND next_poll_at IS NULL
`, OrderStatusUnknown, attemptNo, receipt, immediatePollAt, priceSnapshot.UnitPrice, priceSnapshot.OrderAmount, priceSnapshot.CostAmount, priceSnapshot.ProfitAmount, now, order.ID, OrderStatusProcessing)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil || rows == 0 {
			return err
		}
		insertResult, err := tx.Exec(`
INSERT INTO external_order_attempt (
    order_id, order_no, attempt_no, channel_binding_id, platform_account_id, platform_account_name, platform_subject_id, platform_subject_name,
    provider_code, supplier_goods_no, supplier_goods_name, supplier_us_order_no, supplier_order_no,
    supplier_status, refund_status, request_snapshot, response_snapshot, receipt, status,
    submitted_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, order.ID, order.OrderNo, attemptNo, candidate.BindingID, candidate.PlatformAccountID, candidate.PlatformAccountName,
			candidate.PlatformSubjectID, candidate.PlatformSubjectName, candidate.ProviderCode, candidate.SupplierGoodsNo, candidate.SupplierGoodsName, supplierUSOrderNo, "",
			"", "", "", "", receipt, OrderAttemptStatusUnknown,
			nil, now, now)
		if err != nil {
			return err
		}
		attemptID, err := insertResult.LastInsertId()
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
UPDATE external_order
SET current_attempt_id = ?, updated_at = ?
WHERE id = ?
`, attemptID, now, order.ID)
		return err
	})
}
