package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type GroupController struct{ svc service.GroupService }

func NewGroup(svc service.GroupService) *GroupController { return &GroupController{svc: svc} }

func (c *GroupController) List(ctx context.Context, req *v1.GroupListReq) (res *v1.GroupListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *GroupController) Create(ctx context.Context, req *v1.GroupCreateReq) (res *v1.GroupCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Update(ctx context.Context, req *v1.GroupUpdateReq) (res *v1.GroupUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Delete(ctx context.Context, req *v1.GroupDeleteReq) (res *v1.GroupDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) Status(ctx context.Context, req *v1.GroupStatusReq) (res *v1.GroupStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) GetPermissions(ctx context.Context, req *v1.GroupPermissionsGetReq) (res *v1.GroupPermissionsGetRes, err error) {
	return c.svc.AuthGet(ctx, req)
}

func (c *GroupController) SavePermissions(ctx context.Context, req *v1.GroupPermissionsSaveReq) (res *v1.GroupPermissionsSaveRes, err error) {
	return c.svc.AuthSave(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *GroupController) MenuTree(ctx context.Context, req *v1.MenuTreeReq) (res *v1.MenuTreeRes, err error) {
	return c.svc.MenuTree(ctx, req)
}
