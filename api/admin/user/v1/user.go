package v1

type ListReq struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type AddReq struct {
	Username        string `json:"username"`
	ConfirmUsername string `json:"confirm_username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	RealName        string `json:"real_name"`
	Phone           string `json:"phone"`
	GroupID         int64  `json:"group_id"`
}

type EditReq struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	RealName        string `json:"real_name"`
	Phone           string `json:"phone"`
	GroupID         int64  `json:"group_id"`
}

type StatusReq struct {
	Status int `json:"status"`
}

type NotifyReq struct {
	BalanceNotify int `json:"balance_notify"`
}

type BusinessReq struct {
	IDs []int64 `json:"ids"`
}
