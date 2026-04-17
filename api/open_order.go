package api

import "github.com/gogf/gf/v2/frame/g"

// OpenOrderCreateReq 用于开放接口创建交易订单（Header 签名鉴权）。
type OpenOrderCreateReq struct {
	g.Meta        `path:"/orders" method:"post" tags:"开放下单" summary:"创建订单" dc:"开放接口创建交易订单（Header 签名鉴权）"`
	ClientOrderNo string         `json:"client_order_no" v:"required#client_order_no不能为空" dc:"调用方订单号（幂等）"`
	GoodsCode     string         `json:"goods_code" v:"required#goods_code不能为空" dc:"商品编码"`
	Quantity      int            `json:"quantity" v:"required|min:1#quantity不能为空|quantity错误" dc:"购买数量"`
	Payload       map[string]any `json:"payload" dc:"业务参数payload"`
}

// OpenOrderCreateRes 返回创建订单结果（对外仅三态：processing/success/failed）。
type OpenOrderCreateRes struct {
	OrderNo       string `json:"order_no" dc:"内部订单号"`
	ClientOrderNo string `json:"client_order_no" dc:"调用方订单号"`
	Status        string `json:"status" dc:"订单状态 processing/success/failed"`
	GoodsCode     string `json:"goods_code" dc:"商品编码"`
	GoodsName     string `json:"goods_name" dc:"商品名称"`
	Quantity      int    `json:"quantity" dc:"购买数量"`
	SalePrice     string `json:"sale_price" dc:"单价（保留4位小数）"`
	TotalAmount   string `json:"total_amount" dc:"总金额（保留4位小数）"`
	CreatedAt     string `json:"created_at" dc:"创建时间 YYYY-MM-DD HH:MM:SS"`
}

// OpenOrderGetReq 用于按内部订单号查询订单。
type OpenOrderGetReq struct {
	g.Meta  `path:"/orders/{order_no}" method:"get" tags:"开放下单" summary:"按内部订单号查单" dc:"按内部订单号查询订单详情"`
	OrderNo string `json:"order_no" in:"path" v:"required#order_no不能为空" dc:"内部订单号"`
}

// OpenOrderGetByClientReq 用于按调用方订单号查询订单。
type OpenOrderGetByClientReq struct {
	g.Meta        `path:"/orders/by-client/{client_order_no}" method:"get" tags:"开放下单" summary:"按调用方订单号查单" dc:"按调用方订单号查询订单详情"`
	ClientOrderNo string `json:"client_order_no" in:"path" v:"required#client_order_no不能为空" dc:"调用方订单号"`
}

// OpenUpstreamOrderItem 表示对外暴露的上游 attempt 视图。
type OpenUpstreamOrderItem struct {
	FulfillmentNo           string `json:"fulfillment_no" dc:"履约分片编号"`
	AttemptNo               int    `json:"attempt_no" dc:"attempt编号（从1开始）"`
	AttemptQuantity         int    `json:"attempt_quantity" dc:"attempt数量"`
	Status                  string `json:"status" dc:"attempt状态 processing/success/failed"`
	BindingChannelName      string `json:"binding_channel_name" dc:"命中绑定渠道名称快照"`
	BindingSupplierGoodsNo  string `json:"binding_supplier_goods_no" dc:"命中绑定上游商品编号快照"`
	ChannelOrderNo          string `json:"channel_order_no" dc:"上游订单号"`
	ProviderRequestOrderNo  string `json:"provider_request_order_no" dc:"请求上游的外部订单号（用于回调定位）"`
	UpstreamStatus          string `json:"upstream_status" dc:"上游原始状态码/文案快照"`
	ErrorCategory           string `json:"error_category" dc:"错误分类"`
	ErrorMessage            string `json:"error_message" dc:"错误摘要"`
	FinishedAt              string `json:"finished_at" dc:"完成时间（若已最终态）"`
}

// OpenOrderGetRes 返回订单详情。
type OpenOrderGetRes struct {
	OrderNo        string `json:"order_no" dc:"内部订单号"`
	ClientOrderNo  string `json:"client_order_no" dc:"调用方订单号"`
	Status         string `json:"status" dc:"订单状态 processing/success/failed"`
	GoodsCode      string `json:"goods_code" dc:"商品编码"`
	GoodsName      string `json:"goods_name" dc:"商品名称"`
	Quantity       int    `json:"quantity" dc:"购买数量"`
	SuccessQuantity int   `json:"success_quantity" dc:"成功数量"`
	FailedQuantity  int   `json:"failed_quantity" dc:"失败数量"`
	SalePrice      string `json:"sale_price" dc:"单价（保留4位小数）"`
	TotalAmount    string `json:"total_amount" dc:"总金额（保留4位小数）"`
	FailureReason  string `json:"failure_reason" dc:"失败原因摘要"`
	CreatedAt      string `json:"created_at" dc:"创建时间 YYYY-MM-DD HH:MM:SS"`
	FinishedAt     string `json:"finished_at" dc:"完成时间 YYYY-MM-DD HH:MM:SS"`
	UpstreamOrders []OpenUpstreamOrderItem `json:"upstream_orders" dc:"上游订单attempt视图"`
}

