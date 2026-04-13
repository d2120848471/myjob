package entity

import (
	"database/sql"
	"time"
)

type AdminUser struct {
	ID            int64        `db:"id" json:"id"`
	Username      string       `db:"username" json:"username"`
	PasswordHash  string       `db:"password_hash"`
	RealName      string       `db:"real_name" json:"real_name"`
	Phone         string       `db:"phone" json:"phone"`
	GroupID       int64        `db:"group_id" json:"group_id"`
	Status        int          `db:"status" json:"status"`
	BalanceNotify int          `db:"balance_notify" json:"balance_notify"`
	IsBusiness    int          `db:"is_business" json:"is_business"`
	IsDeleted     int          `db:"is_deleted"`
	LastLoginIP   string       `db:"last_login_ip"`
	LastLoginAt   sql.NullTime `db:"last_login_at"`
	TokenVersion  int          `db:"token_version"`
	DeletedAt     sql.NullTime `db:"deleted_at"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
}

type UserListItem struct {
	ID            int64  `db:"id" json:"id"`
	Username      string `db:"username" json:"username"`
	RealName      string `db:"real_name" json:"real_name"`
	Phone         string `db:"phone" json:"phone"`
	GroupID       int64  `db:"group_id" json:"group_id"`
	GroupName     string `db:"group_name" json:"group_name"`
	Status        int    `db:"status" json:"status"`
	BalanceNotify int    `db:"balance_notify" json:"balance_notify"`
	IsBusiness    int    `db:"is_business" json:"is_business"`
}

type AdminGroup struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Status      int       `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type GroupListItem struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	Status      int    `db:"status" json:"status"`
	UserCount   int    `db:"user_count" json:"user_count"`
}

type AdminMenu struct {
	ID        int64        `db:"id" json:"id"`
	ParentID  int64        `db:"parent_id" json:"parent_id"`
	Name      string       `db:"name" json:"name"`
	Code      string       `db:"code" json:"code"`
	MenuType  string       `db:"menu_type" json:"menu_type"`
	MenuLevel int          `db:"menu_level" json:"-"`
	Status    int          `db:"status" json:"-"`
	SuperOnly int          `db:"super_only" json:"-"`
	Sort      int          `db:"sort" json:"sort"`
	Children  []*AdminMenu `json:"children,omitempty"`
}

type AdminSubject struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	HasTax    int       `db:"has_tax" json:"has_tax"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ProductBrand struct {
	ID              int64     `db:"id" json:"id"`
	ParentID        int64     `db:"parent_id" json:"parent_id"`
	Name            string    `db:"name" json:"name"`
	Icon            string    `db:"icon" json:"icon"`
	CredentialImage string    `db:"credential_image" json:"credential_image"`
	Description     string    `db:"description" json:"description"`
	IsVisible       int       `db:"is_visible" json:"is_visible"`
	Sort            int       `db:"sort" json:"sort"`
	GoodsCount      int       `db:"goods_count" json:"goods_count"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type BrandListItem struct {
	ID              int64           `db:"id" json:"id"`
	ParentID        int64           `db:"parent_id" json:"parent_id"`
	Name            string          `db:"name" json:"name"`
	Icon            string          `db:"icon" json:"icon"`
	CredentialImage string          `db:"credential_image" json:"credential_image"`
	Description     string          `db:"description" json:"description"`
	IsVisible       int             `db:"is_visible" json:"is_visible"`
	Sort            int             `db:"sort" json:"sort"`
	GoodsCount      int             `db:"goods_count" json:"goods_count"`
	HasChildren     bool            `db:"has_children" json:"has_children"`
	Children        []BrandListItem `json:"children"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type ProductIndustry struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Sort       int       `db:"sort" json:"sort"`
	BrandCount int       `db:"brand_count" json:"brand_count"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type IndustryListItem struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Sort       int       `db:"sort" json:"sort"`
	BrandCount int       `db:"brand_count" json:"brand_count"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type IndustryBrandRelationItem struct {
	ID        int64  `db:"id" json:"id"`
	BrandID   int64  `db:"brand_id" json:"brand_id"`
	BrandName string `db:"brand_name" json:"brand_name"`
	BrandIcon string `db:"brand_icon" json:"brand_icon"`
	Sort      int    `db:"sort" json:"sort"`
}

type ProductTemplate struct {
	ID           int64     `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	TemplateType string    `db:"template_type" json:"type"`
	IsShared     int       `db:"is_shared" json:"is_shared"`
	AccountName  string    `db:"account_name" json:"account_name"`
	ValidateType int       `db:"validate_type" json:"validate_type"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type ProductTemplateListItem struct {
	ID                int64     `db:"id" json:"id"`
	Title             string    `db:"title" json:"title"`
	Type              string    `db:"type" json:"type"`
	TypeLabel         string    `db:"type_label" json:"type_label"`
	IsShared          int       `db:"is_shared" json:"is_shared"`
	IsSharedLabel     string    `db:"is_shared_label" json:"is_shared_label"`
	AccountName       string    `db:"account_name" json:"account_name"`
	ValidateType      int       `db:"validate_type" json:"validate_type"`
	ValidateTypeLabel string    `db:"validate_type_label" json:"validate_type_label"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type ProductTemplateValidateTypeItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type PurchaseLimitStrategy struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	LimitType  int       `db:"limit_type" json:"limit_type"`
	PeriodType int       `db:"period_type" json:"period_type"`
	Period     int       `db:"period" json:"period"`
	LimitNums  int       `db:"limit_nums" json:"limit_nums"`
	LimitTimes int       `db:"limit_times" json:"limit_times"`
	Status     int       `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type PurchaseLimitStrategyListItem struct {
	ID              int64     `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	LimitType       int       `db:"limit_type" json:"limit_type"`
	LimitTypeLabel  string    `db:"limit_type_label" json:"limit_type_label"`
	PeriodType      int       `db:"period_type" json:"period_type"`
	PeriodTypeLabel string    `db:"period_type_label" json:"period_type_label"`
	Period          int       `db:"period" json:"period"`
	LimitNums       int       `db:"limit_nums" json:"limit_nums"`
	LimitTimes      int       `db:"limit_times" json:"limit_times"`
	Status          int       `db:"status" json:"status"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type PurchaseLimitEnumItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type BrandSelectorItem struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Icon string `db:"icon" json:"icon"`
}

type OperationLog struct {
	ID          int64     `db:"id" json:"id"`
	AdminID     int64     `db:"admin_id" json:"admin_id"`
	AdminName   string    `db:"admin_name" json:"admin_name"`
	Description string    `db:"description" json:"description"`
	IP          string    `db:"ip" json:"ip"`
	IPRegion    string    `db:"ip_region" json:"ip_region"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type LoginLog struct {
	ID        int64     `db:"id" json:"id"`
	AdminID   int64     `db:"admin_id" json:"admin_id"`
	AdminName string    `db:"admin_name" json:"admin_name"`
	IP        string    `db:"ip" json:"ip"`
	IPRegion  string    `db:"ip_region" json:"ip_region"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
