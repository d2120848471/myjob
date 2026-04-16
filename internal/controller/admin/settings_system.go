package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
)

// GetSystem 读取系统参数配置，支持按分组读取或一次返回全部分组。
func (c *SettingsController) GetSystem(ctx context.Context, req *adminapi.SettingsSystemGetReq) (res *adminapi.SettingsSystemGetRes, err error) {
	return c.systemSvc.Get(ctx, req)
}

// SaveSystem 保存系统参数配置，并记录操作人与客户端 IP。
func (c *SettingsController) SaveSystem(ctx context.Context, req *adminapi.SettingsSystemSaveReq) (res *adminapi.SettingsSystemSaveRes, err error) {
	return c.systemSvc.Save(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
