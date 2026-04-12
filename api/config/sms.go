package config

type SMSConfigSaveReq struct {
	AccessKey           string `json:"access_key"`
	AccessKeySecret     string `json:"access_key_secret"`
	SignName            string `json:"sign_name"`
	TemplateCode        string `json:"template_code"`
	ExpireMinutes       int    `json:"expire_minutes"`
	IntervalMinutes     int    `json:"interval_minutes"`
	KeepAccessKey       bool   `json:"keep_access_key"`
	KeepAccessKeySecret bool   `json:"keep_access_key_secret"`
}

type SMSConfigGetRes struct {
	AccessKeyMasked           string `json:"access_key_masked"`
	AccessKeySecretMasked     string `json:"access_key_secret_masked"`
	AccessKeyConfigured       bool   `json:"access_key_configured"`
	AccessKeySecretConfigured bool   `json:"access_key_secret_configured"`
	SignName                  string `json:"sign_name"`
	TemplateCode              string `json:"template_code"`
	ExpireMinutes             int    `json:"expire_minutes"`
	IntervalMinutes           int    `json:"interval_minutes"`
	UpdatedAt                 string `json:"updated_at,omitempty"`
}
