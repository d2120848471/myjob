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
