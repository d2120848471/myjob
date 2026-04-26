package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// ProductGoodsChannelPriceChangeController 提供商品渠道改价记录查询 HTTP handler。
type ProductGoodsChannelPriceChangeController struct {
	svc service.ProductGoodsChannelPriceChangeService
}

// NewProductGoodsChannelPriceChange 创建商品渠道改价记录控制器。
func NewProductGoodsChannelPriceChange(svc service.ProductGoodsChannelPriceChangeService) *ProductGoodsChannelPriceChangeController {
	return &ProductGoodsChannelPriceChangeController{svc: svc}
}

// List 返回商品渠道自动改价记录分页列表。
func (c *ProductGoodsChannelPriceChangeController) List(ctx context.Context, req *adminapi.ProductGoodsChannelPriceChangeListReq) (res *adminapi.ProductGoodsChannelPriceChangeListRes, err error) {
	return c.svc.ListProductGoodsChannelPriceChanges(ctx, req)
}
