package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// SMSConfigService 定义短信配置读写能力（super-only）。
type SMSConfigService interface {
	Get(ctx context.Context, req *adminapi.SettingsSMSGetReq) (*adminapi.SettingsSMSGetRes, error)
	Save(ctx context.Context, req *adminapi.SettingsSMSSaveReq, actor entity.AdminUser, ip string) (*adminapi.SettingsSMSSaveRes, error)
}
