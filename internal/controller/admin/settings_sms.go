package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
)

// GetSMS 读取短信发送配置（脱敏），供后台“短信配置”页面初始化使用。
func (c *SettingsController) GetSMS(ctx context.Context, req *adminapi.SettingsSMSGetReq) (res *adminapi.SettingsSMSGetRes, err error) {
	return c.smsSvc.Get(ctx, req)
}

// SaveSMS 保存短信发送配置，并记录操作人与客户端 IP。
func (c *SettingsController) SaveSMS(ctx context.Context, req *adminapi.SettingsSMSSaveReq) (res *adminapi.SettingsSMSSaveRes, err error) {
	return c.smsSvc.Save(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
