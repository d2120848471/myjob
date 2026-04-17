package providercontroller

import (
	"context"
	"strings"

	providerapi "myjob/api"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// ProviderPriceNotifyController 提供上游价格通知入口（不做统一 JSON 包装）。
type ProviderPriceNotifyController struct {
	svc service.TradePriceNotifyService
}

// NewProviderPriceNotify 创建 ProviderPriceNotifyController。
func NewProviderPriceNotify(svc service.TradePriceNotifyService) *ProviderPriceNotifyController {
	return &ProviderPriceNotifyController{svc: svc}
}

// PriceNotify 接收上游价格通知并原样返回 ACK。
func (c *ProviderPriceNotifyController) PriceNotify(ctx context.Context, req *providerapi.ProviderPriceNotifyReq) (res *providerapi.ProviderPriceNotifyRes, err error) {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return nil, nil
	}
	ackBody, contentType, notifyErr := c.svc.HandleProviderPriceNotify(ctx, strings.TrimSpace(req.ProviderCode), cloneHeader(request.Header), request.GetBody())
	if strings.TrimSpace(contentType) != "" {
		request.Response.Header().Set("Content-Type", strings.TrimSpace(contentType))
	}
	if len(ackBody) > 0 {
		request.Response.Write(ackBody)
	}
	_ = notifyErr
	return nil, nil
}

