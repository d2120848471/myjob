package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// Add 新增商品模板，并写入操作日志。
func (l *ProductTemplateLogic) Add(ctx context.Context, req *adminapi.ProductTemplateCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateCreateRes, error) {
	normalized, err := normalizeProductTemplateInput(req.Title, req.Type, req.IsShared, req.AccountName, req.ValidateType)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_template (
    title, template_type, is_shared, account_name, validate_type, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
`, normalized.Title, normalized.Type, normalized.IsShared, normalized.AccountName, normalized.ValidateType, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品模板：%s", normalized.Title), ip)
	return &adminapi.ProductTemplateCreateRes{ID: id}, nil
}

// Edit 编辑商品模板，并写入操作日志。
func (l *ProductTemplateLogic) Edit(ctx context.Context, req *adminapi.ProductTemplateUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateUpdateRes, error) {
	if _, err := l.getTemplate(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	normalized, err := normalizeProductTemplateInput(req.Title, req.Type, req.IsShared, req.AccountName, req.ValidateType)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_template
SET title = ?, template_type = ?, is_shared = ?, account_name = ?, validate_type = ?, updated_at = ?
WHERE id = ?
`, normalized.Title, normalized.Type, normalized.IsShared, normalized.AccountName, normalized.ValidateType, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品模板：%d -> %s", req.ID, normalized.Title), ip)
	return &adminapi.ProductTemplateUpdateRes{}, nil
}

// Delete 删除商品模板（要求未被商品引用），并写入操作日志。
func (l *ProductTemplateLogic) Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateDeleteRes, error) {
	template, err := l.getTemplate(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	goodsRefCount, err := countActiveGoodsReference(ctx, l.core.DB(), "product_template_id", req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除校验失败")
	}
	if goodsRefCount > 0 {
		return nil, apiErr(consts.CodeConflict, "该商品模板已被商品引用，请先处理关联商品")
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_template WHERE id = ?`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品模板：%d -> %s", req.ID, template.Title), ip)
	return &adminapi.ProductTemplateDeleteRes{}, nil
}

// BatchDelete 批量删除商品模板（要求均未被商品引用），并写入操作日志。
func (l *ProductTemplateLogic) BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateBatchDeleteRes, error) {
	if len(req.IDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "请至少选择一个商品模板")
	}
	ids, err := uniquePositiveInt64s(req.IDs, "模板ID")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	existing, err := l.loadTemplateIDs(ctx, ids)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板查询失败")
	}
	if len(existing) != len(ids) {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	referenced, err := hasActiveGoodsReferences(ctx, l.core.DB(), "product_template_id", ids)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除校验失败")
	}
	if referenced {
		return nil, apiErr(consts.CodeConflict, "选中的商品模板中存在被商品引用的记录，请先处理关联商品")
	}

	// 批量删除前先做全量存在性校验，避免部分删除成功、部分模板不存在导致结果不确定。
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_template WHERE id IN (`+sqlPlaceholders(len(ids))+`)`, args...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板批量删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量删除商品模板：%d项", len(ids)), ip)
	return &adminapi.ProductTemplateBatchDeleteRes{}, nil
}
