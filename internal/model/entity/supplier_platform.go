package entity

import (
	"database/sql"
	"time"
)

type SupplierPlatformTypeItem struct {
	ID           int       `db:"id" json:"id"`
	TypeName     string    `db:"type_name" json:"type_name"`
	ProviderCode string    `db:"provider_code" json:"provider_code"`
	Status       int       `db:"status" json:"status"`
	Sort         int       `db:"sort" json:"sort"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type SupplierPlatformAccount struct {
	ID                 int64          `db:"id" json:"id"`
	Name               string         `db:"name" json:"name"`
	ProviderCode       string         `db:"provider_code" json:"provider_code"`
	ProviderName       string         `db:"provider_name" json:"provider_name"`
	TypeID             int            `db:"type_id" json:"type_id"`
	SubjectID          int64          `db:"subject_id" json:"subject_id"`
	HasTax             int            `db:"has_tax" json:"has_tax"`
	Domain             string         `db:"domain" json:"domain"`
	BackupDomain       string         `db:"backup_domain" json:"backup_domain"`
	TokenID            string         `db:"token_id" json:"token_id"`
	SecretKey          string         `db:"secret_key" json:"secret_key"`
	ExtraConfig        string         `db:"extra_config" json:"extra_config"`
	ThresholdAmount    string         `db:"threshold_amount" json:"threshold_amount"`
	Sort               int            `db:"sort" json:"sort"`
	CrowdName          string         `db:"crowd_name" json:"crowd_name"`
	LastBalance        sql.NullString `db:"last_balance" json:"last_balance"`
	LastBalanceStatus  int            `db:"last_balance_status" json:"last_balance_status"`
	LastBalanceMessage string         `db:"last_balance_message" json:"last_balance_message"`
	LastBalanceAt      sql.NullTime   `db:"last_balance_at" json:"last_balance_at"`
	LastBalanceTraceID string         `db:"last_balance_trace_id" json:"last_balance_trace_id"`
	IsDeleted          int            `db:"is_deleted" json:"is_deleted"`
	DeletedAt          sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at"`
}

type SupplierPlatformListItem struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Domain             string `json:"domain"`
	BackupDomain       string `json:"backup_domain"`
	ProviderCode       string `json:"provider_code"`
	ProviderName       string `json:"provider_name"`
	TypeID             int    `json:"type_id"`
	TypeName           string `json:"type_name"`
	SubjectID          int64  `json:"subject_id"`
	SubjectName        string `json:"subject_name"`
	HasTax             int    `json:"has_tax"`
	LastBalance        string `json:"last_balance"`
	ThresholdAmount    string `json:"threshold_amount"`
	BalanceWarning     int    `json:"balance_warning"`
	ConnectStatus      int    `json:"connect_status"`
	ConnectStatusText  string `json:"connect_status_text"`
	LastBalanceMessage string `json:"last_balance_message"`
	LastBalanceAt      string `json:"last_balance_at"`
	Sort               int    `json:"sort"`
	CrowdName          string `json:"crowd_name"`
}

type SupplierPlatformBalanceLog struct {
	ID               int64     `db:"id" json:"id"`
	PlatformID       int64     `db:"platform_id" json:"platform_id"`
	OperatorID       int64     `db:"operator_id" json:"operator_id"`
	OperatorName     string    `db:"operator_name" json:"operator_name"`
	ProviderCode     string    `db:"provider_code" json:"provider_code"`
	RequestURL       string    `db:"request_url" json:"request_url"`
	RequestMethod    string    `db:"request_method" json:"request_method"`
	RequestSnapshot  string    `db:"request_snapshot" json:"request_snapshot"`
	ResponseSnapshot string    `db:"response_snapshot" json:"response_snapshot"`
	HTTPStatus       int       `db:"http_status" json:"http_status"`
	Success          int       `db:"success" json:"success"`
	BalanceAmount    string    `db:"balance_amount" json:"balance_amount"`
	Message          string    `db:"message" json:"message"`
	DurationMS       int       `db:"duration_ms" json:"duration_ms"`
	TraceID          string    `db:"trace_id" json:"trace_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}
