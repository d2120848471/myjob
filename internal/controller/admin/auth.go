package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// AuthController 提供后台认证相关 HTTP handler（账号密码登录与短信二验）。
type AuthController struct{ svc service.AuthService }

// NewAuth 创建 AuthController。
func NewAuth(svc service.AuthService) *AuthController { return &AuthController{svc: svc} }

// Login 执行账号密码登录；在需要短信二次验证时会返回临时 login_token。
func (c *AuthController) Login(ctx context.Context, req *adminapi.AuthLoginReq) (res *adminapi.AuthLoginRes, err error) {
	return c.svc.Login(ctx, req, clientIP(ctx))
}

// SendSMS 在需要短信二验时发送登录验证码。
func (c *AuthController) SendSMS(ctx context.Context, req *adminapi.AuthSMSSendReq) (res *adminapi.AuthSMSSendRes, err error) {
	return c.svc.LoginSMSSend(ctx, req)
}

// VerifySMS 校验短信验证码并换取正式登录 token。
func (c *AuthController) VerifySMS(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (res *adminapi.AuthSMSVerifyRes, err error) {
	return c.svc.LoginSMSVerify(ctx, req)
}
