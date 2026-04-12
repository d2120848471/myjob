package api

import "github.com/gogf/gf/v2/frame/g"

type IndustryListReq struct {
	g.Meta   `path:"/industries" method:"get" tags:"行业" summary:"行业列表" security:"BearerAuth" dc:"分页查询行业列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Name     string `json:"name" dc:"行业名称"`
}

type IndustryListRes struct {
	List       []IndustryListItem `json:"list" dc:"行业列表"`
	Pagination PaginationRes      `json:"pagination" dc:"分页信息"`
}

type IndustryCreateReq struct {
	g.Meta   `path:"/industries" method:"post" tags:"行业" summary:"新增行业" security:"BearerAuth" dc:"新增行业并可关联品牌"`
	Name     string  `json:"name" dc:"行业名称"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

type IndustryCreateRes struct {
	ID int64 `json:"id" dc:"行业ID"`
}

type IndustryUpdateReq struct {
	g.Meta   `path:"/industries/{id}" method:"put" tags:"行业" summary:"编辑行业" security:"BearerAuth" dc:"编辑行业信息与关联品牌"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Name     string  `json:"name" dc:"行业名称"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

type IndustryUpdateRes struct{}

type IndustryDeleteReq struct {
	g.Meta `path:"/industries/{id}" method:"delete" tags:"行业" summary:"删除行业" security:"BearerAuth" dc:"删除空行业"`
	ID     int64 `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
}

type IndustryDeleteRes struct{}

type IndustrySortReq struct {
	g.Meta `path:"/industries/{id}/sort" method:"patch" tags:"行业" summary:"调整行业排序" security:"BearerAuth" dc:"调整行业排序"`
	ID     int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Action string `json:"action" dc:"排序动作"`
}

type IndustrySortRes struct{}

type IndustryBrandSelectorReq struct {
	g.Meta `path:"/industries/brand-selector" method:"get" tags:"行业" summary:"品牌选择器" security:"BearerAuth" dc:"查询可供行业关联的一级品牌"`
	Name   string `json:"name" dc:"品牌名称"`
}

type IndustryBrandSelectorRes struct {
	List []BrandSelectorItem `json:"list" dc:"品牌列表"`
}

type IndustryBrandListReq struct {
	g.Meta `path:"/industries/{id}/brands" method:"get" tags:"行业" summary:"行业关联品牌列表" security:"BearerAuth" dc:"查询行业已关联品牌列表"`
	ID     int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Name   string `json:"name" dc:"品牌名称"`
}

type IndustryBrandListRes struct {
	IndustryID   int64                       `json:"industry_id" dc:"行业ID"`
	IndustryName string                      `json:"industry_name" dc:"行业名称"`
	List         []IndustryBrandRelationItem `json:"list" dc:"关联品牌列表"`
}

type IndustryBrandAddReq struct {
	g.Meta   `path:"/industries/{id}/brands" method:"post" tags:"行业" summary:"添加行业关联品牌" security:"BearerAuth" dc:"给行业添加一级品牌关联"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

type IndustryBrandAddRes struct{}

type IndustryBrandDeleteReq struct {
	g.Meta   `path:"/industries/{id}/brands" method:"delete" tags:"行业" summary:"删除行业关联品牌" security:"BearerAuth" dc:"删除行业下一个或多个品牌关联"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

type IndustryBrandDeleteRes struct{}

type IndustryBrandSortReq struct {
	g.Meta  `path:"/industries/{id}/brands/{brand_id}/sort" method:"patch" tags:"行业" summary:"调整行业内品牌排序" security:"BearerAuth" dc:"调整指定行业内品牌排序"`
	ID      int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandID int64  `json:"brand_id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
	Action  string `json:"action" dc:"排序动作"`
}

type IndustryBrandSortRes struct{}
