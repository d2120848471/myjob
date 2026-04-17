package api

import "github.com/gogf/gf/v2/frame/g"

// ProductGoodsChannelConfigGetReq 用于读取指定商品的渠道配置（商品级）。
type ProductGoodsChannelConfigGetReq struct {
	g.Meta  `path:"/product-goods/{goods_id}/channel-config" method:"get" tags:"商品渠道配置" summary:"商品渠道配置详情" security:"BearerAuth" dc:"读取商品级渠道配置与绑定摘要"`
	GoodsID int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsChannelConfigGetRes 返回商品渠道配置（商品级）与绑定摘要信息。
type ProductGoodsChannelConfigGetRes struct {
	GoodsID          int64  `json:"goods_id" dc:"商品ID"`
	GoodsCode        string `json:"goods_code" dc:"商品编码"`
	GoodsName        string `json:"goods_name" dc:"商品名称"`
	SubjectID        *int64 `json:"subject_id" dc:"主体ID"`
	SubjectName      string `json:"subject_name" dc:"主体名称"`
	DefaultSellPrice string `json:"default_sell_price" dc:"默认售价"`

	SmartReplenishEnabled bool    `json:"smart_replenish_enabled" dc:"是否开启智能补单"`
	AttemptTimeoutEnabled bool    `json:"attempt_timeout_enabled" dc:"是否开启 attempt 等待超时"`
	AttemptTimeoutMinutes int     `json:"attempt_timeout_minutes" dc:"attempt 等待超时分钟数"`
	RouteMode             string  `json:"route_mode" dc:"选路模式"`
	SyncCostEnabled       bool    `json:"sync_cost_enabled" dc:"是否同步进价"`
	SyncGoodsNameEnabled  bool    `json:"sync_goods_name_enabled" dc:"是否同步商品名称"`
	AllowLoss             bool    `json:"allow_loss" dc:"是否允许亏本销售"`
	MaxLossAmount         *string `json:"max_loss_amount" dc:"最大允许亏损额（为空表示不限制）"`
	IsBundle              bool    `json:"is_bundle" dc:"是否组合商品标记"`

	BoundChannelCount      int    `json:"bound_channel_count" dc:"已启用绑定数量"`
	PrimaryChannelName     string `json:"primary_channel_name" dc:"主渠道名称展示"`
	MinChannelCost         string `json:"min_channel_cost" dc:"绑定聚合后的最低比较成本价"`
	ChannelAutoPriceStatus bool   `json:"channel_auto_price_status" dc:"是否存在已启用自动改价绑定"`
}

// ProductGoodsChannelConfigUpdateReq 用于更新指定商品的渠道配置（商品级）。
type ProductGoodsChannelConfigUpdateReq struct {
	g.Meta  `path:"/product-goods/{goods_id}/channel-config" method:"patch" tags:"商品渠道配置" summary:"更新商品渠道配置" security:"BearerAuth" dc:"更新商品级渠道配置并刷新绑定摘要"`
	GoodsID int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`

	SmartReplenishEnabled bool `json:"smart_replenish_enabled" dc:"是否开启智能补单"`

	AttemptTimeoutEnabled bool `json:"attempt_timeout_enabled" dc:"是否开启 attempt 等待超时"`
	AttemptTimeoutMinutes int  `json:"attempt_timeout_minutes" dc:"attempt 等待超时分钟数"`

	RouteMode            string `json:"route_mode" dc:"选路模式"`
	SyncCostEnabled      bool   `json:"sync_cost_enabled" dc:"是否同步进价"`
	SyncGoodsNameEnabled bool   `json:"sync_goods_name_enabled" dc:"是否同步商品名称"`

	AllowLoss     bool   `json:"allow_loss" dc:"是否允许亏本销售"`
	MaxLossAmount string `json:"max_loss_amount" dc:"最大允许亏损额（为空表示不限制）"`

	IsBundle bool `json:"is_bundle" dc:"是否组合商品标记"`
}

// ProductGoodsChannelConfigUpdateRes 表示更新成功（返回体为空）。
type ProductGoodsChannelConfigUpdateRes struct{}
