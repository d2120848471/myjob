package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

// AuthService 定义认证、会话与短信二次校验相关能力。
type AuthService interface {
	Login(ctx context.Context, req *adminapi.AuthLoginReq, ip string) (*adminapi.AuthLoginRes, error)
	LoginSMSSend(ctx context.Context, req *adminapi.AuthSMSSendReq) (*adminapi.AuthSMSSendRes, error)
	LoginSMSVerify(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (*adminapi.AuthSMSVerifyRes, error)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthMeRes, error)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthSessionDeleteRes, error)
}
