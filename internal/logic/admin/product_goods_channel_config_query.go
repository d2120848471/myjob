package adminlogic

import (
	"context"
	"database/sql"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// GetInventoryConfig 返回商品库存配置详情、默认值与策略选项。
func (l *ProductGoodsLogic) GetInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigGetReq) (*adminapi.ProductGoodsInventoryConfigGetRes, error) {
	goods, err := l.getActiveProduct(ctx, req.GoodsId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品库存配置查询失败")
	}
	if err := ensureChannelSupplyProduct(goods); err != nil {
		return nil, err
	}

	goodsSummary, err := l.getProductGoodsChannelGoodsSummary(ctx, req.GoodsId)
	if err != nil {
		return nil, err
	}
	state, err := l.loadProductGoodsInventoryConfigState(ctx, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品库存配置查询失败")
	}

	return &adminapi.ProductGoodsInventoryConfigGetRes{
		Goods:                goodsSummary,
		Config:               state.toAPIConfig(),
		OrderStrategyOptions: productGoodsOrderStrategyOptions(),
	}, nil
}

func (l *ProductGoodsLogic) loadProductGoodsInventoryConfigState(ctx context.Context, goodsID int64) (productGoodsInventoryConfigState, error) {
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    goods_id,
    smart_reorder_enabled,
    reorder_timeout_enabled,
    reorder_timeout_minutes,
    order_strategy,
    sync_cost_price_enabled,
    sync_goods_name_enabled,
    allow_loss_sale_enabled,
    max_loss_amount,
    combo_goods_enabled,
    created_at,
    updated_at
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return productGoodsInventoryConfigState{}, err
	}
	if row == nil || len(row) == 0 {
		return defaultProductGoodsInventoryConfigState(), nil
	}
	entityRow := entity.ProductGoodsChannelConfig{
		GoodsID:               row["goods_id"].Int64(),
		SmartReorderEnabled:   row["smart_reorder_enabled"].Int(),
		ReorderTimeoutEnabled: row["reorder_timeout_enabled"].Int(),
		ReorderTimeoutMinutes: row["reorder_timeout_minutes"].Int(),
		OrderStrategy:         productGoodsRecordString(row, "order_strategy"),
		SyncCostPriceEnabled:  row["sync_cost_price_enabled"].Int(),
		SyncGoodsNameEnabled:  row["sync_goods_name_enabled"].Int(),
		AllowLossSaleEnabled:  row["allow_loss_sale_enabled"].Int(),
		MaxLossAmount:         productGoodsRecordMoney(row, "max_loss_amount"),
		ComboGoodsEnabled:     row["combo_goods_enabled"].Int(),
		CreatedAt:             parseRecordTime(row, "created_at"),
		UpdatedAt:             parseRecordTime(row, "updated_at"),
	}
	state := productGoodsInventoryConfigStateFromRow(entityRow)
	if state.OrderStrategy == "" {
		state.OrderStrategy = productGoodsOrderStrategyFixedOrder
	}
	if state.MaxLossAmount == "" {
		state.MaxLossAmount = "0.0000"
	}
	return state, nil
}
