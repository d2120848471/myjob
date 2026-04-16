package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

// List 分页查询行业列表，支持按名称模糊搜索。
func (l *IndustryLogic) List(ctx context.Context, req *adminapi.IndustryListReq) (*adminapi.IndustryListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	name := strings.TrimSpace(req.Name)
	likeName := "%" + name + "%"

	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_industry
WHERE (? = '' OR name LIKE ?)
`, name, likeName)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业列表查询失败")
	}

	items := make([]app.IndustryListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT id, name, sort, brand_count, created_at, updated_at
FROM product_industry
WHERE (? = '' OR name LIKE ?)
ORDER BY sort ASC, id ASC
LIMIT ? OFFSET ?
`, name, likeName, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业列表查询失败")
	}
	return &adminapi.IndustryListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()}}, nil
}

// BrandSelector 查询可绑定的一级品牌选择器列表。
func (l *IndustryLogic) BrandSelector(ctx context.Context, req *adminapi.IndustryBrandSelectorReq) (*adminapi.IndustryBrandSelectorRes, error) {
	name := strings.TrimSpace(req.Name)
	likeName := "%" + name + "%"
	items := make([]app.BrandSelectorItem, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT id, name, icon
FROM product_brand
WHERE parent_id = 0
  AND (? = '' OR name LIKE ?)
ORDER BY sort ASC, id ASC
`, name, likeName); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌选择器查询失败")
	}
	return &adminapi.IndustryBrandSelectorRes{List: items}, nil
}

// BrandList 查询行业已绑定的品牌列表。
func (l *IndustryLogic) BrandList(ctx context.Context, req *adminapi.IndustryBrandListReq) (*adminapi.IndustryBrandListRes, error) {
	industry, err := l.getIndustry(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	name := strings.TrimSpace(req.Name)
	likeName := "%" + name + "%"
	items := make([]app.IndustryBrandRelationItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT
    ib.id,
    ib.brand_id,
    b.name AS brand_name,
    b.icon AS brand_icon,
    ib.sort
FROM product_industry_brand ib
JOIN product_brand b ON b.id = ib.brand_id
WHERE ib.industry_id = ?
  AND (? = '' OR b.name LIKE ?)
ORDER BY ib.sort ASC, ib.id ASC
`, req.ID, name, likeName); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业品牌列表查询失败")
	}
	return &adminapi.IndustryBrandListRes{IndustryID: industry.ID, IndustryName: industry.Name, List: items}, nil
}

func (l *IndustryLogic) getIndustry(ctx context.Context, id int64) (app.ProductIndustry, error) {
	industry := app.ProductIndustry{}
	err := l.core.DB().GetCore().GetScan(ctx, &industry, `SELECT id, name, sort, brand_count, created_at, updated_at FROM product_industry WHERE id = ?`, id)
	return industry, err
}

func (l *IndustryLogic) industryNameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_industry WHERE name = ? AND id <> ?`, name, excludeID)
	if err != nil {
		return false, err
	}
	return count.Int() > 0, nil
}

func (l *IndustryLogic) nextIndustrySort(ctx context.Context) (int, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COALESCE(MAX(sort), 0) + 1 FROM product_industry`)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

func (l *IndustryLogic) loadIndustryOrder(ctx context.Context) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id FROM product_industry ORDER BY sort ASC, id ASC`); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (l *IndustryLogic) loadIndustryOrderTx(tx gdb.TX) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := tx.GetScan(&rows, `SELECT id FROM product_industry ORDER BY sort ASC, id ASC`); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (l *IndustryLogic) loadIndustryBrandOrder(ctx context.Context, industryID int64) ([]int64, error) {
	rows := make([]struct {
		BrandID int64 `db:"brand_id"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT brand_id FROM product_industry_brand WHERE industry_id = ? ORDER BY sort ASC, id ASC`, industryID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.BrandID)
	}
	return ids, nil
}

func (l *IndustryLogic) loadIndustryBrandOrderTx(tx gdb.TX, industryID int64) ([]int64, error) {
	rows := make([]struct {
		BrandID int64 `db:"brand_id"`
	}, 0)
	if err := tx.GetScan(&rows, `SELECT brand_id FROM product_industry_brand WHERE industry_id = ? ORDER BY sort ASC, id ASC`, industryID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.BrandID)
	}
	return ids, nil
}
