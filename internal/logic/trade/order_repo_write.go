package tradelogic

import (
	"context"
	"strings"
	"time"

	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

func (l *TradeOrderLogic) insertTradeOrder(ctx context.Context, tx gdb.TX, orderNo string, input CreateTradeOrderInput, prepared CandidateBuildOutput, firstBinding CandidateBinding, lockedSalePrice decimal.Decimal, totalAmount decimal.Decimal, lossAmount decimal.Decimal) (int64, error) {
	now := l.core.Now()
	lossOrder := 0
	if lossAmount.GreaterThan(decimal.Zero) {
		lossOrder = 1
	}

	result, err := tx.Exec(`
INSERT INTO trade_order (
    order_no, caller_id, client_order_no, goods_id, goods_code_snapshot, goods_name_snapshot,
    binding_id, platform_account_id, route_mode_snapshot, quantity,
    payload_json, sale_price, total_amount,
    source_cost_price_snapshot, cost_price_snapshot, tax_adjust_direction, tax_adjust_rate, tax_adjust_amount,
    loss_order, loss_amount,
    status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		orderNo,
		input.CallerID,
		input.ClientOrderNo,
		prepared.Goods.ID,
		prepared.Goods.GoodsCode,
		prepared.Goods.GoodsName,
		firstBinding.ID,
		firstBinding.PlatformAccountID,
		strings.TrimSpace(prepared.Config.RouteMode),
		input.Quantity,
		input.PayloadJSON,
		MoneyString(lockedSalePrice),
		MoneyString(totalAmount),
		MoneyString(firstBinding.SourceCostPrice),
		MoneyString(firstBinding.CostPrice),
		strings.TrimSpace(firstBinding.TaxAdjustDirection),
		MoneyString(firstBinding.TaxAdjustRate),
		MoneyString(firstBinding.TaxAdjustAmount),
		lossOrder,
		MoneyString(lossAmount),
		"processing",
		now,
		now,
	)
	if err != nil {
		return 0, apiErr(consts.CodeInternalError, "创建订单失败")
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (l *TradeOrderLogic) insertTradeOrderAttempt(ctx context.Context, tx gdb.TX, orderID int64, providerRequestOrderNo string, plan FulfillmentPlanItem, binding CandidateBinding, subjectName string, lockedSalePrice decimal.Decimal, lossAmount decimal.Decimal, attemptNo int) (int64, error) {
	now := l.core.Now()
	channelName := buildBindingChannelNameSnapshot(binding, subjectName)
	if attemptNo <= 0 {
		attemptNo = 1
	}
	result, err := tx.Exec(`
INSERT INTO trade_order_attempt (
    order_id, binding_id, platform_account_id, provider_code,
    fulfillment_no, attempt_quantity, attempt_no, provider_request_order_no,
    binding_channel_name_snapshot, binding_supplier_goods_no_snapshot,
    source_cost_price_snapshot, cost_price_snapshot, sale_price_snapshot, loss_amount_snapshot,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		orderID,
		binding.ID,
		binding.PlatformAccountID,
		strings.TrimSpace(binding.ProviderCode),
		plan.FulfillmentNo,
		plan.AttemptQuantity,
		attemptNo,
		providerRequestOrderNo,
		channelName,
		strings.TrimSpace(binding.SupplierGoodsNo),
		MoneyString(binding.SourceCostPrice),
		MoneyString(binding.CostPrice),
		MoneyString(lockedSalePrice),
		MoneyString(lossAmount),
		now,
		now,
	)
	if err != nil {
		return 0, apiErr(consts.CodeInternalError, "创建订单失败")
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func buildBindingChannelNameSnapshot(binding CandidateBinding, subjectName string) string {
	base := strings.TrimSpace(binding.SupplierGoodsName)
	if base == "" {
		base = strings.TrimSpace(binding.SupplierGoodsNo)
	}
	subjectName = strings.TrimSpace(subjectName)
	providerName := strings.TrimSpace(binding.ProviderName)
	return strings.TrimSpace(base + " / " + subjectName + " / " + providerName)
}

func nullableTimeArg(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
