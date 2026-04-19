package adminlogic

import (
	"context"
	"database/sql"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

// SaveInventoryConfig 保存商品库存配置，并写入操作日志。
func (l *ProductGoodsLogic) SaveInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigSaveReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsInventoryConfigSaveRes, error) {
	goods, err := l.getActiveProduct(ctx, req.GoodsId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品库存配置保存失败")
	}

	normalized, err := l.normalizeProductGoodsInventoryConfigInput(ctx, goods, req)
	if err != nil {
		return nil, err
	}

	now := l.core.Now()
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 库存配置允许重复保存相同内容，因此必须使用单条 upsert，
		// 不能依赖 RowsAffected 判断“记录是否存在”。
		query, args := buildProductGoodsInventoryConfigUpsertSQL(l.core.Config().Database.Driver, goods.ID, normalized, now)
		_, txErr := tx.Exec(query, args...)
		return txErr
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品库存配置保存失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("修改商品库存配置：goods=%d, name=%s", goods.ID, goods.Name), ip)
	return &adminapi.ProductGoodsInventoryConfigSaveRes{}, nil
}

func buildProductGoodsInventoryConfigUpsertSQL(driver string, goodsID int64, normalized normalizedProductGoodsInventoryConfigInput, now any) (string, []any) {
	insertArgs := []any{
		goodsID,
		normalized.SmartReorderEnabled,
		normalized.ReorderTimeoutEnabled,
		normalized.ReorderTimeoutMinutes,
		normalized.OrderStrategy,
		normalized.SyncCostPriceEnabled,
		normalized.SyncGoodsNameEnabled,
		normalized.AllowLossSaleEnabled,
		normalized.MaxLossAmount,
		normalized.ComboGoodsEnabled,
		now,
		now,
	}

	if lDriver := driver; lDriver == "sqlite" {
		return `
INSERT INTO product_goods_channel_config (
    goods_id, smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy,
    sync_cost_price_enabled, sync_goods_name_enabled, allow_loss_sale_enabled, max_loss_amount, combo_goods_enabled,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(goods_id) DO UPDATE SET
    smart_reorder_enabled = excluded.smart_reorder_enabled,
    reorder_timeout_enabled = excluded.reorder_timeout_enabled,
    reorder_timeout_minutes = excluded.reorder_timeout_minutes,
    order_strategy = excluded.order_strategy,
    sync_cost_price_enabled = excluded.sync_cost_price_enabled,
    sync_goods_name_enabled = excluded.sync_goods_name_enabled,
    allow_loss_sale_enabled = excluded.allow_loss_sale_enabled,
    max_loss_amount = excluded.max_loss_amount,
    combo_goods_enabled = excluded.combo_goods_enabled,
    updated_at = excluded.updated_at
`, insertArgs
	}

	return `
INSERT INTO product_goods_channel_config (
    goods_id, smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy,
    sync_cost_price_enabled, sync_goods_name_enabled, allow_loss_sale_enabled, max_loss_amount, combo_goods_enabled,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    smart_reorder_enabled = VALUES(smart_reorder_enabled),
    reorder_timeout_enabled = VALUES(reorder_timeout_enabled),
    reorder_timeout_minutes = VALUES(reorder_timeout_minutes),
    order_strategy = VALUES(order_strategy),
    sync_cost_price_enabled = VALUES(sync_cost_price_enabled),
    sync_goods_name_enabled = VALUES(sync_goods_name_enabled),
    allow_loss_sale_enabled = VALUES(allow_loss_sale_enabled),
    max_loss_amount = VALUES(max_loss_amount),
    combo_goods_enabled = VALUES(combo_goods_enabled),
    updated_at = VALUES(updated_at)
`, insertArgs
}
