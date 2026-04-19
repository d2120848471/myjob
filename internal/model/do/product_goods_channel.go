package do

import "time"

// ProductGoodsChannelBinding 定义商品渠道绑定 DO。
type ProductGoodsChannelBinding struct {
	ID                 any
	GoodsID            any
	PlatformAccountID  any
	SupplierGoodsNo    any
	SupplierGoodsName  any
	SourceCostPrice    any
	CostPrice          any
	TaxAdjustDirection any
	TaxAdjustRate      any
	TaxAdjustAmount    any
	DockStatus         any
	Sort               any
	OrderWeight        any
	OrderTimeStart     any
	OrderTimeEnd       any
	ValidateTemplateID any
	IsAutoChange       any
	AddType            any
	DefaultPrice       any
	IsDeleted          any
	DeletedAt          *time.Time
	CreatedAt          any
	UpdatedAt          any
}

// ProductGoodsChannelConfig 定义商品库存配置 DO。
type ProductGoodsChannelConfig struct {
	GoodsID               any
	SmartReorderEnabled   any
	ReorderTimeoutEnabled any
	ReorderTimeoutMinutes any
	OrderStrategy         any
	SyncCostPriceEnabled  any
	SyncGoodsNameEnabled  any
	AllowLossSaleEnabled  any
	MaxLossAmount         any
	ComboGoodsEnabled     any
	CreatedAt             any
	UpdatedAt             any
}
