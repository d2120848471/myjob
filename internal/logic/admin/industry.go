package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

// IndustryLogic 提供行业管理及行业-品牌关联管理相关业务能力。
type IndustryLogic struct{ core *app.Core }

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

// Add 新增行业，并可同时绑定品牌（在同一事务内同步 brand_count）。
func (l *IndustryLogic) Add(ctx context.Context, req *adminapi.IndustryCreateReq, actor app.AdminUser, ip string) (*adminapi.IndustryCreateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(consts.CodeBadRequest, "行业名称不能为空")
	}
	brandIDs, err := uniqueInt64s(req.BrandIDs)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if len(brandIDs) > 0 {
		if err := l.validateTopLevelBrands(ctx, brandIDs); err != nil {
			return nil, err
		}
	}
	exists, err := l.industryNameExists(ctx, req.Name, 0)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "行业名称已存在")
	}
	sortValue, err := l.nextIndustrySort(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业排序初始化失败")
	}
	var industryID int64
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, txErr := tx.Exec(`INSERT INTO product_industry (name, sort, brand_count, created_at, updated_at) VALUES (?, ?, 0, ?, ?)`, req.Name, sortValue, l.core.Now(), l.core.Now())
		if txErr != nil {
			return txErr
		}
		industryID, _ = result.LastInsertId()
		if txErr = l.replaceIndustryBrands(tx, industryID, brandIDs); txErr != nil {
			return txErr
		}
		return l.syncIndustryBrandCount(tx, industryID)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业新增失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加行业：%s", req.Name), ip)
	return &adminapi.IndustryCreateRes{ID: industryID}, nil
}

// Edit 编辑行业信息；当传入 BrandIDs 时会在事务内替换关联并同步 brand_count。
func (l *IndustryLogic) Edit(ctx context.Context, req *adminapi.IndustryUpdateReq, actor app.AdminUser, ip string) (*adminapi.IndustryUpdateRes, error) {
	industry, err := l.getIndustry(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(consts.CodeBadRequest, "行业名称不能为空")
	}
	brandIDs, err := uniqueInt64s(req.BrandIDs)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if req.BrandIDs != nil && len(brandIDs) > 0 {
		if err = l.validateTopLevelBrands(ctx, brandIDs); err != nil {
			return nil, err
		}
	}
	exists, err := l.industryNameExists(ctx, req.Name, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "行业名称已存在")
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`UPDATE product_industry SET name = ?, updated_at = ? WHERE id = ?`, req.Name, l.core.Now(), req.ID); txErr != nil {
			return txErr
		}
		if req.BrandIDs != nil {
			// 这里依赖 nil/empty slice 语义区分“未传品牌列表”和“明确清空关联”。
			if txErr := l.replaceIndustryBrands(tx, req.ID, brandIDs); txErr != nil {
				return txErr
			}
			return l.syncIndustryBrandCount(tx, req.ID)
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑行业：%d -> %s", industry.ID, req.Name), ip)
	return &adminapi.IndustryUpdateRes{}, nil
}

// Delete 删除行业（要求无关联品牌），并重新压实行业排序。
func (l *IndustryLogic) Delete(ctx context.Context, req *adminapi.IndustryDeleteReq, actor app.AdminUser, ip string) (*adminapi.IndustryDeleteRes, error) {
	industry, err := l.getIndustry(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	relationCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_industry_brand WHERE industry_id = ?`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业删除校验失败")
	}
	if industry.BrandCount > 0 || relationCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "该行业下存在关联品牌，请先解除关联后再删除")
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM product_industry WHERE id = ?`, req.ID); txErr != nil {
			return txErr
		}
		orderedIDs, txErr := l.loadIndustryOrderTx(tx)
		if txErr != nil {
			return txErr
		}
		return l.rebuildIndustrySort(tx, orderedIDs)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除行业：%d", req.ID), ip)
	return &adminapi.IndustryDeleteRes{}, nil
}

// Sort 调整行业排序（top/up/down/bottom）。
func (l *IndustryLogic) Sort(ctx context.Context, req *adminapi.IndustrySortReq, actor app.AdminUser, ip string) (*adminapi.IndustrySortRes, error) {
	if _, err := l.getIndustry(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	action, err := normalizeSortAction(req.Action)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	orderedIDs, err := l.loadIndustryOrder(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业排序读取失败")
	}
	newOrder := moveIDByAction(orderedIDs, req.ID, action)
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return l.rebuildIndustrySort(tx, newOrder)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业排序失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("调整行业排序：%d -> %s", req.ID, action), ip)
	return &adminapi.IndustrySortRes{}, nil
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

// BrandAdd 批量为行业新增品牌关联，并同步 brand_count。
func (l *IndustryLogic) BrandAdd(ctx context.Context, req *adminapi.IndustryBrandAddReq, actor app.AdminUser, ip string) (*adminapi.IndustryBrandAddRes, error) {
	if _, err := l.getIndustry(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	brandIDs, err := uniqueInt64s(req.BrandIDs)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if len(brandIDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "品牌ID列表不能为空")
	}
	if err := l.validateTopLevelBrands(ctx, brandIDs); err != nil {
		return nil, err
	}
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		rows := make([]struct {
			BrandID int64 `db:"brand_id"`
		}, 0)
		if txErr := tx.GetScan(&rows, `SELECT brand_id FROM product_industry_brand WHERE industry_id = ? ORDER BY sort ASC, id ASC`, req.ID); txErr != nil {
			return txErr
		}
		existing := make(map[int64]struct{}, len(rows))
		for _, row := range rows {
			existing[row.BrandID] = struct{}{}
		}
		nextSortValue, txErr := l.maxIndustryBrandSort(tx, req.ID)
		if txErr != nil {
			return txErr
		}
		addedCount := 0
		for _, brandID := range brandIDs {
			if _, ok := existing[brandID]; ok {
				continue
			}
			nextSortValue++
			addedCount++
			if _, txErr = tx.Exec(`INSERT INTO product_industry_brand (industry_id, brand_id, sort, created_at) VALUES (?, ?, ?, ?)`, req.ID, brandID, nextSortValue, l.core.Now()); txErr != nil {
				return txErr
			}
		}
		if addedCount == 0 {
			return nil
		}
		return l.syncIndustryBrandCount(tx, req.ID)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业添加品牌失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("行业添加品牌：industry=%d, count=%d", req.ID, len(brandIDs)), ip)
	return &adminapi.IndustryBrandAddRes{}, nil
}

// BrandDelete 批量删除行业-品牌关联，并同步 brand_count 与关联排序。
func (l *IndustryLogic) BrandDelete(ctx context.Context, req *adminapi.IndustryBrandDeleteReq, actor app.AdminUser, ip string) (*adminapi.IndustryBrandDeleteRes, error) {
	if _, err := l.getIndustry(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	brandIDs, err := uniqueInt64s(req.BrandIDs)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if len(brandIDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "品牌ID列表不能为空")
	}
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		args := make([]any, 0, len(brandIDs)+1)
		args = append(args, req.ID)
		for _, brandID := range brandIDs {
			args = append(args, brandID)
		}
		if _, txErr := tx.Exec(`DELETE FROM product_industry_brand WHERE industry_id = ? AND brand_id IN (`+sqlPlaceholders(len(brandIDs))+`)`, args...); txErr != nil {
			return txErr
		}
		orderedIDs, txErr := l.loadIndustryBrandOrderTx(tx, req.ID)
		if txErr != nil {
			return txErr
		}
		if txErr = l.rebuildIndustryBrandSort(tx, req.ID, orderedIDs); txErr != nil {
			return txErr
		}
		return l.syncIndustryBrandCount(tx, req.ID)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业删除品牌失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("行业删除品牌：industry=%d, count=%d", req.ID, len(brandIDs)), ip)
	return &adminapi.IndustryBrandDeleteRes{}, nil
}

// BrandSort 调整行业下品牌关联的排序（top/up/down/bottom）。
func (l *IndustryLogic) BrandSort(ctx context.Context, req *adminapi.IndustryBrandSortReq, actor app.AdminUser, ip string) (*adminapi.IndustryBrandSortRes, error) {
	if _, err := l.getIndustry(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "行业不存在")
	}
	action, err := normalizeSortAction(req.Action)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	orderedIDs, err := l.loadIndustryBrandOrder(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业品牌排序读取失败")
	}
	if indexOfID(orderedIDs, req.BrandID) < 0 {
		return nil, apiErr(consts.CodeBadRequest, "品牌未关联到该行业")
	}
	newOrder := moveIDByAction(orderedIDs, req.BrandID, action)
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return l.rebuildIndustryBrandSort(tx, req.ID, newOrder)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "行业品牌排序失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("行业品牌排序：industry=%d, brand=%d, action=%s", req.ID, req.BrandID, action), ip)
	return &adminapi.IndustryBrandSortRes{}, nil
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

func (l *IndustryLogic) rebuildIndustrySort(tx gdb.TX, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.Exec(`UPDATE product_industry SET sort = ?, updated_at = ? WHERE id = ?`, index+1, l.core.Now(), id); err != nil {
			return err
		}
	}
	return nil
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

func (l *IndustryLogic) rebuildIndustryBrandSort(tx gdb.TX, industryID int64, orderedBrandIDs []int64) error {
	for index, brandID := range orderedBrandIDs {
		if _, err := tx.Exec(`UPDATE product_industry_brand SET sort = ? WHERE industry_id = ? AND brand_id = ?`, index+1, industryID, brandID); err != nil {
			return err
		}
	}
	return nil
}

func (l *IndustryLogic) replaceIndustryBrands(tx gdb.TX, industryID int64, brandIDs []int64) error {
	if _, err := tx.Exec(`DELETE FROM product_industry_brand WHERE industry_id = ?`, industryID); err != nil {
		return err
	}
	for index, brandID := range brandIDs {
		if _, err := tx.Exec(`INSERT INTO product_industry_brand (industry_id, brand_id, sort, created_at) VALUES (?, ?, ?, ?)`, industryID, brandID, index+1, l.core.Now()); err != nil {
			return err
		}
	}
	return nil
}

func (l *IndustryLogic) syncIndustryBrandCount(tx gdb.TX, industryID int64) error {
	count, err := tx.GetValue(`SELECT COUNT(*) FROM product_industry_brand WHERE industry_id = ?`, industryID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE product_industry SET brand_count = ?, updated_at = ? WHERE id = ?`, count.Int(), l.core.Now(), industryID)
	return err
}

func (l *IndustryLogic) validateTopLevelBrands(ctx context.Context, brandIDs []int64) error {
	if len(brandIDs) == 0 {
		return nil
	}
	args := make([]any, 0, len(brandIDs))
	for _, brandID := range brandIDs {
		args = append(args, brandID)
	}
	count, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE id IN (`+sqlPlaceholders(len(brandIDs))+`)
  AND parent_id = 0
`, args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "品牌校验失败")
	}
	if count.Int() != len(brandIDs) {
		return apiErr(consts.CodeBadRequest, "行业仅允许关联一级品牌")
	}
	return nil
}

func (l *IndustryLogic) maxIndustryBrandSort(tx gdb.TX, industryID int64) (int, error) {
	value, err := tx.GetValue(`SELECT COALESCE(MAX(sort), 0) FROM product_industry_brand WHERE industry_id = ?`, industryID)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}
