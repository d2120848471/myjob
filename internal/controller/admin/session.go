package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SessionController struct{ svc service.AuthService }

func NewSession(svc service.AuthService) *SessionController { return &SessionController{svc: svc} }

func (c *SessionController) Me(ctx context.Context, req *v1.AuthMeReq) (res *v1.AuthMeRes, err error) {
	return c.svc.Me(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}

func (c *SessionController) Delete(ctx context.Context, req *v1.AuthSessionDeleteReq) (res *v1.AuthSessionDeleteRes, err error) {
	return c.svc.Logout(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}
