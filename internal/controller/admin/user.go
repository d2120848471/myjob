package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type UserController struct{ svc service.UserService }

func NewUser(svc service.UserService) *UserController { return &UserController{svc: svc} }

func (c *UserController) List(ctx context.Context, req *v1.UserListReq) (res *v1.UserListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *UserController) Trash(ctx context.Context, req *v1.UserTrashReq) (res *v1.UserTrashRes, err error) {
	return c.svc.Trash(ctx, req)
}

func (c *UserController) Create(ctx context.Context, req *v1.UserCreateReq) (res *v1.UserCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Update(ctx context.Context, req *v1.UserUpdateReq) (res *v1.UserUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Delete(ctx context.Context, req *v1.UserDeleteReq) (res *v1.UserDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Restore(ctx context.Context, req *v1.UserRestoreReq) (res *v1.UserRestoreRes, err error) {
	return c.svc.Restore(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Status(ctx context.Context, req *v1.UserStatusReq) (res *v1.UserStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) Notify(ctx context.Context, req *v1.UserNotifyReq) (res *v1.UserNotifyRes, err error) {
	return c.svc.Notify(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) AssignBusiness(ctx context.Context, req *v1.UserBusinessAssignReq) (res *v1.UserBusinessAssignRes, err error) {
	return c.svc.SetBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *UserController) CancelBusiness(ctx context.Context, req *v1.UserBusinessCancelReq) (res *v1.UserBusinessCancelRes, err error) {
	return c.svc.CancelBusiness(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
