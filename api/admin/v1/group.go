package v1

import "github.com/gogf/gf/v2/frame/g"

type GroupListReq struct {
	g.Meta   `path:"/groups" method:"get" tags:"用户组" summary:"用户组列表" security:"BearerAuth" dc:"分页查询用户组列表"`
	Page     int `json:"page" dc:"页码"`
	PageSize int `json:"page_size" dc:"每页条数"`
}

type GroupListRes struct {
	List       []GroupListItem `json:"list" dc:"用户组列表"`
	Pagination PaginationRes   `json:"pagination" dc:"分页信息"`
}

type GroupCreateReq struct {
	g.Meta      `path:"/groups" method:"post" tags:"用户组" summary:"新增用户组" security:"BearerAuth" dc:"新增用户组"`
	Name        string `json:"name" dc:"用户组名称"`
	Description string `json:"description" dc:"描述"`
}

type GroupCreateRes struct {
	ID int64 `json:"id" dc:"用户组ID"`
}

type GroupUpdateReq struct {
	g.Meta      `path:"/groups/{id}" method:"put" tags:"用户组" summary:"编辑用户组" security:"BearerAuth" dc:"编辑用户组"`
	ID          int64  `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	Name        string `json:"name" dc:"用户组名称"`
	Description string `json:"description" dc:"描述"`
}

type GroupUpdateRes struct{}

type GroupDeleteReq struct {
	g.Meta `path:"/groups/{id}" method:"delete" tags:"用户组" summary:"删除用户组" security:"BearerAuth" dc:"删除用户组"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
}

type GroupDeleteRes struct{}

type GroupStatusReq struct {
	g.Meta `path:"/groups/{id}/status" method:"patch" tags:"用户组" summary:"切换用户组状态" security:"BearerAuth" dc:"切换用户组启停状态"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	Status int   `json:"status" dc:"状态值"`
}

type GroupStatusRes struct{}

type GroupPermissionsGetReq struct {
	g.Meta `path:"/groups/{id}/permissions" method:"get" tags:"用户组" summary:"读取用户组权限" security:"BearerAuth" dc:"读取用户组权限菜单ID"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
}

type GroupPermissionsGetRes struct {
	MenuIDs []int64 `json:"menu_ids" dc:"菜单ID列表"`
}

type GroupPermissionsSaveReq struct {
	g.Meta  `path:"/groups/{id}/permissions" method:"patch" tags:"用户组" summary:"保存用户组权限" security:"BearerAuth" dc:"保存用户组权限菜单ID"`
	ID      int64   `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	MenuIDs []int64 `json:"menu_ids" dc:"菜单ID列表"`
}

type GroupPermissionsSaveRes struct{}

type MenuTreeReq struct {
	g.Meta `path:"/menus/tree" method:"get" tags:"用户组" summary:"权限菜单树" security:"BearerAuth" dc:"获取可授权菜单树"`
}

type MenuTreeRes struct {
	List []*MenuItem `json:"list" dc:"菜单树"`
}
