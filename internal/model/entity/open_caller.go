package entity

import (
	"database/sql"
	"time"
)

// OpenCaller 对应 open_caller 对外调用方主数据表。
type OpenCaller struct {
	ID            int64        `db:"id" json:"id"`
	Name          string       `db:"name" json:"name"`
	AppKey        string       `db:"app_key" json:"app_key"`
	AppSecret     string       `db:"app_secret" json:"app_secret"`
	Status        string       `db:"status" json:"status"`
	AllowedIPList string       `db:"allowed_ip_list" json:"allowed_ip_list"`
	SignVersion   string       `db:"sign_version" json:"sign_version"`
	Remark        string       `db:"remark" json:"remark"`
	IsDeleted     int          `db:"is_deleted" json:"is_deleted"`
	DeletedAt     sql.NullTime `db:"deleted_at" json:"deleted_at"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}
