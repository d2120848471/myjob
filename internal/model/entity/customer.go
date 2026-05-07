package entity

import (
	"database/sql"
	"time"
)

// CustomerUser 是客户账号实体，包含登录、支付密码和后台管理状态。
type CustomerUser struct {
	ID              int64        `db:"id" json:"id"`
	CompanyName     string       `db:"company_name" json:"company_name"`
	Phone           string       `db:"phone" json:"phone"`
	PasswordHash    string       `db:"password_hash"`
	PayPasswordHash string       `db:"pay_password_hash"`
	Status          int          `db:"status" json:"status"`
	IsDeleted       int          `db:"is_deleted"`
	LastLoginIP     string       `db:"last_login_ip"`
	LastLoginAt     sql.NullTime `db:"last_login_at"`
	TokenVersion    int          `db:"token_version"`
	DeletedAt       sql.NullTime `db:"deleted_at"`
	CreatedAt       time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at" json:"updated_at"`
}

// CustomerListItem 是后台客户列表和回收站列表的展示项。
type CustomerListItem struct {
	ID          int64  `db:"id" json:"id"`
	CompanyName string `db:"company_name" json:"company_name"`
	Phone       string `db:"phone" json:"phone"`
	Status      int    `db:"status" json:"status"`
	LastLoginIP string `db:"last_login_ip" json:"last_login_ip"`
	LastLoginAt string `db:"last_login_at" json:"last_login_at"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
}
