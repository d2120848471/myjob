package service

import (
	"context"

	customerapi "myjob/api"
)

// CustomerAuthService 定义客户注册、登录、短信验证码和忘记密码能力。
type CustomerAuthService interface {
	SendSMS(ctx context.Context, req *customerapi.CustomerAuthSMSSendReq) (*customerapi.CustomerAuthSMSSendRes, error)
	Register(ctx context.Context, req *customerapi.CustomerRegisterReq, ip string) (*customerapi.CustomerRegisterRes, error)
	Login(ctx context.Context, req *customerapi.CustomerLoginReq, ip string) (*customerapi.CustomerLoginRes, error)
	ForgotPassword(ctx context.Context, req *customerapi.CustomerForgotPasswordReq) (*customerapi.CustomerForgotPasswordRes, error)
}
