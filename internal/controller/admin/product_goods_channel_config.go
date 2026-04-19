package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
)

// GetInventoryConfig 返回指定商品的库存配置与策略选项。
func (c *ProductGoodsController) GetInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigGetReq) (res *adminapi.ProductGoodsInventoryConfigGetRes, err error) {
	return c.svc.GetInventoryConfig(ctx, req)
}

// SaveInventoryConfig 保存指定商品的库存配置，并记录操作人与客户端 IP。
func (c *ProductGoodsController) SaveInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigSaveReq) (res *adminapi.ProductGoodsInventoryConfigSaveRes, err error) {
	return c.svc.SaveInventoryConfig(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
