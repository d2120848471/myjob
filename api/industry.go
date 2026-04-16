package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// IndustryListItem 是行业列表展示项。
type IndustryListItem = entity.IndustryListItem

// IndustryBrandRelationItem 是行业与品牌关联关系的展示项。
type IndustryBrandRelationItem = entity.IndustryBrandRelationItem

// BrandSelectorItem 是品牌选择器的展示项（当前仅返回一级品牌）。
type BrandSelectorItem = entity.BrandSelectorItem

// IndustryListReq 用于分页查询行业列表，支持按名称关键词过滤。
type IndustryListReq struct {
	g.Meta   `path:"/industries" method:"get" tags:"行业" summary:"行业列表" security:"BearerAuth" dc:"分页查询行业列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Name     string `json:"name" dc:"行业名称"`
}

// IndustryListRes 返回行业列表与分页信息。
type IndustryListRes struct {
	List       []IndustryListItem `json:"list" dc:"行业列表"`
	Pagination PaginationRes      `json:"pagination" dc:"分页信息"`
}

// IndustryCreateReq 用于新增行业，并可同时关联一级品牌。
type IndustryCreateReq struct {
	g.Meta   `path:"/industries" method:"post" tags:"行业" summary:"新增行业" security:"BearerAuth" dc:"新增行业并可关联品牌"`
	Name     string  `json:"name" dc:"行业名称"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

// IndustryCreateRes 返回新增后的行业 ID。
type IndustryCreateRes struct {
	ID int64 `json:"id" dc:"行业ID"`
}

// IndustryUpdateReq 用于编辑行业信息与关联品牌。
type IndustryUpdateReq struct {
	g.Meta   `path:"/industries/{id}" method:"put" tags:"行业" summary:"编辑行业" security:"BearerAuth" dc:"编辑行业信息与关联品牌"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Name     string  `json:"name" dc:"行业名称"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

// IndustryUpdateRes 表示行业编辑成功（返回体为空）。
type IndustryUpdateRes struct{}

// IndustryDeleteReq 用于删除行业（仅允许删除空行业）。
type IndustryDeleteReq struct {
	g.Meta `path:"/industries/{id}" method:"delete" tags:"行业" summary:"删除行业" security:"BearerAuth" dc:"删除空行业"`
	ID     int64 `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
}

// IndustryDeleteRes 表示行业删除成功（返回体为空）。
type IndustryDeleteRes struct{}

// IndustrySortReq 用于调整行业排序。
type IndustrySortReq struct {
	g.Meta `path:"/industries/{id}/sort" method:"patch" tags:"行业" summary:"调整行业排序" security:"BearerAuth" dc:"调整行业排序"`
	ID     int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Action string `json:"action" dc:"排序动作"`
}

// IndustrySortRes 表示排序调整成功（返回体为空）。
type IndustrySortRes struct{}

// IndustryBrandSelectorReq 用于查询可供行业关联的品牌选择器数据（当前仅返回一级品牌）。
type IndustryBrandSelectorReq struct {
	g.Meta `path:"/industries/brand-selector" method:"get" tags:"行业" summary:"品牌选择器" security:"BearerAuth" dc:"查询可供行业关联的一级品牌"`
	Name   string `json:"name" dc:"品牌名称"`
}

// IndustryBrandSelectorRes 返回品牌选择器列表。
type IndustryBrandSelectorRes struct {
	List []BrandSelectorItem `json:"list" dc:"品牌列表"`
}

// IndustryBrandListReq 用于查询行业已关联品牌列表（支持按品牌名称过滤）。
type IndustryBrandListReq struct {
	g.Meta `path:"/industries/{id}/brands" method:"get" tags:"行业" summary:"行业关联品牌列表" security:"BearerAuth" dc:"查询行业已关联品牌列表"`
	ID     int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	Name   string `json:"name" dc:"品牌名称"`
}

// IndustryBrandListRes 返回行业信息与关联品牌列表。
type IndustryBrandListRes struct {
	IndustryID   int64                       `json:"industry_id" dc:"行业ID"`
	IndustryName string                      `json:"industry_name" dc:"行业名称"`
	List         []IndustryBrandRelationItem `json:"list" dc:"关联品牌列表"`
}

// IndustryBrandAddReq 用于给行业添加一级品牌关联。
type IndustryBrandAddReq struct {
	g.Meta   `path:"/industries/{id}/brands" method:"post" tags:"行业" summary:"添加行业关联品牌" security:"BearerAuth" dc:"给行业添加一级品牌关联"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

// IndustryBrandAddRes 表示关联添加成功（返回体为空）。
type IndustryBrandAddRes struct{}

// IndustryBrandDeleteReq 用于删除行业下一个或多个品牌关联。
type IndustryBrandDeleteReq struct {
	g.Meta   `path:"/industries/{id}/brands" method:"delete" tags:"行业" summary:"删除行业关联品牌" security:"BearerAuth" dc:"删除行业下一个或多个品牌关联"`
	ID       int64   `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandIDs []int64 `json:"brand_ids" dc:"品牌ID列表"`
}

// IndustryBrandDeleteRes 表示关联删除成功（返回体为空）。
type IndustryBrandDeleteRes struct{}

// IndustryBrandSortReq 用于调整指定行业内的品牌排序。
type IndustryBrandSortReq struct {
	g.Meta  `path:"/industries/{id}/brands/{brand_id}/sort" method:"patch" tags:"行业" summary:"调整行业内品牌排序" security:"BearerAuth" dc:"调整指定行业内品牌排序"`
	ID      int64  `json:"id" in:"path" v:"required#行业ID不能为空" dc:"行业ID"`
	BrandID int64  `json:"brand_id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
	Action  string `json:"action" dc:"排序动作"`
}

// IndustryBrandSortRes 表示关联品牌排序调整成功（返回体为空）。
type IndustryBrandSortRes struct{}
