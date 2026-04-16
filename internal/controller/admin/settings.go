package admincontroller

import (
	"myjob/internal/service"
)

// SettingsController 提供后台设置类接口的 HTTP handler。
//
// 当前设置域已拆分为短信配置与系统参数配置，两类 handler 实现在 settings_sms.go/settings_system.go。
type SettingsController struct {
	smsSvc    service.SMSConfigService
	systemSvc service.SystemConfigService
}

// NewSettings 创建 SettingsController，供 bootstrap 在路由层注册使用。
func NewSettings(smsSvc service.SMSConfigService, systemSvc service.SystemConfigService) *SettingsController {
	return &SettingsController{smsSvc: smsSvc, systemSvc: systemSvc}
}
