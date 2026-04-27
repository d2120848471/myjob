package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// RechargeRiskService 定义充值账号风控规则和拦截记录后台能力。
type RechargeRiskService interface {
	ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error)
	CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleCreateRes, error)
	UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleUpdateRes, error)
	UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleStatusRes, error)
	DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleDeleteRes, error)
	ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error)
}
