package api

import "github.com/gogf/gf/v2/frame/g"

// ProductGoodsChannelPriceChangeListReq 用于分页查询商品渠道自动改价记录。
type ProductGoodsChannelPriceChangeListReq struct {
	g.Meta          `path:"/product-goods-channel-price-changes" method:"get" tags:"商品管理" summary:"自动改价记录" security:"BearerAuth" dc:"分页查询监控或推送触发的商品渠道改价记录"`
	Page            int    `json:"page" dc:"页码"`
	PageSize        int    `json:"page_size" dc:"每页条数"`
	Source          string `json:"source" dc:"来源类型"`
	Keyword         string `json:"keyword" dc:"本地商品编号或名称"`
	SupplierGoodsNo string `json:"supplier_goods_no" dc:"上游商品编号"`
	PlatformID      int64  `json:"platform_id" dc:"平台账号ID"`
	StartAt         string `json:"start_at" dc:"开始时间"`
	EndAt           string `json:"end_at" dc:"结束时间"`
}

// ProductGoodsChannelPriceChangeListRes 返回商品渠道自动改价记录列表。
type ProductGoodsChannelPriceChangeListRes struct {
	List       []ProductGoodsChannelPriceChangeItem `json:"list" dc:"改价记录列表"`
	Pagination PaginationRes                        `json:"pagination" dc:"分页信息"`
}

// ProductGoodsChannelPriceChangeItem 是自动改价记录列表单行数据。
type ProductGoodsChannelPriceChangeItem struct {
	ID                    int64  `json:"id" dc:"改价记录ID"`
	Source                string `json:"source" dc:"来源"`
	ProviderCode          string `json:"provider_code" dc:"供应商编码"`
	PlatformAccountID     int64  `json:"platform_account_id" dc:"平台账号ID"`
	PlatformAccountName   string `json:"platform_account_name" dc:"平台账号名称"`
	BindingID             int64  `json:"binding_id" dc:"渠道绑定ID"`
	GoodsID               int64  `json:"goods_id" dc:"本地商品ID"`
	GoodsCode             string `json:"goods_code" dc:"本地商品编码"`
	GoodsName             string `json:"goods_name" dc:"本地商品名称"`
	GoodsIcon             string `json:"goods_icon" dc:"商品图标"`
	SupplierGoodsNo       string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName     string `json:"supplier_goods_name" dc:"上游商品名称"`
	OldSourceCostPrice    string `json:"old_source_cost_price" dc:"变动前原始进货价"`
	NewSourceCostPrice    string `json:"new_source_cost_price" dc:"变动后原始进货价"`
	OldCostPrice          string `json:"old_cost_price" dc:"变动前比较成本价"`
	NewCostPrice          string `json:"new_cost_price" dc:"变动后比较成本价"`
	OldEffectiveSellPrice string `json:"old_effective_sell_price" dc:"变动前利润后价格"`
	NewEffectiveSellPrice string `json:"new_effective_sell_price" dc:"变动后利润后价格"`
	ChangeAmount          string `json:"change_amount" dc:"利润后价格变化值"`
	Description           string `json:"description" dc:"变动描述"`
	RawPayload            string `json:"raw_payload" dc:"原始载荷"`
	ChangedAt             string `json:"changed_at" dc:"变动时间"`
}
