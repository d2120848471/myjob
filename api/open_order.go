package api

import "github.com/gogf/gf/v2/frame/g"

// OpenOrderCreateReq 用于外部调用方创建一笔直充渠道订单。
type OpenOrderCreateReq struct {
	g.Meta   `path:"/orders" method:"post" tags:"开放订单" summary:"开放下单" dc:"外部调用方通过 token 创建一笔待异步提交的订单"`
	Token    string `json:"token" v:"required#token不能为空" dc:"开放下单token"`
	GoodsID  string `json:"goods_id" v:"required#商品ID不能为空" dc:"对外商品ID，对应商品编码"`
	Account  string `json:"account" v:"required#充值账号不能为空" dc:"充值账号"`
	Quantity int    `json:"quantity" v:"required#数量不能为空" dc:"购买数量"`
}

// OpenOrderCreateRes 返回我方订单号与创建后的初始状态。
type OpenOrderCreateRes struct {
	OrderNo    string `json:"order_no" dc:"我方订单号"`
	StatusCode string `json:"status_code" dc:"订单状态码"`
	StatusText string `json:"status_text" dc:"订单状态文案"`
	CreatedAt  string `json:"created_at" dc:"创建时间"`
}

// OpenOrderPathQueryReq 用于通过路径订单号查询开放订单状态。
type OpenOrderPathQueryReq struct {
	g.Meta  `path:"/orders/{orderNo}" method:"get" tags:"开放订单" summary:"开放查单" dc:"通过路径订单号查询订单状态"`
	OrderNo string `json:"orderNo" in:"path" v:"required#订单号不能为空" dc:"我方订单号"`
	Token   string `json:"token" in:"query" v:"required#token不能为空" dc:"开放下单token"`
}

// OpenOrderQueryReq 用于通过 query 参数查询开放订单状态。
type OpenOrderQueryReq struct {
	g.Meta  `path:"/orders" method:"get" tags:"开放订单" summary:"开放查单Query" dc:"通过query订单号查询订单状态"`
	OrderNo string `json:"order_no" in:"query" v:"required#订单号不能为空" dc:"我方订单号"`
	Token   string `json:"token" in:"query" v:"required#token不能为空" dc:"开放下单token"`
}

// OpenOrderQueryRes 返回外部可见的订单状态信息。
type OpenOrderQueryRes struct {
	OrderNo    string `json:"order_no" dc:"我方订单号"`
	StatusCode string `json:"status_code" dc:"订单状态码"`
	StatusText string `json:"status_text" dc:"订单状态文案"`
	GoodsID    string `json:"goods_id" dc:"对外商品ID"`
	GoodsName  string `json:"goods_name" dc:"商品名称"`
	Account    string `json:"account" dc:"充值账号"`
	Quantity   int    `json:"quantity" dc:"购买数量"`
	CreatedAt  string `json:"created_at" dc:"创建时间"`
	UpdatedAt  string `json:"updated_at" dc:"更新时间"`
}
