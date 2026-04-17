package api

import "github.com/gogf/gf/v2/frame/g"

// ProductGoodsChannelBindingListReq 用于查询指定商品的渠道绑定列表。
type ProductGoodsChannelBindingListReq struct {
	g.Meta  `path:"/product-goods/{goods_id}/channel-bindings" method:"get" tags:"商品渠道绑定" summary:"绑定列表" security:"BearerAuth" dc:"查询指定商品的渠道绑定列表"`
	GoodsID int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsChannelBindingListRes 返回指定商品的绑定列表。
type ProductGoodsChannelBindingListRes struct {
	List []ProductGoodsChannelBindingItem `json:"list" dc:"绑定列表"`
}

// ProductGoodsChannelBindingItem 表示商品渠道绑定列表项。
type ProductGoodsChannelBindingItem struct {
	ID          int64  `json:"id" dc:"绑定ID"`
	DisplayName string `json:"display_name" dc:"展示名称"`
	DockStatus  string `json:"dock_status" dc:"启停状态 enabled/disabled"`

	PlatformAccountID   int64  `json:"platform_account_id" dc:"渠道账号ID"`
	PlatformAccountName string `json:"platform_account_name" dc:"渠道账号名称"`
	ProviderCode        string `json:"provider_code" dc:"Provider 编码"`
	ProviderName        string `json:"provider_name" dc:"Provider 名称"`

	SupplierGoodsNo   string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName string `json:"supplier_goods_name" dc:"上游商品名称快照"`

	SourceCostPrice    string `json:"source_cost_price" dc:"原始进货价"`
	CostPrice          string `json:"cost_price" dc:"比较成本价（税态换算后）"`
	TaxAdjustDirection string `json:"tax_adjust_direction" dc:"税态换算方向 none/untaxed_to_taxed/taxed_to_untaxed"`
	TaxAdjustRate      string `json:"tax_adjust_rate" dc:"税点（百分比）"`
	TaxAdjustAmount    string `json:"tax_adjust_amount" dc:"税额"`

	Sort      int    `json:"sort" dc:"排序值"`
	Weight    int    `json:"weight" dc:"权重"`
	StartTime string `json:"start_time" dc:"时段开始 HH:MM"`
	EndTime   string `json:"end_time" dc:"时段结束 HH:MM"`

	ValidateTemplateID   *int64 `json:"validate_template_id" dc:"模板ID"`
	ValidateTemplateName string `json:"validate_template_name" dc:"模板名称"`

	IsAutoChange int    `json:"is_auto_change" dc:"是否开启自动改价"`
	AddType      string `json:"add_type" dc:"利润类型 fixed/percent"`
	DefaultPrice string `json:"default_price" dc:"利润值"`
	LockPrice    string `json:"lock_price" dc:"锁价（一期仅存储回显）"`
	SymbolPrice  string `json:"symbol_price" dc:"符号价（一期仅存储回显）"`
	MaxPrice     string `json:"max_price" dc:"最高价（一期仅存储回显）"`
	MinPrice     string `json:"min_price" dc:"最低价（一期仅存储回显）"`
}

// ProductGoodsChannelBindingCreateReq 用于新增渠道绑定。
type ProductGoodsChannelBindingCreateReq struct {
	g.Meta  `path:"/product-goods/{goods_id}/channel-bindings" method:"post" tags:"商品渠道绑定" summary:"新增绑定" security:"BearerAuth" dc:"新增商品渠道绑定并计算比较成本价"`
	GoodsID int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`

	PlatformAccountID  int64  `json:"platform_account_id" dc:"渠道账号ID"`
	SupplierGoodsNo    string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName  string `json:"supplier_goods_name" dc:"上游商品名称"`
	SourceCostPrice    string `json:"source_cost_price" dc:"原始进货价"`
	DockStatus         string `json:"dock_status" dc:"启停状态 enabled/disabled"`
	Sort               int    `json:"sort" dc:"排序值"`
	Weight             int    `json:"weight" dc:"权重"`
	StartTime          string `json:"start_time" dc:"时段开始 HH:MM"`
	EndTime            string `json:"end_time" dc:"时段结束 HH:MM"`
	ValidateTemplateID *int64 `json:"validate_template_id" dc:"模板ID"`
}

// ProductGoodsChannelBindingCreateRes 返回新增绑定的 ID。
type ProductGoodsChannelBindingCreateRes struct {
	ID int64 `json:"id" dc:"绑定ID"`
}

// ProductGoodsChannelBindingUpdateReq 用于更新绑定基础字段。
type ProductGoodsChannelBindingUpdateReq struct {
	g.Meta    `path:"/product-goods/{goods_id}/channel-bindings/{binding_id}" method:"patch" tags:"商品渠道绑定" summary:"更新绑定" security:"BearerAuth" dc:"更新绑定基础字段并重新计算比较成本价"`
	GoodsID   int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingID int64 `json:"binding_id" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`

	PlatformAccountID  int64  `json:"platform_account_id" dc:"渠道账号ID"`
	SupplierGoodsNo    string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName  string `json:"supplier_goods_name" dc:"上游商品名称"`
	SourceCostPrice    string `json:"source_cost_price" dc:"原始进货价"`
	DockStatus         string `json:"dock_status" dc:"启停状态 enabled/disabled"`
	Sort               int    `json:"sort" dc:"排序值"`
	Weight             int    `json:"weight" dc:"权重"`
	StartTime          string `json:"start_time" dc:"时段开始 HH:MM"`
	EndTime            string `json:"end_time" dc:"时段结束 HH:MM"`
	ValidateTemplateID *int64 `json:"validate_template_id" dc:"模板ID"`
}

// ProductGoodsChannelBindingUpdateRes 表示更新成功（返回体为空）。
type ProductGoodsChannelBindingUpdateRes struct{}

// ProductGoodsChannelBindingDeleteReq 用于删除绑定（软删除）。
type ProductGoodsChannelBindingDeleteReq struct {
	g.Meta    `path:"/product-goods/{goods_id}/channel-bindings/{binding_id}" method:"delete" tags:"商品渠道绑定" summary:"删除绑定" security:"BearerAuth" dc:"软删除指定绑定"`
	GoodsID   int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingID int64 `json:"binding_id" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`
}

// ProductGoodsChannelBindingDeleteRes 表示删除成功（返回体为空）。
type ProductGoodsChannelBindingDeleteRes struct{}

// ProductGoodsChannelBindingBatchStatusReq 用于批量启停绑定。
type ProductGoodsChannelBindingBatchStatusReq struct {
	g.Meta     `path:"/product-goods/{goods_id}/channel-bindings:batch-status" method:"post" tags:"商品渠道绑定" summary:"批量启停" security:"BearerAuth" dc:"批量更新绑定启停状态"`
	GoodsID    int64   `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingIDs []int64 `json:"binding_ids" dc:"绑定ID列表"`
	DockStatus string  `json:"dock_status" dc:"启停状态 enabled/disabled"`
}

// ProductGoodsChannelBindingBatchStatusRes 表示批量启停成功（返回体为空）。
type ProductGoodsChannelBindingBatchStatusRes struct{}

// ProductGoodsChannelBindingBatchDeleteReq 用于批量删除绑定（软删除）。
type ProductGoodsChannelBindingBatchDeleteReq struct {
	g.Meta     `path:"/product-goods/{goods_id}/channel-bindings:batch-delete" method:"post" tags:"商品渠道绑定" summary:"批量删除" security:"BearerAuth" dc:"批量软删除绑定"`
	GoodsID    int64   `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingIDs []int64 `json:"binding_ids" dc:"绑定ID列表"`
}

// ProductGoodsChannelBindingBatchDeleteRes 表示批量删除成功（返回体为空）。
type ProductGoodsChannelBindingBatchDeleteRes struct{}

// ProductGoodsChannelBindingReorderReq 用于一键排序绑定。
type ProductGoodsChannelBindingReorderReq struct {
	g.Meta  `path:"/product-goods/{goods_id}/channel-bindings:reorder" method:"post" tags:"商品渠道绑定" summary:"一键排序" security:"BearerAuth" dc:"按成本与排序规则重写 sort 为 10/20/30..."`
	GoodsID int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsChannelBindingReorderRes 表示排序成功（返回体为空）。
type ProductGoodsChannelBindingReorderRes struct{}

// ProductGoodsChannelBindingAutoPriceUpdateReq 用于更新单条绑定的自动改价字段。
type ProductGoodsChannelBindingAutoPriceUpdateReq struct {
	g.Meta    `path:"/product-goods/{goods_id}/channel-bindings/{binding_id}/auto-price" method:"patch" tags:"商品渠道绑定" summary:"更新自动改价" security:"BearerAuth" dc:"仅更新自动改价字段，不修改绑定基础字段"`
	GoodsID   int64 `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingID int64 `json:"binding_id" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`

	IsAutoChange int    `json:"is_auto_change" dc:"是否开启自动改价"`
	AddType      string `json:"add_type" dc:"利润类型 fixed/percent"`
	DefaultPrice string `json:"default_price" dc:"利润值"`
	LockPrice    string `json:"lock_price" dc:"锁价（一期仅存储回显）"`
	SymbolPrice  string `json:"symbol_price" dc:"符号价（一期仅存储回显）"`
	MaxPrice     string `json:"max_price" dc:"最高价（一期仅存储回显）"`
	MinPrice     string `json:"min_price" dc:"最低价（一期仅存储回显）"`
}

// ProductGoodsChannelBindingAutoPriceUpdateRes 表示更新成功（返回体为空）。
type ProductGoodsChannelBindingAutoPriceUpdateRes struct{}

// ProductGoodsChannelBindingAutoPriceBatchReq 用于批量更新自动改价字段。
type ProductGoodsChannelBindingAutoPriceBatchReq struct {
	g.Meta     `path:"/product-goods/{goods_id}/channel-bindings:auto-price-batch" method:"post" tags:"商品渠道绑定" summary:"批量更新自动改价" security:"BearerAuth" dc:"批量更新自动改价字段"`
	GoodsID    int64   `json:"goods_id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingIDs []int64 `json:"binding_ids" dc:"绑定ID列表"`

	IsAutoChange int    `json:"is_auto_change" dc:"是否开启自动改价"`
	AddType      string `json:"add_type" dc:"利润类型 fixed/percent"`
	DefaultPrice string `json:"default_price" dc:"利润值"`
	LockPrice    string `json:"lock_price" dc:"锁价（一期仅存储回显）"`
	SymbolPrice  string `json:"symbol_price" dc:"符号价（一期仅存储回显）"`
	MaxPrice     string `json:"max_price" dc:"最高价（一期仅存储回显）"`
	MinPrice     string `json:"min_price" dc:"最低价（一期仅存储回显）"`
}

// ProductGoodsChannelBindingAutoPriceBatchRes 表示批量更新成功（返回体为空）。
type ProductGoodsChannelBindingAutoPriceBatchRes struct{}
