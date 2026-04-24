package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// OrderController 提供后台订单记录查询能力。
type OrderController struct{ svc service.OrderService }

// NewOrder 创建后台订单控制器。
func NewOrder(svc service.OrderService) *OrderController { return &OrderController{svc: svc} }

// List 返回后台订单记录列表与统计。
func (c *OrderController) List(ctx context.Context, req *adminapi.OrderListReq) (*adminapi.OrderListRes, error) {
	return c.svc.ListAdminOrders(ctx, req)
}
