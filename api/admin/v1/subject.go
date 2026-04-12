package v1

import "github.com/gogf/gf/v2/frame/g"

type SubjectListReq struct {
	g.Meta `path:"/subjects" method:"get" tags:"主体" summary:"主体列表" security:"BearerAuth" dc:"查询主体列表"`
}

type SubjectListRes struct {
	List []SubjectItem `json:"list" dc:"主体列表"`
}

type SubjectCreateReq struct {
	g.Meta `path:"/subjects" method:"post" tags:"主体" summary:"新增主体" security:"BearerAuth" dc:"新增主体"`
	Name   string `json:"name" dc:"主体名称"`
	HasTax int    `json:"has_tax" dc:"是否含税"`
}

type SubjectCreateRes struct {
	ID int64 `json:"id" dc:"主体ID"`
}

type SubjectUpdateReq struct {
	g.Meta `path:"/subjects/{id}" method:"put" tags:"主体" summary:"编辑主体" security:"BearerAuth" dc:"编辑主体"`
	ID     int64  `json:"id" in:"path" v:"required#主体ID不能为空" dc:"主体ID"`
	Name   string `json:"name" dc:"主体名称"`
	HasTax int    `json:"has_tax" dc:"是否含税"`
}

type SubjectUpdateRes struct{}
