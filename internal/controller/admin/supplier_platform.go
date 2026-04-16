package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// SupplierPlatformController 提供第三方供货平台对接相关 HTTP handler。
type SupplierPlatformController struct {
	svc service.SupplierPlatformService
}

// NewSupplierPlatform 创建 SupplierPlatformController。
func NewSupplierPlatform(svc service.SupplierPlatformService) *SupplierPlatformController {
	return &SupplierPlatformController{svc: svc}
}

// TypeList 返回第三方供货平台类型字典。
func (c *SupplierPlatformController) TypeList(ctx context.Context, req *adminapi.SupplierPlatformTypeListReq) (res *adminapi.SupplierPlatformTypeListRes, err error) {
	return c.svc.TypeList(ctx, req)
}

// List 返回第三方平台账号分页列表。
func (c *SupplierPlatformController) List(ctx context.Context, req *adminapi.SupplierPlatformListReq) (res *adminapi.SupplierPlatformListRes, err error) {
	return c.svc.List(ctx, req)
}

// Detail 返回指定平台账号详情。
func (c *SupplierPlatformController) Detail(ctx context.Context, req *adminapi.SupplierPlatformDetailReq) (res *adminapi.SupplierPlatformDetailRes, err error) {
	return c.svc.Detail(ctx, req)
}

// Create 新增平台账号，并记录操作人与客户端 IP。
func (c *SupplierPlatformController) Create(ctx context.Context, req *adminapi.SupplierPlatformCreateReq) (res *adminapi.SupplierPlatformCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑平台账号，并记录操作人与客户端 IP。
func (c *SupplierPlatformController) Update(ctx context.Context, req *adminapi.SupplierPlatformUpdateReq) (res *adminapi.SupplierPlatformUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 软删除平台账号，并记录操作人与客户端 IP。
func (c *SupplierPlatformController) Delete(ctx context.Context, req *adminapi.SupplierPlatformDeleteReq) (res *adminapi.SupplierPlatformDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// RefreshBalance 手动刷新平台余额，并记录操作人与客户端 IP。
func (c *SupplierPlatformController) RefreshBalance(ctx context.Context, req *adminapi.SupplierPlatformRefreshBalanceReq) (res *adminapi.SupplierPlatformRefreshBalanceRes, err error) {
	return c.svc.RefreshBalance(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
