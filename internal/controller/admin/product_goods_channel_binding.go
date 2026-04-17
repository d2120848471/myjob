package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// ProductGoodsChannelBindingController 提供商品渠道绑定（绑定级）相关 HTTP handler。
type ProductGoodsChannelBindingController struct {
	svc service.ProductGoodsChannelBindingService
}

// NewProductGoodsChannelBinding 创建 ProductGoodsChannelBindingController。
func NewProductGoodsChannelBinding(svc service.ProductGoodsChannelBindingService) *ProductGoodsChannelBindingController {
	return &ProductGoodsChannelBindingController{svc: svc}
}

// List 返回指定商品的绑定列表。
func (c *ProductGoodsChannelBindingController) List(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (res *adminapi.ProductGoodsChannelBindingListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增绑定，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) Create(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq) (res *adminapi.ProductGoodsChannelBindingCreateRes, err error) {
	return c.svc.Create(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 更新绑定基础字段，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) Update(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq) (res *adminapi.ProductGoodsChannelBindingUpdateRes, err error) {
	return c.svc.Update(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除绑定（软删），并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) Delete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq) (res *adminapi.ProductGoodsChannelBindingDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BatchStatus 批量启停绑定，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) BatchStatus(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchStatusReq) (res *adminapi.ProductGoodsChannelBindingBatchStatusRes, err error) {
	return c.svc.BatchStatus(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BatchDelete 批量删除绑定（软删），并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) BatchDelete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchDeleteReq) (res *adminapi.ProductGoodsChannelBindingBatchDeleteRes, err error) {
	return c.svc.BatchDelete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Reorder 一键排序绑定，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) Reorder(ctx context.Context, req *adminapi.ProductGoodsChannelBindingReorderReq) (res *adminapi.ProductGoodsChannelBindingReorderRes, err error) {
	return c.svc.Reorder(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// AutoPriceUpdate 更新单条绑定自动改价字段，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) AutoPriceUpdate(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq) (res *adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, err error) {
	return c.svc.AutoPriceUpdate(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// AutoPriceBatch 批量更新自动改价字段，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelBindingController) AutoPriceBatch(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceBatchReq) (res *adminapi.ProductGoodsChannelBindingAutoPriceBatchRes, err error) {
	return c.svc.AutoPriceBatch(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
