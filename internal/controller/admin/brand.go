package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type BrandController struct{ svc service.BrandService }

func NewBrand(svc service.BrandService) *BrandController { return &BrandController{svc: svc} }

func (c *BrandController) List(ctx context.Context, req *adminapi.BrandListReq) (res *adminapi.BrandListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *BrandController) Children(ctx context.Context, req *adminapi.BrandChildrenReq) (res *adminapi.BrandChildrenRes, err error) {
	return c.svc.Children(ctx, req)
}

func (c *BrandController) Create(ctx context.Context, req *adminapi.BrandCreateReq) (res *adminapi.BrandCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *BrandController) Update(ctx context.Context, req *adminapi.BrandUpdateReq) (res *adminapi.BrandUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *BrandController) Delete(ctx context.Context, req *adminapi.BrandDeleteReq) (res *adminapi.BrandDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *BrandController) Sort(ctx context.Context, req *adminapi.BrandSortReq) (res *adminapi.BrandSortRes, err error) {
	return c.svc.Sort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *BrandController) Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq) (res *adminapi.BrandVisibilityRes, err error) {
	return c.svc.Visibility(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *BrandController) Upload(ctx context.Context, req *adminapi.BrandUploadReq) (res *adminapi.BrandUploadRes, err error) {
	return c.svc.Upload(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
