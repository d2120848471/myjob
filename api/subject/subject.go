package subject

type AddReq struct {
	Name   string `json:"name"`
	HasTax int    `json:"has_tax"`
}

type EditReq struct {
	Name   string `json:"name"`
	HasTax int    `json:"has_tax"`
}
