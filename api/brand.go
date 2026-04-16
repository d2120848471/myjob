package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// BrandListItem 是品牌列表展示项（用于一级品牌分页与子品牌列表）。
type BrandListItem = entity.BrandListItem

// BrandListReq 用于分页查询一级品牌列表，支持按名称关键词过滤。
type BrandListReq struct {
	g.Meta   `path:"/brands" method:"get" tags:"品牌" summary:"品牌列表" security:"BearerAuth" dc:"分页查询一级品牌列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Name     string `json:"name" dc:"品牌名称"`
}

// BrandListRes 返回一级品牌列表与分页信息。
type BrandListRes struct {
	List       []BrandListItem `json:"list" dc:"品牌列表"`
	Pagination PaginationRes   `json:"pagination" dc:"分页信息"`
}

// BrandChildrenReq 用于查询指定品牌的直接子品牌列表。
type BrandChildrenReq struct {
	g.Meta `path:"/brands/{id}/children" method:"get" tags:"品牌" summary:"子品牌列表" security:"BearerAuth" dc:"查询指定品牌的直接子品牌列表"`
	ID     int64 `json:"id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
}

// BrandChildrenRes 返回指定品牌的子品牌列表。
type BrandChildrenRes struct {
	List []BrandListItem `json:"list" dc:"子品牌列表"`
}

// BrandCreateReq 用于新增品牌，支持新增一级/二级/三级品牌。
//
// ParentID 为 0 表示新增一级品牌；非 0 表示新增该品牌的子品牌。
type BrandCreateReq struct {
	g.Meta          `path:"/brands" method:"post" tags:"品牌" summary:"新增品牌" security:"BearerAuth" dc:"新增一级、二级或三级品牌"`
	ParentID        int64  `json:"parent_id" dc:"父品牌ID"`
	Name            string `json:"name" dc:"品牌名称"`
	Icon            string `json:"icon" dc:"品牌图标"`
	CredentialImage string `json:"credential_image" dc:"资质图片"`
	Description     string `json:"description" dc:"品牌描述"`
	IsVisible       int    `json:"is_visible" dc:"显示状态"`
}

// BrandCreateRes 返回新增后的品牌 ID。
type BrandCreateRes struct {
	ID int64 `json:"id" dc:"品牌ID"`
}

// BrandUpdateReq 用于编辑品牌信息。
type BrandUpdateReq struct {
	g.Meta          `path:"/brands/{id}" method:"put" tags:"品牌" summary:"编辑品牌" security:"BearerAuth" dc:"编辑品牌信息"`
	ID              int64  `json:"id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
	Name            string `json:"name" dc:"品牌名称"`
	Icon            string `json:"icon" dc:"品牌图标"`
	CredentialImage string `json:"credential_image" dc:"资质图片"`
	Description     string `json:"description" dc:"品牌描述"`
	IsVisible       int    `json:"is_visible" dc:"显示状态"`
}

// BrandUpdateRes 表示品牌编辑成功（返回体为空）。
type BrandUpdateRes struct{}

// BrandDeleteReq 用于删除品牌，支持删除一级/二级/三级品牌。
type BrandDeleteReq struct {
	g.Meta `path:"/brands/{id}" method:"delete" tags:"品牌" summary:"删除品牌" security:"BearerAuth" dc:"删除一级、二级或三级品牌"`
	ID     int64 `json:"id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
}

// BrandDeleteRes 表示品牌删除成功（返回体为空）。
type BrandDeleteRes struct{}

// BrandSortReq 用于同级品牌排序调整。
type BrandSortReq struct {
	g.Meta `path:"/brands/{id}/sort" method:"patch" tags:"品牌" summary:"调整品牌排序" security:"BearerAuth" dc:"同级品牌排序"`
	ID     int64  `json:"id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
	Action string `json:"action" dc:"排序动作"`
}

// BrandSortRes 表示排序调整成功（返回体为空）。
type BrandSortRes struct{}

// BrandVisibilityReq 用于切换品牌显示/隐藏状态。
type BrandVisibilityReq struct {
	g.Meta    `path:"/brands/{id}/visibility" method:"patch" tags:"品牌" summary:"切换品牌显隐" security:"BearerAuth" dc:"切换品牌显示状态"`
	ID        int64 `json:"id" in:"path" v:"required#品牌ID不能为空" dc:"品牌ID"`
	IsVisible int   `json:"is_visible" dc:"显示状态"`
}

// BrandVisibilityRes 表示显隐切换成功（返回体为空）。
type BrandVisibilityRes struct{}

// BrandUploadReq 用于上传品牌图片（icon 或资质图片），以 multipart/form-data 提交。
type BrandUploadReq struct {
	g.Meta `path:"/brands/upload" method:"post" mime:"multipart/form-data" tags:"品牌" summary:"上传品牌图片" security:"BearerAuth" dc:"上传品牌 icon 或资质图片"`
	Type   string            `json:"type" form:"type" dc:"图片用途"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required#请上传图片" dc:"图片文件"`
}

// BrandUploadRes 返回上传后图片的访问 URL 与文件信息。
type BrandUploadRes struct {
	URL      string `json:"url" dc:"访问地址"`
	FileName string `json:"file_name" dc:"文件名"`
	Size     int64  `json:"size" dc:"文件大小"`
}
