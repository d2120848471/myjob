package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// ProductGoodsService 定义商品管理相关能力。
type ProductGoodsService interface {
	List(ctx context.Context, req *adminapi.ProductGoodsListReq) (*adminapi.ProductGoodsListRes, error)
	Detail(ctx context.Context, req *adminapi.ProductGoodsDetailReq) (*adminapi.ProductGoodsDetailRes, error)
	FormOptions(ctx context.Context, req *adminapi.ProductGoodsFormOptionsReq) (*adminapi.ProductGoodsFormOptionsRes, error)
	ChannelBindingList(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (*adminapi.ProductGoodsChannelBindingListRes, error)
	ChannelBindingFormOptions(ctx context.Context, req *adminapi.ProductGoodsChannelBindingFormOptionsReq) (*adminapi.ProductGoodsChannelBindingFormOptionsRes, error)
	GetInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigGetReq) (*adminapi.ProductGoodsInventoryConfigGetRes, error)
	Add(ctx context.Context, req *adminapi.ProductGoodsCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsCreateRes, error)
	Edit(ctx context.Context, req *adminapi.ProductGoodsUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsDeleteRes, error)
	Status(ctx context.Context, req *adminapi.ProductGoodsStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsStatusRes, error)
	SaveInventoryConfig(ctx context.Context, req *adminapi.ProductGoodsInventoryConfigSaveReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsInventoryConfigSaveRes, error)
	CreateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingCreateRes, error)
	UpdateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingUpdateRes, error)
	DeleteChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingDeleteRes, error)
	UpdateChannelBindingAutoPrice(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, error)
}
