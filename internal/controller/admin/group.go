package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type GroupController struct{ svc service.GroupService }

func NewGroup(svc service.GroupService) *GroupController { return &GroupController{svc: svc} }

func (c *GroupController) List(ctx context.Context, req *adminapi.GroupListReq) (res *adminapi.GroupListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *GroupController) Create(ctx context.Context, req *adminapi.GroupCreateReq) (res *adminapi.GroupCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Update(ctx context.Context, req *adminapi.GroupUpdateReq) (res *adminapi.GroupUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Delete(ctx context.Context, req *adminapi.GroupDeleteReq) (res *adminapi.GroupDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Status(ctx context.Context, req *adminapi.GroupStatusReq) (res *adminapi.GroupStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) GetPermissions(ctx context.Context, req *adminapi.GroupPermissionsGetReq) (res *adminapi.GroupPermissionsGetRes, err error) {
	return c.svc.AuthGet(ctx, req)
}

func (c *GroupController) SavePermissions(ctx context.Context, req *adminapi.GroupPermissionsSaveReq) (res *adminapi.GroupPermissionsSaveRes, err error) {
	return c.svc.AuthSave(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) MenuTree(ctx context.Context, req *adminapi.MenuTreeReq) (res *adminapi.MenuTreeRes, err error) {
	return c.svc.MenuTree(ctx, req)
}
