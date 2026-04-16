package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

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

func (l *IndustryLogic) maxIndustryBrandSort(tx gdb.TX, industryID int64) (int, error) {
	value, err := tx.GetValue(`SELECT COALESCE(MAX(sort), 0) FROM product_industry_brand WHERE industry_id = ?`, industryID)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}
