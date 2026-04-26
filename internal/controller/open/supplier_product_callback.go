package opencontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// SupplierProductCallbackController 提供供应商商品信息变动开放回调。
type SupplierProductCallbackController struct {
	svc service.SupplierProductCallbackService
}

// NewSupplierProductCallback 创建供应商商品变动回调控制器。
func NewSupplierProductCallback(svc service.SupplierProductCallbackService) *SupplierProductCallbackController {
	return &SupplierProductCallbackController{svc: svc}
}

// ProductChange 接收供应商商品变动推送，成功时按上游要求返回纯文本 ok。
func (c *SupplierProductCallbackController) ProductChange(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq) (*adminapi.SupplierProductChangeCallbackRes, error) {
	request := g.RequestFromCtx(ctx)
	body := request.GetBody()
	if err := c.svc.HandleSupplierProductChangeCallback(ctx, req, body); err != nil {
		return nil, err
	}
	request.Response.Write("ok")
	return &adminapi.SupplierProductChangeCallbackRes{}, nil
}
