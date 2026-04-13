package api

import "github.com/gogf/gf/v2/frame/g"

type ProductTemplateListReq struct {
	g.Meta   `path:"/product-templates" method:"get" tags:"商品模板" summary:"商品模板列表" security:"BearerAuth" dc:"分页查询本地商品模板列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Keyword  string `json:"keyword" dc:"关键词"`
	Type     string `json:"type" dc:"模板类型"`
	IsShared string `json:"is_shared" dc:"共享状态"`
}

type ProductTemplateListRes struct {
	List       []ProductTemplateListItem `json:"list" dc:"模板列表"`
	Pagination PaginationRes             `json:"pagination" dc:"分页信息"`
}

type ProductTemplateCreateReq struct {
	g.Meta       `path:"/product-templates" method:"post" tags:"商品模板" summary:"新增商品模板" security:"BearerAuth" dc:"新增本地商品模板"`
	Title        string `json:"title" dc:"模板名称"`
	Type         string `json:"type" dc:"模板类型"`
	IsShared     int    `json:"is_shared" dc:"是否共享"`
	AccountName  string `json:"account_name" dc:"充值账号名称"`
	ValidateType int    `json:"validate_type" dc:"验证类型"`
}

type ProductTemplateCreateRes struct {
	ID int64 `json:"id" dc:"模板ID"`
}

type ProductTemplateUpdateReq struct {
	g.Meta       `path:"/product-templates/{id}" method:"put" tags:"商品模板" summary:"编辑商品模板" security:"BearerAuth" dc:"编辑本地商品模板"`
	ID           int64  `json:"id" in:"path" v:"required#模板ID不能为空" dc:"模板ID"`
	Title        string `json:"title" dc:"模板名称"`
	Type         string `json:"type" dc:"模板类型"`
	IsShared     int    `json:"is_shared" dc:"是否共享"`
	AccountName  string `json:"account_name" dc:"充值账号名称"`
	ValidateType int    `json:"validate_type" dc:"验证类型"`
}

type ProductTemplateUpdateRes struct{}

type ProductTemplateDeleteReq struct {
	g.Meta `path:"/product-templates/{id}" method:"delete" tags:"商品模板" summary:"删除商品模板" security:"BearerAuth" dc:"删除单个商品模板"`
	ID     int64 `json:"id" in:"path" v:"required#模板ID不能为空" dc:"模板ID"`
}

type ProductTemplateDeleteRes struct{}

type ProductTemplateBatchDeleteReq struct {
	g.Meta `path:"/product-templates" method:"delete" tags:"商品模板" summary:"批量删除商品模板" security:"BearerAuth" dc:"批量删除本地商品模板"`
	IDs    []int64 `json:"ids" dc:"模板ID列表"`
}

type ProductTemplateBatchDeleteRes struct{}

type ProductTemplateValidateTypeListReq struct {
	g.Meta `path:"/product-templates/validate-types" method:"get" tags:"商品模板" summary:"验证方式枚举" security:"BearerAuth" dc:"获取商品模板支持的验证方式"`
}

type ProductTemplateValidateTypeListRes struct {
	List []ProductTemplateValidateTypeItem `json:"list" dc:"验证方式列表"`
}
