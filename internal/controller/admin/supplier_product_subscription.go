package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// SupplierProductSubscriptionController 提供供应商商品订阅记录相关 HTTP handler。
type SupplierProductSubscriptionController struct {
	svc service.SupplierProductSubscriptionService
}

// NewSupplierProductSubscription 创建供应商商品订阅记录控制器。
func NewSupplierProductSubscription(svc service.SupplierProductSubscriptionService) *SupplierProductSubscriptionController {
	return &SupplierProductSubscriptionController{svc: svc}
}

// List 返回供应商商品订阅记录分页列表。
func (c *SupplierProductSubscriptionController) List(ctx context.Context, req *adminapi.SupplierProductSubscriptionListReq) (res *adminapi.SupplierProductSubscriptionListRes, err error) {
	return c.svc.ListSupplierProductSubscriptions(ctx, req)
}

// Cancel 取消指定供应商商品订阅，并记录操作人与客户端 IP。
func (c *SupplierProductSubscriptionController) Cancel(ctx context.Context, req *adminapi.SupplierProductSubscriptionCancelReq) (res *adminapi.SupplierProductSubscriptionCancelRes, err error) {
	return c.svc.CancelSupplierProductSubscription(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Resubscribe 重新订阅指定供应商商品，并记录操作人与客户端 IP。
func (c *SupplierProductSubscriptionController) Resubscribe(ctx context.Context, req *adminapi.SupplierProductSubscriptionResubscribeReq) (res *adminapi.SupplierProductSubscriptionResubscribeRes, err error) {
	return c.svc.ResubscribeSupplierProductSubscription(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
