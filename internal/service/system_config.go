package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// SystemConfigService 定义系统参数配置读写能力（super-only）。
type SystemConfigService interface {
	Get(ctx context.Context, req *adminapi.SettingsSystemGetReq) (*adminapi.SettingsSystemGetRes, error)
	Save(ctx context.Context, req *adminapi.SettingsSystemSaveReq, actor entity.AdminUser, ip string) (*adminapi.SettingsSystemSaveRes, error)
}
