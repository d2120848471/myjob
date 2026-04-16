package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// UserController 提供员工管理相关 HTTP handler。
type UserController struct{ svc service.UserService }

// NewUser 创建 UserController。
func NewUser(svc service.UserService) *UserController { return &UserController{svc: svc} }

// List 返回员工分页列表。
func (c *UserController) List(ctx context.Context, req *adminapi.UserListReq) (res *adminapi.UserListRes, err error) {
	return c.svc.List(ctx, req)
}

// Trash 返回员工回收站分页列表。
func (c *UserController) Trash(ctx context.Context, req *adminapi.UserTrashReq) (res *adminapi.UserTrashRes, err error) {
	return c.svc.Trash(ctx, req)
}

// Create 新增员工账号，并记录操作人与客户端 IP。
func (c *UserController) Create(ctx context.Context, req *adminapi.UserCreateReq) (res *adminapi.UserCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑员工账号，并记录操作人与客户端 IP。
func (c *UserController) Update(ctx context.Context, req *adminapi.UserUpdateReq) (res *adminapi.UserUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除员工（移入回收站），并记录操作人与客户端 IP。
func (c *UserController) Delete(ctx context.Context, req *adminapi.UserDeleteReq) (res *adminapi.UserDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Restore 从回收站恢复员工，并记录操作人与客户端 IP。
func (c *UserController) Restore(ctx context.Context, req *adminapi.UserRestoreReq) (res *adminapi.UserRestoreRes, err error) {
	return c.svc.Restore(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Status 启用/禁用员工，并记录操作人与客户端 IP。
func (c *UserController) Status(ctx context.Context, req *adminapi.UserStatusReq) (res *adminapi.UserStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Notify 切换员工余额通知开关，并记录操作人与客户端 IP。
func (c *UserController) Notify(ctx context.Context, req *adminapi.UserNotifyReq) (res *adminapi.UserNotifyRes, err error) {
	return c.svc.Notify(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// AssignBusiness 批量设置商务员工，并记录操作人与客户端 IP。
func (c *UserController) AssignBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq) (res *adminapi.UserBusinessAssignRes, err error) {
	return c.svc.SetBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// CancelBusiness 批量取消商务员工标记，并记录操作人与客户端 IP。
func (c *UserController) CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq) (res *adminapi.UserBusinessCancelRes, err error) {
	return c.svc.CancelBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
