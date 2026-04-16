package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// SessionController 提供登录态相关 HTTP handler（读取当前用户信息与退出登录）。
type SessionController struct{ svc service.AuthService }

// NewSession 创建 SessionController。
func NewSession(svc service.AuthService) *SessionController { return &SessionController{svc: svc} }

// Me 返回当前登录用户信息与权限码列表。
func (c *SessionController) Me(ctx context.Context, req *adminapi.AuthMeReq) (res *adminapi.AuthMeRes, err error) {
	return c.svc.Me(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}

// Delete 退出当前登录会话。
func (c *SessionController) Delete(ctx context.Context, req *adminapi.AuthSessionDeleteReq) (res *adminapi.AuthSessionDeleteRes, err error) {
	return c.svc.Logout(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}
