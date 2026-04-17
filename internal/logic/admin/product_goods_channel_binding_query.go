package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/consts"
)

// List 查询指定商品的渠道绑定列表（不包含已软删除绑定）。
func (l *ProductGoodsChannelBindingLogic) List(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (*adminapi.ProductGoodsChannelBindingListRes, error) {
	goods, err := l.getActiveGoodsForBinding(ctx, req.GoodsID)
	if err != nil {
		return nil, err
	}
	if goods.SupplyType != "channel" {
		return nil, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}

	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    b.id,
    b.dock_status,
    b.platform_account_id,
    COALESCE(a.name, '') AS platform_account_name,
    COALESCE(a.provider_code, '') AS provider_code,
    COALESCE(a.provider_name, '') AS provider_name,
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.source_cost_price,
    b.cost_price,
    b.tax_adjust_direction,
    b.tax_adjust_rate,
    b.tax_adjust_amount,
    b.sort,
    b.weight,
    b.start_time,
    b.end_time,
    b.validate_template_id,
    COALESCE(t.title, '') AS validate_template_name,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    b.lock_price,
    b.symbol_price,
    b.max_price,
    b.min_price
FROM product_goods_channel_binding b
LEFT JOIN supplier_platform_account a ON a.id = b.platform_account_id
LEFT JOIN product_template t ON t.id = b.validate_template_id
WHERE b.goods_id = ? AND b.is_deleted = 0
ORDER BY b.sort ASC, b.id ASC
`, req.GoodsID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "绑定列表查询失败")
	}

	items := make([]adminapi.ProductGoodsChannelBindingItem, 0, len(rows))
	for _, row := range rows {
		supplierGoodsNo := row["supplier_goods_no"].String()
		supplierGoodsName := productGoodsRecordString(row, "supplier_goods_name")
		providerName := productGoodsRecordString(row, "provider_name")
		displayName := buildBindingDisplayName(supplierGoodsName, supplierGoodsNo, goods.SubjectName, providerName)

		items = append(items, adminapi.ProductGoodsChannelBindingItem{
			ID:                   row["id"].Int64(),
			DisplayName:          displayName,
			DockStatus:           row["dock_status"].String(),
			PlatformAccountID:    row["platform_account_id"].Int64(),
			PlatformAccountName:  productGoodsRecordString(row, "platform_account_name"),
			ProviderCode:         productGoodsRecordString(row, "provider_code"),
			ProviderName:         providerName,
			SupplierGoodsNo:      supplierGoodsNo,
			SupplierGoodsName:    supplierGoodsName,
			SourceCostPrice:      formatMoney(productGoodsRecordString(row, "source_cost_price")),
			CostPrice:            formatMoney(productGoodsRecordString(row, "cost_price")),
			TaxAdjustDirection:   productGoodsRecordString(row, "tax_adjust_direction"),
			TaxAdjustRate:        formatMoney(productGoodsRecordString(row, "tax_adjust_rate")),
			TaxAdjustAmount:      formatMoney(productGoodsRecordString(row, "tax_adjust_amount")),
			Sort:                 row["sort"].Int(),
			Weight:               row["weight"].Int(),
			StartTime:            productGoodsRecordString(row, "start_time"),
			EndTime:              productGoodsRecordString(row, "end_time"),
			ValidateTemplateID:   nullableInt64Pointer(productGoodsRecordNullInt64(row, "validate_template_id")),
			ValidateTemplateName: productGoodsRecordString(row, "validate_template_name"),
			IsAutoChange:         row["is_auto_change"].Int(),
			AddType:              productGoodsRecordString(row, "add_type"),
			DefaultPrice:         formatMoney(productGoodsRecordString(row, "default_price")),
			LockPrice:            formatMoney(productGoodsRecordString(row, "lock_price")),
			SymbolPrice:          formatMoney(productGoodsRecordString(row, "symbol_price")),
			MaxPrice:             formatMoney(productGoodsRecordString(row, "max_price")),
			MinPrice:             formatMoney(productGoodsRecordString(row, "min_price")),
		})
	}

	return &adminapi.ProductGoodsChannelBindingListRes{List: items}, nil
}
