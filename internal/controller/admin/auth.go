package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	"myjob/internal/service"
)

type AuthController struct{ svc service.AuthService }

func NewAuth(svc service.AuthService) *AuthController { return &AuthController{svc: svc} }

func (c *AuthController) Login(ctx context.Context, req *v1.AuthLoginReq) (res *v1.AuthLoginRes, err error) {
	return c.svc.Login(ctx, req, clientIP(ctx))
}

func (c *AuthController) SendSMS(ctx context.Context, req *v1.AuthSMSSendReq) (res *v1.AuthSMSSendRes, err error) {
	return c.svc.LoginSMSSend(ctx, req)
}

func (c *AuthController) VerifySMS(ctx context.Context, req *v1.AuthSMSVerifyReq) (res *v1.AuthSMSVerifyRes, err error) {
	return c.svc.LoginSMSVerify(ctx, req)
}
