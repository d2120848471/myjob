package entity

import (
	"database/sql"
	"time"
)

// ProductGoodsChannelConfig 对应 product_goods_channel_config 商品级渠道配置表。
type ProductGoodsChannelConfig struct {
	ID                             int64          `db:"id" json:"id"`
	GoodsID                        int64          `db:"goods_id" json:"goods_id"`
	SmartReplenishEnabled          int            `db:"smart_replenish_enabled" json:"smart_replenish_enabled"`
	AttemptTimeoutEnabled          int            `db:"attempt_timeout_enabled" json:"attempt_timeout_enabled"`
	AttemptTimeoutMinutes          int            `db:"attempt_timeout_minutes" json:"attempt_timeout_minutes"`
	RouteMode                      string         `db:"route_mode" json:"route_mode"`
	SyncCostEnabled                int            `db:"sync_cost_enabled" json:"sync_cost_enabled"`
	SyncGoodsNameEnabled           int            `db:"sync_goods_name_enabled" json:"sync_goods_name_enabled"`
	AllowLoss                      int            `db:"allow_loss" json:"allow_loss"`
	MaxLossAmount                  sql.NullString `db:"max_loss_amount" json:"max_loss_amount"`
	IsBundle                       int            `db:"is_bundle" json:"is_bundle"`
	MinChannelCostSnapshot         sql.NullString `db:"min_channel_cost_snapshot" json:"min_channel_cost_snapshot"`
	BoundChannelCountSnapshot      int            `db:"bound_channel_count_snapshot" json:"bound_channel_count_snapshot"`
	PrimaryBindingID               sql.NullInt64  `db:"primary_binding_id" json:"primary_binding_id"`
	PrimaryChannelNameSnapshot     string         `db:"primary_channel_name_snapshot" json:"primary_channel_name_snapshot"`
	ChannelAutoPriceStatusSnapshot int            `db:"channel_auto_price_status_snapshot" json:"channel_auto_price_status_snapshot"`
	CreatedAt                      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt                      time.Time      `db:"updated_at" json:"updated_at"`
}

// ProductGoodsChannelBinding 对应 product_goods_channel_binding 商品-渠道绑定表。
type ProductGoodsChannelBinding struct {
	ID                 int64         `db:"id" json:"id"`
	GoodsID            int64         `db:"goods_id" json:"goods_id"`
	PlatformAccountID  int64         `db:"platform_account_id" json:"platform_account_id"`
	SupplierGoodsNo    string        `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName  string        `db:"supplier_goods_name" json:"supplier_goods_name"`
	SourceCostPrice    string        `db:"source_cost_price" json:"source_cost_price"`
	CostPrice          string        `db:"cost_price" json:"cost_price"`
	TaxAdjustDirection string        `db:"tax_adjust_direction" json:"tax_adjust_direction"`
	TaxAdjustRate      string        `db:"tax_adjust_rate" json:"tax_adjust_rate"`
	TaxAdjustAmount    string        `db:"tax_adjust_amount" json:"tax_adjust_amount"`
	DockStatus         string        `db:"dock_status" json:"dock_status"`
	Sort               int           `db:"sort" json:"sort"`
	Weight             int           `db:"weight" json:"weight"`
	StartTime          string        `db:"start_time" json:"start_time"`
	EndTime            string        `db:"end_time" json:"end_time"`
	ValidateTemplateID sql.NullInt64 `db:"validate_template_id" json:"validate_template_id"`
	IsAutoChange       int           `db:"is_auto_change" json:"is_auto_change"`
	AddType            string        `db:"add_type" json:"add_type"`
	DefaultPrice       string        `db:"default_price" json:"default_price"`
	LockPrice          string        `db:"lock_price" json:"lock_price"`
	SymbolPrice        string        `db:"symbol_price" json:"symbol_price"`
	MaxPrice           string        `db:"max_price" json:"max_price"`
	MinPrice           string        `db:"min_price" json:"min_price"`
	IsDeleted          int           `db:"is_deleted" json:"is_deleted"`
	DeletedAt          sql.NullTime  `db:"deleted_at" json:"deleted_at"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `db:"updated_at" json:"updated_at"`
}
