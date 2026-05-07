package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// CustomerController 提供后台客户管理相关 HTTP handler。
type CustomerController struct{ svc service.CustomerService }

// NewCustomer 创建 CustomerController。
func NewCustomer(svc service.CustomerService) *CustomerController {
	return &CustomerController{svc: svc}
}

// List 返回未删除客户分页列表。
func (c *CustomerController) List(ctx context.Context, req *adminapi.CustomerListReq) (res *adminapi.CustomerListRes, err error) {
	return c.svc.List(ctx, req)
}

// Trash 返回客户回收站分页列表。
func (c *CustomerController) Trash(ctx context.Context, req *adminapi.CustomerTrashReq) (res *adminapi.CustomerTrashRes, err error) {
	return c.svc.Trash(ctx, req)
}

// Detail 返回客户详情。
func (c *CustomerController) Detail(ctx context.Context, req *adminapi.CustomerDetailReq) (res *adminapi.CustomerDetailRes, err error) {
	return c.svc.Detail(ctx, req)
}

// Create 新增客户，并记录操作人与客户端 IP。
func (c *CustomerController) Create(ctx context.Context, req *adminapi.CustomerCreateReq) (res *adminapi.CustomerCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑客户基础资料，并记录操作人与客户端 IP。
func (c *CustomerController) Update(ctx context.Context, req *adminapi.CustomerUpdateReq) (res *adminapi.CustomerUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Status 启用或禁用客户，并记录操作人与客户端 IP。
func (c *CustomerController) Status(ctx context.Context, req *adminapi.CustomerStatusReq) (res *adminapi.CustomerStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 将客户移入回收站，并记录操作人与客户端 IP。
func (c *CustomerController) Delete(ctx context.Context, req *adminapi.CustomerDeleteReq) (res *adminapi.CustomerDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Restore 从回收站恢复客户，并记录操作人与客户端 IP。
func (c *CustomerController) Restore(ctx context.Context, req *adminapi.CustomerRestoreReq) (res *adminapi.CustomerRestoreRes, err error) {
	return c.svc.Restore(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// ResetPassword 重置客户登录密码，并记录操作人与客户端 IP。
func (c *CustomerController) ResetPassword(ctx context.Context, req *adminapi.CustomerPasswordResetReq) (res *adminapi.CustomerPasswordResetRes, err error) {
	return c.svc.ResetPassword(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// ResetPayPassword 重置客户支付密码，并记录操作人与客户端 IP。
func (c *CustomerController) ResetPayPassword(ctx context.Context, req *adminapi.CustomerPayPasswordResetReq) (res *adminapi.CustomerPayPasswordResetRes, err error) {
	return c.svc.ResetPayPassword(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
