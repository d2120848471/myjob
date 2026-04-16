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

// Add 新增品牌，并写入操作日志。
func (l *BrandLogic) Add(ctx context.Context, req *adminapi.BrandCreateReq, actor app.AdminUser, ip string) (*adminapi.BrandCreateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	req.CredentialImage = strings.TrimSpace(req.CredentialImage)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || (req.IsVisible != 0 && req.IsVisible != 1) || req.ParentID < 0 {
		return nil, apiErr(consts.CodeBadRequest, "品牌参数错误")
	}
	_, level, err := l.validateParent(ctx, req.ParentID)
	if err != nil {
		return nil, err
	}
	exists, err := l.siblingNameExists(ctx, req.ParentID, req.Name, 0)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "同级品牌名称已存在")
	}
	sortValue, err := l.nextSort(ctx, req.ParentID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序初始化失败")
	}
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_brand (
    parent_id, name, icon, credential_image, description, is_visible, sort, goods_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
`, req.ParentID, req.Name, req.Icon, req.CredentialImage, req.Description, req.IsVisible, sortValue, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, l.buildBrandCreateLog(level, req.Name, req.ParentID), ip)
	return &adminapi.BrandCreateRes{ID: id}, nil
}

// Edit 编辑品牌信息，并写入操作日志。
func (l *BrandLogic) Edit(ctx context.Context, req *adminapi.BrandUpdateReq, actor app.AdminUser, ip string) (*adminapi.BrandUpdateRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	req.CredentialImage = strings.TrimSpace(req.CredentialImage)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || (req.IsVisible != 0 && req.IsVisible != 1) {
		return nil, apiErr(consts.CodeBadRequest, "品牌参数错误")
	}
	exists, err := l.siblingNameExists(ctx, brand.ParentID, req.Name, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "同级品牌名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_brand
SET name = ?, icon = ?, credential_image = ?, description = ?, is_visible = ?, updated_at = ?
WHERE id = ?
`, req.Name, req.Icon, req.CredentialImage, req.Description, req.IsVisible, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑品牌：%d -> %s", req.ID, req.Name), ip)
	return &adminapi.BrandUpdateRes{}, nil
}

// Delete 删除品牌（要求无子品牌且未被行业/商品引用），并写入操作日志。
func (l *BrandLogic) Delete(ctx context.Context, req *adminapi.BrandDeleteReq, actor app.AdminUser, ip string) (*adminapi.BrandDeleteRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	childCount, countErr := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_brand WHERE parent_id = ?`, req.ID)
	if countErr != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除校验失败")
	}
	if childCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "该品牌下存在子品牌，请先删除子品牌")
	}
	industryRefCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_industry_brand WHERE brand_id = ?`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除校验失败")
	}
	if industryRefCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "该品牌已被行业关联，请先解除关联")
	}
	goodsRefCount, err := countActiveGoodsReference(ctx, l.core.DB(), "brand_id", req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除校验失败")
	}
	if goodsRefCount > 0 {
		return nil, apiErr(consts.CodeConflict, "该品牌已被商品引用，请先处理关联商品")
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM product_brand WHERE id = ?`, req.ID); txErr != nil {
			return txErr
		}
		orderedIDs, txErr := l.loadSiblingOrderTx(tx, brand.ParentID)
		if txErr != nil {
			return txErr
		}
		// 删除后重新压实同级排序，避免留下空洞 sort。
		return l.reorderSiblingSortTx(tx, orderedIDs)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除品牌：%d", req.ID), ip)
	return &adminapi.BrandDeleteRes{}, nil
}

// Sort 调整同级品牌排序（top/up/down/bottom）。
func (l *BrandLogic) Sort(ctx context.Context, req *adminapi.BrandSortReq, actor app.AdminUser, ip string) (*adminapi.BrandSortRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	action, err := normalizeSortAction(req.Action)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	orderedIDs, err := l.loadSiblingOrder(ctx, brand.ParentID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序读取失败")
	}
	newOrder := moveIDByAction(orderedIDs, req.ID, action)
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 品牌排序统一走“内存换位 + 事务回写”，避免 MySQL/SQLite 方言差异。
		return l.reorderSiblingSortTx(tx, newOrder)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("调整品牌排序：%d -> %s", req.ID, action), ip)
	return &adminapi.BrandSortRes{}, nil
}

// Visibility 切换品牌显示状态，并写入操作日志。
func (l *BrandLogic) Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq, actor app.AdminUser, ip string) (*adminapi.BrandVisibilityRes, error) {
	if req.IsVisible != 0 && req.IsVisible != 1 {
		return nil, apiErr(consts.CodeBadRequest, "显示状态错误")
	}
	if _, err := l.getBrand(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE product_brand SET is_visible = ?, updated_at = ? WHERE id = ?`, req.IsVisible, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌显隐更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换品牌显示状态：%d -> %d", req.ID, req.IsVisible), ip)
	return &adminapi.BrandVisibilityRes{}, nil
}

func (l *BrandLogic) reorderSiblingSortTx(tx gdb.TX, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.Exec(`UPDATE product_brand SET sort = ?, updated_at = ? WHERE id = ?`, index+1, l.core.Now(), id); err != nil {
			return err
		}
	}
	return nil
}
