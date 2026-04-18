package entity

import (
	"database/sql"
	"time"
)

// ProductGoodsChannelBinding 表示商品与渠道账号之间的单条绑定关系。
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
	DockStatus         int           `db:"dock_status" json:"dock_status"`
	Sort               int           `db:"sort" json:"sort"`
	ValidateTemplateID sql.NullInt64 `db:"validate_template_id" json:"validate_template_id"`
	IsAutoChange       int           `db:"is_auto_change" json:"is_auto_change"`
	AddType            string        `db:"add_type" json:"add_type"`
	DefaultPrice       string        `db:"default_price" json:"default_price"`
	IsDeleted          int           `db:"is_deleted" json:"is_deleted"`
	DeletedAt          sql.NullTime  `db:"deleted_at" json:"deleted_at"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `db:"updated_at" json:"updated_at"`
}
