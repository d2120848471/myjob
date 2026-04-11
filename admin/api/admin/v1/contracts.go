package v1

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginSMSSendReq struct {
	LoginToken string `json:"login_token"`
}

type LoginSMSVerifyReq struct {
	LoginToken string `json:"login_token"`
	SMSCode    string `json:"sms_code"`
}

type GroupListReq struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type LogListReq struct {
	AdminID   string `json:"admin_id"`
	Keyword   string `json:"keyword"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

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
