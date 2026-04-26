package entity

import (
	"database/sql"
	"time"
)

// SupplierProductSubscription 表示供应商商品推送订阅的本地状态。
type SupplierProductSubscription struct {
	ID                  int64        `db:"id" json:"id"`
	ProviderCode        string       `db:"provider_code" json:"provider_code"`
	PlatformAccountID   int64        `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName string       `db:"platform_account_name" json:"platform_account_name"`
	GoodsID             int64        `db:"goods_id" json:"goods_id"`
	BindingID           int64        `db:"binding_id" json:"binding_id"`
	SupplierGoodsNo     string       `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName   string       `db:"supplier_goods_name" json:"supplier_goods_name"`
	CallbackURL         string       `db:"callback_url" json:"callback_url"`
	Status              string       `db:"status" json:"status"`
	LastAction          string       `db:"last_action" json:"last_action"`
	LastError           string       `db:"last_error" json:"last_error"`
	RequestSnapshot     string       `db:"request_snapshot" json:"request_snapshot"`
	ResponseSnapshot    string       `db:"response_snapshot" json:"response_snapshot"`
	SubscribedAt        sql.NullTime `db:"subscribed_at" json:"subscribed_at"`
	CanceledAt          sql.NullTime `db:"canceled_at" json:"canceled_at"`
	CreatedAt           time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time    `db:"updated_at" json:"updated_at"`
}

// ProductGoodsChannelPriceChangeLog 表示一次由监控或推送触发的渠道价格变化。
type ProductGoodsChannelPriceChangeLog struct {
	ID                    int64     `db:"id" json:"id"`
	Source                string    `db:"source" json:"source"`
	ProviderCode          string    `db:"provider_code" json:"provider_code"`
	PlatformAccountID     int64     `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName   string    `db:"platform_account_name" json:"platform_account_name"`
	BindingID             int64     `db:"binding_id" json:"binding_id"`
	GoodsID               int64     `db:"goods_id" json:"goods_id"`
	GoodsCode             string    `db:"goods_code" json:"goods_code"`
	GoodsName             string    `db:"goods_name" json:"goods_name"`
	GoodsIcon             string    `db:"goods_icon" json:"goods_icon"`
	SupplierGoodsNo       string    `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName     string    `db:"supplier_goods_name" json:"supplier_goods_name"`
	OldSourceCostPrice    string    `db:"old_source_cost_price" json:"old_source_cost_price"`
	NewSourceCostPrice    string    `db:"new_source_cost_price" json:"new_source_cost_price"`
	OldCostPrice          string    `db:"old_cost_price" json:"old_cost_price"`
	NewCostPrice          string    `db:"new_cost_price" json:"new_cost_price"`
	OldEffectiveSellPrice string    `db:"old_effective_sell_price" json:"old_effective_sell_price"`
	NewEffectiveSellPrice string    `db:"new_effective_sell_price" json:"new_effective_sell_price"`
	ChangeAmount          string    `db:"change_amount" json:"change_amount"`
	Description           string    `db:"description" json:"description"`
	RawPayload            string    `db:"raw_payload" json:"raw_payload"`
	ChangedAt             time.Time `db:"changed_at" json:"changed_at"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}
