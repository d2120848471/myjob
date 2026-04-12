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
