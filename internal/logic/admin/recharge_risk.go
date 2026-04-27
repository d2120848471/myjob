package adminlogic

import (
	"myjob/internal/app"
	"myjob/internal/service"
)

// RechargeRiskLogic 提供充值账号风控规则和拦截记录后台业务能力。
type RechargeRiskLogic struct{ core *app.Core }

var _ service.RechargeRiskService = (*RechargeRiskLogic)(nil)
