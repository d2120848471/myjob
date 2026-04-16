package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// SupplierPlatformTypeItem 是第三方供货平台类型字典项。
type SupplierPlatformTypeItem = entity.SupplierPlatformTypeItem

// SupplierPlatformListItem 是第三方平台账号列表展示项。
type SupplierPlatformListItem = entity.SupplierPlatformListItem

// SupplierPlatformTypeListReq 用于读取第三方供货平台类型字典。
type SupplierPlatformTypeListReq struct {
	g.Meta `path:"/supplier-platform-types" method:"get" tags:"第三方对接" summary:"平台类型字典" security:"BearerAuth" dc:"获取第三方供货平台类型字典"`
}

// SupplierPlatformTypeListRes 返回平台类型字典列表。
type SupplierPlatformTypeListRes struct {
	List []SupplierPlatformTypeItem `json:"list" dc:"平台类型列表"`
}

// SupplierPlatformListReq 用于分页查询第三方供货平台账号列表，支持多维度筛选。
type SupplierPlatformListReq struct {
	g.Meta        `path:"/supplier-platforms" method:"get" tags:"第三方对接" summary:"平台账号列表" security:"BearerAuth" dc:"分页查询第三方供货平台账号"`
	Page          int    `json:"page" dc:"页码"`
	PageSize      int    `json:"page_size" dc:"每页条数"`
	Keyword       string `json:"keyword" dc:"平台名称关键词"`
	TypeID        int    `json:"type_id" dc:"平台类型ID"`
	SubjectID     int64  `json:"subject_id" dc:"主体ID"`
	HasTax        string `json:"has_tax" dc:"含税筛选"`
	ConnectStatus string `json:"connect_status" dc:"对接状态筛选"`
}

// SupplierPlatformListRes 返回平台账号列表与分页信息。
type SupplierPlatformListRes struct {
	List       []SupplierPlatformListItem `json:"list" dc:"平台账号列表"`
	Pagination PaginationRes              `json:"pagination" dc:"分页信息"`
}

// SupplierPlatformDetailReq 用于读取指定平台账号详情。
type SupplierPlatformDetailReq struct {
	g.Meta `path:"/supplier-platforms/{id}" method:"get" tags:"第三方对接" summary:"平台账号详情" security:"BearerAuth" dc:"获取第三方供货平台账号详情"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

// SupplierPlatformDetailRes 返回平台账号详情（包含域名、主体关联与密钥等配置）。
type SupplierPlatformDetailRes struct {
	ID              int64          `json:"id" dc:"平台ID"`
	Name            string         `json:"name" dc:"平台名称"`
	Domain          string         `json:"domain" dc:"主域名"`
	BackupDomain    string         `json:"backup_domain" dc:"备用域名"`
	TypeID          int            `json:"type_id" dc:"平台类型ID"`
	SubjectID       int64          `json:"subject_id" dc:"主体ID"`
	HasTax          int            `json:"has_tax" dc:"是否含税"`
	TokenID         string         `json:"token_id" dc:"商户号或平台账号ID"`
	SecretKey       string         `json:"secret_key" dc:"平台密钥"`
	ThresholdAmount string         `json:"threshold_amount" dc:"余额阈值"`
	Sort            int            `json:"sort" dc:"排序值"`
	CrowdName       string         `json:"crowd_name" dc:"群名或备注"`
	ProviderCode    string         `json:"provider_code" dc:"内部适配器编码"`
	ProviderName    string         `json:"provider_name" dc:"内部适配器名称"`
	ExtraConfig     map[string]any `json:"extra_config" dc:"扩展配置"`
}

// SupplierPlatformCreateReq 用于新增第三方供货平台账号。
type SupplierPlatformCreateReq struct {
	g.Meta          `path:"/supplier-platforms" method:"post" tags:"第三方对接" summary:"新增平台账号" security:"BearerAuth" dc:"新增第三方供货平台账号"`
	Name            string `json:"name" dc:"平台名称"`
	Domain          string `json:"domain" dc:"主域名"`
	BackupDomain    string `json:"backup_domain" dc:"备用域名"`
	TypeID          int    `json:"type_id" dc:"平台类型ID"`
	SubjectID       int64  `json:"subject_id" dc:"主体ID"`
	HasTax          int    `json:"has_tax" dc:"是否含税"`
	TokenID         string `json:"token_id" dc:"商户号或平台账号ID"`
	SecretKey       string `json:"secret_key" dc:"平台密钥"`
	ThresholdAmount string `json:"threshold_amount" dc:"余额阈值"`
	Sort            int    `json:"sort" dc:"排序值"`
	CrowdName       string `json:"crowd_name" dc:"群名或备注"`
}

// SupplierPlatformCreateRes 返回新增后的平台账号 ID。
type SupplierPlatformCreateRes struct {
	ID int64 `json:"id" dc:"平台ID"`
}

// SupplierPlatformUpdateReq 用于编辑第三方供货平台账号信息。
type SupplierPlatformUpdateReq struct {
	g.Meta          `path:"/supplier-platforms/{id}" method:"put" tags:"第三方对接" summary:"编辑平台账号" security:"BearerAuth" dc:"编辑第三方供货平台账号"`
	ID              int64  `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
	Name            string `json:"name" dc:"平台名称"`
	Domain          string `json:"domain" dc:"主域名"`
	BackupDomain    string `json:"backup_domain" dc:"备用域名"`
	TypeID          int    `json:"type_id" dc:"平台类型ID"`
	SubjectID       int64  `json:"subject_id" dc:"主体ID"`
	HasTax          int    `json:"has_tax" dc:"是否含税"`
	TokenID         string `json:"token_id" dc:"商户号或平台账号ID"`
	SecretKey       string `json:"secret_key" dc:"平台密钥"`
	ThresholdAmount string `json:"threshold_amount" dc:"余额阈值"`
	Sort            int    `json:"sort" dc:"排序值"`
	CrowdName       string `json:"crowd_name" dc:"群名或备注"`
}

// SupplierPlatformUpdateRes 表示平台账号编辑成功（返回体为空）。
type SupplierPlatformUpdateRes struct{}

// SupplierPlatformDeleteReq 用于软删除第三方供货平台账号。
type SupplierPlatformDeleteReq struct {
	g.Meta `path:"/supplier-platforms/{id}" method:"delete" tags:"第三方对接" summary:"删除平台账号" security:"BearerAuth" dc:"软删除第三方供货平台账号"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

// SupplierPlatformDeleteRes 表示平台账号删除成功（返回体为空）。
type SupplierPlatformDeleteRes struct{}

// SupplierPlatformRefreshBalanceReq 用于手动刷新指定平台账号余额。
type SupplierPlatformRefreshBalanceReq struct {
	g.Meta `path:"/supplier-platforms/{id}/balance/refresh" method:"post" tags:"第三方对接" summary:"刷新平台余额" security:"BearerAuth" dc:"手动刷新第三方供货平台余额"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

// SupplierPlatformRefreshBalanceRes 返回刷新后的余额与对接状态信息。
type SupplierPlatformRefreshBalanceRes struct {
	ID                int64  `json:"id" dc:"平台ID"`
	Balance           string `json:"balance" dc:"当前余额"`
	ConnectStatus     int    `json:"connect_status" dc:"对接状态"`
	ConnectStatusText string `json:"connect_status_text" dc:"对接状态文案"`
	Message           string `json:"message" dc:"刷新说明"`
	RefreshedAt       string `json:"refreshed_at" dc:"刷新时间"`
	TraceID           string `json:"trace_id" dc:"链路追踪ID"`
}
