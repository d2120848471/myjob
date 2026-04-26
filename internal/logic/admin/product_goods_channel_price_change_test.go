package adminlogic

import (
	"context"
	"testing"

	"myjob/internal/app"
	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestApplyProductGoodsChannelPriceChangePushUpdatesPriceAndWritesLog(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 0, "qqlogin.yxp8.cn", 1, 0)
	logic := NewProductGoodsLogic(core)

	candidate := loadSinglePriceChangeCandidate(t, logic, goodsID)
	result, err := logic.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
		SupplierGoodsNo: candidate.SupplierGoodsNo,
		GoodsName:       "推送后名称",
		GoodsPrice:      decimal.RequireFromString("12.0000"),
		GoodsPriceValid: true,
		Raw:             `{"goodsid":"SKU-100","goodsprice":"12.0000"}`,
	}, productGoodsChannelPriceChangeSourcePush)
	require.NoError(t, err)
	require.True(t, result.Updated)
	require.True(t, result.PriceChanged)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "12.0000", row.SourceCostPrice)
	require.Equal(t, "12.5400", row.CostPrice)

	count, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE source = 'push' AND binding_id = ?`, candidate.BindingID)
	require.NoError(t, err)
	require.Equal(t, 1, count.Int())
}

func TestApplyProductGoodsChannelPriceChangeDoesNotLogWhenPriceUnchanged(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 0, "qqlogin.yxp8.cn", 1, 1)
	logic := NewProductGoodsLogic(core)
	candidate := loadSinglePriceChangeCandidate(t, logic, goodsID)

	result, err := logic.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
		SupplierGoodsNo: candidate.SupplierGoodsNo,
		GoodsPrice:      decimal.RequireFromString("10.0000"),
		GoodsPriceValid: true,
		Raw:             `{"goodsid":"SKU-100","goodsprice":"10.0000"}`,
	}, productGoodsChannelPriceChangeSourcePush)
	require.NoError(t, err)
	require.False(t, result.PriceChanged)

	count, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE binding_id = ?`, candidate.BindingID)
	require.NoError(t, err)
	require.Equal(t, 0, count.Int())
}

func TestApplyProductGoodsChannelPriceChangeKeepsPriceWhenLogInsertFails(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 0, "qqlogin.yxp8.cn", 1, 0)
	logic := NewProductGoodsLogic(core)
	candidate := loadSinglePriceChangeCandidate(t, logic, goodsID)

	_, err = core.DB().Exec(ctx, `DROP TRIGGER IF EXISTS fail_price_change_log_insert`)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
CREATE TRIGGER fail_price_change_log_insert
BEFORE INSERT ON product_goods_channel_price_change_log
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'forced price log failure'
`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = core.DB().Exec(context.Background(), `DROP TRIGGER IF EXISTS fail_price_change_log_insert`)
	})

	result, err := logic.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
		SupplierGoodsNo: candidate.SupplierGoodsNo,
		GoodsName:       "推送后名称",
		GoodsPrice:      decimal.RequireFromString("12.0000"),
		GoodsPriceValid: true,
		Raw:             `{"goodsid":"SKU-100","goodsprice":"12.0000"}`,
	}, productGoodsChannelPriceChangeSourcePush)
	require.NoError(t, err)
	require.True(t, result.Updated)
	require.True(t, result.PriceChanged)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "12.0000", row.SourceCostPrice)
	require.Equal(t, "12.5400", row.CostPrice)

	count, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE binding_id = ?`, candidate.BindingID)
	require.NoError(t, err)
	require.Equal(t, 0, count.Int())
}

func loadSinglePriceChangeCandidate(t *testing.T, logic *ProductGoodsLogic, goodsID int64) productGoodsChannelSyncCandidate {
	t.Helper()
	candidates, err := logic.loadProductGoodsChannelSyncCandidates(context.Background(), ProductGoodsChannelSyncOptions{GoodsID: goodsID}, 0, 10)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	return candidates[0]
}
