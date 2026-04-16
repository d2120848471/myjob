package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// ProductTemplateController 提供商品模板管理相关 HTTP handler。
type ProductTemplateController struct {
	svc service.ProductTemplateService
}

// NewProductTemplate 创建 ProductTemplateController。
func NewProductTemplate(svc service.ProductTemplateService) *ProductTemplateController {
	return &ProductTemplateController{svc: svc}
}

// List 返回商品模板分页列表。
func (c *ProductTemplateController) List(ctx context.Context, req *adminapi.ProductTemplateListReq) (res *adminapi.ProductTemplateListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增商品模板，并记录操作人与客户端 IP。
func (c *ProductTemplateController) Create(ctx context.Context, req *adminapi.ProductTemplateCreateReq) (res *adminapi.ProductTemplateCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑商品模板，并记录操作人与客户端 IP。
func (c *ProductTemplateController) Update(ctx context.Context, req *adminapi.ProductTemplateUpdateReq) (res *adminapi.ProductTemplateUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除单个商品模板，并记录操作人与客户端 IP。
func (c *ProductTemplateController) Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq) (res *adminapi.ProductTemplateDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BatchDelete 批量删除商品模板，并记录操作人与客户端 IP。
func (c *ProductTemplateController) BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq) (res *adminapi.ProductTemplateBatchDeleteRes, err error) {
	return c.svc.BatchDelete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// ValidateTypes 返回商品模板支持的验证方式枚举列表。
func (c *ProductTemplateController) ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (res *adminapi.ProductTemplateValidateTypeListRes, err error) {
	return c.svc.ValidateTypes(ctx, req)
}
