package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SettingsController struct {
	smsSvc    service.SMSConfigService
	systemSvc service.SystemConfigService
}

func NewSettings(smsSvc service.SMSConfigService, systemSvc service.SystemConfigService) *SettingsController {
	return &SettingsController{smsSvc: smsSvc, systemSvc: systemSvc}
}

func (c *SettingsController) GetSMS(ctx context.Context, req *adminapi.SettingsSMSGetReq) (res *adminapi.SettingsSMSGetRes, err error) {
	return c.smsSvc.Get(ctx, req)
}

func (c *SettingsController) SaveSMS(ctx context.Context, req *adminapi.SettingsSMSSaveReq) (res *adminapi.SettingsSMSSaveRes, err error) {
	return c.smsSvc.Save(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SettingsController) GetSystem(ctx context.Context, req *adminapi.SettingsSystemGetReq) (res *adminapi.SettingsSystemGetRes, err error) {
	return c.systemSvc.Get(ctx, req)
}

func (c *SettingsController) SaveSystem(ctx context.Context, req *adminapi.SettingsSystemSaveReq) (res *adminapi.SettingsSystemSaveRes, err error) {
	return c.systemSvc.Save(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
