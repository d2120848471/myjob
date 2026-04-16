package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// PurchaseLimitController 提供购买数量限制策略管理相关 HTTP handler。
type PurchaseLimitController struct {
	svc service.PurchaseLimitService
}

// NewPurchaseLimit 创建 PurchaseLimitController。
func NewPurchaseLimit(svc service.PurchaseLimitService) *PurchaseLimitController {
	return &PurchaseLimitController{svc: svc}
}

// List 返回购买数量限制策略分页列表。
func (c *PurchaseLimitController) List(ctx context.Context, req *adminapi.PurchaseLimitStrategyListReq) (res *adminapi.PurchaseLimitStrategyListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增购买数量限制策略，并记录操作人与客户端 IP。
func (c *PurchaseLimitController) Create(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq) (res *adminapi.PurchaseLimitStrategyCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑购买数量限制策略，并记录操作人与客户端 IP。
func (c *PurchaseLimitController) Update(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq) (res *adminapi.PurchaseLimitStrategyUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除购买数量限制策略，并记录操作人与客户端 IP。
func (c *PurchaseLimitController) Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq) (res *adminapi.PurchaseLimitStrategyDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Status 切换购买数量限制策略启停状态，并记录操作人与客户端 IP。
func (c *PurchaseLimitController) Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq) (res *adminapi.PurchaseLimitStrategyStatusRes, err error) {
	return c.svc.Status(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Enums 返回购买数量限制策略相关的枚举数据。
func (c *PurchaseLimitController) Enums(ctx context.Context, req *adminapi.PurchaseLimitStrategyEnumsReq) (res *adminapi.PurchaseLimitStrategyEnumsRes, err error) {
	return c.svc.Enums(ctx, req)
}
