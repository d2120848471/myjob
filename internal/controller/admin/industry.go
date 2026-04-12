package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type IndustryController struct{ svc service.IndustryService }

func NewIndustry(svc service.IndustryService) *IndustryController {
	return &IndustryController{svc: svc}
}

func (c *IndustryController) List(ctx context.Context, req *adminapi.IndustryListReq) (res *adminapi.IndustryListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *IndustryController) Create(ctx context.Context, req *adminapi.IndustryCreateReq) (res *adminapi.IndustryCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) Update(ctx context.Context, req *adminapi.IndustryUpdateReq) (res *adminapi.IndustryUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) Delete(ctx context.Context, req *adminapi.IndustryDeleteReq) (res *adminapi.IndustryDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) Sort(ctx context.Context, req *adminapi.IndustrySortReq) (res *adminapi.IndustrySortRes, err error) {
	return c.svc.Sort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) BrandSelector(ctx context.Context, req *adminapi.IndustryBrandSelectorReq) (res *adminapi.IndustryBrandSelectorRes, err error) {
	return c.svc.BrandSelector(ctx, req)
}

func (c *IndustryController) BrandList(ctx context.Context, req *adminapi.IndustryBrandListReq) (res *adminapi.IndustryBrandListRes, err error) {
	return c.svc.BrandList(ctx, req)
}

func (c *IndustryController) BrandAdd(ctx context.Context, req *adminapi.IndustryBrandAddReq) (res *adminapi.IndustryBrandAddRes, err error) {
	return c.svc.BrandAdd(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) BrandDelete(ctx context.Context, req *adminapi.IndustryBrandDeleteReq) (res *adminapi.IndustryBrandDeleteRes, err error) {
	return c.svc.BrandDelete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *IndustryController) BrandSort(ctx context.Context, req *adminapi.IndustryBrandSortReq) (res *adminapi.IndustryBrandSortRes, err error) {
	return c.svc.BrandSort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
