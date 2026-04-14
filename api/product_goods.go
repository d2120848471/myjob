package api

import "github.com/gogf/gf/v2/frame/g"

type ProductGoodsListReq struct {
	g.Meta    `path:"/products" method:"get" tags:"商品管理" summary:"商品列表" security:"BearerAuth" dc:"分页查询商品列表"`
	Page      int    `json:"page" dc:"页码"`
	PageSize  int    `json:"page_size" dc:"每页条数"`
	Keyword   string `json:"keyword" dc:"关键词"`
	BrandID   int64  `json:"brand_id" dc:"品牌ID"`
	GoodsType string `json:"goods_type" dc:"商品类型"`
	HasTax    string `json:"has_tax" dc:"含税筛选"`
	Status    string `json:"status" dc:"状态筛选"`
}

type ProductGoodsListRes struct {
	List       []ProductGoodsListItem `json:"list" dc:"商品列表"`
	Pagination PaginationRes          `json:"pagination" dc:"分页信息"`
}

type ProductGoodsDetailReq struct {
	g.Meta `path:"/products/{id}" method:"get" tags:"商品管理" summary:"商品详情" security:"BearerAuth" dc:"获取商品详情"`
	ID     int64 `json:"id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

type ProductGoodsDetailRes struct {
	ID                          int64  `json:"id" dc:"商品ID"`
	GoodsCode                   string `json:"goods_code" dc:"商品业务编码"`
	BrandID                     int64  `json:"brand_id" dc:"品牌ID"`
	BrandName                   string `json:"brand_name" dc:"品牌名称"`
	Name                        string `json:"name" dc:"商品名称"`
	GoodsType                   string `json:"goods_type" dc:"商品类型"`
	SupplyType                  string `json:"supply_type" dc:"供货方式"`
	IsExport                    int    `json:"is_export" dc:"是否可导出"`
	IsDouyin                    int    `json:"is_douyin" dc:"是否可抖音"`
	HasTax                      int    `json:"has_tax" dc:"是否含税"`
	ExceptionNotify             int    `json:"exception_notify" dc:"是否异常提醒"`
	SubjectID                   *int64 `json:"subject_id" dc:"主体ID"`
	SubjectName                 string `json:"subject_name" dc:"主体名称"`
	ProductTemplateID           *int64 `json:"product_template_id" dc:"商品模板ID"`
	ProductTemplateTitle        string `json:"product_template_title" dc:"商品模板标题"`
	PurchaseLimitStrategyID     *int64 `json:"purchase_limit_strategy_id" dc:"购买数量限制策略ID"`
	PurchaseLimitStrategyName   string `json:"purchase_limit_strategy_name" dc:"购买数量限制策略名称"`
	PurchaseLimitStrategyStatus int    `json:"purchase_limit_strategy_status" dc:"购买数量限制策略状态"`
	PurchaseNotice              string `json:"purchase_notice" dc:"购买须知"`
	TerminalPriceLimit          string `json:"terminal_price_limit" dc:"终端限价"`
	BalanceLimit                string `json:"balance_limit" dc:"余额限制"`
	DefaultSellPrice            string `json:"default_sell_price" dc:"默认售价"`
	MinPurchaseQty              int    `json:"min_purchase_qty" dc:"最小购买数量"`
	MaxPurchaseQty              int    `json:"max_purchase_qty" dc:"最大购买数量"`
	Status                      int    `json:"status" dc:"状态"`
	CreatedAt                   string `json:"created_at" dc:"创建时间"`
	UpdatedAt                   string `json:"updated_at" dc:"更新时间"`
}

type ProductGoodsCreateReq struct {
	g.Meta                  `path:"/products" method:"post" tags:"商品管理" summary:"新增商品" security:"BearerAuth" dc:"新增商品"`
	BrandID                 int64  `json:"brand_id" dc:"品牌ID"`
	Name                    string `json:"name" dc:"商品名称"`
	GoodsType               string `json:"goods_type" dc:"商品类型"`
	SupplyType              string `json:"supply_type" dc:"供货方式"`
	IsExport                int    `json:"is_export" dc:"是否可导出"`
	IsDouyin                int    `json:"is_douyin" dc:"是否可抖音"`
	HasTax                  int    `json:"has_tax" dc:"是否含税"`
	SubjectID               *int64 `json:"subject_id" dc:"主体ID"`
	ExceptionNotify         int    `json:"exception_notify" dc:"是否异常提醒"`
	ProductTemplateID       *int64 `json:"product_template_id" dc:"商品模板ID"`
	PurchaseLimitStrategyID *int64 `json:"purchase_limit_strategy_id" dc:"购买数量限制策略ID"`
	PurchaseNotice          string `json:"purchase_notice" dc:"购买须知"`
	TerminalPriceLimit      string `json:"terminal_price_limit" dc:"终端限价"`
	BalanceLimit            string `json:"balance_limit" dc:"余额限制"`
	DefaultSellPrice        string `json:"default_sell_price" dc:"默认售价"`
	MinPurchaseQty          int    `json:"min_purchase_qty" dc:"最小购买数量"`
	MaxPurchaseQty          int    `json:"max_purchase_qty" dc:"最大购买数量"`
	Status                  int    `json:"status" dc:"状态"`
}

type ProductGoodsCreateRes struct {
	ID        int64  `json:"id" dc:"商品ID"`
	GoodsCode string `json:"goods_code" dc:"商品业务编码"`
}

type ProductGoodsUpdateReq struct {
	g.Meta                  `path:"/products/{id}" method:"put" tags:"商品管理" summary:"编辑商品" security:"BearerAuth" dc:"编辑商品"`
	ID                      int64  `json:"id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
	BrandID                 int64  `json:"brand_id" dc:"品牌ID"`
	Name                    string `json:"name" dc:"商品名称"`
	GoodsType               string `json:"goods_type" dc:"商品类型"`
	SupplyType              string `json:"supply_type" dc:"供货方式"`
	IsExport                int    `json:"is_export" dc:"是否可导出"`
	IsDouyin                int    `json:"is_douyin" dc:"是否可抖音"`
	HasTax                  int    `json:"has_tax" dc:"是否含税"`
	SubjectID               *int64 `json:"subject_id" dc:"主体ID"`
	ExceptionNotify         int    `json:"exception_notify" dc:"是否异常提醒"`
	ProductTemplateID       *int64 `json:"product_template_id" dc:"商品模板ID"`
	PurchaseLimitStrategyID *int64 `json:"purchase_limit_strategy_id" dc:"购买数量限制策略ID"`
	PurchaseNotice          string `json:"purchase_notice" dc:"购买须知"`
	TerminalPriceLimit      string `json:"terminal_price_limit" dc:"终端限价"`
	BalanceLimit            string `json:"balance_limit" dc:"余额限制"`
	DefaultSellPrice        string `json:"default_sell_price" dc:"默认售价"`
	MinPurchaseQty          int    `json:"min_purchase_qty" dc:"最小购买数量"`
	MaxPurchaseQty          int    `json:"max_purchase_qty" dc:"最大购买数量"`
	Status                  int    `json:"status" dc:"状态"`
}

type ProductGoodsUpdateRes struct{}

type ProductGoodsDeleteReq struct {
	g.Meta `path:"/products/{id}" method:"delete" tags:"商品管理" summary:"删除商品" security:"BearerAuth" dc:"软删除商品"`
	ID     int64 `json:"id" in:"path" v:"required#商品ID不能为空" dc:"商品ID"`
}

type ProductGoodsDeleteRes struct{}

type ProductGoodsFormOptionsReq struct {
	g.Meta `path:"/products/form-options" method:"get" tags:"商品管理" summary:"商品表单下拉数据" security:"BearerAuth" dc:"获取商品表单聚合下拉数据"`
}

type ProductGoodsFormOptionsRes struct {
	Brands                  []ProductGoodsBrandTreeItem  `json:"brands" dc:"品牌树"`
	Templates               []ProductGoodsTemplateOption `json:"templates" dc:"模板下拉"`
	PurchaseLimitStrategies []ProductGoodsStrategyOption `json:"purchase_limit_strategies" dc:"策略下拉"`
	Subjects                []ProductGoodsSubjectOption  `json:"subjects" dc:"主体下拉"`
	GoodsTypes              []ProductGoodsStringOption   `json:"goods_types" dc:"商品类型"`
	SupplyTypes             []ProductGoodsStringOption   `json:"supply_types" dc:"供货方式"`
	BooleanOptions          []ProductGoodsIntOption      `json:"boolean_options" dc:"布尔选项"`
	StatusOptions           []ProductGoodsIntOption      `json:"status_options" dc:"状态选项"`
}

type ProductGoodsListItem struct {
	ID                        int64  `json:"id" dc:"商品ID"`
	GoodsCode                 string `json:"goods_code" dc:"商品业务编码"`
	BrandID                   int64  `json:"brand_id" dc:"品牌ID"`
	BrandName                 string `json:"brand_name" dc:"品牌名称"`
	BrandIcon                 string `json:"brand_icon" dc:"品牌图片"`
	SubjectID                 *int64 `json:"subject_id" dc:"主体ID"`
	SubjectName               string `json:"subject_name" dc:"主体名称"`
	Name                      string `json:"name" dc:"商品名称"`
	GoodsType                 string `json:"goods_type" dc:"商品类型"`
	SupplyType                string `json:"supply_type" dc:"供货方式"`
	IsExport                  int    `json:"is_export" dc:"是否可导出"`
	IsDouyin                  int    `json:"is_douyin" dc:"是否可抖音"`
	HasTax                    int    `json:"has_tax" dc:"是否含税"`
	ExceptionNotify           int    `json:"exception_notify" dc:"是否异常提醒"`
	ProductTemplateID         int64  `json:"product_template_id" dc:"商品模板ID"`
	ProductTemplateTitle      string `json:"product_template_title" dc:"商品模板标题"`
	PurchaseLimitStrategyID   int64  `json:"purchase_limit_strategy_id" dc:"购买数量限制策略ID"`
	PurchaseLimitStrategyName string `json:"purchase_limit_strategy_name" dc:"购买数量限制策略名称"`
	DefaultSellPrice          string `json:"default_sell_price" dc:"默认售价"`
	TerminalPriceLimit        string `json:"terminal_price_limit" dc:"终端限价"`
	Status                    int    `json:"status" dc:"状态"`
	CreatedAt                 string `json:"created_at" dc:"创建时间"`
}

type ProductGoodsBrandTreeItem struct {
	ID       int64                       `json:"id" dc:"品牌ID"`
	Name     string                      `json:"name" dc:"品牌名称"`
	IsLeaf   bool                        `json:"is_leaf" dc:"是否末级品牌"`
	Children []ProductGoodsBrandTreeItem `json:"children" dc:"子品牌"`
}

type ProductGoodsTemplateOption struct {
	ID    int64  `json:"id" dc:"模板ID"`
	Title string `json:"title" dc:"模板标题"`
}

type ProductGoodsStrategyOption struct {
	ID   int64  `json:"id" dc:"策略ID"`
	Name string `json:"name" dc:"策略名称"`
}

type ProductGoodsSubjectOption struct {
	ID   int64  `json:"id" dc:"主体ID"`
	Name string `json:"name" dc:"主体名称"`
}

type ProductGoodsStringOption struct {
	Value string `json:"value" dc:"值"`
	Label string `json:"label" dc:"文案"`
}

type ProductGoodsIntOption struct {
	Value int    `json:"value" dc:"值"`
	Label string `json:"label" dc:"文案"`
}
