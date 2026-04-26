package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// SupplierProductCallbackService 定义开放供应商商品变动回调处理能力。
type SupplierProductCallbackService interface {
	HandleSupplierProductChangeCallback(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq, body []byte) error
}

// SupplierProductSubscriptionService 定义管理端商品订阅记录查询和变更能力。
type SupplierProductSubscriptionService interface {
	ListSupplierProductSubscriptions(ctx context.Context, req *adminapi.SupplierProductSubscriptionListReq) (*adminapi.SupplierProductSubscriptionListRes, error)
	CancelSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionCancelReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionCancelRes, error)
	ResubscribeSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionResubscribeReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionResubscribeRes, error)
}

// ProductGoodsChannelPriceChangeService 定义商品渠道改价记录查询能力。
type ProductGoodsChannelPriceChangeService interface {
	ListProductGoodsChannelPriceChanges(ctx context.Context, req *adminapi.ProductGoodsChannelPriceChangeListReq) (*adminapi.ProductGoodsChannelPriceChangeListRes, error)
}
