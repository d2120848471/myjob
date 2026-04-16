package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// GroupListItem 是用户组列表展示项。
type GroupListItem = entity.GroupListItem

// MenuItem 是菜单树节点，用于用户组授权与菜单树返回。
type MenuItem = entity.AdminMenu

// GroupListReq 用于分页查询用户组列表。
type GroupListReq struct {
	g.Meta   `path:"/groups" method:"get" tags:"用户组" summary:"用户组列表" security:"BearerAuth" dc:"分页查询用户组列表"`
	Page     int `json:"page" dc:"页码"`
	PageSize int `json:"page_size" dc:"每页条数"`
}

// GroupListRes 返回用户组列表与分页信息。
type GroupListRes struct {
	List       []GroupListItem `json:"list" dc:"用户组列表"`
	Pagination PaginationRes   `json:"pagination" dc:"分页信息"`
}

// GroupCreateReq 用于新增用户组。
type GroupCreateReq struct {
	g.Meta      `path:"/groups" method:"post" tags:"用户组" summary:"新增用户组" security:"BearerAuth" dc:"新增用户组"`
	Name        string `json:"name" dc:"用户组名称"`
	Description string `json:"description" dc:"描述"`
}

// GroupCreateRes 返回新增后的用户组 ID。
type GroupCreateRes struct {
	ID int64 `json:"id" dc:"用户组ID"`
}

// GroupUpdateReq 用于编辑用户组基础信息。
type GroupUpdateReq struct {
	g.Meta      `path:"/groups/{id}" method:"put" tags:"用户组" summary:"编辑用户组" security:"BearerAuth" dc:"编辑用户组"`
	ID          int64  `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	Name        string `json:"name" dc:"用户组名称"`
	Description string `json:"description" dc:"描述"`
}

// GroupUpdateRes 表示用户组编辑成功（返回体为空）。
type GroupUpdateRes struct{}

// GroupDeleteReq 用于删除用户组。
type GroupDeleteReq struct {
	g.Meta `path:"/groups/{id}" method:"delete" tags:"用户组" summary:"删除用户组" security:"BearerAuth" dc:"删除用户组"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
}

// GroupDeleteRes 表示用户组删除成功（返回体为空）。
type GroupDeleteRes struct{}

// GroupStatusReq 用于切换用户组启停状态。
type GroupStatusReq struct {
	g.Meta `path:"/groups/{id}/status" method:"patch" tags:"用户组" summary:"切换用户组状态" security:"BearerAuth" dc:"切换用户组启停状态"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	Status int   `json:"status" dc:"状态值"`
}

// GroupStatusRes 表示用户组状态切换成功（返回体为空）。
type GroupStatusRes struct{}

// GroupPermissionsGetReq 用于读取用户组已授权的菜单 ID 列表。
type GroupPermissionsGetReq struct {
	g.Meta `path:"/groups/{id}/permissions" method:"get" tags:"用户组" summary:"读取用户组权限" security:"BearerAuth" dc:"读取用户组权限菜单ID"`
	ID     int64 `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
}

// GroupPermissionsGetRes 返回用户组已授权的菜单 ID 列表。
type GroupPermissionsGetRes struct {
	MenuIDs []int64 `json:"menu_ids" dc:"菜单ID列表"`
}

// GroupPermissionsSaveReq 用于保存用户组菜单授权（以菜单 ID 列表的形式提交）。
type GroupPermissionsSaveReq struct {
	g.Meta  `path:"/groups/{id}/permissions" method:"patch" tags:"用户组" summary:"保存用户组权限" security:"BearerAuth" dc:"保存用户组权限菜单ID"`
	ID      int64   `json:"id" in:"path" v:"required#用户组ID不能为空" dc:"用户组ID"`
	MenuIDs []int64 `json:"menu_ids" dc:"菜单ID列表"`
}

// GroupPermissionsSaveRes 表示用户组菜单授权保存成功（返回体为空）。
type GroupPermissionsSaveRes struct{}

// MenuTreeReq 用于读取可授权的菜单树（用于用户组授权与回显）。
type MenuTreeReq struct {
	g.Meta `path:"/menus/tree" method:"get" tags:"用户组" summary:"权限菜单树" security:"BearerAuth" dc:"获取可授权菜单树"`
}

// MenuTreeRes 返回可授权菜单树结构。
type MenuTreeRes struct {
	List []*MenuItem `json:"list" dc:"菜单树"`
}
