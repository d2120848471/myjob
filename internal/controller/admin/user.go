package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type UserController struct{ svc service.UserService }

func NewUser(svc service.UserService) *UserController { return &UserController{svc: svc} }

func (c *UserController) List(ctx context.Context, req *adminapi.UserListReq) (res *adminapi.UserListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *UserController) Trash(ctx context.Context, req *adminapi.UserTrashReq) (res *adminapi.UserTrashRes, err error) {
	return c.svc.Trash(ctx, req)
}

func (c *UserController) Create(ctx context.Context, req *adminapi.UserCreateReq) (res *adminapi.UserCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Update(ctx context.Context, req *adminapi.UserUpdateReq) (res *adminapi.UserUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Delete(ctx context.Context, req *adminapi.UserDeleteReq) (res *adminapi.UserDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Restore(ctx context.Context, req *adminapi.UserRestoreReq) (res *adminapi.UserRestoreRes, err error) {
	return c.svc.Restore(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Status(ctx context.Context, req *adminapi.UserStatusReq) (res *adminapi.UserStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Notify(ctx context.Context, req *adminapi.UserNotifyReq) (res *adminapi.UserNotifyRes, err error) {
	return c.svc.Notify(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) AssignBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq) (res *adminapi.UserBusinessAssignRes, err error) {
	return c.svc.SetBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq) (res *adminapi.UserBusinessCancelRes, err error) {
	return c.svc.CancelBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
