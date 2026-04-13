package api

import "github.com/gogf/gf/v2/frame/g"

type SettingsSMSGetReq struct {
	g.Meta `path:"/settings/sms" method:"get" tags:"设置" summary:"读取短信配置" security:"BearerAuth" dc:"读取短信发送配置"`
}

type SettingsSMSGetRes struct {
	AccessKeyMasked           string `json:"access_key_masked" dc:"脱敏AccessKey"`
	AccessKeySecretMasked     string `json:"access_key_secret_masked" dc:"脱敏AccessKeySecret"`
	AccessKeyConfigured       bool   `json:"access_key_configured" dc:"是否已配置AccessKey"`
	AccessKeySecretConfigured bool   `json:"access_key_secret_configured" dc:"是否已配置AccessKeySecret"`
	SignName                  string `json:"sign_name" dc:"短信签名"`
	TemplateCode              string `json:"template_code" dc:"短信模板"`
	ExpireMinutes             int    `json:"expire_minutes" dc:"验证码有效期"`
	IntervalMinutes           int    `json:"interval_minutes" dc:"发送间隔"`
	UpdatedAt                 string `json:"updated_at,omitempty" dc:"更新时间"`
}

type SettingsSMSSaveReq struct {
	g.Meta              `path:"/settings/sms" method:"put" tags:"设置" summary:"保存短信配置" security:"BearerAuth" dc:"保存短信发送配置"`
	AccessKey           string `json:"access_key" dc:"AccessKey"`
	KeepAccessKey       bool   `json:"keep_access_key" dc:"是否保留原AccessKey"`
	AccessKeySecret     string `json:"access_key_secret" dc:"AccessKeySecret"`
	KeepAccessKeySecret bool   `json:"keep_access_key_secret" dc:"是否保留原AccessKeySecret"`
	SignName            string `json:"sign_name" dc:"短信签名"`
	TemplateCode        string `json:"template_code" dc:"短信模板"`
	ExpireMinutes       int    `json:"expire_minutes" dc:"验证码有效期"`
	IntervalMinutes     int    `json:"interval_minutes" dc:"发送间隔"`
}

type SettingsSMSSaveRes struct{}

type SettingsSystemGetReq struct {
	g.Meta `path:"/settings/system" method:"get" tags:"设置" summary:"读取系统参数配置" security:"BearerAuth" dc:"按分组读取系统参数配置"`
	Group  string `json:"group" dc:"配置分组，为空时返回全部分组"`
}

type SettingsSystemItem struct {
	Key        string `json:"key" dc:"组内键名"`
	Label      string `json:"label" dc:"展示名称"`
	Value      string `json:"value" dc:"参数值"`
	ValueType  string `json:"value_type" dc:"参数类型"`
	Unit       string `json:"unit,omitempty" dc:"单位"`
	Required   bool   `json:"required" dc:"是否必填"`
	Configured bool   `json:"configured" dc:"是否已配置"`
	UpdatedAt  string `json:"updated_at,omitempty" dc:"更新时间"`
}

type SettingsSystemGroup struct {
	Group string               `json:"group" dc:"配置分组"`
	Label string               `json:"label,omitempty" dc:"分组名称"`
	Items []SettingsSystemItem `json:"items" dc:"参数列表"`
}

type SettingsSystemGetRes struct {
	Group  string                `json:"group,omitempty" dc:"配置分组"`
	Label  string                `json:"label,omitempty" dc:"分组名称"`
	Items  []SettingsSystemItem  `json:"items,omitempty" dc:"参数列表"`
	Groups []SettingsSystemGroup `json:"groups,omitempty" dc:"分组参数列表"`
}

type SettingsSystemSaveReq struct {
	g.Meta `path:"/settings/system" method:"put" tags:"设置" summary:"保存系统参数配置" security:"BearerAuth" dc:"按分组批量保存系统参数配置"`
	Group  string                    `json:"group" dc:"配置分组，兼容旧单组写法"`
	Items  []SettingsSystemSaveItem  `json:"items" dc:"参数列表，兼容旧单组写法"`
	Groups []SettingsSystemSaveGroup `json:"groups" dc:"分组参数列表"`
}

type SettingsSystemSaveItem struct {
	Key   string `json:"key" v:"required#key不能为空" dc:"组内键名"`
	Value string `json:"value" dc:"参数值"`
}

type SettingsSystemSaveGroup struct {
	Group string                   `json:"group" v:"required#group不能为空" dc:"配置分组"`
	Items []SettingsSystemSaveItem `json:"items" v:"required#items不能为空" dc:"参数列表"`
}

type SettingsSystemSaveRes struct{}
