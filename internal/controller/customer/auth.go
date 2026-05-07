package customercontroller

import (
	"context"

	customerapi "myjob/api"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// AuthController 提供客户侧认证 HTTP handler。
type AuthController struct{ svc service.CustomerAuthService }

// NewAuth 创建客户侧 AuthController。
func NewAuth(svc service.CustomerAuthService) *AuthController { return &AuthController{svc: svc} }

// SendSMS 发送客户注册或找回密码验证码。
func (c *AuthController) SendSMS(ctx context.Context, req *customerapi.CustomerAuthSMSSendReq) (res *customerapi.CustomerAuthSMSSendRes, err error) {
	return c.svc.SendSMS(ctx, req)
}

// Register 注册客户并直接返回客户 token。
func (c *AuthController) Register(ctx context.Context, req *customerapi.CustomerRegisterReq) (res *customerapi.CustomerRegisterRes, err error) {
	return c.svc.Register(ctx, req, clientIP(ctx))
}

// Login 使用手机号和登录密码登录客户账号。
func (c *AuthController) Login(ctx context.Context, req *customerapi.CustomerLoginReq) (res *customerapi.CustomerLoginRes, err error) {
	return c.svc.Login(ctx, req, clientIP(ctx))
}

// ForgotPassword 通过短信验证码重置客户登录密码。
func (c *AuthController) ForgotPassword(ctx context.Context, req *customerapi.CustomerForgotPasswordReq) (res *customerapi.CustomerForgotPasswordRes, err error) {
	return c.svc.ForgotPassword(ctx, req)
}

func clientIP(ctx context.Context) string {
	if request := g.RequestFromCtx(ctx); request != nil {
		return request.GetClientIp()
	}
	return ""
}
