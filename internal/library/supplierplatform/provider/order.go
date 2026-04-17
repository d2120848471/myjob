package supplierprovider

import (
	"context"
	"net/http"
)

// CreateOrderInput 是交易层传递给 Provider 的建单输入。
type CreateOrderInput struct {
	ProviderRequestOrderNo string         // 本地生成的 provider_request_order_no（用于回调/查单定位 attempt）
	SupplierGoodsNo        string         // 上游商品编号
	Quantity               int            // 本次 attempt 的数量
	Payload                map[string]any // 下游 payload（已解码）
}

// CreateOrderResult 是 Provider 解析建单响应后的统一结果。
type CreateOrderResult struct {
	Accepted     bool
	FinalSuccess bool
	FinalFailed  bool
	Uncertain    bool

	ChannelOrderNo  string
	UpstreamStatus  string
	ErrorCategory   string
	ErrorCode       string
	ErrorMessage    string
	RawPayload      string
	ProviderTraceID string
}

// QueryOrderInput 是交易层传递给 Provider 的查单输入。
type QueryOrderInput struct {
	ProviderRequestOrderNo string
	ChannelOrderNo         string
}

// QueryOrderResult 是 Provider 解析查单响应后的统一结果。
type QueryOrderResult struct {
	Processing   bool
	FinalSuccess bool
	FinalFailed  bool

	ChannelOrderNo string
	UpstreamStatus string

	ErrorCategory string
	ErrorCode     string
	ErrorMessage  string
	RawPayload    string
}

// OrderProvider 定义可真实下单的 Provider 能力。
type OrderProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	SupportsNativeQuantity() bool
	BuildCreateOrderRequest(ctx context.Context, account AccountConfig, input CreateOrderInput, baseURL string) (*http.Request, error)
	ParseCreateOrderResponse(statusCode int, body []byte) (*CreateOrderResult, error)
	BuildQueryOrderRequest(ctx context.Context, account AccountConfig, input QueryOrderInput, baseURL string) (*http.Request, error)
	ParseQueryOrderResponse(statusCode int, body []byte) (*QueryOrderResult, error)
}
