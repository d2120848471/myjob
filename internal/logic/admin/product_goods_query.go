package adminlogic

import (
	"context"
	"database/sql"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// List 分页查询商品列表，支持品牌/类型/含税/状态/关键字等筛选条件。
func (l *ProductGoodsLogic) List(ctx context.Context, req *adminapi.ProductGoodsListReq) (*adminapi.ProductGoodsListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	keyword := strings.TrimSpace(req.Keyword)
	goodsType := strings.TrimSpace(req.GoodsType)
	hasTax, hasHasTaxFilter, err := normalizeProductGoodsTriState(req.HasTax, "含税状态")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	status, hasStatusFilter, err := normalizeProductGoodsTriState(req.Status, "状态")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	conditions := []string{"p.is_deleted = 0"}
	args := make([]any, 0, 12)
	if keyword != "" {
		conditions = append(conditions, "(p.goods_code LIKE ? OR p.name LIKE ?)")
		likeKeyword := "%" + keyword + "%"
		args = append(args, likeKeyword, likeKeyword)
	}
	if req.BrandID > 0 {
		brandRows, loadErr := l.loadBrandRows(ctx)
		if loadErr != nil {
			return nil, apiErr(consts.CodeInternalError, "商品列表查询失败")
		}
		brandIDs := expandBrandIDs(brandRows, req.BrandID)
		if len(brandIDs) == 0 {
			return &adminapi.ProductGoodsListRes{
				List:       []adminapi.ProductGoodsListItem{},
				Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: 0},
			}, nil
		}
		conditions = append(conditions, "p.brand_id IN ("+sqlPlaceholders(len(brandIDs))+")")
		for _, brandID := range brandIDs {
			args = append(args, brandID)
		}
	}
	if goodsType != "" {
		conditions = append(conditions, "p.goods_type = ?")
		args = append(args, goodsType)
	}
	if hasHasTaxFilter {
		conditions = append(conditions, "p.has_tax = ?")
		args = append(args, hasTax)
	}
	if hasStatusFilter {
		conditions = append(conditions, "p.status = ?")
		args = append(args, status)
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods p WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品列表查询失败")
	}

	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
	SELECT
	    p.id,
	    p.goods_code,
	    p.brand_id,
    COALESCE(b.name, '') AS brand_name,
    COALESCE(NULLIF(b.icon, ''), NULLIF(pb.icon, ''), NULLIF(gb.icon, ''), '') AS brand_icon,
    p.subject_id,
    COALESCE(sub.name, '') AS subject_name,
    p.name,
    p.goods_type,
    p.supply_type,
    p.is_export,
    p.is_douyin,
    p.has_tax,
    p.exception_notify,
    p.product_template_id,
    t.title AS product_template_title,
    p.purchase_limit_strategy_id,
    s.name AS purchase_limit_strategy_name,
    p.default_sell_price,
    p.terminal_price_limit,
    p.status,
    p.created_at
FROM product_goods p
LEFT JOIN product_brand b ON b.id = p.brand_id
LEFT JOIN product_brand pb ON pb.id = b.parent_id
LEFT JOIN product_brand gb ON gb.id = pb.parent_id
LEFT JOIN admin_subject sub ON sub.id = p.subject_id
LEFT JOIN product_template t ON t.id = p.product_template_id
LEFT JOIN product_purchase_limit_strategy s ON s.id = p.purchase_limit_strategy_id
WHERE `+whereClause+`
	ORDER BY p.created_at DESC, p.id DESC
	LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品列表查询失败")
	}

	items := make([]adminapi.ProductGoodsListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.ProductGoodsListItem{
			ID:                           row["id"].Int64(),
			GoodsCode:                    row["goods_code"].String(),
			BrandID:                      row["brand_id"].Int64(),
			BrandName:                    row["brand_name"].String(),
			BrandIcon:                    productGoodsRecordString(row, "brand_icon"),
			SubjectID:                    nullableInt64Pointer(productGoodsRecordNullInt64(row, "subject_id")),
			SubjectName:                  productGoodsRecordString(row, "subject_name"),
			Name:                         row["name"].String(),
			GoodsType:                    row["goods_type"].String(),
			SupplyType:                   row["supply_type"].String(),
			IsExport:                     row["is_export"].Int(),
			IsDouyin:                     row["is_douyin"].Int(),
			HasTax:                       row["has_tax"].Int(),
			ExceptionNotify:              row["exception_notify"].Int(),
			ProductTemplateID:            productGoodsRecordInt64(row, "product_template_id"),
			ProductTemplateTitle:         productGoodsRecordString(row, "product_template_title"),
			PurchaseLimitStrategyID:      productGoodsRecordInt64(row, "purchase_limit_strategy_id"),
			PurchaseLimitStrategyName:    productGoodsRecordString(row, "purchase_limit_strategy_name"),
			DefaultSellPrice:             productGoodsRecordMoney(row, "default_sell_price"),
			TerminalPriceLimit:           productGoodsRecordMoney(row, "terminal_price_limit"),
			BoundChannels:                []string{},
			MinChannelCost:               "",
			MinChannelEffectiveSellPrice: "",
			Status:                       row["status"].Int(),
			CreatedAt:                    formatAppTime(parseRecordTime(row, "created_at")),
		})
	}

	if len(items) > 0 {
		goodsIDs := make([]int64, 0, len(items))
		indexByID := make(map[int64]int, len(items))
		for idx, item := range items {
			goodsIDs = append(goodsIDs, item.ID)
			indexByID[item.ID] = idx
		}

		summaries, summaryErr := l.loadProductGoodsChannelSummaries(ctx, goodsIDs)
		if summaryErr != nil {
			return nil, apiErr(consts.CodeInternalError, "商品列表查询失败")
		}
		for goodsID, summary := range summaries {
			idx, ok := indexByID[goodsID]
			if !ok {
				continue
			}
			items[idx].BoundChannels = summary.BoundChannels
			items[idx].BoundChannelCount = summary.BoundChannelCount
			items[idx].PrimaryChannelName = summary.PrimaryChannelName
			items[idx].MinChannelCost = summary.MinChannelCost
			items[idx].MinChannelEffectiveSellPrice = summary.MinChannelEffectiveSellPrice
			items[idx].ChannelAutoPriceStatus = summary.ChannelAutoPriceStatus
		}
	}

	return &adminapi.ProductGoodsListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

// Detail 查询商品详情（用于编辑回显等场景）。
func (l *ProductGoodsLogic) Detail(ctx context.Context, req *adminapi.ProductGoodsDetailReq) (*adminapi.ProductGoodsDetailRes, error) {
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    p.id,
    p.goods_code,
    p.brand_id,
    COALESCE(b.name, '') AS brand_name,
    p.name,
    p.goods_type,
    p.supply_type,
    p.is_export,
    p.is_douyin,
    p.has_tax,
    p.subject_id,
    COALESCE(sub.name, '') AS subject_name,
    p.exception_notify,
    p.product_template_id,
    t.title AS product_template_title,
    p.purchase_limit_strategy_id,
    s.name AS purchase_limit_strategy_name,
    s.status AS purchase_limit_strategy_status,
    p.purchase_notice,
    p.terminal_price_limit,
    p.balance_limit,
    p.default_sell_price,
    p.min_purchase_qty,
    p.max_purchase_qty,
    p.status,
    p.created_at,
    p.updated_at
FROM product_goods p
	LEFT JOIN product_brand b ON b.id = p.brand_id
	LEFT JOIN admin_subject sub ON sub.id = p.subject_id
	LEFT JOIN product_template t ON t.id = p.product_template_id
	LEFT JOIN product_purchase_limit_strategy s ON s.id = p.purchase_limit_strategy_id
	WHERE p.id = ? AND p.is_deleted = 0
`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品详情查询失败")
	}
	if row == nil || len(row) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	return &adminapi.ProductGoodsDetailRes{
		ID:                          row["id"].Int64(),
		GoodsCode:                   row["goods_code"].String(),
		BrandID:                     row["brand_id"].Int64(),
		BrandName:                   row["brand_name"].String(),
		Name:                        row["name"].String(),
		GoodsType:                   row["goods_type"].String(),
		SupplyType:                  row["supply_type"].String(),
		IsExport:                    row["is_export"].Int(),
		IsDouyin:                    row["is_douyin"].Int(),
		HasTax:                      row["has_tax"].Int(),
		SubjectID:                   nullableInt64Pointer(productGoodsRecordNullInt64(row, "subject_id")),
		SubjectName:                 productGoodsRecordString(row, "subject_name"),
		ExceptionNotify:             row["exception_notify"].Int(),
		ProductTemplateID:           nullableInt64Pointer(productGoodsRecordNullInt64(row, "product_template_id")),
		ProductTemplateTitle:        productGoodsRecordString(row, "product_template_title"),
		PurchaseLimitStrategyID:     nullableInt64Pointer(productGoodsRecordNullInt64(row, "purchase_limit_strategy_id")),
		PurchaseLimitStrategyName:   productGoodsRecordString(row, "purchase_limit_strategy_name"),
		PurchaseLimitStrategyStatus: int(productGoodsRecordInt64(row, "purchase_limit_strategy_status")),
		PurchaseNotice:              productGoodsRecordString(row, "purchase_notice"),
		TerminalPriceLimit:          productGoodsRecordMoney(row, "terminal_price_limit"),
		BalanceLimit:                formatMoney(row["balance_limit"].String()),
		DefaultSellPrice:            productGoodsRecordMoney(row, "default_sell_price"),
		MinPurchaseQty:              row["min_purchase_qty"].Int(),
		MaxPurchaseQty:              row["max_purchase_qty"].Int(),
		Status:                      row["status"].Int(),
		CreatedAt:                   formatAppTime(parseRecordTime(row, "created_at")),
		UpdatedAt:                   formatAppTime(parseRecordTime(row, "updated_at")),
	}, nil
}

func (l *ProductGoodsLogic) getActiveProduct(ctx context.Context, id int64) (entity.ProductGoods, error) {
	product := entity.ProductGoods{}
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    id,
    goods_code,
    brand_id,
    name,
    goods_type,
    supply_type,
    is_export,
    is_douyin,
    has_tax,
    subject_id,
    exception_notify,
    product_template_id,
    purchase_limit_strategy_id,
    purchase_notice,
    terminal_price_limit,
    balance_limit,
    default_sell_price,
    min_purchase_qty,
    max_purchase_qty,
    status,
    is_deleted,
    deleted_at,
    created_at,
    updated_at
FROM product_goods
WHERE id = ? AND is_deleted = 0
`, id)
	if err != nil {
		return product, err
	}
	if row == nil || len(row) == 0 {
		return product, sql.ErrNoRows
	}
	product.ID = row["id"].Int64()
	product.GoodsCode = row["goods_code"].String()
	product.BrandID = row["brand_id"].Int64()
	product.Name = row["name"].String()
	product.GoodsType = row["goods_type"].String()
	product.SupplyType = row["supply_type"].String()
	product.IsExport = row["is_export"].Int()
	product.IsDouyin = row["is_douyin"].Int()
	product.HasTax = row["has_tax"].Int()
	product.SubjectID = productGoodsRecordNullInt64(row, "subject_id")
	product.ExceptionNotify = row["exception_notify"].Int()
	product.ProductTemplateID = productGoodsRecordNullInt64(row, "product_template_id")
	product.PurchaseLimitStrategyID = productGoodsRecordNullInt64(row, "purchase_limit_strategy_id")
	product.PurchaseNotice = productGoodsRecordNullString(row, "purchase_notice")
	product.TerminalPriceLimit = productGoodsRecordNullString(row, "terminal_price_limit")
	product.BalanceLimit = row["balance_limit"].String()
	product.DefaultSellPrice = productGoodsRecordNullString(row, "default_sell_price")
	product.MinPurchaseQty = row["min_purchase_qty"].Int()
	product.MaxPurchaseQty = row["max_purchase_qty"].Int()
	product.Status = row["status"].Int()
	product.IsDeleted = row["is_deleted"].Int()
	product.DeletedAt = nullableTimeFromRecord(row, "deleted_at")
	product.CreatedAt = parseRecordTime(row, "created_at")
	product.UpdatedAt = parseRecordTime(row, "updated_at")
	return product, nil
}
