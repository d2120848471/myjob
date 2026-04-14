package entity

import (
	"database/sql"
	"time"
)

type ProductGoods struct {
	ID                      int64          `db:"id" json:"id"`
	GoodsCode               string         `db:"goods_code" json:"goods_code"`
	BrandID                 int64          `db:"brand_id" json:"brand_id"`
	Name                    string         `db:"name" json:"name"`
	GoodsType               string         `db:"goods_type" json:"goods_type"`
	SupplyType              string         `db:"supply_type" json:"supply_type"`
	IsExport                int            `db:"is_export" json:"is_export"`
	IsDouyin                int            `db:"is_douyin" json:"is_douyin"`
	HasTax                  int            `db:"has_tax" json:"has_tax"`
	ExceptionNotify         int            `db:"exception_notify" json:"exception_notify"`
	ProductTemplateID       sql.NullInt64  `db:"product_template_id" json:"product_template_id"`
	PurchaseLimitStrategyID sql.NullInt64  `db:"purchase_limit_strategy_id" json:"purchase_limit_strategy_id"`
	PurchaseNotice          sql.NullString `db:"purchase_notice" json:"purchase_notice"`
	TerminalPriceLimit      sql.NullString `db:"terminal_price_limit" json:"terminal_price_limit"`
	BalanceLimit            string         `db:"balance_limit" json:"balance_limit"`
	DefaultSellPrice        sql.NullString `db:"default_sell_price" json:"default_sell_price"`
	MinPurchaseQty          int            `db:"min_purchase_qty" json:"min_purchase_qty"`
	MaxPurchaseQty          int            `db:"max_purchase_qty" json:"max_purchase_qty"`
	Status                  int            `db:"status" json:"status"`
	IsDeleted               int            `db:"is_deleted" json:"is_deleted"`
	DeletedAt               sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	CreatedAt               time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time      `db:"updated_at" json:"updated_at"`
}
