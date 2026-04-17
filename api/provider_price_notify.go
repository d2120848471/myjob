package api

import "github.com/gogf/gf/v2/frame/g"

// ProviderPriceNotifyReq 用于接收上游价格通知（原样透传 ACK，不做统一 JSON 包装）。
type ProviderPriceNotifyReq struct {
	g.Meta       `path:"/{provider_code}/price-notify" method:"post" tags:"上游回调" summary:"价格通知" dc:"接收上游价格通知（原样透传 ACK，不做统一 JSON 包装）"`
	ProviderCode string `json:"provider_code" in:"path" v:"required#provider_code不能为空" dc:"Provider 编码"`
}

// ProviderPriceNotifyRes 表示价格通知处理完成（返回体由 Provider 约定决定）。
//
// 注意：本接口的真实响应体可能是纯文本或平台约定格式，因此 controller 会直接写入响应并跳过统一包装。
type ProviderPriceNotifyRes struct{}

