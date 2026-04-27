package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// RechargeRiskController 提供充值账号风控规则和拦截记录后台 HTTP handler。
type RechargeRiskController struct {
	svc service.RechargeRiskService
}

// NewRechargeRisk 创建 RechargeRiskController。
func NewRechargeRisk(svc service.RechargeRiskService) *RechargeRiskController {
	return &RechargeRiskController{svc: svc}
}

// ListRules 返回充值账号风控规则分页列表。
func (c *RechargeRiskController) ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error) {
	return c.svc.ListRules(ctx, req)
}

// CreateRule 新增充值账号风控规则。
func (c *RechargeRiskController) CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq) (*adminapi.RechargeRiskRuleCreateRes, error) {
	return c.svc.CreateRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateRule 编辑充值账号风控规则。
func (c *RechargeRiskController) UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq) (*adminapi.RechargeRiskRuleUpdateRes, error) {
	return c.svc.UpdateRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateRuleStatus 启用或停用充值账号风控规则。
func (c *RechargeRiskController) UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq) (*adminapi.RechargeRiskRuleStatusRes, error) {
	return c.svc.UpdateRuleStatus(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// DeleteRule 软删除充值账号风控规则。
func (c *RechargeRiskController) DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq) (*adminapi.RechargeRiskRuleDeleteRes, error) {
	return c.svc.DeleteRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// ListRecords 返回充值账号风控拦截记录分页列表。
func (c *RechargeRiskController) ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error) {
	return c.svc.ListRecords(ctx, req)
}
