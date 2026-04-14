package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SupplierPlatformController struct {
	svc service.SupplierPlatformService
}

func NewSupplierPlatform(svc service.SupplierPlatformService) *SupplierPlatformController {
	return &SupplierPlatformController{svc: svc}
}

func (c *SupplierPlatformController) TypeList(ctx context.Context, req *adminapi.SupplierPlatformTypeListReq) (res *adminapi.SupplierPlatformTypeListRes, err error) {
	return c.svc.TypeList(ctx, req)
}

func (c *SupplierPlatformController) List(ctx context.Context, req *adminapi.SupplierPlatformListReq) (res *adminapi.SupplierPlatformListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *SupplierPlatformController) Detail(ctx context.Context, req *adminapi.SupplierPlatformDetailReq) (res *adminapi.SupplierPlatformDetailRes, err error) {
	return c.svc.Detail(ctx, req)
}

func (c *SupplierPlatformController) Create(ctx context.Context, req *adminapi.SupplierPlatformCreateReq) (res *adminapi.SupplierPlatformCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SupplierPlatformController) Update(ctx context.Context, req *adminapi.SupplierPlatformUpdateReq) (res *adminapi.SupplierPlatformUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SupplierPlatformController) Delete(ctx context.Context, req *adminapi.SupplierPlatformDeleteReq) (res *adminapi.SupplierPlatformDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SupplierPlatformController) RefreshBalance(ctx context.Context, req *adminapi.SupplierPlatformRefreshBalanceReq) (res *adminapi.SupplierPlatformRefreshBalanceRes, err error) {
	return c.svc.RefreshBalance(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
