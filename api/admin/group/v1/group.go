package v1

type ListReq struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type AddReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EditReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StatusReq struct {
	Status int `json:"status"`
}

type AuthSaveReq struct {
	MenuIDs []int64 `json:"menu_ids"`
}
