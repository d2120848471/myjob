package orderlogic

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	runtimeapp "myjob/internal/app"
	"myjob/internal/library/channelpricing"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

type createSubmitResult struct {
	OrderStatus       string
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

// SubmitPendingOnce 扫描一批待提交订单，并逐单提交到一个可用云发卡渠道。
func (l *OrderLogic) SubmitPendingOnce(ctx context.Context) error {
	rows := make([]entity.ExternalOrder, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT *
FROM external_order
WHERE status = ?
ORDER BY id ASC
LIMIT 20
`, OrderStatusPendingSubmit); err != nil {
		return err
	}
	var firstErr error
	for _, row := range rows {
		if err := l.submitOrder(ctx, row); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (l *OrderLogic) submitOrder(ctx context.Context, order entity.ExternalOrder) error {
	attempted, err := l.loadAttemptedBindingIDs(ctx, order.ID)
	if err != nil {
		return err
	}
	candidates, err := l.loadCandidateChannels(ctx, order.GoodsID, attempted)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return l.markOrderFailed(ctx, order.ID, 0, "暂无可用云发卡渠道")
	}
	config, err := l.loadReorderConfig(ctx, order.GoodsID)
	if err != nil {
		return err
	}
	candidate := selectCandidate(candidates, attempted, config.OrderStrategy, l.core.Now())
	if candidate.BindingID == 0 {
		return l.markOrderFailed(ctx, order.ID, 0, "暂无可用云发卡渠道")
	}
	account, err := l.loadPlatformAccount(ctx, candidate.PlatformAccountID)
	if err != nil {
		return err
	}
	provider, ok := supplierprovider.LookupOrder(account.ProviderCode)
	if !ok {
		return l.markOrderFailed(ctx, order.ID, 0, "平台适配器不存在")
	}
	attemptNo := order.AttemptCount + 1
	supplierUSOrderNo := order.OrderNo + "-T" + intToString(attemptNo)
	attempt, claimed, err := l.claimOrderWithPendingAttempt(ctx, order, candidate, attemptNo, supplierUSOrderNo)
	if err != nil || !claimed {
		return err
	}
	result, err := l.executeCreateOrder(ctx, provider, account, supplierprovider.CreateOrderInput{
		SupplierGoodsNo:   candidate.SupplierGoodsNo,
		Quantity:          order.Quantity,
		Account:           order.Account,
		SupplierUSOrderNo: supplierUSOrderNo,
	})
	if err != nil && result.OrderStatus == "" {
		result = createSubmitResult{
			OrderStatus:       OrderStatusUnknown,
			AttemptStatus:     OrderAttemptStatusUnknown,
			SupplierUSOrderNo: supplierUSOrderNo,
			Message:           err.Error(),
		}
	}
	if strings.TrimSpace(result.SupplierUSOrderNo) == "" {
		result.SupplierUSOrderNo = supplierUSOrderNo
	}
	if err := l.applySubmitResultToAttempt(ctx, order, candidate, attempt, attemptNo, result); err != nil {
		return err
	}
	if result.OrderStatus == OrderStatusFailed {
		nextOrder := order
		nextOrder.Status = OrderStatusFailed
		nextOrder.CurrentAttemptID = sql.NullInt64{Int64: attempt.ID, Valid: true}
		nextOrder.AttemptCount = attemptNo
		return l.handleAttemptFailed(ctx, nextOrder, attempt, defaultOrderMessage(result.Receipt, result.Message))
	}
	return nil
}

func (l *OrderLogic) executeCreateOrder(ctx context.Context, provider supplierprovider.OrderProvider, account entity.SupplierPlatformAccount, input supplierprovider.CreateOrderInput) (createSubmitResult, error) {
	providerAccount := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  map[string]any{},
	}
	now := l.core.Now()
	client := l.httpClientForOrderProvider(account.ProviderCode)
	result := createSubmitResult{
		OrderStatus:       OrderStatusUnknown,
		AttemptStatus:     OrderAttemptStatusUnknown,
		SupplierUSOrderNo: input.SupplierUSOrderNo,
		Message:           "上游响应无法确认",
	}
	var lastErr error
	for _, baseURL := range provider.CandidateBaseURLs(providerAccount) {
		request, buildErr := provider.BuildCreateOrderRequest(context.Background(), providerAccount, now, baseURL, input)
		if buildErr != nil {
			return createSubmitResult{}, buildErr
		}
		requestSnapshot, snapshotErr := snapshotOrderRequest(request, account)
		if snapshotErr != nil {
			return createSubmitResult{}, snapshotErr
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
			return createSubmitResult{}, readErr
		}
		result.ResponseSnapshot = truncateOrderSnapshot(sanitizeOrderSnapshot(string(body), account))
		parsed, parseErr := provider.ParseCreateOrderResponse(response.StatusCode, body)
		if parseErr != nil {
			result.SupplierOrderNo = parsed.SupplierOrderNo
			result.SupplierUSOrderNo = input.SupplierUSOrderNo
			result.SupplierStatus = parsed.SupplierStatus
			result.Message = defaultOrderMessage(parsed.Message, parseErr.Error())
			if errors.Is(parseErr, supplierprovider.ErrSupplierUnknownResponse) {
				return result, nil
			}
			result.OrderStatus = OrderStatusFailed
			result.AttemptStatus = OrderAttemptStatusFailed
			return result, nil
		}
		result.OrderStatus = OrderStatusProcessing
		result.AttemptStatus = OrderAttemptStatusSubmitted
		result.SupplierOrderNo = parsed.SupplierOrderNo
		result.SupplierUSOrderNo = input.SupplierUSOrderNo
		result.SupplierStatus = parsed.SupplierStatus
		result.Message = defaultOrderMessage(parsed.Message, "下单成功")
		return result, nil
	}
	if lastErr != nil {
		return result, nil
	}
	return result, nil
}

func (l *OrderLogic) handleAttemptFailed(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, receipt string) error {
	config, err := l.loadReorderConfig(ctx, order.GoodsID)
	if err != nil {
		return err
	}
	if !canReorder(config, order.CreatedAt, l.core.Now()) {
		return l.markOrderFailed(ctx, order.ID, attempt.ID, receipt)
	}
	candidates, err := l.loadCandidateChannels(ctx, order.GoodsID, map[int64]struct{}{attempt.ChannelBindingID: {}})
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return l.markOrderFailed(ctx, order.ID, attempt.ID, receipt)
	}
	return l.submitOrder(ctx, order)
}

func (l *OrderLogic) claimOrderWithPendingAttempt(ctx context.Context, order entity.ExternalOrder, candidate orderChannelCandidate, attemptNo int, supplierUSOrderNo string) (entity.ExternalOrderAttempt, bool, error) {
	now := l.core.Now()
	priceSnapshot, err := channelpricing.OrderSnapshot(candidate.pricingRule(), order.Quantity)
	if err != nil {
		return entity.ExternalOrderAttempt{}, false, err
	}
	nextPollAt := now.Add(pollIntervalDuration(l.core.Config().OpenOrder.PollIntervalSeconds))
	attempt := entity.ExternalOrderAttempt{
		OrderID:             order.ID,
		OrderNo:             order.OrderNo,
		AttemptNo:           attemptNo,
		ChannelBindingID:    candidate.BindingID,
		PlatformAccountID:   candidate.PlatformAccountID,
		PlatformAccountName: candidate.PlatformAccountName,
		PlatformSubjectID:   candidate.PlatformSubjectID,
		PlatformSubjectName: candidate.PlatformSubjectName,
		ProviderCode:        candidate.ProviderCode,
		SupplierGoodsNo:     candidate.SupplierGoodsNo,
		SupplierGoodsName:   candidate.SupplierGoodsName,
		SupplierUSOrderNo:   supplierUSOrderNo,
		Receipt:             "等待上游提交结果",
		Status:              OrderAttemptStatusPending,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		claimed, err := l.markOrderSubmitting(ctx, tx, order, attemptNo, nextPollAt, priceSnapshot, now)
		if err != nil || !claimed {
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
			"", "", "", "", attempt.Receipt, OrderAttemptStatusPending,
			nil, now, now)
		if err != nil {
			return err
		}
		attempt.ID, err = insertResult.LastInsertId()
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
UPDATE external_order
SET current_attempt_id = ?, updated_at = ?
WHERE id = ?
`, attempt.ID, now, order.ID)
		return err
	})
	if err != nil {
		return entity.ExternalOrderAttempt{}, false, err
	}
	if attempt.ID == 0 {
		return entity.ExternalOrderAttempt{}, false, nil
	}
	return attempt, true, nil
}

func (l *OrderLogic) markOrderSubmitting(ctx context.Context, tx gdb.TX, order entity.ExternalOrder, attemptNo int, nextPollAt time.Time, priceSnapshot channelpricing.OrderPriceSnapshot, now time.Time) (bool, error) {
	sqlText := `
UPDATE external_order
SET status = ?, attempt_count = ?, last_receipt = ?, next_poll_at = ?,
    unit_price = ?, order_amount = ?, cost_amount = ?, profit_amount = ?, updated_at = ?
WHERE id = ?
`
	args := []any{OrderStatusProcessing, attemptNo, "正在提交上游订单", nextPollAt, priceSnapshot.UnitPrice, priceSnapshot.OrderAmount, priceSnapshot.CostAmount, priceSnapshot.ProfitAmount, now, order.ID}
	switch {
	case order.Status == OrderStatusPendingSubmit:
		sqlText += ` AND status = ? AND (current_attempt_id IS NULL OR current_attempt_id = 0)`
		args = append(args, OrderStatusPendingSubmit)
	case order.CurrentAttemptID.Valid:
		sqlText += ` AND status = ? AND current_attempt_id = ?`
		args = append(args, order.Status, order.CurrentAttemptID.Int64)
	default:
		sqlText += ` AND status = ? AND (current_attempt_id IS NULL OR current_attempt_id = 0)`
		args = append(args, order.Status)
	}
	result, err := tx.Exec(sqlText, args...)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}

func (l *OrderLogic) applySubmitResultToAttempt(ctx context.Context, order entity.ExternalOrder, candidate orderChannelCandidate, attempt entity.ExternalOrderAttempt, attemptNo int, result createSubmitResult) error {
	now := l.core.Now()
	priceSnapshot, err := channelpricing.OrderSnapshot(candidate.pricingRule(), order.Quantity)
	if err != nil {
		return err
	}
	nextPollAt := any(nil)
	if result.OrderStatus == OrderStatusProcessing || result.OrderStatus == OrderStatusUnknown {
		nextPollAt = now.Add(pollIntervalDuration(l.core.Config().OpenOrder.PollIntervalSeconds))
	}
	return l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Exec(`
UPDATE external_order_attempt
SET supplier_order_no = ?, supplier_us_order_no = ?, supplier_status = ?, refund_status = ?,
    request_snapshot = ?, response_snapshot = ?, receipt = ?, status = ?, submitted_at = ?, updated_at = ?
WHERE id = ?
`, result.SupplierOrderNo, result.SupplierUSOrderNo, result.SupplierStatus, result.RefundStatus,
			result.RequestSnapshot, result.ResponseSnapshot, result.Receipt, result.AttemptStatus, now, now, attempt.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
UPDATE external_order
SET status = ?, current_attempt_id = ?, attempt_count = ?, last_receipt = ?, next_poll_at = ?,
    unit_price = ?, order_amount = ?, cost_amount = ?, profit_amount = ?, updated_at = ?
WHERE id = ?
`, result.OrderStatus, attempt.ID, attemptNo, defaultOrderMessage(result.Receipt, result.Message), nextPollAt, priceSnapshot.UnitPrice, priceSnapshot.OrderAmount, priceSnapshot.CostAmount, priceSnapshot.ProfitAmount, now, order.ID)
		return err
	})
}

func (l *OrderLogic) loadAttemptedBindingIDs(ctx context.Context, orderID int64) (map[int64]struct{}, error) {
	rows := make([]struct {
		ChannelBindingID int64 `db:"channel_binding_id"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT channel_binding_id FROM external_order_attempt WHERE order_id = ?`, orderID); err != nil {
		return nil, err
	}
	result := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		result[row.ChannelBindingID] = struct{}{}
	}
	return result, nil
}

func (l *OrderLogic) loadPlatformAccount(ctx context.Context, id int64) (entity.SupplierPlatformAccount, error) {
	account := entity.SupplierPlatformAccount{}
	err := l.core.DB().GetCore().GetScan(ctx, &account, `SELECT * FROM supplier_platform_account WHERE id = ? AND is_deleted = 0`, id)
	return account, err
}

func (l *OrderLogic) markOrderFailed(ctx context.Context, orderID, attemptID int64, receipt string) error {
	now := l.core.Now()
	currentAttemptID := any(nil)
	if attemptID > 0 {
		currentAttemptID = attemptID
	}
	_, err := l.core.DB().Exec(ctx, `
UPDATE external_order
SET status = ?, current_attempt_id = ?, last_receipt = ?, next_poll_at = NULL, updated_at = ?
WHERE id = ?
`, OrderStatusFailed, currentAttemptID, receipt, now, orderID)
	return err
}

func (l *OrderLogic) loadReorderConfig(ctx context.Context, goodsID int64) (reorderConfig, error) {
	row := struct {
		SmartReorderEnabled   int    `db:"smart_reorder_enabled"`
		ReorderTimeoutEnabled int    `db:"reorder_timeout_enabled"`
		ReorderTimeoutMinutes int    `db:"reorder_timeout_minutes"`
		OrderStrategy         string `db:"order_strategy"`
	}{OrderStrategy: "fixed_order"}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return reorderConfig{OrderStrategy: "fixed_order"}, nil
	}
	if strings.TrimSpace(row.OrderStrategy) == "" {
		row.OrderStrategy = "fixed_order"
	}
	return reorderConfig{
		SmartEnabled:   row.SmartReorderEnabled,
		TimeoutEnabled: row.ReorderTimeoutEnabled,
		TimeoutMinutes: row.ReorderTimeoutMinutes,
		OrderStrategy:  row.OrderStrategy,
	}, nil
}

func (l *OrderLogic) httpClientForOrderProvider(providerCode string) *http.Client {
	if providerCode != ProviderCodeKakayun {
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

func snapshotOrderRequest(request *http.Request, account entity.SupplierPlatformAccount) (string, error) {
	body := []byte{}
	if request.Body != nil {
		raw, err := io.ReadAll(request.Body)
		if err != nil {
			return "", err
		}
		body = raw
		request.Body = io.NopCloser(bytes.NewReader(raw))
		request.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(raw)), nil
		}
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
		"body":    string(body),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return truncateOrderSnapshot(sanitizeOrderSnapshot(string(raw), account)), nil
}

func sanitizeOrderSnapshot(value string, account entity.SupplierPlatformAccount) string {
	value = strings.ReplaceAll(value, account.SecretKey, runtimeapp.MaskSecret(account.SecretKey))
	value = strings.ReplaceAll(value, account.TokenID, runtimeapp.MaskSecret(account.TokenID))
	return value
}

func truncateOrderSnapshot(value string) string {
	if len(value) <= 4096 {
		return value
	}
	return value[:4096]
}

func defaultOrderMessage(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func intToString(value int) string {
	return decimal.NewFromInt(int64(value)).String()
}
