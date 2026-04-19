package api

import "github.com/gogf/gf/v2/frame/g"

// ProductGoodsInventoryConfigGetReq 用于读取商品库存配置。
type ProductGoodsInventoryConfigGetReq struct {
	g.Meta  `path:"/products/{goodsId}/inventory-config" method:"get" tags:"商品管理" summary:"商品库存配置详情" security:"BearerAuth" dc:"读取指定商品的库存配置与策略选项"`
	GoodsId int64 `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsInventoryConfigGetRes 返回商品摘要、库存配置与策略选项。
type ProductGoodsInventoryConfigGetRes struct {
	Goods                ProductGoodsChannelGoodsSummary `json:"goods" dc:"商品摘要"`
	Config               ProductGoodsInventoryConfig     `json:"config" dc:"库存配置"`
	OrderStrategyOptions []ProductGoodsStringOption      `json:"order_strategy_options" dc:"下单策略选项"`
}

// ProductGoodsInventoryConfigSaveReq 用于保存商品库存配置。
type ProductGoodsInventoryConfigSaveReq struct {
	g.Meta                `path:"/products/{goodsId}/inventory-config" method:"put" tags:"商品管理" summary:"保存商品库存配置" security:"BearerAuth" dc:"保存指定商品的库存配置与智能补单规则"`
	GoodsId               int64  `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	SmartReorderEnabled   int    `json:"smart_reorder_enabled" dc:"智能补单开关"`
	ReorderTimeoutEnabled int    `json:"reorder_timeout_enabled" dc:"补单超时开关"`
	ReorderTimeoutMinutes int    `json:"reorder_timeout_minutes" dc:"补单超时分钟数"`
	OrderStrategy         string `json:"order_strategy" dc:"下单策略"`
	SyncCostPriceEnabled  int    `json:"sync_cost_price_enabled" dc:"同步进价开关"`
	SyncGoodsNameEnabled  int    `json:"sync_goods_name_enabled" dc:"同步商品名称开关"`
	AllowLossSaleEnabled  int    `json:"allow_loss_sale_enabled" dc:"亏本销售开关"`
	MaxLossAmount         string `json:"max_loss_amount" dc:"允许亏本金额"`
	ComboGoodsEnabled     int    `json:"combo_goods_enabled" dc:"组合商品开关"`
}

// ProductGoodsInventoryConfigSaveRes 表示库存配置保存成功。
type ProductGoodsInventoryConfigSaveRes struct{}

// ProductGoodsInventoryConfig 表示商品维度的库存配置完整回显。
type ProductGoodsInventoryConfig struct {
	SmartReorderEnabled   int    `json:"smart_reorder_enabled" dc:"智能补单开关"`
	ReorderTimeoutEnabled int    `json:"reorder_timeout_enabled" dc:"补单超时开关"`
	ReorderTimeoutMinutes int    `json:"reorder_timeout_minutes" dc:"补单超时分钟数"`
	OrderStrategy         string `json:"order_strategy" dc:"下单策略"`
	SyncCostPriceEnabled  int    `json:"sync_cost_price_enabled" dc:"同步进价开关"`
	SyncGoodsNameEnabled  int    `json:"sync_goods_name_enabled" dc:"同步商品名称开关"`
	AllowLossSaleEnabled  int    `json:"allow_loss_sale_enabled" dc:"亏本销售开关"`
	MaxLossAmount         string `json:"max_loss_amount" dc:"允许亏本金额"`
	ComboGoodsEnabled     int    `json:"combo_goods_enabled" dc:"组合商品开关"`
}

// ProductGoodsInventoryConfigSummary 表示渠道弹窗顶部的库存配置摘要。
type ProductGoodsInventoryConfigSummary struct {
	SmartReorderEnabled   int    `json:"smart_reorder_enabled" dc:"智能补单开关"`
	ReorderTimeoutEnabled int    `json:"reorder_timeout_enabled" dc:"补单超时开关"`
	ReorderTimeoutMinutes int    `json:"reorder_timeout_minutes" dc:"补单超时分钟数"`
	OrderStrategy         string `json:"order_strategy" dc:"下单策略"`
	SyncCostPriceEnabled  int    `json:"sync_cost_price_enabled" dc:"同步进价开关"`
	SyncGoodsNameEnabled  int    `json:"sync_goods_name_enabled" dc:"同步商品名称开关"`
	AllowLossSaleEnabled  int    `json:"allow_loss_sale_enabled" dc:"亏本销售开关"`
	MaxLossAmount         string `json:"max_loss_amount" dc:"允许亏本金额"`
	ComboGoodsEnabled     int    `json:"combo_goods_enabled" dc:"组合商品开关"`
}
