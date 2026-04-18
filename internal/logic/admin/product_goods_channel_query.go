package adminlogic

import (
	"context"
	"database/sql"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

type productGoodsChannelSummary struct {
	BoundChannels                []string
	BoundChannelCount            int
	PrimaryChannelName           string
	MinChannelCost               string
	MinChannelEffectiveSellPrice string
	ChannelAutoPriceStatus       int
}

// ChannelBindingList 返回指定商品的渠道弹窗数据。
func (l *ProductGoodsLogic) ChannelBindingList(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (*adminapi.ProductGoodsChannelBindingListRes, error) {
	goods, err := l.getProductGoodsChannelGoodsSummary(ctx, req.GoodsId)
	if err != nil {
		return nil, err
	}

	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    b.id,
    b.goods_id,
    b.platform_account_id,
    a.name AS platform_account_name,
    a.subject_id AS platform_subject_id,
    s.name AS platform_subject_name,
    a.has_tax AS platform_has_tax,
    a.last_balance_status,
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.source_cost_price,
    b.cost_price,
    b.tax_adjust_direction,
    b.tax_adjust_rate,
    b.tax_adjust_amount,
    b.validate_template_id,
    COALESCE(t.title, '') AS validate_template_title,
    b.dock_status,
    b.sort,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    b.created_at,
    b.updated_at
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
JOIN admin_subject s ON s.id = a.subject_id
LEFT JOIN product_template t ON t.id = b.validate_template_id
WHERE b.goods_id = ? AND b.is_deleted = 0 AND a.is_deleted = 0
ORDER BY b.sort ASC, b.id ASC
`, req.GoodsId)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定查询失败")
	}

	items := make([]adminapi.ProductGoodsChannelBindingItem, 0, len(rows))
	for _, row := range rows {
		effectiveSellPrice, priceErr := computeChannelEffectiveSellPrice(goods.DefaultSellPrice, productGoodsRecordMoney(row, "cost_price"), row["is_auto_change"].Int(), row["add_type"].String(), productGoodsRecordMoney(row, "default_price"))
		if priceErr != nil {
			return nil, apiErr(consts.CodeInternalError, "商品渠道绑定价格计算失败")
		}
		displayName := strings.TrimSpace(strings.Join([]string{
			productGoodsRecordString(row, "supplier_goods_name"),
			productGoodsRecordString(row, "platform_subject_name"),
			productGoodsRecordString(row, "platform_account_name"),
		}, " "))
		items = append(items, adminapi.ProductGoodsChannelBindingItem{
			ID:                    row["id"].Int64(),
			PlatformAccountID:     row["platform_account_id"].Int64(),
			PlatformAccountName:   row["platform_account_name"].String(),
			PlatformHasTax:        row["platform_has_tax"].Int(),
			ConnectStatus:         row["last_balance_status"].Int(),
			ConnectStatusText:     connectStatusText(row["last_balance_status"].Int()),
			SupplierGoodsNo:       row["supplier_goods_no"].String(),
			SupplierGoodsName:     row["supplier_goods_name"].String(),
			DisplayName:           displayName,
			SourceCostPrice:       productGoodsRecordMoney(row, "source_cost_price"),
			CostPrice:             productGoodsRecordMoney(row, "cost_price"),
			EffectiveSellPrice:    effectiveSellPrice,
			TaxAdjustDirection:    productGoodsRecordString(row, "tax_adjust_direction"),
			TaxAdjustRate:         productGoodsRecordMoney(row, "tax_adjust_rate"),
			TaxAdjustAmount:       productGoodsRecordMoney(row, "tax_adjust_amount"),
			ValidateTemplateID:    nullableInt64Pointer(productGoodsRecordNullInt64(row, "validate_template_id")),
			ValidateTemplateTitle: productGoodsRecordString(row, "validate_template_title"),
			DockStatus:            row["dock_status"].Int(),
			Sort:                  row["sort"].Int(),
			IsAutoChange:          row["is_auto_change"].Int(),
			AddType:               productGoodsRecordString(row, "add_type"),
			DefaultPrice:          productGoodsRecordMoney(row, "default_price"),
			CreatedAt:             formatAppTime(parseRecordTime(row, "created_at")),
			UpdatedAt:             formatAppTime(parseRecordTime(row, "updated_at")),
		})
	}

	return &adminapi.ProductGoodsChannelBindingListRes{
		Goods: goods,
		List:  items,
	}, nil
}

// ChannelBindingFormOptions 返回商品渠道绑定弹窗的表单选项。
func (l *ProductGoodsLogic) ChannelBindingFormOptions(ctx context.Context, req *adminapi.ProductGoodsChannelBindingFormOptionsReq) (*adminapi.ProductGoodsChannelBindingFormOptionsRes, error) {
	if _, err := l.getActiveProduct(ctx, req.GoodsId); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定选项查询失败")
	}

	platformRows := make([]gdb.Record, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &platformRows, `
SELECT
    a.id,
    a.name,
    a.subject_id,
    s.name AS subject_name,
    a.has_tax,
    a.last_balance_status
FROM supplier_platform_account a
JOIN admin_subject s ON s.id = a.subject_id
WHERE a.is_deleted = 0
ORDER BY a.sort ASC, a.id DESC
`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定选项查询失败")
	}
	platformAccounts := make([]adminapi.ProductGoodsChannelPlatformAccountOption, 0, len(platformRows))
	for _, row := range platformRows {
		connectStatus := row["last_balance_status"].Int()
		platformAccounts = append(platformAccounts, adminapi.ProductGoodsChannelPlatformAccountOption{
			ID:                row["id"].Int64(),
			Name:              row["name"].String(),
			SubjectID:         row["subject_id"].Int64(),
			SubjectName:       row["subject_name"].String(),
			HasTax:            row["has_tax"].Int(),
			ConnectStatus:     connectStatus,
			ConnectStatusText: connectStatusText(connectStatus),
		})
	}

	templateRows := make([]struct {
		ID    int64  `db:"id"`
		Title string `db:"title"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &templateRows, `SELECT id, title FROM product_template ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道绑定选项查询失败")
	}
	validateTemplates := make([]adminapi.ProductGoodsTemplateOption, 0, len(templateRows))
	for _, row := range templateRows {
		validateTemplates = append(validateTemplates, adminapi.ProductGoodsTemplateOption{ID: row.ID, Title: row.Title})
	}

	return &adminapi.ProductGoodsChannelBindingFormOptionsRes{
		PlatformAccounts:  platformAccounts,
		ValidateTemplates: validateTemplates,
		DockStatusOptions: []adminapi.ProductGoodsIntOption{
			{Value: consts.StatusEnabled, Label: "正常"},
			{Value: consts.StatusDisabled, Label: "关闭"},
		},
		AutoPriceTypeOptions: []adminapi.ProductGoodsStringOption{
			{Value: autoPriceAddTypeFixed, Label: "固定值"},
			{Value: autoPriceAddTypePercent, Label: "百分比"},
		},
	}, nil
}

func (l *ProductGoodsLogic) getProductGoodsChannelGoodsSummary(ctx context.Context, goodsID int64) (adminapi.ProductGoodsChannelGoodsSummary, error) {
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    p.id,
    p.goods_code,
    p.name,
    COALESCE(b.name, '') AS brand_name,
    p.subject_id,
    COALESCE(s.name, '') AS subject_name,
    p.has_tax,
    p.default_sell_price
FROM product_goods p
LEFT JOIN product_brand b ON b.id = p.brand_id
LEFT JOIN admin_subject s ON s.id = p.subject_id
WHERE p.id = ? AND p.is_deleted = 0
`, goodsID)
	if err != nil {
		return adminapi.ProductGoodsChannelGoodsSummary{}, apiErr(consts.CodeInternalError, "商品渠道绑定查询失败")
	}
	if row == nil || len(row) == 0 {
		return adminapi.ProductGoodsChannelGoodsSummary{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	return adminapi.ProductGoodsChannelGoodsSummary{
		ID:               row["id"].Int64(),
		GoodsCode:        row["goods_code"].String(),
		Name:             row["name"].String(),
		BrandName:        productGoodsRecordString(row, "brand_name"),
		SubjectID:        nullableInt64Pointer(productGoodsRecordNullInt64(row, "subject_id")),
		SubjectName:      productGoodsRecordString(row, "subject_name"),
		HasTax:           row["has_tax"].Int(),
		DefaultSellPrice: productGoodsRecordMoney(row, "default_sell_price"),
	}, nil
}

func (l *ProductGoodsLogic) loadProductGoodsChannelSummaries(ctx context.Context, goodsIDs []int64) (map[int64]productGoodsChannelSummary, error) {
	if len(goodsIDs) == 0 {
		return map[int64]productGoodsChannelSummary{}, nil
	}

	args := make([]any, 0, len(goodsIDs))
	for _, goodsID := range goodsIDs {
		args = append(args, goodsID)
	}
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    b.goods_id,
    b.id,
    a.name AS platform_account_name,
    b.cost_price,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    p.default_sell_price
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
JOIN product_goods p ON p.id = b.goods_id
WHERE b.is_deleted = 0 AND b.dock_status = 1 AND a.is_deleted = 0 AND b.goods_id IN (`+sqlPlaceholders(len(goodsIDs))+`)
ORDER BY b.goods_id ASC, b.sort ASC, b.id ASC
`, args...)
	if err != nil {
		return nil, err
	}

	summaries := make(map[int64]productGoodsChannelSummary, len(goodsIDs))
	channelNameSetByGoods := make(map[int64]map[string]struct{}, len(goodsIDs))
	minCostByGoods := make(map[int64]decimal.Decimal, len(goodsIDs))
	minEffectiveByGoods := make(map[int64]decimal.Decimal, len(goodsIDs))

	for _, row := range rows {
		goodsID := row["goods_id"].Int64()
		summary := summaries[goodsID]
		summary.BoundChannelCount++

		if summary.PrimaryChannelName == "" {
			summary.PrimaryChannelName = row["platform_account_name"].String()
		}

		if _, ok := channelNameSetByGoods[goodsID]; !ok {
			channelNameSetByGoods[goodsID] = map[string]struct{}{}
		}
		channelName := row["platform_account_name"].String()
		if _, exists := channelNameSetByGoods[goodsID][channelName]; !exists {
			channelNameSetByGoods[goodsID][channelName] = struct{}{}
			summary.BoundChannels = append(summary.BoundChannels, channelName)
		}

		costAmount, parseErr := decimal.NewFromString(productGoodsRecordMoney(row, "cost_price"))
		if parseErr != nil {
			return nil, parseErr
		}
		if current, ok := minCostByGoods[goodsID]; !ok || costAmount.LessThan(current) {
			minCostByGoods[goodsID] = costAmount
			summary.MinChannelCost = costAmount.StringFixed(4)
		}

		effectiveSellPrice, priceErr := computeChannelEffectiveSellPrice(productGoodsRecordMoney(row, "default_sell_price"), productGoodsRecordMoney(row, "cost_price"), row["is_auto_change"].Int(), row["add_type"].String(), productGoodsRecordMoney(row, "default_price"))
		if priceErr != nil {
			return nil, priceErr
		}
		if strings.TrimSpace(effectiveSellPrice) != "" {
			effectiveAmount, parseErr := decimal.NewFromString(effectiveSellPrice)
			if parseErr != nil {
				return nil, parseErr
			}
			if current, ok := minEffectiveByGoods[goodsID]; !ok || effectiveAmount.LessThan(current) {
				minEffectiveByGoods[goodsID] = effectiveAmount
				summary.MinChannelEffectiveSellPrice = effectiveAmount.StringFixed(4)
			}
		}

		if row["is_auto_change"].Int() == 1 {
			summary.ChannelAutoPriceStatus = 1
		}
		summaries[goodsID] = summary
	}

	return summaries, nil
}
