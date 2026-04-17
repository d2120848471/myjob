package api

import "github.com/gogf/gf/v2/frame/g"

// ProviderOrderCallbackReq 用于接收上游订单回调（原样透传 ACK，不做统一 JSON 包装）。
type ProviderOrderCallbackReq struct {
	g.Meta       `path:"/{provider_code}/order-callback" method:"post" tags:"上游回调" summary:"订单回调" dc:"接收上游订单回调（原样透传 ACK，不做统一 JSON 包装）"`
	ProviderCode string `json:"provider_code" in:"path" v:"required#provider_code不能为空" dc:"Provider 编码"`
}

// ProviderOrderCallbackRes 表示回调处理完成（返回体由 Provider 约定决定）。
//
// 注意：本接口的真实响应体可能是纯文本或平台约定格式，因此 controller 会直接写入响应并跳过统一包装。
type ProviderOrderCallbackRes struct{}

