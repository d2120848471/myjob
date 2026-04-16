package adminlogic

import (
	"context"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

// List 分页查询一级品牌列表（parent_id=0），支持按名称模糊搜索。
func (l *BrandLogic) List(ctx context.Context, req *adminapi.BrandListReq) (*adminapi.BrandListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	name := strings.TrimSpace(req.Name)
	likeName := "%" + name + "%"

	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE parent_id = 0
  AND (? = '' OR name LIKE ?)
`, name, likeName)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌列表查询失败")
	}

	type brandRow struct {
		ID              int64     `db:"id"`
		ParentID        int64     `db:"parent_id"`
		Name            string    `db:"name"`
		Icon            string    `db:"icon"`
		CredentialImage string    `db:"credential_image"`
		Description     string    `db:"description"`
		IsVisible       int       `db:"is_visible"`
		Sort            int       `db:"sort"`
		GoodsCount      int       `db:"goods_count"`
		HasChildren     int       `db:"has_children"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}
	rows := make([]brandRow, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id,
    b.parent_id,
    b.name,
    b.icon,
    b.credential_image,
    COALESCE(b.description, '') AS description,
    b.is_visible,
    b.sort,
    b.goods_count,
    CASE WHEN EXISTS(SELECT 1 FROM product_brand c WHERE c.parent_id = b.id) THEN 1 ELSE 0 END AS has_children,
    b.created_at,
    b.updated_at
FROM product_brand b
WHERE b.parent_id = 0
  AND (? = '' OR b.name LIKE ?)
ORDER BY b.sort ASC, b.id ASC
LIMIT ? OFFSET ?
`, name, likeName, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌列表查询失败")
	}

	items := make([]adminapi.BrandListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.BrandListItem{
			ID:              row.ID,
			ParentID:        row.ParentID,
			Name:            row.Name,
			Icon:            row.Icon,
			CredentialImage: row.CredentialImage,
			Description:     row.Description,
			IsVisible:       row.IsVisible,
			Sort:            row.Sort,
			GoodsCount:      row.GoodsCount,
			HasChildren:     row.HasChildren == 1,
			Children:        make([]adminapi.BrandListItem, 0),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return &adminapi.BrandListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()}}, nil
}

// Children 查询指定品牌下的子品牌列表。
func (l *BrandLogic) Children(ctx context.Context, req *adminapi.BrandChildrenReq) (*adminapi.BrandChildrenRes, error) {
	if _, err := l.getBrand(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	type brandRow struct {
		ID              int64     `db:"id"`
		ParentID        int64     `db:"parent_id"`
		Name            string    `db:"name"`
		Icon            string    `db:"icon"`
		CredentialImage string    `db:"credential_image"`
		Description     string    `db:"description"`
		IsVisible       int       `db:"is_visible"`
		Sort            int       `db:"sort"`
		GoodsCount      int       `db:"goods_count"`
		HasChildren     int       `db:"has_children"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}
	rows := make([]brandRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id,
    b.parent_id,
    b.name,
    b.icon,
    b.credential_image,
    COALESCE(b.description, '') AS description,
    b.is_visible,
    b.sort,
    b.goods_count,
    CASE WHEN EXISTS(SELECT 1 FROM product_brand c WHERE c.parent_id = b.id) THEN 1 ELSE 0 END AS has_children,
    b.created_at,
    b.updated_at
FROM product_brand b
WHERE b.parent_id = ?
ORDER BY sort ASC, id ASC
`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "子品牌查询失败")
	}
	items := make([]app.BrandListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, app.BrandListItem{
			ID:              row.ID,
			ParentID:        row.ParentID,
			Name:            row.Name,
			Icon:            row.Icon,
			CredentialImage: row.CredentialImage,
			Description:     row.Description,
			IsVisible:       row.IsVisible,
			Sort:            row.Sort,
			GoodsCount:      row.GoodsCount,
			HasChildren:     row.HasChildren == 1,
			Children:        make([]app.BrandListItem, 0),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return &adminapi.BrandChildrenRes{List: items}, nil
}

func (l *BrandLogic) getBrand(ctx context.Context, id int64) (app.ProductBrand, error) {
	brand := app.ProductBrand{}
	err := l.core.DB().GetCore().GetScan(ctx, &brand, `
SELECT id, parent_id, name, icon, credential_image, COALESCE(description, '') AS description,
       is_visible, sort, goods_count, created_at, updated_at
FROM product_brand
WHERE id = ?
`, id)
	return brand, err
}

func (l *BrandLogic) siblingNameExists(ctx context.Context, parentID int64, name string, excludeID int64) (bool, error) {
	count, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE parent_id = ? AND name = ? AND id <> ?
`, parentID, name, excludeID)
	if err != nil {
		return false, err
	}
	return count.Int() > 0, nil
}

func (l *BrandLogic) nextSort(ctx context.Context, parentID int64) (int, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COALESCE(MAX(sort), 0) + 1 FROM product_brand WHERE parent_id = ?`, parentID)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

func (l *BrandLogic) loadSiblingOrder(ctx context.Context, parentID int64) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id FROM product_brand WHERE parent_id = ? ORDER BY sort ASC, id ASC`, parentID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (l *BrandLogic) loadSiblingOrderTx(tx gdb.TX, parentID int64) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := tx.GetScan(&rows, `SELECT id FROM product_brand WHERE parent_id = ? ORDER BY sort ASC, id ASC`, parentID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}
