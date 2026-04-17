package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// ProductGoodsChannelConfigController 提供商品渠道配置（商品级）相关 HTTP handler。
type ProductGoodsChannelConfigController struct {
	svc service.ProductGoodsChannelConfigService
}

// NewProductGoodsChannelConfig 创建 ProductGoodsChannelConfigController。
func NewProductGoodsChannelConfig(svc service.ProductGoodsChannelConfigService) *ProductGoodsChannelConfigController {
	return &ProductGoodsChannelConfigController{svc: svc}
}

// Get 返回指定商品的渠道配置与绑定摘要。
func (c *ProductGoodsChannelConfigController) Get(ctx context.Context, req *adminapi.ProductGoodsChannelConfigGetReq) (res *adminapi.ProductGoodsChannelConfigGetRes, err error) {
	return c.svc.Get(ctx, req)
}

// Update 更新指定商品的渠道配置，并记录操作人与客户端 IP。
func (c *ProductGoodsChannelConfigController) Update(ctx context.Context, req *adminapi.ProductGoodsChannelConfigUpdateReq) (res *adminapi.ProductGoodsChannelConfigUpdateRes, err error) {
	return c.svc.Update(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
