package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type ProductTemplateController struct {
	svc service.ProductTemplateService
}

func NewProductTemplate(svc service.ProductTemplateService) *ProductTemplateController {
	return &ProductTemplateController{svc: svc}
}

func (c *ProductTemplateController) List(ctx context.Context, req *adminapi.ProductTemplateListReq) (res *adminapi.ProductTemplateListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *ProductTemplateController) Create(ctx context.Context, req *adminapi.ProductTemplateCreateReq) (res *adminapi.ProductTemplateCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductTemplateController) Update(ctx context.Context, req *adminapi.ProductTemplateUpdateReq) (res *adminapi.ProductTemplateUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductTemplateController) Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq) (res *adminapi.ProductTemplateDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductTemplateController) BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq) (res *adminapi.ProductTemplateBatchDeleteRes, err error) {
	return c.svc.BatchDelete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductTemplateController) ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (res *adminapi.ProductTemplateValidateTypeListRes, err error) {
	return c.svc.ValidateTypes(ctx, req)
}
