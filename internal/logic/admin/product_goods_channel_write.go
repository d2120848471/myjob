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

// CreateChannelBinding 新增单条商品渠道绑定，并写入操作日志。
func (l *ProductGoodsLogic) CreateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingCreateRes, error) {
	goods, err := l.getActiveProduct(ctx, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	normalized, err := l.normalizeProductGoodsChannelBindingInput(ctx, goods, req.PlatformAccountID, req.SupplierGoodsNo, req.SupplierGoodsName, req.SourceCostPrice, req.ValidateTemplateID, req.DockStatus, req.Sort, req.OrderWeight, req.OrderTimeStart, req.OrderTimeEnd, nil)
	if err != nil {
		return nil, err
	}

	createdID := int64(0)
	now := l.core.Now()
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, txErr := tx.Exec(`
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price,
    tax_adjust_direction, tax_adjust_rate, tax_adjust_amount, dock_status, sort, order_weight, order_time_start, order_time_end, validate_template_id,
    is_auto_change, add_type, default_price, is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, '', '0.0000', 0, ?, ?)
`, goods.ID, normalized.PlatformAccountID, normalized.SupplierGoodsNo, normalized.SupplierGoodsName, normalized.SourceCostPrice, normalized.CostSnapshot.CostPrice, normalized.CostSnapshot.TaxAdjustDirection, normalized.CostSnapshot.TaxAdjustRate, normalized.CostSnapshot.TaxAdjustAmount, normalized.DockStatus, normalized.Sort, normalized.OrderWeight, nullableStringArg(normalized.OrderTimeStart), nullableStringArg(normalized.OrderTimeEnd), nullableInt64Arg(normalized.ValidateTemplateID), now, now)
		if txErr != nil {
			return txErr
		}
		createdID, _ = result.LastInsertId()
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定新增失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品渠道绑定：goods=%d, binding=%d", goods.ID, createdID), ip)
	l.triggerProductGoodsChannelAutoSubscription(ctx, createdID)
	return &adminapi.ProductGoodsChannelBindingCreateRes{ID: createdID}, nil
}

// UpdateChannelBinding 编辑单条商品渠道绑定基础字段，并写入操作日志。
func (l *ProductGoodsLogic) UpdateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingUpdateRes, error) {
	goods, err := l.getActiveProduct(ctx, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	current, err := l.getActiveProductGoodsChannelBinding(ctx, req.GoodsId, req.BindingId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定查询失败")
	}

	normalized, err := l.normalizeProductGoodsChannelBindingInput(ctx, goods, req.PlatformAccountID, req.SupplierGoodsNo, req.SupplierGoodsName, req.SourceCostPrice, req.ValidateTemplateID, req.DockStatus, req.Sort, req.OrderWeight, req.OrderTimeStart, req.OrderTimeEnd, &current.ID)
	if err != nil {
		return nil, err
	}

	now := l.core.Now()
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, txErr := tx.Exec(`
UPDATE product_goods_channel_binding
SET platform_account_id = ?, supplier_goods_no = ?, supplier_goods_name = ?, source_cost_price = ?, cost_price = ?,
    tax_adjust_direction = ?, tax_adjust_rate = ?, tax_adjust_amount = ?, dock_status = ?, sort = ?, order_weight = ?, order_time_start = ?, order_time_end = ?, validate_template_id = ?,
    updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, normalized.PlatformAccountID, normalized.SupplierGoodsNo, normalized.SupplierGoodsName, normalized.SourceCostPrice, normalized.CostSnapshot.CostPrice, normalized.CostSnapshot.TaxAdjustDirection, normalized.CostSnapshot.TaxAdjustRate, normalized.CostSnapshot.TaxAdjustAmount, normalized.DockStatus, normalized.Sort, normalized.OrderWeight, nullableStringArg(normalized.OrderTimeStart), nullableStringArg(normalized.OrderTimeEnd), nullableInt64Arg(normalized.ValidateTemplateID), now, req.BindingId, req.GoodsId)
		if txErr != nil {
			return txErr
		}
		return ensureMutationAffected(result)
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定编辑失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品渠道绑定：goods=%d, binding=%d", req.GoodsId, req.BindingId), ip)
	l.triggerProductGoodsChannelAutoSubscription(ctx, req.BindingId)
	return &adminapi.ProductGoodsChannelBindingUpdateRes{}, nil
}

// DeleteChannelBinding 软删除单条商品渠道绑定，并写入操作日志。
func (l *ProductGoodsLogic) DeleteChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingDeleteRes, error) {
	if _, err := l.getActiveProductGoodsChannelBinding(ctx, req.GoodsId, req.BindingId); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定查询失败")
	}

	now := l.core.Now()
	result, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_deleted = 1, deleted_at = ?, updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, now, now, req.BindingId, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定删除失败")
	}
	if err := ensureMutationAffected(result); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品渠道绑定：goods=%d, binding=%d", req.GoodsId, req.BindingId), ip)
	return &adminapi.ProductGoodsChannelBindingDeleteRes{}, nil
}

// UpdateChannelBindingAutoPrice 编辑单条绑定的自动改价规则，并写入操作日志。
func (l *ProductGoodsLogic) UpdateChannelBindingAutoPrice(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, error) {
	if _, err := l.getActiveProductGoodsChannelBinding(ctx, req.GoodsId, req.BindingId); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定查询失败")
	}

	normalized, err := l.normalizeProductGoodsChannelAutoPriceInput(req.IsAutoChange, req.AddType, req.DefaultPrice)
	if err != nil {
		return nil, err
	}

	result, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_auto_change = ?, add_type = ?, default_price = ?, updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, normalized.IsAutoChange, normalized.AddType, normalized.DefaultPrice, l.core.Now(), req.BindingId, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定自动改价保存失败")
	}
	if err := ensureMutationAffected(result); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "渠道绑定不存在")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品渠道绑定自动改价：goods=%d, binding=%d", req.GoodsId, req.BindingId), ip)
	return &adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes{}, nil
}

func (l *ProductGoodsLogic) getActiveProductGoodsChannelBinding(ctx context.Context, goodsID, bindingID int64) (entity.ProductGoodsChannelBinding, error) {
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    id, goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price,
    tax_adjust_direction, tax_adjust_rate, tax_adjust_amount, dock_status, sort, order_weight, order_time_start, order_time_end, validate_template_id,
    is_auto_change, add_type, default_price, is_deleted, deleted_at, created_at, updated_at
FROM product_goods_channel_binding
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, bindingID, goodsID)
	if err != nil {
		return entity.ProductGoodsChannelBinding{}, err
	}
	if row == nil || len(row) == 0 {
		return entity.ProductGoodsChannelBinding{}, sql.ErrNoRows
	}
	return entity.ProductGoodsChannelBinding{
		ID:                 row["id"].Int64(),
		GoodsID:            row["goods_id"].Int64(),
		PlatformAccountID:  row["platform_account_id"].Int64(),
		SupplierGoodsNo:    row["supplier_goods_no"].String(),
		SupplierGoodsName:  row["supplier_goods_name"].String(),
		SourceCostPrice:    productGoodsRecordMoney(row, "source_cost_price"),
		CostPrice:          productGoodsRecordMoney(row, "cost_price"),
		TaxAdjustDirection: productGoodsRecordString(row, "tax_adjust_direction"),
		TaxAdjustRate:      productGoodsRecordMoney(row, "tax_adjust_rate"),
		TaxAdjustAmount:    productGoodsRecordMoney(row, "tax_adjust_amount"),
		DockStatus:         row["dock_status"].Int(),
		Sort:               row["sort"].Int(),
		OrderWeight:        productGoodsRecordMoney(row, "order_weight"),
		OrderTimeStart:     productGoodsRecordNullString(row, "order_time_start"),
		OrderTimeEnd:       productGoodsRecordNullString(row, "order_time_end"),
		ValidateTemplateID: productGoodsRecordNullInt64(row, "validate_template_id"),
		IsAutoChange:       row["is_auto_change"].Int(),
		AddType:            productGoodsRecordString(row, "add_type"),
		DefaultPrice:       productGoodsRecordMoney(row, "default_price"),
		IsDeleted:          row["is_deleted"].Int(),
		DeletedAt:          nullableTimeFromRecord(row, "deleted_at"),
		CreatedAt:          parseRecordTime(row, "created_at"),
		UpdatedAt:          parseRecordTime(row, "updated_at"),
	}, nil
}
