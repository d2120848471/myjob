package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type PurchaseLimitController struct {
	svc service.PurchaseLimitService
}

func NewPurchaseLimit(svc service.PurchaseLimitService) *PurchaseLimitController {
	return &PurchaseLimitController{svc: svc}
}

func (c *PurchaseLimitController) List(ctx context.Context, req *adminapi.PurchaseLimitStrategyListReq) (res *adminapi.PurchaseLimitStrategyListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *PurchaseLimitController) Create(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq) (res *adminapi.PurchaseLimitStrategyCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *PurchaseLimitController) Update(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq) (res *adminapi.PurchaseLimitStrategyUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *PurchaseLimitController) Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq) (res *adminapi.PurchaseLimitStrategyDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *PurchaseLimitController) Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq) (res *adminapi.PurchaseLimitStrategyStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *PurchaseLimitController) Enums(ctx context.Context, req *adminapi.PurchaseLimitStrategyEnumsReq) (res *adminapi.PurchaseLimitStrategyEnumsRes, err error) {
	return c.svc.Enums(ctx, req)
}
