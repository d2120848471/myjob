package adminlogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

const (
	productGoodsTypeCardSecret     = "card_secret"
	productGoodsTypeDirectRecharge = "direct_recharge"
	productGoodsSupplyTypeChannel  = "channel"
)

var productGoodsTypeLabels = map[string]string{
	productGoodsTypeCardSecret:     "卡密",
	productGoodsTypeDirectRecharge: "直充",
}

type ProductGoodsLogic struct{ core *app.Core }

type normalizedProductGoodsInput struct {
	BrandID                 int64
	Name                    string
	GoodsType               string
	SupplyType              string
	IsExport                int
	IsDouyin                int
	HasTax                  int
	SubjectID               *int64
	ExceptionNotify         int
	ProductTemplateID       *int64
	PurchaseLimitStrategyID *int64
	PurchaseNotice          string
	TerminalPriceLimit      string
	BalanceLimit            string
	DefaultSellPrice        string
	MinPurchaseQty          int
	MaxPurchaseQty          int
	Status                  int
}

type productBrandTreeRow struct {
	ID       int64  `db:"id"`
	ParentID int64  `db:"parent_id"`
	Name     string `db:"name"`
	Sort     int    `db:"sort"`
}

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

	rows := make([]gdb.Record, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if rows, err = l.core.DB().GetCore().GetAll(ctx, `
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
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品列表查询失败")
	}

	items := make([]adminapi.ProductGoodsListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.ProductGoodsListItem{
			ID:                        row["id"].Int64(),
			GoodsCode:                 row["goods_code"].String(),
			BrandID:                   row["brand_id"].Int64(),
			BrandName:                 row["brand_name"].String(),
			BrandIcon:                 productGoodsRecordString(row, "brand_icon"),
			SubjectID:                 nullableInt64Pointer(productGoodsRecordNullInt64(row, "subject_id")),
			SubjectName:               productGoodsRecordString(row, "subject_name"),
			Name:                      row["name"].String(),
			GoodsType:                 row["goods_type"].String(),
			SupplyType:                row["supply_type"].String(),
			IsExport:                  row["is_export"].Int(),
			IsDouyin:                  row["is_douyin"].Int(),
			HasTax:                    row["has_tax"].Int(),
			ExceptionNotify:           row["exception_notify"].Int(),
			ProductTemplateID:         productGoodsRecordInt64(row, "product_template_id"),
			ProductTemplateTitle:      productGoodsRecordString(row, "product_template_title"),
			PurchaseLimitStrategyID:   productGoodsRecordInt64(row, "purchase_limit_strategy_id"),
			PurchaseLimitStrategyName: productGoodsRecordString(row, "purchase_limit_strategy_name"),
			DefaultSellPrice:          productGoodsRecordMoney(row, "default_sell_price"),
			TerminalPriceLimit:        productGoodsRecordMoney(row, "terminal_price_limit"),
			Status:                    row["status"].Int(),
			CreatedAt:                 formatAppTime(parseRecordTime(row, "created_at")),
		})
	}

	return &adminapi.ProductGoodsListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

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

func (l *ProductGoodsLogic) FormOptions(ctx context.Context, _ *adminapi.ProductGoodsFormOptionsReq) (*adminapi.ProductGoodsFormOptionsRes, error) {
	brandRows, err := l.loadBrandRows(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}

	templateRows := make([]struct {
		ID    int64  `db:"id"`
		Title string `db:"title"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &templateRows, `SELECT id, title FROM product_template ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	templates := make([]adminapi.ProductGoodsTemplateOption, 0, len(templateRows))
	for _, row := range templateRows {
		templates = append(templates, adminapi.ProductGoodsTemplateOption{ID: row.ID, Title: row.Title})
	}

	strategyRows := make([]struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &strategyRows, `SELECT id, name FROM product_purchase_limit_strategy WHERE status = ? ORDER BY id DESC`, consts.StatusEnabled); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	strategies := make([]adminapi.ProductGoodsStrategyOption, 0, len(strategyRows))
	for _, row := range strategyRows {
		strategies = append(strategies, adminapi.ProductGoodsStrategyOption{ID: row.ID, Name: row.Name})
	}

	subjectRows := make([]struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &subjectRows, `SELECT id, name FROM admin_subject WHERE has_tax = 1 ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品表单下拉查询失败")
	}
	subjects := make([]adminapi.ProductGoodsSubjectOption, 0, len(subjectRows))
	for _, row := range subjectRows {
		subjects = append(subjects, adminapi.ProductGoodsSubjectOption{ID: row.ID, Name: row.Name})
	}

	return &adminapi.ProductGoodsFormOptionsRes{
		Brands:                  buildProductBrandTree(brandRows),
		Templates:               templates,
		PurchaseLimitStrategies: strategies,
		Subjects:                subjects,
		GoodsTypes: []adminapi.ProductGoodsStringOption{
			{Value: productGoodsTypeCardSecret, Label: productGoodsTypeLabels[productGoodsTypeCardSecret]},
			{Value: productGoodsTypeDirectRecharge, Label: productGoodsTypeLabels[productGoodsTypeDirectRecharge]},
		},
		SupplyTypes: []adminapi.ProductGoodsStringOption{
			{Value: productGoodsSupplyTypeChannel, Label: "渠道供货"},
		},
		BooleanOptions: []adminapi.ProductGoodsIntOption{
			{Value: 1, Label: "是"},
			{Value: 0, Label: "否"},
		},
		StatusOptions: []adminapi.ProductGoodsIntOption{
			{Value: 1, Label: "启用"},
			{Value: 0, Label: "停用"},
		},
	}, nil
}

func (l *ProductGoodsLogic) Add(ctx context.Context, req *adminapi.ProductGoodsCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsCreateRes, error) {
	normalized, err := l.normalizeProductGoodsInput(ctx, req.BrandID, req.Name, req.GoodsType, req.SupplyType, req.IsExport, req.IsDouyin, req.HasTax, req.SubjectID, req.ExceptionNotify, req.ProductTemplateID, req.PurchaseLimitStrategyID, req.PurchaseNotice, req.TerminalPriceLimit, req.BalanceLimit, req.DefaultSellPrice, req.MinPurchaseQty, req.MaxPurchaseQty, req.Status, nil)
	if err != nil {
		return nil, err
	}

	createdID := int64(0)
	createdCode := ""
	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tempGoodsCode := temporaryProductGoodsCode(now)
		result, txErr := tx.Exec(`
INSERT INTO product_goods (
    goods_code, brand_id, name, goods_type, supply_type, is_export, is_douyin, has_tax, subject_id, exception_notify,
    product_template_id, purchase_limit_strategy_id, purchase_notice, terminal_price_limit, balance_limit,
    default_sell_price, min_purchase_qty, max_purchase_qty, status, is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
`, tempGoodsCode, normalized.BrandID, normalized.Name, normalized.GoodsType, normalized.SupplyType, normalized.IsExport, normalized.IsDouyin, normalized.HasTax, nullableInt64Arg(normalized.SubjectID), normalized.ExceptionNotify, nullableInt64Arg(normalized.ProductTemplateID), nullableInt64Arg(normalized.PurchaseLimitStrategyID), nullableStringArg(normalized.PurchaseNotice), nullableMoneyArg(normalized.TerminalPriceLimit), normalized.BalanceLimit, nullableMoneyArg(normalized.DefaultSellPrice), normalized.MinPurchaseQty, normalized.MaxPurchaseQty, normalized.Status, now, now)
		if txErr != nil {
			return txErr
		}
		id, _ := result.LastInsertId()
		goodsCode := buildGoodsCode(id)
		if _, txErr = tx.Exec(`UPDATE product_goods SET goods_code = ?, updated_at = ? WHERE id = ?`, goodsCode, now, id); txErr != nil {
			return txErr
		}
		if txErr = adjustBrandGoodsCountTx(tx, normalized.BrandID, 1, now); txErr != nil {
			return txErr
		}
		createdID = id
		createdCode = goodsCode
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品新增失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品：%s", normalized.Name), ip)
	return &adminapi.ProductGoodsCreateRes{ID: createdID, GoodsCode: createdCode}, nil
}

func (l *ProductGoodsLogic) Edit(ctx context.Context, req *adminapi.ProductGoodsUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsUpdateRes, error) {
	current, err := l.getActiveProduct(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	currentStrategyID := nullableInt64Pointer(current.PurchaseLimitStrategyID)
	normalized, err := l.normalizeProductGoodsInput(ctx, req.BrandID, req.Name, req.GoodsType, req.SupplyType, req.IsExport, req.IsDouyin, req.HasTax, req.SubjectID, req.ExceptionNotify, req.ProductTemplateID, req.PurchaseLimitStrategyID, req.PurchaseNotice, req.TerminalPriceLimit, req.BalanceLimit, req.DefaultSellPrice, req.MinPurchaseQty, req.MaxPurchaseQty, req.Status, currentStrategyID)
	if err != nil {
		return nil, err
	}

	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先确认商品仍是未删除状态，再处理品牌计数，避免并发软删把 goods_count 改漂。
		result, txErr := tx.Exec(`
UPDATE product_goods
SET brand_id = ?, name = ?, goods_type = ?, supply_type = ?, is_export = ?, is_douyin = ?, has_tax = ?, subject_id = ?, exception_notify = ?,
    product_template_id = ?, purchase_limit_strategy_id = ?, purchase_notice = ?, terminal_price_limit = ?, balance_limit = ?,
    default_sell_price = ?, min_purchase_qty = ?, max_purchase_qty = ?, status = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, normalized.BrandID, normalized.Name, normalized.GoodsType, normalized.SupplyType, normalized.IsExport, normalized.IsDouyin, normalized.HasTax, nullableInt64Arg(normalized.SubjectID), normalized.ExceptionNotify, nullableInt64Arg(normalized.ProductTemplateID), nullableInt64Arg(normalized.PurchaseLimitStrategyID), nullableStringArg(normalized.PurchaseNotice), nullableMoneyArg(normalized.TerminalPriceLimit), normalized.BalanceLimit, nullableMoneyArg(normalized.DefaultSellPrice), normalized.MinPurchaseQty, normalized.MaxPurchaseQty, normalized.Status, now, req.ID)
		if txErr != nil {
			return txErr
		}
		if txErr = ensureMutationAffected(result); txErr != nil {
			return txErr
		}
		if normalized.BrandID != current.BrandID {
			if txErr = adjustBrandGoodsCountTx(tx, current.BrandID, -1, now); txErr != nil {
				return txErr
			}
			if txErr = adjustBrandGoodsCountTx(tx, normalized.BrandID, 1, now); txErr != nil {
				return txErr
			}
		}
		return nil
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品编辑失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.ProductGoodsUpdateRes{}, nil
}

func (l *ProductGoodsLogic) Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsDeleteRes, error) {
	current, err := l.getActiveProduct(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 删除先校验命中行数，避免重复删除时继续扣减品牌计数。
		result, txErr := tx.Exec(`UPDATE product_goods SET is_deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`, now, now, req.ID)
		if txErr != nil {
			return txErr
		}
		if txErr = ensureMutationAffected(result); txErr != nil {
			return txErr
		}
		return adjustBrandGoodsCountTx(tx, current.BrandID, -1, now)
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品删除失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品：%d -> %s", req.ID, current.Name), ip)
	return &adminapi.ProductGoodsDeleteRes{}, nil
}

func (l *ProductGoodsLogic) Status(ctx context.Context, req *adminapi.ProductGoodsStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsStatusRes, error) {
	if len(req.IDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "请至少选择一个商品")
	}
	ids, err := uniquePositiveInt64s(req.IDs, "商品ID")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if req.Status != consts.StatusEnabled && req.Status != consts.StatusDisabled {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}

	successIDs := make([]int64, 0, len(ids))
	failed := make([]adminapi.ProductGoodsStatusFailedItem, 0)
	now := l.core.Now()
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		existingIDs, txErr := loadActiveProductIDSetTx(tx, ids, l.core.Config().Database.Driver)
		if txErr != nil {
			return txErr
		}
		for _, id := range ids {
			if _, ok := existingIDs[id]; ok {
				successIDs = append(successIDs, id)
				continue
			}
			failed = append(failed, adminapi.ProductGoodsStatusFailedItem{
				ID:     id,
				Reason: "商品不存在",
			})
		}
		if len(successIDs) == 0 {
			return nil
		}

		args := make([]any, 0, len(successIDs)+2)
		args = append(args, req.Status, now)
		for _, id := range successIDs {
			args = append(args, id)
		}
		// 事务内先锁定、再更新，避免并发软删把未命中的商品误算进 success_ids。
		if _, txErr = tx.Exec(`
UPDATE product_goods
SET status = ?, updated_at = ?
WHERE is_deleted = 0 AND id IN (`+sqlPlaceholders(len(successIDs))+`)
`, args...); txErr != nil {
			return txErr
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品状态更新失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量修改商品状态：status=%d, total=%d, success=%d, failed=%d", req.Status, len(ids), len(successIDs), len(failed)), ip)
	return &adminapi.ProductGoodsStatusRes{
		SuccessIDs:   successIDs,
		SuccessCount: len(successIDs),
		FailedCount:  len(failed),
		Failed:       failed,
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

func loadActiveProductIDSetTx(tx gdb.TX, ids []int64, driver string) (map[int64]struct{}, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if err := tx.GetScan(&rows, productGoodsStatusSelectSQL(driver, len(ids)), args...); err != nil {
		return nil, err
	}
	result := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		result[row.ID] = struct{}{}
	}
	return result, nil
}

func productGoodsStatusSelectSQL(driver string, idCount int) string {
	query := `SELECT id FROM product_goods WHERE is_deleted = 0 AND id IN (` + sqlPlaceholders(idCount) + `)`
	if strings.EqualFold(strings.TrimSpace(driver), "mysql") {
		return query + ` FOR UPDATE`
	}
	return query
}

func (l *ProductGoodsLogic) normalizeProductGoodsInput(ctx context.Context, brandID int64, name, goodsType, supplyType string, isExport, isDouyin, hasTax int, subjectID *int64, exceptionNotify int, productTemplateID, purchaseLimitStrategyID *int64, purchaseNotice, terminalPriceLimit, balanceLimit, defaultSellPrice string, minPurchaseQty, maxPurchaseQty, status int, allowDisabledCurrentStrategyID *int64) (normalizedProductGoodsInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "商品名称不能为空")
	}
	if _, ok := productGoodsTypeLabels[strings.TrimSpace(goodsType)]; !ok {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "商品类型错误")
	}
	supplyType = strings.TrimSpace(strings.ToLower(supplyType))
	if supplyType != productGoodsSupplyTypeChannel {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "供货方式错误")
	}
	if err := validateBooleanFlag(isExport, "可导出"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(isDouyin, "可抖音"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(hasTax, "含税标识"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(exceptionNotify, "异常提醒"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if status != consts.StatusEnabled && status != consts.StatusDisabled {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if _, err := l.validateLeafBrand(ctx, brandID); err != nil {
		return normalizedProductGoodsInput{}, err
	}

	normalizedSubjectID := normalizeOptionalID(subjectID)
	if hasTax == 1 {
		if normalizedSubjectID == nil {
			return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "含税商品必须选择主体")
		}
		// 含税商品后续会走开票链路，这里提前锁死主体存在且可开票。
		if err := l.ensureTaxSubjectUsable(ctx, *normalizedSubjectID); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	} else {
		// 不含税商品不保留主体，避免后续出现脏数据。
		normalizedSubjectID = nil
	}

	normalizedTemplateID := normalizeOptionalID(productTemplateID)
	if normalizedTemplateID != nil {
		if err := l.ensureTemplateExists(ctx, *normalizedTemplateID); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	}

	normalizedStrategyID := normalizeOptionalID(purchaseLimitStrategyID)
	if normalizedStrategyID != nil {
		allowDisabled := allowDisabledCurrentStrategyID != nil && *allowDisabledCurrentStrategyID == *normalizedStrategyID
		if err := l.ensureStrategyUsable(ctx, *normalizedStrategyID, allowDisabled); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	}

	normalizedTerminalPriceLimit, err := normalizeOptionalMoney(terminalPriceLimit)
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "终端限价格式错误")
	}
	normalizedBalanceLimit, err := normalizeDefaultMoney(balanceLimit, "0")
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "余额限制格式错误")
	}
	normalizedDefaultSellPrice, err := normalizeOptionalMoney(defaultSellPrice)
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "默认售价格式错误")
	}
	if minPurchaseQty < 1 {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "最小购买数量必须大于等于1")
	}
	if maxPurchaseQty < minPurchaseQty {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "最大购买数量不能小于最小购买数量")
	}

	return normalizedProductGoodsInput{
		BrandID:                 brandID,
		Name:                    name,
		GoodsType:               strings.TrimSpace(goodsType),
		SupplyType:              supplyType,
		IsExport:                isExport,
		IsDouyin:                isDouyin,
		HasTax:                  hasTax,
		SubjectID:               normalizedSubjectID,
		ExceptionNotify:         exceptionNotify,
		ProductTemplateID:       normalizedTemplateID,
		PurchaseLimitStrategyID: normalizedStrategyID,
		PurchaseNotice:          strings.TrimSpace(purchaseNotice),
		TerminalPriceLimit:      normalizedTerminalPriceLimit,
		BalanceLimit:            normalizedBalanceLimit,
		DefaultSellPrice:        normalizedDefaultSellPrice,
		MinPurchaseQty:          minPurchaseQty,
		MaxPurchaseQty:          maxPurchaseQty,
		Status:                  status,
	}, nil
}

func (l *ProductGoodsLogic) validateLeafBrand(ctx context.Context, brandID int64) (entity.ProductBrand, error) {
	if brandID <= 0 {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌不能为空")
	}
	brand := entity.ProductBrand{}
	if err := l.core.DB().GetCore().GetScan(ctx, &brand, `SELECT id, parent_id, name, icon, credential_image, COALESCE(description, '') AS description, is_visible, sort, goods_count, created_at, updated_at FROM product_brand WHERE id = ?`, brandID); err != nil {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	childCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_brand WHERE parent_id = ?`, brandID)
	if err != nil {
		return entity.ProductBrand{}, apiErr(consts.CodeInternalError, "品牌校验失败")
	}
	if childCount.Int() > 0 {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌必须选择末级品牌")
	}
	return brand, nil
}

func (l *ProductGoodsLogic) ensureTemplateExists(ctx context.Context, id int64) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_template WHERE id = ?`, id)
	if err != nil {
		return apiErr(consts.CodeInternalError, "商品模板校验失败")
	}
	if count.Int() == 0 {
		return apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	return nil
}

func (l *ProductGoodsLogic) ensureTaxSubjectUsable(ctx context.Context, id int64) error {
	row := struct {
		ID     int64 `db:"id"`
		HasTax int   `db:"has_tax"`
	}{}
	if err := l.core.DB().GetCore().GetScan(ctx, &row, `SELECT id, has_tax FROM admin_subject WHERE id = ?`, id); err != nil {
		return apiErr(consts.CodeBadRequest, "主体不存在")
	}
	if row.HasTax != 1 {
		return apiErr(consts.CodeBadRequest, "含税商品必须选择含税主体")
	}
	return nil
}

func (l *ProductGoodsLogic) ensureStrategyUsable(ctx context.Context, id int64, allowDisabled bool) error {
	row := struct {
		ID     int64 `db:"id"`
		Status int   `db:"status"`
	}{}
	if err := l.core.DB().GetCore().GetScan(ctx, &row, `SELECT id, status FROM product_purchase_limit_strategy WHERE id = ?`, id); err != nil {
		return apiErr(consts.CodeBadRequest, "购买数量限制策略不存在")
	}
	if row.Status != consts.StatusEnabled && !allowDisabled {
		return apiErr(consts.CodeBadRequest, "购买数量限制策略必须为启用状态")
	}
	return nil
}

func (l *ProductGoodsLogic) loadBrandRows(ctx context.Context) ([]productBrandTreeRow, error) {
	rows := make([]productBrandTreeRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id, parent_id, name, sort FROM product_brand ORDER BY sort ASC, id ASC`); err != nil {
		return nil, err
	}
	return rows, nil
}

func buildProductBrandTree(rows []productBrandTreeRow) []adminapi.ProductGoodsBrandTreeItem {
	childrenByParent := make(map[int64][]productBrandTreeRow, len(rows))
	for _, row := range rows {
		childrenByParent[row.ParentID] = append(childrenByParent[row.ParentID], row)
	}

	var build func(parentID int64) []adminapi.ProductGoodsBrandTreeItem
	build = func(parentID int64) []adminapi.ProductGoodsBrandTreeItem {
		children := childrenByParent[parentID]
		items := make([]adminapi.ProductGoodsBrandTreeItem, 0, len(children))
		for _, child := range children {
			grandChildren := build(child.ID)
			items = append(items, adminapi.ProductGoodsBrandTreeItem{
				ID:       child.ID,
				Name:     child.Name,
				IsLeaf:   len(grandChildren) == 0,
				Children: grandChildren,
			})
		}
		return items
	}

	return build(0)
}

func expandBrandIDs(rows []productBrandTreeRow, rootID int64) []int64 {
	childrenByParent := make(map[int64][]int64, len(rows))
	exists := false
	for _, row := range rows {
		if row.ID == rootID {
			exists = true
		}
		childrenByParent[row.ParentID] = append(childrenByParent[row.ParentID], row.ID)
	}
	if !exists {
		return nil
	}
	result := make([]int64, 0, 8)
	queue := []int64{rootID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, childrenByParent[current]...)
	}
	return result
}

func normalizeProductGoodsTriState(value string, fieldName string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("%s筛选错误", fieldName)
	}
}

func validateBooleanFlag(value int, fieldName string) error {
	if value != 0 && value != 1 {
		return apiErr(consts.CodeBadRequest, fieldName+"错误")
	}
	return nil
}

func normalizeOptionalID(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	id := *value
	return &id
}

func normalizeOptionalMoney(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.IsNegative() {
		return "", fmt.Errorf("money format invalid")
	}
	return amount.StringFixed(4), nil
}

func normalizeDefaultMoney(value string, defaultValue string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultValue
	}
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.IsNegative() {
		return "", fmt.Errorf("money format invalid")
	}
	return amount.StringFixed(4), nil
}

func nullableInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	id := value.Int64
	return &id
}

func nullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableMoney(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return formatMoney(value.String)
}

func productGoodsRecordString(row gdb.Record, key string) string {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return ""
	}
	return value.String()
}

func productGoodsRecordInt64(row gdb.Record, key string) int64 {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return 0
	}
	return value.Int64()
}

func productGoodsRecordMoney(row gdb.Record, key string) string {
	value := productGoodsRecordString(row, key)
	if value == "" {
		return ""
	}
	return formatMoney(value)
}

func productGoodsRecordNullString(row gdb.Record, key string) sql.NullString {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func productGoodsRecordNullInt64(row gdb.Record, key string) sql.NullInt64 {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value.Int64(), Valid: true}
}

func nullableStringArg(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableMoneyArg(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableInt64Arg(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatAppTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func buildGoodsCode(id int64) string {
	return fmt.Sprintf("GD%010d", id)
}

func temporaryProductGoodsCode(now time.Time) string {
	return fmt.Sprintf("TMP%d", now.UnixNano())
}

func adjustBrandGoodsCountTx(tx gdb.TX, brandID int64, delta int, now time.Time) error {
	if delta == 0 {
		return nil
	}
	if _, err := tx.Exec(`
UPDATE product_brand
SET goods_count = CASE WHEN goods_count + ? < 0 THEN 0 ELSE goods_count + ? END,
    updated_at = ?
WHERE id = ?
`, delta, delta, now, brandID); err != nil {
		return err
	}
	return nil
}

func ensureMutationAffected(result sql.Result) error {
	if result == nil {
		return sql.ErrNoRows
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func countActiveGoodsReference(ctx context.Context, db gdb.DB, column string, id int64) (int, error) {
	value, err := db.GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods WHERE is_deleted = 0 AND `+column+` = ?`, id)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

func hasActiveGoodsReferences(ctx context.Context, db gdb.DB, column string, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if err := db.GetCore().GetScan(ctx, &rows, `SELECT DISTINCT `+column+` AS id FROM product_goods WHERE is_deleted = 0 AND `+column+` IN (`+sqlPlaceholders(len(ids))+`) LIMIT 1`, args...); err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}
