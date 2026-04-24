package api

import "github.com/gogf/gf/v2/frame/g"

// OrderListReq 用于后台查询外部订单记录列表。
type OrderListReq struct {
	g.Meta     `path:"/orders" method:"get" tags:"订单记录" summary:"订单记录列表" security:"BearerAuth" dc:"后台分页查询外部订单记录与基础统计"`
	Page       int    `json:"page" dc:"页码"`
	PageSize   int    `json:"page_size" dc:"每页条数"`
	Keyword    string `json:"keyword" dc:"关键词"`
	KeywordBy  string `json:"keyword_by" dc:"关键词类型：order_no/account/goods_name"`
	Status     string `json:"status" dc:"订单状态"`
	HasTax     string `json:"has_tax" dc:"商品含税筛选"`
	ChannelID  int64  `json:"channel_id" dc:"当前渠道平台账号ID"`
	IsCard     string `json:"is_card" dc:"是否卡密"`
	StartTime  string `json:"start_time" dc:"创建开始时间"`
	EndTime    string `json:"end_time" dc:"创建结束时间"`
	QuickRange string `json:"quick_range" dc:"快捷时间：yesterday/today/week/month/three_months；month 为本月，three_months 为本月及前两个月自然月"`
}

// OrderListRes 返回后台订单列表、分页和统计。
type OrderListRes struct {
	List       []OrderListItem `json:"list" dc:"订单列表"`
	Pagination PaginationRes   `json:"pagination" dc:"分页信息"`
	Stats      OrderStats      `json:"stats" dc:"订单统计"`
}

// OrderStats 表示订单列表顶部基础统计。
type OrderStats struct {
	TodayOrderCount      int    `json:"today_order_count" dc:"今日订单数"`
	TodayOrderAmount     string `json:"today_order_amount" dc:"今日交易额"`
	YesterdayOrderCount  int    `json:"yesterday_order_count" dc:"昨日订单数"`
	YesterdayOrderAmount string `json:"yesterday_order_amount" dc:"昨日交易额"`
}

// OrderListItem 是后台订单列表单行数据。
type OrderListItem struct {
	ID                 int64  `json:"id" dc:"订单ID"`
	SalesSubjectName   string `json:"sales_subject_name" dc:"销售主体"`
	OrderNo            string `json:"order_no" dc:"订单号"`
	GoodsID            string `json:"goods_id" dc:"对外商品ID"`
	GoodsName          string `json:"goods_name" dc:"商品名称"`
	Account            string `json:"account" dc:"充值账号"`
	Quantity           int    `json:"quantity" dc:"购买数量"`
	OrderAmount        string `json:"order_amount" dc:"订单金额"`
	CostAmount         string `json:"cost_amount" dc:"成本金额"`
	ProfitAmount       string `json:"profit_amount" dc:"利润金额"`
	CurrentChannelID   int64  `json:"current_channel_id" dc:"当前渠道平台账号ID"`
	CurrentChannelName string `json:"current_channel_name" dc:"当前渠道名称"`
	SupplierOrderNo    string `json:"supplier_order_no" dc:"上游订单号"`
	AttemptCount       int    `json:"attempt_count" dc:"尝试次数"`
	LastReceipt        string `json:"last_receipt" dc:"失败或回执摘要"`
	StatusCode         string `json:"status_code" dc:"状态码"`
	StatusText         string `json:"status_text" dc:"状态文案"`
	CreatedAt          string `json:"created_at" dc:"创建时间"`
	UpdatedAt          string `json:"updated_at" dc:"更新时间"`
}
