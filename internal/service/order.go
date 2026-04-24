package service

import (
	"context"

	adminapi "myjob/api"
)

// OrderService 定义开放下单、查单、后台订单列表与异步履约能力。
type OrderService interface {
	CreateOpenOrder(ctx context.Context, req *adminapi.OpenOrderCreateReq) (*adminapi.OpenOrderCreateRes, error)
	QueryOpenOrder(ctx context.Context, token, orderNo string) (*adminapi.OpenOrderQueryRes, error)
	ListAdminOrders(ctx context.Context, req *adminapi.OrderListReq) (*adminapi.OrderListRes, error)
	TriggerSubmit(orderNo string)
	SubmitPendingOnce(ctx context.Context) error
	PollDueOnce(ctx context.Context) error
}
