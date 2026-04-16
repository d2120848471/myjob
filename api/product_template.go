package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ProductTemplateListItem 是商品模板列表展示项。
type ProductTemplateListItem = entity.ProductTemplateListItem

// ProductTemplateValidateTypeItem 是商品模板支持的验证方式枚举项。
type ProductTemplateValidateTypeItem = entity.ProductTemplateValidateTypeItem

// ProductTemplateListReq 用于分页查询商品模板列表，支持关键词与筛选条件。
type ProductTemplateListReq struct {
	g.Meta   `path:"/product-templates" method:"get" tags:"商品模板" summary:"商品模板列表" security:"BearerAuth" dc:"分页查询本地商品模板列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Keyword  string `json:"keyword" dc:"关键词"`
	Type     string `json:"type" dc:"模板类型"`
	IsShared string `json:"is_shared" dc:"共享状态"`
}

// ProductTemplateListRes 返回模板列表与分页信息。
type ProductTemplateListRes struct {
	List       []ProductTemplateListItem `json:"list" dc:"模板列表"`
	Pagination PaginationRes             `json:"pagination" dc:"分页信息"`
}

// ProductTemplateCreateReq 用于新增本地商品模板。
type ProductTemplateCreateReq struct {
	g.Meta       `path:"/product-templates" method:"post" tags:"商品模板" summary:"新增商品模板" security:"BearerAuth" dc:"新增本地商品模板"`
	Title        string `json:"title" dc:"模板名称"`
	Type         string `json:"type" dc:"模板类型"`
	IsShared     int    `json:"is_shared" dc:"是否共享"`
	AccountName  string `json:"account_name" dc:"充值账号名称"`
	ValidateType int    `json:"validate_type" dc:"验证类型"`
}

// ProductTemplateCreateRes 返回新增后的模板 ID。
type ProductTemplateCreateRes struct {
	ID int64 `json:"id" dc:"模板ID"`
}

// ProductTemplateUpdateReq 用于编辑本地商品模板。
type ProductTemplateUpdateReq struct {
	g.Meta       `path:"/product-templates/{id}" method:"put" tags:"商品模板" summary:"编辑商品模板" security:"BearerAuth" dc:"编辑本地商品模板"`
	ID           int64  `json:"id" in:"path" v:"required#模板ID不能为空" dc:"模板ID"`
	Title        string `json:"title" dc:"模板名称"`
	Type         string `json:"type" dc:"模板类型"`
	IsShared     int    `json:"is_shared" dc:"是否共享"`
	AccountName  string `json:"account_name" dc:"充值账号名称"`
	ValidateType int    `json:"validate_type" dc:"验证类型"`
}

// ProductTemplateUpdateRes 表示模板编辑成功（返回体为空）。
type ProductTemplateUpdateRes struct{}

// ProductTemplateDeleteReq 用于删除单个商品模板。
type ProductTemplateDeleteReq struct {
	g.Meta `path:"/product-templates/{id}" method:"delete" tags:"商品模板" summary:"删除商品模板" security:"BearerAuth" dc:"删除单个商品模板"`
	ID     int64 `json:"id" in:"path" v:"required#模板ID不能为空" dc:"模板ID"`
}

// ProductTemplateDeleteRes 表示模板删除成功（返回体为空）。
type ProductTemplateDeleteRes struct{}

// ProductTemplateBatchDeleteReq 用于批量删除商品模板。
type ProductTemplateBatchDeleteReq struct {
	g.Meta `path:"/product-templates" method:"delete" tags:"商品模板" summary:"批量删除商品模板" security:"BearerAuth" dc:"批量删除本地商品模板"`
	IDs    []int64 `json:"ids" dc:"模板ID列表"`
}

// ProductTemplateBatchDeleteRes 表示批量删除成功（返回体为空）。
type ProductTemplateBatchDeleteRes struct{}

// ProductTemplateValidateTypeListReq 用于读取商品模板支持的验证方式枚举列表。
type ProductTemplateValidateTypeListReq struct {
	g.Meta `path:"/product-templates/validate-types" method:"get" tags:"商品模板" summary:"验证方式枚举" security:"BearerAuth" dc:"获取商品模板支持的验证方式"`
}

// ProductTemplateValidateTypeListRes 返回验证方式枚举列表。
type ProductTemplateValidateTypeListRes struct {
	List []ProductTemplateValidateTypeItem `json:"list" dc:"验证方式列表"`
}
