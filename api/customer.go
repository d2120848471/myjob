package api

import (
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// CustomerListItem 是后台客户列表和回收站列表展示项。
type CustomerListItem = entity.CustomerListItem

// CustomerListReq 用于分页查询未删除客户列表。
type CustomerListReq struct {
	g.Meta   `path:"/customers" method:"get" tags:"客户管理" summary:"客户列表" security:"BearerAuth" dc:"分页查询客户列表"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Keyword  string `json:"keyword" dc:"公司名或手机号关键字"`
	Status   *int   `json:"status" dc:"状态筛选"`
}

// CustomerListRes 返回客户列表与分页信息。
type CustomerListRes struct {
	List       []CustomerListItem `json:"list" dc:"客户列表"`
	Pagination PaginationRes      `json:"pagination" dc:"分页信息"`
}

// CustomerTrashReq 用于分页查询客户回收站。
type CustomerTrashReq struct {
	g.Meta   `path:"/customers/trash" method:"get" tags:"客户管理" summary:"客户回收站" security:"BearerAuth" dc:"分页查询已删除客户"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Keyword  string `json:"keyword" dc:"公司名或手机号关键字"`
}

// CustomerTrashRes 返回回收站客户列表与分页信息。
type CustomerTrashRes struct {
	List       []CustomerListItem `json:"list" dc:"客户列表"`
	Pagination PaginationRes      `json:"pagination" dc:"分页信息"`
}

// CustomerDetailReq 用于读取客户详情。
type CustomerDetailReq struct {
	g.Meta `path:"/customers/{id}" method:"get" tags:"客户管理" summary:"客户详情" security:"BearerAuth" dc:"读取客户详情"`
	ID     int64 `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
}

// CustomerDetailRes 返回客户详情。
type CustomerDetailRes struct {
	ID          int64  `json:"id" dc:"客户ID"`
	CompanyName string `json:"company_name" dc:"店铺或公司名称"`
	Phone       string `json:"phone" dc:"手机号"`
	Status      int    `json:"status" dc:"状态"`
	LastLoginIP string `json:"last_login_ip" dc:"最后登录IP"`
	LastLoginAt string `json:"last_login_at" dc:"最后登录时间"`
	CreatedAt   string `json:"created_at" dc:"创建时间"`
	UpdatedAt   string `json:"updated_at" dc:"更新时间"`
}

// CustomerCreateReq 用于后台新增客户账号。
type CustomerCreateReq struct {
	g.Meta             `path:"/customers" method:"post" tags:"客户管理" summary:"新增客户" security:"BearerAuth" dc:"后台新增客户账号"`
	CompanyName        string `json:"company_name" dc:"店铺或公司名称"`
	Phone              string `json:"phone" dc:"手机号"`
	Password           string `json:"password" dc:"登录密码"`
	ConfirmPassword    string `json:"confirm_password" dc:"确认登录密码"`
	PayPassword        string `json:"pay_password" dc:"支付密码"`
	ConfirmPayPassword string `json:"confirm_pay_password" dc:"确认支付密码"`
	Status             int    `json:"status" dc:"状态"`
}

// CustomerCreateRes 返回新增后的客户 ID。
type CustomerCreateRes struct {
	ID int64 `json:"id" dc:"客户ID"`
}

// CustomerUpdateReq 用于编辑客户基础资料。
type CustomerUpdateReq struct {
	g.Meta      `path:"/customers/{id}" method:"put" tags:"客户管理" summary:"编辑客户" security:"BearerAuth" dc:"编辑客户基础资料"`
	ID          int64  `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
	CompanyName string `json:"company_name" dc:"店铺或公司名称"`
	Phone       string `json:"phone" dc:"手机号"`
	Status      int    `json:"status" dc:"状态"`
}

// CustomerUpdateRes 表示客户编辑成功（返回体为空）。
type CustomerUpdateRes struct{}

// CustomerStatusReq 用于启用或禁用客户账号。
type CustomerStatusReq struct {
	g.Meta `path:"/customers/{id}/status" method:"patch" tags:"客户管理" summary:"切换客户状态" security:"BearerAuth" dc:"启用或禁用客户"`
	ID     int64 `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
	Status int   `json:"status" dc:"状态"`
}

// CustomerStatusRes 表示客户状态切换成功（返回体为空）。
type CustomerStatusRes struct{}

// CustomerDeleteReq 用于软删除客户。
type CustomerDeleteReq struct {
	g.Meta `path:"/customers/{id}" method:"delete" tags:"客户管理" summary:"删除客户" security:"BearerAuth" dc:"将客户移入回收站"`
	ID     int64 `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
}

// CustomerDeleteRes 表示客户删除成功（返回体为空）。
type CustomerDeleteRes struct{}

// CustomerRestoreReq 用于从回收站恢复客户。
type CustomerRestoreReq struct {
	g.Meta `path:"/customers/{id}/restore" method:"patch" tags:"客户管理" summary:"恢复客户" security:"BearerAuth" dc:"从回收站恢复客户"`
	ID     int64 `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
}

// CustomerRestoreRes 表示客户恢复成功（返回体为空）。
type CustomerRestoreRes struct{}

// CustomerPasswordResetReq 用于后台重置客户登录密码。
type CustomerPasswordResetReq struct {
	g.Meta          `path:"/customers/{id}/password" method:"patch" tags:"客户管理" summary:"重置客户登录密码" security:"BearerAuth" dc:"后台重置客户登录密码"`
	ID              int64  `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
	Password        string `json:"password" dc:"新登录密码"`
	ConfirmPassword string `json:"confirm_password" dc:"确认新登录密码"`
}

// CustomerPasswordResetRes 表示登录密码重置成功（返回体为空）。
type CustomerPasswordResetRes struct{}

// CustomerPayPasswordResetReq 用于后台重置客户支付密码。
type CustomerPayPasswordResetReq struct {
	g.Meta             `path:"/customers/{id}/pay-password" method:"patch" tags:"客户管理" summary:"重置客户支付密码" security:"BearerAuth" dc:"后台重置客户支付密码"`
	ID                 int64  `json:"id" in:"path" v:"required#客户ID不能为空" dc:"客户ID"`
	PayPassword        string `json:"pay_password" dc:"新支付密码"`
	ConfirmPayPassword string `json:"confirm_pay_password" dc:"确认支付密码"`
}

// CustomerPayPasswordResetRes 表示支付密码重置成功（返回体为空）。
type CustomerPayPasswordResetRes struct{}
