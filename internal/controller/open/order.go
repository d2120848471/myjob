package opencontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// OrderController 提供开放订单接口，不依赖后台登录态。
type OrderController struct{ svc service.OrderService }

// NewOrder 创建开放订单控制器。
func NewOrder(svc service.OrderService) *OrderController { return &OrderController{svc: svc} }

// Create 创建一笔外部订单并触发异步提交。
func (c *OrderController) Create(ctx context.Context, req *adminapi.OpenOrderCreateReq) (*adminapi.OpenOrderCreateRes, error) {
	return c.svc.CreateOpenOrder(ctx, req)
}

// QueryByPath 通过路径订单号查询订单状态。
func (c *OrderController) QueryByPath(ctx context.Context, req *adminapi.OpenOrderPathQueryReq) (*adminapi.OpenOrderQueryRes, error) {
	return c.svc.QueryOpenOrder(ctx, req.Token, req.OrderNo)
}

// Query 通过 query 订单号查询订单状态。
func (c *OrderController) Query(ctx context.Context, req *adminapi.OpenOrderQueryReq) (*adminapi.OpenOrderQueryRes, error) {
	return c.svc.QueryOpenOrder(ctx, req.Token, req.OrderNo)
}
