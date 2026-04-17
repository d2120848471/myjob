package tradelogic

import (
	"context"
	"testing"
	"time"

	"myjob/internal/app"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestBuildCandidateBindings_Basic(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, err := subjectResult.LastInsertId()
	require.NoError(t, err)

	templateResult, err := core.DB().Exec(ctx, `
INSERT INTO product_template (title, template_type, is_shared, account_name, validate_type, created_at, updated_at)
VALUES ('手机号模板', 'local', 0, '手机号', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	templateID, err := templateResult.LastInsertId()
	require.NoError(t, err)

	goodsCode := "P-TRADE-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品A', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, err := goodsResult.LastInsertId()
	require.NoError(t, err)

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'test', '测试平台', 6, ?, 0, 'https://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, err := accountResult.LastInsertId()
	require.NoError(t, err)

	// 绑定表中的 cost_price 已在管理端保存时计算好；此处直接写入快照即可。
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name,
    source_cost_price, cost_price, tax_adjust_direction, tax_adjust_rate, tax_adjust_amount,
    dock_status, sort, weight, start_time, end_time, validate_template_id,
    is_auto_change, add_type, default_price,
    created_at, updated_at
) VALUES (
    ?, ?, 'G001', '上游商品A',
    '100.0000', '113.0000', 'untaxed_to_taxed', '13.0000', '13.0000',
    'enabled', 10, 0, '', '', NULL,
    1, 'fixed', '2.0000',
    ?, ?
)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	out, err := BuildCandidateBindings(ctx, core, goodsCode, 2, `{"mobile":"13800138000"}`)
	require.NoError(t, err)
	require.Equal(t, goodsID, out.Goods.ID)
	require.Equal(t, goodsCode, out.Goods.GoodsCode)
	require.Equal(t, int64(subjectID), out.Goods.SubjectID)
	require.Equal(t, "fixed_order", out.Config.RouteMode) // config 行不存在时会自动插入默认值
	require.Len(t, out.Candidates, 1)
	require.Equal(t, accountID, out.Candidates[0].PlatformAccountID)
	require.Equal(t, "test", out.Candidates[0].ProviderCode)
	require.Equal(t, "测试平台", out.Candidates[0].ProviderName)
	require.Equal(t, "113.0000", MoneyString(out.Candidates[0].CostPrice))
	require.Equal(t, "2.0000", MoneyString(out.Candidates[0].DefaultPrice))
	require.True(t, out.Candidates[0].IsAutoChange)
}

func TestBuildCandidateBindings_QuantityOutOfRange(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	goodsCode := "P-TRADE-002"
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品A', 'card_secret', 'channel', 1, ?, '29.9000', 1, 1, 1, ?, ?)
`, goodsCode, subjectID, now, now)
	require.NoError(t, err)

	_, err = BuildCandidateBindings(ctx, core, goodsCode, 2, `{"mobile":"13800138000"}`)
	require.Error(t, err)
}

func TestBuildCandidateBindings_BindingTemplateMismatch_FilterOut(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	goodsCode := "P-TRADE-003"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品A', 'card_secret', 'channel', 1, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	// digits 模板：validate_type=6
	templateResult, err := core.DB().Exec(ctx, `
INSERT INTO product_template (title, template_type, is_shared, account_name, validate_type, created_at, updated_at)
VALUES ('纯数字模板', 'local', 0, '纯数字', 6, ?, ?)
`, now, now)
	require.NoError(t, err)
	templateID, _ := templateResult.LastInsertId()

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'test', '测试平台', 6, ?, 0, 'https://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name,
    source_cost_price, cost_price, tax_adjust_direction, tax_adjust_rate, tax_adjust_amount,
    dock_status, sort, weight, start_time, end_time, validate_template_id,
    is_auto_change, add_type, default_price,
    created_at, updated_at
) VALUES (
    ?, ?, 'G001', '上游商品A',
    '100.0000', '100.0000', 'none', '0.0000', '0.0000',
    'enabled', 10, 0, '', '', ?,
    0, 'fixed', '0.0000',
    ?, ?
)
`, goodsID, accountID, templateID, now, now)
	require.NoError(t, err)

	_, err = BuildCandidateBindings(ctx, core, goodsCode, 1, `{"mobile":"13800138000"}`)
	require.Error(t, err)
}

func TestPickFirstBinding_WithBuiltCandidates_LowestCostFirst(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	goodsCode := "P-TRADE-004"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品A', 'card_secret', 'channel', 1, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	// 配置路由模式：lowest_cost_first
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, route_mode, created_at, updated_at)
VALUES (?, 'lowest_cost_first', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'test', '测试平台', 6, ?, 0, 'https://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, source_cost_price, cost_price, dock_status, sort, created_at, updated_at)
VALUES (?, ?, 'G001', '100.0000', '110.0000', 'enabled', 20, ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, source_cost_price, cost_price, dock_status, sort, created_at, updated_at)
VALUES (?, ?, 'G002', '100.0000', '105.0000', 'enabled', 30, ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	out, err := BuildCandidateBindings(ctx, core, goodsCode, 1, `{"mobile":"13800138000"}`)
	require.NoError(t, err)
	require.Len(t, out.Candidates, 2)

	first, err := PickFirstBinding(out.Config.RouteMode, out.Candidates, time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local), nil)
	require.NoError(t, err)
	require.Equal(t, "105.0000", MoneyString(first.CostPrice))
	require.True(t, first.CostPrice.LessThan(decimal.RequireFromString("110.0000")))
}
