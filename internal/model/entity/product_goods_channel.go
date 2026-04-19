package entity

import (
	"database/sql"
	"time"
)

// ProductGoodsChannelBinding 表示商品与渠道账号之间的单条绑定关系。
type ProductGoodsChannelBinding struct {
	ID                 int64          `db:"id" json:"id"`
	GoodsID            int64          `db:"goods_id" json:"goods_id"`
	PlatformAccountID  int64          `db:"platform_account_id" json:"platform_account_id"`
	SupplierGoodsNo    string         `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName  string         `db:"supplier_goods_name" json:"supplier_goods_name"`
	SourceCostPrice    string         `db:"source_cost_price" json:"source_cost_price"`
	CostPrice          string         `db:"cost_price" json:"cost_price"`
	TaxAdjustDirection string         `db:"tax_adjust_direction" json:"tax_adjust_direction"`
	TaxAdjustRate      string         `db:"tax_adjust_rate" json:"tax_adjust_rate"`
	TaxAdjustAmount    string         `db:"tax_adjust_amount" json:"tax_adjust_amount"`
	DockStatus         int            `db:"dock_status" json:"dock_status"`
	Sort               int            `db:"sort" json:"sort"`
	OrderWeight        string         `db:"order_weight" json:"order_weight"`
	OrderTimeStart     sql.NullString `db:"order_time_start" json:"order_time_start"`
	OrderTimeEnd       sql.NullString `db:"order_time_end" json:"order_time_end"`
	ValidateTemplateID sql.NullInt64  `db:"validate_template_id" json:"validate_template_id"`
	IsAutoChange       int            `db:"is_auto_change" json:"is_auto_change"`
	AddType            string         `db:"add_type" json:"add_type"`
	DefaultPrice       string         `db:"default_price" json:"default_price"`
	IsDeleted          int            `db:"is_deleted" json:"is_deleted"`
	DeletedAt          sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at"`
}

// ProductGoodsChannelConfig 表示商品维度的渠道库存配置。
type ProductGoodsChannelConfig struct {
	GoodsID               int64     `db:"goods_id" json:"goods_id"`
	SmartReorderEnabled   int       `db:"smart_reorder_enabled" json:"smart_reorder_enabled"`
	ReorderTimeoutEnabled int       `db:"reorder_timeout_enabled" json:"reorder_timeout_enabled"`
	ReorderTimeoutMinutes int       `db:"reorder_timeout_minutes" json:"reorder_timeout_minutes"`
	OrderStrategy         string    `db:"order_strategy" json:"order_strategy"`
	SyncCostPriceEnabled  int       `db:"sync_cost_price_enabled" json:"sync_cost_price_enabled"`
	SyncGoodsNameEnabled  int       `db:"sync_goods_name_enabled" json:"sync_goods_name_enabled"`
	AllowLossSaleEnabled  int       `db:"allow_loss_sale_enabled" json:"allow_loss_sale_enabled"`
	MaxLossAmount         string    `db:"max_loss_amount" json:"max_loss_amount"`
	ComboGoodsEnabled     int       `db:"combo_goods_enabled" json:"combo_goods_enabled"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}
