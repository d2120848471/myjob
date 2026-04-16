package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// ProductGoodsController 提供商品管理相关 HTTP handler。
type ProductGoodsController struct {
	svc service.ProductGoodsService
}

// NewProductGoods 创建 ProductGoodsController。
func NewProductGoods(svc service.ProductGoodsService) *ProductGoodsController {
	return &ProductGoodsController{svc: svc}
}

// List 返回商品分页列表。
func (c *ProductGoodsController) List(ctx context.Context, req *adminapi.ProductGoodsListReq) (res *adminapi.ProductGoodsListRes, err error) {
	return c.svc.List(ctx, req)
}

// Detail 返回指定商品详情。
func (c *ProductGoodsController) Detail(ctx context.Context, req *adminapi.ProductGoodsDetailReq) (res *adminapi.ProductGoodsDetailRes, err error) {
	return c.svc.Detail(ctx, req)
}

// FormOptions 返回商品表单所需的聚合下拉数据。
func (c *ProductGoodsController) FormOptions(ctx context.Context, req *adminapi.ProductGoodsFormOptionsReq) (res *adminapi.ProductGoodsFormOptionsRes, err error) {
	return c.svc.FormOptions(ctx, req)
}

// Create 新增商品，并记录操作人与客户端 IP。
func (c *ProductGoodsController) Create(ctx context.Context, req *adminapi.ProductGoodsCreateReq) (res *adminapi.ProductGoodsCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑商品，并记录操作人与客户端 IP。
func (c *ProductGoodsController) Update(ctx context.Context, req *adminapi.ProductGoodsUpdateReq) (res *adminapi.ProductGoodsUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除商品（软删），并记录操作人与客户端 IP。
func (c *ProductGoodsController) Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq) (res *adminapi.ProductGoodsDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Status 批量切换商品状态，并记录操作人与客户端 IP。
func (c *ProductGoodsController) Status(ctx context.Context, req *adminapi.ProductGoodsStatusReq) (res *adminapi.ProductGoodsStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
