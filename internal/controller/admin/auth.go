package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

type AuthController struct{ svc service.AuthService }

func NewAuth(svc service.AuthService) *AuthController { return &AuthController{svc: svc} }

func (c *AuthController) Login(ctx context.Context, req *adminapi.AuthLoginReq) (res *adminapi.AuthLoginRes, err error) {
	return c.svc.Login(ctx, req, clientIP(ctx))
}

func (c *AuthController) SendSMS(ctx context.Context, req *adminapi.AuthSMSSendReq) (res *adminapi.AuthSMSSendRes, err error) {
	return c.svc.LoginSMSSend(ctx, req)
}

func (c *AuthController) VerifySMS(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (res *adminapi.AuthSMSVerifyRes, err error) {
	return c.svc.LoginSMSVerify(ctx, req)
}
