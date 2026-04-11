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
