package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SettingsController struct{ svc service.SMSConfigService }

func NewSettings(svc service.SMSConfigService) *SettingsController {
	return &SettingsController{svc: svc}
}

func (c *SettingsController) GetSMS(ctx context.Context, req *adminapi.SettingsSMSGetReq) (res *adminapi.SettingsSMSGetRes, err error) {
	return c.svc.Get(ctx, req)
}

func (c *SettingsController) SaveSMS(ctx context.Context, req *adminapi.SettingsSMSSaveReq) (res *adminapi.SettingsSMSSaveRes, err error) {
	return c.svc.Save(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
