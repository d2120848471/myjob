package runtime

import (
	"time"
)

type APIError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

type ResponseEnvelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Principal struct {
	UserID       int64
	GroupID      int64
	TokenVersion int
	JTI          string
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
