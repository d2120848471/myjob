package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// GroupController 提供用户组与菜单授权相关 HTTP handler。
type GroupController struct{ svc service.GroupService }

// NewGroup 创建 GroupController。
func NewGroup(svc service.GroupService) *GroupController { return &GroupController{svc: svc} }

// List 返回用户组分页列表。
func (c *GroupController) List(ctx context.Context, req *adminapi.GroupListReq) (res *adminapi.GroupListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增用户组，并记录操作人与客户端 IP。
func (c *GroupController) Create(ctx context.Context, req *adminapi.GroupCreateReq) (res *adminapi.GroupCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑用户组，并记录操作人与客户端 IP。
func (c *GroupController) Update(ctx context.Context, req *adminapi.GroupUpdateReq) (res *adminapi.GroupUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除用户组，并记录操作人与客户端 IP。
func (c *GroupController) Delete(ctx context.Context, req *adminapi.GroupDeleteReq) (res *adminapi.GroupDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Status 切换用户组启停状态，并记录操作人与客户端 IP。
func (c *GroupController) Status(ctx context.Context, req *adminapi.GroupStatusReq) (res *adminapi.GroupStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// GetPermissions 读取用户组已授权菜单 ID 列表。
func (c *GroupController) GetPermissions(ctx context.Context, req *adminapi.GroupPermissionsGetReq) (res *adminapi.GroupPermissionsGetRes, err error) {
	return c.svc.AuthGet(ctx, req)
}

// SavePermissions 保存用户组菜单授权，并记录操作人与客户端 IP。
func (c *GroupController) SavePermissions(ctx context.Context, req *adminapi.GroupPermissionsSaveReq) (res *adminapi.GroupPermissionsSaveRes, err error) {
	return c.svc.AuthSave(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// MenuTree 返回可授权的菜单树。
func (c *GroupController) MenuTree(ctx context.Context, req *adminapi.MenuTreeReq) (res *adminapi.MenuTreeRes, err error) {
	return c.svc.MenuTree(ctx, req)
}
