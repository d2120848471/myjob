package api

import "github.com/gogf/gf/v2/frame/g"

// SupplierProductChangeCallbackReq 用于接收第三方平台商品信息变动推送。
type SupplierProductChangeCallbackReq struct {
	g.Meta            `path:"/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback" method:"post" tags:"开放回调" summary:"供应商商品变动回调" dc:"第三方平台商品价格等信息变动后的通用回调入口"`
	ProviderCode      string `json:"providerCode" in:"path" v:"required#平台编码不能为空" dc:"供应商适配器编码"`
	PlatformAccountID int64  `json:"platformAccountId" in:"path" v:"required#平台账号ID不能为空" dc:"平台账号ID"`
}

// SupplierProductChangeCallbackRes 表示供应商商品变动回调已处理。
type SupplierProductChangeCallbackRes struct{}
