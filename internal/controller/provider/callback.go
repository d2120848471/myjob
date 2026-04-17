package providercontroller

import (
	"context"
	"net/http"
	"strings"

	providerapi "myjob/api"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// ProviderCallbackController 提供上游订单回调入口（不做统一 JSON 包装）。
type ProviderCallbackController struct {
	svc service.TradeCallbackService
}

// NewProviderCallback 创建 ProviderCallbackController。
func NewProviderCallback(svc service.TradeCallbackService) *ProviderCallbackController {
	return &ProviderCallbackController{svc: svc}
}

// OrderCallback 接收上游回调并原样返回 ACK。
func (c *ProviderCallbackController) OrderCallback(ctx context.Context, req *providerapi.ProviderOrderCallbackReq) (res *providerapi.ProviderOrderCallbackRes, err error) {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return nil, nil
	}
	ackBody, contentType, callbackErr := c.svc.HandleProviderOrderCallback(ctx, strings.TrimSpace(req.ProviderCode), cloneHeader(request.Header), request.GetBody())
	if strings.TrimSpace(contentType) != "" {
		request.Response.Header().Set("Content-Type", strings.TrimSpace(contentType))
	}
	if len(ackBody) > 0 {
		request.Response.Write(ackBody)
	}
	// 上游回调通常需要尽快收到 ACK，业务错误也不强制暴露给上游。
	_ = callbackErr
	return nil, nil
}

func cloneHeader(header http.Header) http.Header {
	if header == nil {
		return http.Header{}
	}
	duplicated := make(http.Header, len(header))
	for key, values := range header {
		items := make([]string, len(values))
		copy(items, values)
		duplicated[key] = items
	}
	return duplicated
}

