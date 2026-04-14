package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type ProductGoodsController struct {
	svc service.ProductGoodsService
}

func NewProductGoods(svc service.ProductGoodsService) *ProductGoodsController {
	return &ProductGoodsController{svc: svc}
}

func (c *ProductGoodsController) List(ctx context.Context, req *adminapi.ProductGoodsListReq) (res *adminapi.ProductGoodsListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *ProductGoodsController) Detail(ctx context.Context, req *adminapi.ProductGoodsDetailReq) (res *adminapi.ProductGoodsDetailRes, err error) {
	return c.svc.Detail(ctx, req)
}

func (c *ProductGoodsController) FormOptions(ctx context.Context, req *adminapi.ProductGoodsFormOptionsReq) (res *adminapi.ProductGoodsFormOptionsRes, err error) {
	return c.svc.FormOptions(ctx, req)
}

func (c *ProductGoodsController) Create(ctx context.Context, req *adminapi.ProductGoodsCreateReq) (res *adminapi.ProductGoodsCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductGoodsController) Update(ctx context.Context, req *adminapi.ProductGoodsUpdateReq) (res *adminapi.ProductGoodsUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *ProductGoodsController) Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq) (res *adminapi.ProductGoodsDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
