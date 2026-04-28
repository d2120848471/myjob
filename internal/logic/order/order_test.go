package orderlogic

import (
	"context"
	"strings"
	"testing"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/stretchr/testify/require"
)

func TestCanReorderRequiresSmartAndTimeoutWindow(t *testing.T) {
	created := time.Date(2026, 4, 24, 15, 0, 0, 0, time.Local)
	require.False(t, canReorder(reorderConfig{SmartEnabled: 0, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 0, TimeoutMinutes: 10}, created, created.Add(time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 0}, created, created.Add(time.Minute)))
	require.True(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(9*time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(11*time.Minute)))
}

func TestSelectCandidateFixedAndLowestCost(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 20, CostPrice: "9.0000"},
		{BindingID: 11, Sort: 10, CostPrice: "12.0000"},
	}
	fixed := selectCandidate(candidates, map[int64]struct{}{}, "fixed_order", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 11, fixed.BindingID)
	lowest := selectCandidate(candidates, map[int64]struct{}{}, "lowest_cost", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 10, lowest.BindingID)
}

func TestSelectCandidateSkipsAttempted(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 10, CostPrice: "9.0000"},
		{BindingID: 11, Sort: 20, CostPrice: "12.0000"},
	}
	selected := selectCandidate(candidates, map[int64]struct{}{10: {}}, "fixed_order", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 11, selected.BindingID)
}

func TestSelectCandidateWeightedPercentSkipsZeroWeight(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 10, CostPrice: "9.0000", OrderWeight: "0.0000"},
		{BindingID: 11, Sort: 20, CostPrice: "12.0000", OrderWeight: "100.0000"},
	}

	for range 100 {
		selected := selectCandidate(candidates, map[int64]struct{}{}, "weighted_percent", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
		require.EqualValues(t, 11, selected.BindingID)
	}
}

func TestPollIntervalDurationUsesConfiguredSeconds(t *testing.T) {
	require.Equal(t, 7*time.Second, pollIntervalDuration(7))
	require.Equal(t, 30*time.Second, pollIntervalDuration(0))
}

func TestBuildOrderSegmentsRespectsProviderMaxQuantity(t *testing.T) {
	segments := buildOrderSegments("O20260428123045123456-T1", 3, supplierprovider.OrderProviderCapabilities{MaxQuantityPerCreate: 1})
	require.Equal(t, []orderSegmentPlan{
		{SegmentNo: 1, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S1"},
		{SegmentNo: 2, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S2"},
		{SegmentNo: 3, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S3"},
	}, segments)
}

func TestBuildOrderSegmentsUsesSingleSegmentWhenUnlimited(t *testing.T) {
	segments := buildOrderSegments("O20260428123045123456-T1", 3, supplierprovider.OrderProviderCapabilities{})
	require.Equal(t, []orderSegmentPlan{{SegmentNo: 1, Quantity: 3, SupplierUSOrderNo: "O20260428123045123456-T1-S1"}}, segments)
}

func TestAggregateSegmentStatuses(t *testing.T) {
	success := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess},
		{Status: OrderAttemptStatusSuccess},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusSuccess, success.OrderStatus)
	require.Equal(t, OrderAttemptStatusSuccess, success.AttemptStatus)

	failed := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusFailed, Receipt: "失败"},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusFailed, failed.OrderStatus)
	require.Equal(t, OrderAttemptStatusFailed, failed.AttemptStatus)

	processing := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess},
		{Status: OrderAttemptStatusProcessing},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusProcessing, processing.OrderStatus)
	require.Equal(t, OrderAttemptStatusProcessing, processing.AttemptStatus)
}

func TestAggregateSegmentStatusesPartialFailureBecomesUnknown(t *testing.T) {
	partial := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSubmitted, Receipt: "子单1已受理"},
		{Status: OrderAttemptStatusFailed, Receipt: "子单2失败"},
		{Status: OrderAttemptStatusPending, Receipt: "子单3未提交"},
	})

	require.Equal(t, supplierprovider.SupplierOrderStatusUnknown, partial.OrderStatus)
	require.Equal(t, OrderAttemptStatusUnknown, partial.AttemptStatus)
	require.Contains(t, partial.Message, "部分")

	successThenFailed := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess, Receipt: "子单1成功"},
		{Status: OrderAttemptStatusFailed, Receipt: "子单2失败"},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusUnknown, successThenFailed.OrderStatus)
	require.Equal(t, OrderAttemptStatusUnknown, successThenFailed.AttemptStatus)
}

func TestAggregateSegmentStatusesSingleFailureCanFailAttempt(t *testing.T) {
	failed := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusFailed, Receipt: "失败"},
	})

	require.Equal(t, supplierprovider.SupplierOrderStatusFailed, failed.OrderStatus)
	require.Equal(t, OrderAttemptStatusFailed, failed.AttemptStatus)
}

func TestCreateOpenOrderRetriesOrderNoUniqueConflict(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	fixture := seedOpenOrderCreationFixture(t, core, "G-ORDER-RETRY")
	insertExistingExternalOrderNo(t, core, fixture, "ORETRYDUP")

	logic := NewOrderLogic(core)
	orderNos := []string{"ORETRYDUP", "ORETRYOK"}
	logic.orderNoGenerator = func() string {
		next := orderNos[0]
		orderNos = orderNos[1:]
		return next
	}

	res, err := logic.CreateOpenOrder(context.Background(), &adminapi.OpenOrderCreateReq{
		Token:    "test-open-order-token",
		GoodsID:  fixture.goodsCode,
		Account:  "13800138000",
		Quantity: 1,
	})
	require.NoError(t, err)
	require.Equal(t, "ORETRYOK", res.OrderNo)
	require.EqualValues(t, 1, scalarOrderTestInt(t, core, `SELECT COUNT(*) FROM external_order WHERE order_no = ?`, "ORETRYDUP"))
	require.EqualValues(t, 1, scalarOrderTestInt(t, core, `SELECT COUNT(*) FROM external_order WHERE order_no = ?`, "ORETRYOK"))
}

func TestRechargeRiskSnapshotsFitColumnLimits(t *testing.T) {
	longToken := strings.Repeat("token", 30)
	longReason := strings.Repeat("r", 512)

	require.LessOrEqual(t, len([]rune(riskRecordTokenSnapshot(longToken))), 64)
	require.LessOrEqual(t, len([]rune(riskOrderReceipt(longReason))), 512)
}

func TestCreateRiskFailedOpenOrderRechecksRuleStatus(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	fixture := seedOpenOrderCreationFixture(t, core, "G-RISK-RECHECK")
	now := core.Now()
	_, err = core.DB().Exec(ctx, `
INSERT INTO recharge_risk_rule (
    account, goods_keyword, reason, status, hit_count,
    created_by_id, created_by_name, updated_by_id, updated_by_name,
    is_deleted, created_at, updated_at
) VALUES ('risk-recheck-account', '重试', '规则已停用不应拦截', 1, 0, 1, 'admin', 1, 'admin', 0, ?, ?)
`, now, now)
	require.NoError(t, err)

	logic := NewOrderLogic(core)
	goods, err := logic.loadOpenOrderGoods(ctx, fixture.goodsCode)
	require.NoError(t, err)
	match, matched, err := logic.matchRechargeRisk(ctx, "risk-recheck-account", goods.Name)
	require.NoError(t, err)
	require.True(t, matched)
	_, err = core.DB().Exec(ctx, `UPDATE recharge_risk_rule SET status = 0, updated_at = ? WHERE id = ?`, core.Now(), match.RuleID)
	require.NoError(t, err)

	unitPrice, err := normalizeOrderMoney(goods.DefaultSellPrice)
	require.NoError(t, err)
	orderAmount, err := multiplyOrderMoney(unitPrice, 1)
	require.NoError(t, err)
	orderNo, created, err := logic.createRiskFailedOpenOrder(ctx, &adminapi.OpenOrderCreateReq{
		Token:    "test-open-order-token",
		GoodsID:  fixture.goodsCode,
		Account:  "risk-recheck-account",
		Quantity: 1,
	}, goods, "risk-recheck-account", unitPrice, orderAmount, match, core.Now())
	require.NoError(t, err)
	require.False(t, created)
	require.Empty(t, orderNo)
	require.EqualValues(t, 0, scalarOrderTestInt(t, core, `SELECT COUNT(*) FROM external_order WHERE account = ?`, "risk-recheck-account"))
	require.EqualValues(t, 0, scalarOrderTestInt(t, core, `SELECT COUNT(*) FROM recharge_risk_record WHERE account = ?`, "risk-recheck-account"))
	require.EqualValues(t, 0, scalarOrderTestInt(t, core, `SELECT hit_count FROM recharge_risk_rule WHERE id = ?`, match.RuleID))
}

type openOrderCreationFixture struct {
	goodsID   int64
	goodsCode string
}

func seedOpenOrderCreationFixture(t *testing.T, core *app.Core, goodsCode string) openOrderCreationFixture {
	t.Helper()
	now := core.Now()
	result, err := core.DB().Exec(context.Background(), `
INSERT INTO product_goods (
    goods_code, brand_id, name, goods_type, supply_type, is_export, is_douyin, has_tax,
    subject_id, exception_notify, balance_limit, default_sell_price, min_purchase_qty,
    max_purchase_qty, status, is_deleted, created_at, updated_at
) VALUES (?, 1, '订单号重试商品', 'direct_recharge', 'channel', 1, 0, 0, NULL, 1, '0.0000', '20.0000', 1, 1, 1, 0, ?, ?)
`, goodsCode, now, now)
	require.NoError(t, err)
	goodsID, err := result.LastInsertId()
	require.NoError(t, err)

	result, err = core.DB().Exec(context.Background(), `
INSERT INTO supplier_platform_account (
    name, provider_code, provider_name, type_id, subject_id, has_tax, status, domain,
    backup_domain, token_id, secret_key, extra_config, threshold_amount, sort, crowd_name,
    is_deleted, created_at, updated_at
) VALUES ('订单号重试云发卡', 'kakayun', '卡卡云', 6, 1, 0, 1, 'example.test', 'example.test', 'order-retry-token', 'secret-key', '{}', '5000.0000', 1, '订单群', 0, ?, ?)
`, now, now)
	require.NoError(t, err)
	platformID, err := result.LastInsertId()
	require.NoError(t, err)

	_, err = core.DB().Exec(context.Background(), `
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price,
    cost_price, dock_status, sort, order_weight, is_deleted, created_at, updated_at
) VALUES (?, ?, '2478510', '云发卡测试直充商品', '10.0000', '10.0000', 1, 10, '0.0000', 0, ?, ?)
`, goodsID, platformID, now, now)
	require.NoError(t, err)

	return openOrderCreationFixture{goodsID: goodsID, goodsCode: goodsCode}
}

func insertExistingExternalOrderNo(t *testing.T, core *app.Core, fixture openOrderCreationFixture, orderNo string) {
	t.Helper()
	now := core.Now()
	_, err := core.DB().Exec(context.Background(), `
INSERT INTO external_order (
    order_no, goods_id, goods_code, goods_name, goods_type, supply_type, subject_id,
    subject_name, has_tax, account, quantity, unit_price, order_amount, cost_amount,
    profit_amount, status, attempt_count, created_at, updated_at
) VALUES (?, ?, ?, '订单号重试商品', 'direct_recharge', 'channel', NULL, '', 0, '13800138000', 1, '20.0000', '20.0000', '0.0000', '20.0000', 'pending_submit', 0, ?, ?)
`, orderNo, fixture.goodsID, fixture.goodsCode, now, now)
	require.NoError(t, err)
}

func scalarOrderTestInt(t *testing.T, core *app.Core, query string, args ...any) int64 {
	t.Helper()
	value, err := core.DB().GetCore().GetValue(context.Background(), query, args...)
	require.NoError(t, err)
	return value.Int64()
}
