package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// ProductGoodsChannelConfigService 定义商品渠道配置（商品级）相关能力。
type ProductGoodsChannelConfigService interface {
	Get(ctx context.Context, req *adminapi.ProductGoodsChannelConfigGetReq) (*adminapi.ProductGoodsChannelConfigGetRes, error)
	Update(ctx context.Context, req *adminapi.ProductGoodsChannelConfigUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelConfigUpdateRes, error)
}
