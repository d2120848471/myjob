package api

import "github.com/gogf/gf/v2/frame/g"

// ProductGoodsChannelBindingListReq 用于读取指定商品的渠道绑定弹窗数据。
type ProductGoodsChannelBindingListReq struct {
	g.Meta  `path:"/products/{goodsId}/channel-bindings" method:"get" tags:"商品管理" summary:"商品渠道绑定列表" security:"BearerAuth" dc:"读取商品渠道绑定弹窗所需的商品摘要与绑定列表"`
	GoodsId int64 `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsChannelBindingListRes 返回商品渠道绑定弹窗数据。
type ProductGoodsChannelBindingListRes struct {
	Goods ProductGoodsChannelGoodsSummary  `json:"goods" dc:"商品摘要"`
	List  []ProductGoodsChannelBindingItem `json:"list" dc:"绑定列表"`
}

// ProductGoodsChannelBindingFormOptionsReq 用于读取商品渠道绑定表单选项。
type ProductGoodsChannelBindingFormOptionsReq struct {
	g.Meta  `path:"/products/{goodsId}/channel-bindings/form-options" method:"get" tags:"商品管理" summary:"商品渠道绑定表单选项" security:"BearerAuth" dc:"读取商品渠道绑定弹窗的渠道账号、模板和枚举选项"`
	GoodsId int64 `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

// ProductGoodsChannelBindingFormOptionsRes 返回商品渠道绑定表单选项。
type ProductGoodsChannelBindingFormOptionsRes struct {
	PlatformAccounts     []ProductGoodsChannelPlatformAccountOption `json:"platform_accounts" dc:"渠道账号选项"`
	ValidateTemplates    []ProductGoodsTemplateOption               `json:"validate_templates" dc:"充值模板选项"`
	DockStatusOptions    []ProductGoodsIntOption                    `json:"dock_status_options" dc:"对接状态选项"`
	AutoPriceTypeOptions []ProductGoodsStringOption                 `json:"auto_price_type_options" dc:"自动改价类型选项"`
}

// ProductGoodsChannelBindingCreateReq 用于新增单条商品渠道绑定。
type ProductGoodsChannelBindingCreateReq struct {
	g.Meta             `path:"/products/{goodsId}/channel-bindings" method:"post" tags:"商品管理" summary:"新增商品渠道绑定" security:"BearerAuth" dc:"为指定商品新增一条渠道绑定"`
	GoodsId            int64  `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	PlatformAccountID  int64  `json:"platform_account_id" dc:"渠道账号ID"`
	SupplierGoodsNo    string `json:"supplier_goods_no" dc:"对接商品编号"`
	SupplierGoodsName  string `json:"supplier_goods_name" dc:"对接商品名称"`
	SourceCostPrice    string `json:"source_cost_price" dc:"原始进货价"`
	ValidateTemplateID *int64 `json:"validate_template_id" dc:"充值模板ID"`
	DockStatus         int    `json:"dock_status" dc:"对接状态"`
	Sort               int    `json:"sort" dc:"排序值"`
}

// ProductGoodsChannelBindingCreateRes 返回新增后的绑定 ID。
type ProductGoodsChannelBindingCreateRes struct {
	ID int64 `json:"id" dc:"绑定ID"`
}

// ProductGoodsChannelBindingUpdateReq 用于编辑单条商品渠道绑定基础字段。
type ProductGoodsChannelBindingUpdateReq struct {
	g.Meta             `path:"/products/{goodsId}/channel-bindings/{bindingId}" method:"patch" tags:"商品管理" summary:"编辑商品渠道绑定" security:"BearerAuth" dc:"编辑商品渠道绑定的基础字段并重算比较成本价"`
	GoodsId            int64  `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingId          int64  `json:"bindingId" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`
	PlatformAccountID  int64  `json:"platform_account_id" dc:"渠道账号ID"`
	SupplierGoodsNo    string `json:"supplier_goods_no" dc:"对接商品编号"`
	SupplierGoodsName  string `json:"supplier_goods_name" dc:"对接商品名称"`
	SourceCostPrice    string `json:"source_cost_price" dc:"原始进货价"`
	ValidateTemplateID *int64 `json:"validate_template_id" dc:"充值模板ID"`
	DockStatus         int    `json:"dock_status" dc:"对接状态"`
	Sort               int    `json:"sort" dc:"排序值"`
}

// ProductGoodsChannelBindingUpdateRes 表示商品渠道绑定编辑成功。
type ProductGoodsChannelBindingUpdateRes struct{}

// ProductGoodsChannelBindingDeleteReq 用于软删除单条商品渠道绑定。
type ProductGoodsChannelBindingDeleteReq struct {
	g.Meta    `path:"/products/{goodsId}/channel-bindings/{bindingId}" method:"delete" tags:"商品管理" summary:"删除商品渠道绑定" security:"BearerAuth" dc:"软删除指定商品下的单条渠道绑定"`
	GoodsId   int64 `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingId int64 `json:"bindingId" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`
}

// ProductGoodsChannelBindingDeleteRes 表示商品渠道绑定删除成功。
type ProductGoodsChannelBindingDeleteRes struct{}

// ProductGoodsChannelBindingAutoPriceUpdateReq 用于编辑单条绑定的自动改价规则。
type ProductGoodsChannelBindingAutoPriceUpdateReq struct {
	g.Meta       `path:"/products/{goodsId}/channel-bindings/{bindingId}/auto-price" method:"patch" tags:"商品管理" summary:"编辑绑定自动改价" security:"BearerAuth" dc:"编辑指定商品渠道绑定的自动改价规则"`
	GoodsId      int64  `json:"goodsId" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BindingId    int64  `json:"bindingId" in:"path" v:"required#绑定ID不能为空" dc:"绑定ID"`
	IsAutoChange int    `json:"is_auto_change" dc:"是否启用自动改价"`
	AddType      string `json:"add_type" dc:"加价类型"`
	DefaultPrice string `json:"default_price" dc:"利润值"`
}

// ProductGoodsChannelBindingAutoPriceUpdateRes 表示自动改价规则保存成功。
type ProductGoodsChannelBindingAutoPriceUpdateRes struct{}

// ProductGoodsChannelGoodsSummary 是渠道弹窗顶部展示的商品摘要。
type ProductGoodsChannelGoodsSummary struct {
	ID               int64  `json:"id" dc:"商品ID"`
	GoodsCode        string `json:"goods_code" dc:"商品编码"`
	Name             string `json:"name" dc:"商品名称"`
	BrandName        string `json:"brand_name" dc:"品牌名称"`
	SubjectID        *int64 `json:"subject_id" dc:"主体ID"`
	SubjectName      string `json:"subject_name" dc:"主体名称"`
	HasTax           int    `json:"has_tax" dc:"是否含税"`
	DefaultSellPrice string `json:"default_sell_price" dc:"默认售价"`
}

// ProductGoodsChannelBindingItem 是商品渠道绑定弹窗中的单行绑定数据。
type ProductGoodsChannelBindingItem struct {
	ID                    int64  `json:"id" dc:"绑定ID"`
	PlatformAccountID     int64  `json:"platform_account_id" dc:"渠道账号ID"`
	PlatformAccountName   string `json:"platform_account_name" dc:"渠道账号名称"`
	PlatformHasTax        int    `json:"platform_has_tax" dc:"渠道是否含税"`
	ConnectStatus         int    `json:"connect_status" dc:"渠道连接状态"`
	ConnectStatusText     string `json:"connect_status_text" dc:"渠道连接状态文案"`
	SupplierGoodsNo       string `json:"supplier_goods_no" dc:"对接商品编号"`
	SupplierGoodsName     string `json:"supplier_goods_name" dc:"对接商品名称"`
	DisplayName           string `json:"display_name" dc:"列表展示名称"`
	SourceCostPrice       string `json:"source_cost_price" dc:"原始进货价"`
	CostPrice             string `json:"cost_price" dc:"比较成本价"`
	EffectiveSellPrice    string `json:"effective_sell_price" dc:"利润后价格"`
	TaxAdjustDirection    string `json:"tax_adjust_direction" dc:"税额调整方向"`
	TaxAdjustRate         string `json:"tax_adjust_rate" dc:"税率"`
	TaxAdjustAmount       string `json:"tax_adjust_amount" dc:"税额调整值"`
	ValidateTemplateID    *int64 `json:"validate_template_id" dc:"充值模板ID"`
	ValidateTemplateTitle string `json:"validate_template_title" dc:"充值模板标题"`
	DockStatus            int    `json:"dock_status" dc:"对接状态"`
	Sort                  int    `json:"sort" dc:"排序值"`
	IsAutoChange          int    `json:"is_auto_change" dc:"是否启用自动改价"`
	AddType               string `json:"add_type" dc:"加价类型"`
	DefaultPrice          string `json:"default_price" dc:"利润值"`
	CreatedAt             string `json:"created_at" dc:"创建时间"`
	UpdatedAt             string `json:"updated_at" dc:"更新时间"`
}

// ProductGoodsChannelPlatformAccountOption 是商品渠道绑定表单中的渠道账号选项。
type ProductGoodsChannelPlatformAccountOption struct {
	ID                int64  `json:"id" dc:"渠道账号ID"`
	Name              string `json:"name" dc:"渠道账号名称"`
	SubjectID         int64  `json:"subject_id" dc:"主体ID"`
	SubjectName       string `json:"subject_name" dc:"主体名称"`
	HasTax            int    `json:"has_tax" dc:"是否含税"`
	ConnectStatus     int    `json:"connect_status" dc:"连接状态"`
	ConnectStatusText string `json:"connect_status_text" dc:"连接状态文案"`
}
