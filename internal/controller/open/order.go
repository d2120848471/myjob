package opencontroller

import (
	"context"

	openapi "myjob/api"
	openauth "myjob/internal/library/openauth"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// OpenOrderController 提供开放接口下单与查单 HTTP handler。
type OpenOrderController struct {
	svc service.OpenOrderService
}

// NewOpenOrder 创建 OpenOrderController。
func NewOpenOrder(svc service.OpenOrderService) *OpenOrderController {
	return &OpenOrderController{svc: svc}
}

// Create 创建订单（幂等），并返回对外三态视图。
func (c *OpenOrderController) Create(ctx context.Context, req *openapi.OpenOrderCreateReq) (res *openapi.OpenOrderCreateRes, err error) {
	caller := openauth.MustCallerFromCtx(ctx)
	ip := ""
	if request := g.RequestFromCtx(ctx); request != nil {
		ip = request.GetClientIp()
	}
	return c.svc.Create(ctx, req, caller, ip)
}

// Get 按内部订单号查询订单详情。
func (c *OpenOrderController) Get(ctx context.Context, req *openapi.OpenOrderGetReq) (res *openapi.OpenOrderGetRes, err error) {
	return c.svc.Get(ctx, req, openauth.MustCallerFromCtx(ctx))
}

// GetByClient 按调用方订单号查询订单详情。
func (c *OpenOrderController) GetByClient(ctx context.Context, req *openapi.OpenOrderGetByClientReq) (res *openapi.OpenOrderGetRes, err error) {
	return c.svc.GetByClient(ctx, req, openauth.MustCallerFromCtx(ctx))
}

