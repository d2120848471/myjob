package v1

import "github.com/gogf/gf/v2/frame/g"

type UserListReq struct {
	g.Meta   `path:"/users" method:"get" tags:"员工" summary:"员工列表" security:"BearerAuth" dc:"分页查询员工列表"`
	Page     int `json:"page" dc:"页码"`
	PageSize int `json:"page_size" dc:"每页条数"`
}

type UserListRes struct {
	List       []UserListItem `json:"list" dc:"员工列表"`
	Pagination PaginationRes  `json:"pagination" dc:"分页信息"`
}

type UserTrashReq struct {
	g.Meta   `path:"/users/trash" method:"get" tags:"员工" summary:"员工回收站" security:"BearerAuth" dc:"分页查询已删除员工"`
	Page     int `json:"page" dc:"页码"`
	PageSize int `json:"page_size" dc:"每页条数"`
}

type UserTrashRes struct {
	List       []UserListItem `json:"list" dc:"员工列表"`
	Pagination PaginationRes  `json:"pagination" dc:"分页信息"`
}

type UserCreateReq struct {
	g.Meta          `path:"/users" method:"post" tags:"员工" summary:"新增员工" security:"BearerAuth" dc:"新增后台员工账号"`
	Username        string `json:"username" dc:"用户名"`
	ConfirmUsername string `json:"confirm_username" dc:"确认用户名"`
	Password        string `json:"password" dc:"密码"`
	ConfirmPassword string `json:"confirm_password" dc:"确认密码"`
	RealName        string `json:"real_name" dc:"姓名"`
	Phone           string `json:"phone" dc:"手机号"`
	GroupID         int64  `json:"group_id" dc:"用户组ID"`
}

type UserCreateRes struct {
	ID int64 `json:"id" dc:"员工ID"`
}

type UserUpdateReq struct {
	g.Meta          `path:"/users/{id}" method:"put" tags:"员工" summary:"编辑员工" security:"BearerAuth" dc:"编辑后台员工账号"`
	ID              int64  `json:"id" in:"path" v:"required#员工ID不能为空" dc:"员工ID"`
	Password        string `json:"password" dc:"密码"`
	ConfirmPassword string `json:"confirm_password" dc:"确认密码"`
	RealName        string `json:"real_name" dc:"姓名"`
	Phone           string `json:"phone" dc:"手机号"`
	GroupID         int64  `json:"group_id" dc:"用户组ID"`
}

type UserUpdateRes struct{}

type UserDeleteReq struct {
	g.Meta `path:"/users/{id}" method:"delete" tags:"员工" summary:"删除员工" security:"BearerAuth" dc:"将员工移入回收站"`
	ID     int64 `json:"id" in:"path" v:"required#员工ID不能为空" dc:"员工ID"`
}

type UserDeleteRes struct{}

type UserRestoreReq struct {
	g.Meta `path:"/users/{id}/restore" method:"patch" tags:"员工" summary:"恢复员工" security:"BearerAuth" dc:"从回收站恢复员工"`
	ID     int64 `json:"id" in:"path" v:"required#员工ID不能为空" dc:"员工ID"`
}

type UserRestoreRes struct{}

type UserStatusReq struct {
	g.Meta `path:"/users/{id}/status" method:"patch" tags:"员工" summary:"切换员工状态" security:"BearerAuth" dc:"启用或禁用员工"`
	ID     int64 `json:"id" in:"path" v:"required#员工ID不能为空" dc:"员工ID"`
	Status int   `json:"status" dc:"状态值"`
}

type UserStatusRes struct{}

type UserNotifyReq struct {
	g.Meta        `path:"/users/{id}/notify" method:"patch" tags:"员工" summary:"切换余额通知" security:"BearerAuth" dc:"切换员工余额通知开关"`
	ID            int64 `json:"id" in:"path" v:"required#员工ID不能为空" dc:"员工ID"`
	BalanceNotify int   `json:"balance_notify" dc:"余额通知开关"`
}

type UserNotifyRes struct{}

type UserBusinessAssignReq struct {
	g.Meta `path:"/users/business" method:"post" tags:"员工" summary:"批量设置商务" security:"BearerAuth" dc:"批量设置商务员工"`
	IDs    []int64 `json:"ids" dc:"员工ID列表"`
}

type UserBusinessAssignRes struct{}

type UserBusinessCancelReq struct {
	g.Meta `path:"/users/business" method:"delete" tags:"员工" summary:"批量取消商务" security:"BearerAuth" dc:"批量取消商务员工"`
	IDs    []int64 `json:"ids" dc:"员工ID列表"`
}

type UserBusinessCancelRes struct{}
