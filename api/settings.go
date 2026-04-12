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
