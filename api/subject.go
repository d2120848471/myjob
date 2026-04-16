package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// SubjectItem 是主体信息展示项。
type SubjectItem = entity.AdminSubject

// SubjectListReq 用于查询主体列表。
type SubjectListReq struct {
	g.Meta `path:"/subjects" method:"get" tags:"主体" summary:"主体列表" security:"BearerAuth" dc:"查询主体列表"`
}

// SubjectListRes 返回主体列表。
type SubjectListRes struct {
	List []SubjectItem `json:"list" dc:"主体列表"`
}

// SubjectCreateReq 用于新增主体。
type SubjectCreateReq struct {
	g.Meta `path:"/subjects" method:"post" tags:"主体" summary:"新增主体" security:"BearerAuth" dc:"新增主体"`
	Name   string `json:"name" dc:"主体名称"`
	HasTax int    `json:"has_tax" dc:"是否含税"`
}

// SubjectCreateRes 返回新增后的主体 ID。
type SubjectCreateRes struct {
	ID int64 `json:"id" dc:"主体ID"`
}

// SubjectUpdateReq 用于编辑主体信息。
type SubjectUpdateReq struct {
	g.Meta `path:"/subjects/{id}" method:"put" tags:"主体" summary:"编辑主体" security:"BearerAuth" dc:"编辑主体"`
	ID     int64  `json:"id" in:"path" v:"required#主体ID不能为空" dc:"主体ID"`
	Name   string `json:"name" dc:"主体名称"`
	HasTax int    `json:"has_tax" dc:"是否含税"`
}

// SubjectUpdateRes 表示主体编辑成功（返回体为空）。
type SubjectUpdateRes struct{}
