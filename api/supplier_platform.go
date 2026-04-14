package api

import "github.com/gogf/gf/v2/frame/g"

type SupplierPlatformTypeListReq struct {
	g.Meta `path:"/supplier-platform-types" method:"get" tags:"第三方对接" summary:"平台类型字典" security:"BearerAuth" dc:"获取第三方供货平台类型字典"`
}

type SupplierPlatformTypeListRes struct {
	List []SupplierPlatformTypeItem `json:"list" dc:"平台类型列表"`
}

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

type SupplierPlatformListRes struct {
	List       []SupplierPlatformListItem `json:"list" dc:"平台账号列表"`
	Pagination PaginationRes              `json:"pagination" dc:"分页信息"`
}

type SupplierPlatformDetailReq struct {
	g.Meta `path:"/supplier-platforms/{id}" method:"get" tags:"第三方对接" summary:"平台账号详情" security:"BearerAuth" dc:"获取第三方供货平台账号详情"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

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

type SupplierPlatformCreateRes struct {
	ID int64 `json:"id" dc:"平台ID"`
}

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

type SupplierPlatformUpdateRes struct{}

type SupplierPlatformDeleteReq struct {
	g.Meta `path:"/supplier-platforms/{id}" method:"delete" tags:"第三方对接" summary:"删除平台账号" security:"BearerAuth" dc:"软删除第三方供货平台账号"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

type SupplierPlatformDeleteRes struct{}

type SupplierPlatformRefreshBalanceReq struct {
	g.Meta `path:"/supplier-platforms/{id}/balance/refresh" method:"post" tags:"第三方对接" summary:"刷新平台余额" security:"BearerAuth" dc:"手动刷新第三方供货平台余额"`
	ID     int64 `json:"id" in:"path" v:"required#平台ID不能为空" dc:"平台ID"`
}

type SupplierPlatformRefreshBalanceRes struct {
	ID                int64  `json:"id" dc:"平台ID"`
	Balance           string `json:"balance" dc:"当前余额"`
	ConnectStatus     int    `json:"connect_status" dc:"对接状态"`
	ConnectStatusText string `json:"connect_status_text" dc:"对接状态文案"`
	Message           string `json:"message" dc:"刷新说明"`
	RefreshedAt       string `json:"refreshed_at" dc:"刷新时间"`
	TraceID           string `json:"trace_id" dc:"链路追踪ID"`
}
