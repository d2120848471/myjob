package runtime

import "time"

type Principal struct {
	UserID       int64
	GroupID      int64
	TokenVersion int
	JTI          string
}

type LoginUser struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	RealName   string `json:"real_name"`
	GroupID    int64  `json:"group_id"`
	GroupName  string `json:"group_name"`
	IsBusiness int    `json:"is_business"`
}

type SessionPayload struct {
	UserID       int64     `json:"user_id"`
	GroupID      int64     `json:"group_id"`
	TokenVersion int       `json:"token_version"`
	JTI          string    `json:"jti"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TempLoginPayload struct {
	UserID   int64  `json:"user_id"`
	IP       string `json:"ip"`
	Attempts int    `json:"attempts"`
}

type SMSCodePayload struct {
	LoginToken string `json:"login_token"`
	Code       string `json:"code"`
}

type OperationEvent struct {
	AdminID     int64
	AdminName   string
	Description string
	IP          string
	IPRegion    string
}

type SMSConfig struct {
	AccessKey       string `json:"access_key"`
	AccessKeySecret string `json:"access_key_secret"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
	ExpireMinutes   int    `json:"expire_minutes"`
	IntervalMinutes int    `json:"interval_minutes"`
}

type SMSConfigState struct {
	Version                   int       `json:"version"`
	Config                    SMSConfig `json:"config"`
	AccessKeyConfigured       bool      `json:"access_key_configured"`
	AccessKeySecretConfigured bool      `json:"access_key_secret_configured"`
	UpdatedAt                 time.Time `json:"updated_at,omitempty"`
}

type SystemConfigItem struct {
	Key        string    `json:"key"`
	Label      string    `json:"label"`
	Value      string    `json:"value"`
	ValueType  string    `json:"value_type"`
	Unit       string    `json:"unit,omitempty"`
	Required   bool      `json:"required"`
	Configured bool      `json:"configured"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

type SystemConfigGroupState struct {
	Version int                `json:"version"`
	Group   string             `json:"group"`
	Label   string             `json:"label,omitempty"`
	Items   []SystemConfigItem `json:"items"`
}

type FinanceTaxConfig struct {
	TaxExclusiveRate       string `json:"tax_exclusive_rate"`
	TaxExclusiveRateScaled int64  `json:"tax_exclusive_rate_scaled"`
	TaxInclusiveRate       string `json:"tax_inclusive_rate"`
	TaxInclusiveRateScaled int64  `json:"tax_inclusive_rate_scaled"`
}
