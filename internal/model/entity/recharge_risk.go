package entity

import (
	"database/sql"
	"time"
)

// RechargeRiskRule 表示一条充值账号风控规则。
type RechargeRiskRule struct {
	ID            int64        `db:"id" json:"id"`
	Account       string       `db:"account" json:"account"`
	GoodsKeyword  string       `db:"goods_keyword" json:"goods_keyword"`
	Reason        string       `db:"reason" json:"reason"`
	Status        int          `db:"status" json:"status"`
	HitCount      int          `db:"hit_count" json:"hit_count"`
	CreatedByID   int64        `db:"created_by_id" json:"created_by_id"`
	CreatedByName string       `db:"created_by_name" json:"created_by_name"`
	UpdatedByID   int64        `db:"updated_by_id" json:"updated_by_id"`
	UpdatedByName string       `db:"updated_by_name" json:"updated_by_name"`
	IsDeleted     int          `db:"is_deleted" json:"is_deleted"`
	DeletedAt     sql.NullTime `db:"deleted_at" json:"deleted_at"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}

// RechargeRiskRecord 表示一次开放下单被本地风控拦截的流水。
type RechargeRiskRecord struct {
	ID                 int64     `db:"id" json:"id"`
	RuleID             int64     `db:"rule_id" json:"rule_id"`
	OrderID            int64     `db:"order_id" json:"order_id"`
	OrderNo            string    `db:"order_no" json:"order_no"`
	Account            string    `db:"account" json:"account"`
	GoodsID            int64     `db:"goods_id" json:"goods_id"`
	GoodsCode          string    `db:"goods_code" json:"goods_code"`
	GoodsName          string    `db:"goods_name" json:"goods_name"`
	MatchedKeyword     string    `db:"matched_keyword" json:"matched_keyword"`
	Reason             string    `db:"reason" json:"reason"`
	RequestTokenMasked string    `db:"request_token_masked" json:"request_token_masked"`
	InterceptedAt      time.Time `db:"intercepted_at" json:"intercepted_at"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
