package api

import "github.com/gogf/gf/v2/frame/g"

// SupplierProductSubscriptionListReq 用于分页查询供应商商品推送订阅记录。
type SupplierProductSubscriptionListReq struct {
	g.Meta          `path:"/supplier-product-subscriptions" method:"get" tags:"第三方对接" summary:"商品订阅记录" security:"BearerAuth" dc:"分页查询供应商商品推送订阅记录"`
	Page            int    `json:"page" dc:"页码"`
	PageSize        int    `json:"page_size" dc:"每页条数"`
	Keyword         string `json:"keyword" dc:"商品名称关键词"`
	SupplierGoodsNo string `json:"supplier_goods_no" dc:"上游商品编号"`
	PlatformID      int64  `json:"platform_id" dc:"平台账号ID"`
	Status          string `json:"status" dc:"订阅状态"`
	StartAt         string `json:"start_at" dc:"开始时间"`
	EndAt           string `json:"end_at" dc:"结束时间"`
}

// SupplierProductSubscriptionListRes 返回供应商商品推送订阅记录列表。
type SupplierProductSubscriptionListRes struct {
	List       []SupplierProductSubscriptionItem `json:"list" dc:"订阅记录列表"`
	Pagination PaginationRes                     `json:"pagination" dc:"分页信息"`
}

// SupplierProductSubscriptionItem 是订阅记录列表单行数据。
type SupplierProductSubscriptionItem struct {
	ID                  int64  `json:"id" dc:"订阅记录ID"`
	GoodsID             int64  `json:"goods_id" dc:"本地商品ID"`
	GoodsName           string `json:"goods_name" dc:"商品名称"`
	GoodsIcon           string `json:"goods_icon" dc:"商品图标"`
	ProviderCode        string `json:"provider_code" dc:"供应商编码"`
	PlatformAccountID   int64  `json:"platform_account_id" dc:"平台账号ID"`
	PlatformAccountName string `json:"platform_account_name" dc:"平台账号名称"`
	SupplierGoodsNo     string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName   string `json:"supplier_goods_name" dc:"上游商品名称"`
	CallbackURL         string `json:"callback_url" dc:"回调地址"`
	Status              string `json:"status" dc:"订阅状态"`
	LastAction          string `json:"last_action" dc:"最近动作"`
	LastError           string `json:"last_error" dc:"最近失败原因"`
	SubscribedAt        string `json:"subscribed_at" dc:"订阅时间"`
	CanceledAt          string `json:"canceled_at" dc:"取消时间"`
	UpdatedAt           string `json:"updated_at" dc:"更新时间"`
}

// SupplierProductSubscriptionCancelReq 用于取消单条供应商商品订阅。
type SupplierProductSubscriptionCancelReq struct {
	g.Meta `path:"/supplier-product-subscriptions/{id}/cancel" method:"post" tags:"第三方对接" summary:"取消商品订阅" security:"BearerAuth" dc:"取消指定供应商商品推送订阅"`
	ID     int64 `json:"id" in:"path" v:"required#订阅记录ID不能为空" dc:"订阅记录ID"`
}

// SupplierProductSubscriptionCancelRes 表示取消订阅成功。
type SupplierProductSubscriptionCancelRes struct{}

// SupplierProductSubscriptionResubscribeReq 用于重新订阅单条供应商商品。
type SupplierProductSubscriptionResubscribeReq struct {
	g.Meta `path:"/supplier-product-subscriptions/{id}/resubscribe" method:"post" tags:"第三方对接" summary:"重新订阅商品" security:"BearerAuth" dc:"重新订阅指定供应商商品推送"`
	ID     int64 `json:"id" in:"path" v:"required#订阅记录ID不能为空" dc:"订阅记录ID"`
}

// SupplierProductSubscriptionResubscribeRes 表示重新订阅成功。
type SupplierProductSubscriptionResubscribeRes struct{}
