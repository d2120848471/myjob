package entity

import (
	"database/sql"
	"time"
)

// ExternalOrder 表示外部开放下单创建的主订单。
type ExternalOrder struct {
	ID               int64         `db:"id" json:"id"`
	OrderNo          string        `db:"order_no" json:"order_no"`
	GoodsID          int64         `db:"goods_id" json:"goods_id"`
	GoodsCode        string        `db:"goods_code" json:"goods_code"`
	GoodsName        string        `db:"goods_name" json:"goods_name"`
	GoodsType        string        `db:"goods_type" json:"goods_type"`
	SupplyType       string        `db:"supply_type" json:"supply_type"`
	SubjectID        sql.NullInt64 `db:"subject_id" json:"subject_id"`
	SubjectName      string        `db:"subject_name" json:"subject_name"`
	HasTax           int           `db:"has_tax" json:"has_tax"`
	Account          string        `db:"account" json:"account"`
	Quantity         int           `db:"quantity" json:"quantity"`
	UnitPrice        string        `db:"unit_price" json:"unit_price"`
	OrderAmount      string        `db:"order_amount" json:"order_amount"`
	CostAmount       string        `db:"cost_amount" json:"cost_amount"`
	ProfitAmount     string        `db:"profit_amount" json:"profit_amount"`
	Status           string        `db:"status" json:"status"`
	CurrentAttemptID int64         `db:"current_attempt_id" json:"current_attempt_id"`
	AttemptCount     int           `db:"attempt_count" json:"attempt_count"`
	LastReceipt      string        `db:"last_receipt" json:"last_receipt"`
	NextPollAt       time.Time     `db:"next_poll_at" json:"next_poll_at"`
	LastPollAt       time.Time     `db:"last_poll_at" json:"last_poll_at"`
	PollCount        int           `db:"poll_count" json:"poll_count"`
	LastPollError    string        `db:"last_poll_error" json:"last_poll_error"`
	CreatedAt        time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at" json:"updated_at"`
}

// ExternalOrderAttempt 表示一次渠道提交或查单尝试。
type ExternalOrderAttempt struct {
	ID                  int64     `db:"id" json:"id"`
	OrderID             int64     `db:"order_id" json:"order_id"`
	OrderNo             string    `db:"order_no" json:"order_no"`
	AttemptNo           int       `db:"attempt_no" json:"attempt_no"`
	ChannelBindingID    int64     `db:"channel_binding_id" json:"channel_binding_id"`
	PlatformAccountID   int64     `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName string    `db:"platform_account_name" json:"platform_account_name"`
	ProviderCode        string    `db:"provider_code" json:"provider_code"`
	SupplierGoodsNo     string    `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName   string    `db:"supplier_goods_name" json:"supplier_goods_name"`
	SupplierUSOrderNo   string    `db:"supplier_us_order_no" json:"supplier_us_order_no"`
	SupplierOrderNo     string    `db:"supplier_order_no" json:"supplier_order_no"`
	SupplierStatus      string    `db:"supplier_status" json:"supplier_status"`
	RefundStatus        string    `db:"refund_status" json:"refund_status"`
	RequestSnapshot     string    `db:"request_snapshot" json:"request_snapshot"`
	ResponseSnapshot    string    `db:"response_snapshot" json:"response_snapshot"`
	Receipt             string    `db:"receipt" json:"receipt"`
	Status              string    `db:"status" json:"status"`
	SubmittedAt         time.Time `db:"submitted_at" json:"submitted_at"`
	LastCheckedAt       time.Time `db:"last_checked_at" json:"last_checked_at"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}
