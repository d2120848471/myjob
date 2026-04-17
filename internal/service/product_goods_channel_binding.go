package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// ProductGoodsChannelBindingService 定义商品渠道绑定（绑定级）相关能力。
type ProductGoodsChannelBindingService interface {
	List(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (*adminapi.ProductGoodsChannelBindingListRes, error)
	Create(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingCreateRes, error)
	Update(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingDeleteRes, error)
	BatchStatus(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingBatchStatusRes, error)
	BatchDelete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingBatchDeleteRes, error)
	Reorder(ctx context.Context, req *adminapi.ProductGoodsChannelBindingReorderReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingReorderRes, error)
	AutoPriceUpdate(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, error)
	AutoPriceBatch(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceBatchReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceBatchRes, error)
}
