package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SessionController struct{ svc service.AuthService }

func NewSession(svc service.AuthService) *SessionController { return &SessionController{svc: svc} }

func (c *SessionController) Me(ctx context.Context, req *adminapi.AuthMeReq) (res *adminapi.AuthMeRes, err error) {
	return c.svc.Me(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}

func (c *SessionController) Delete(ctx context.Context, req *adminapi.AuthSessionDeleteReq) (res *adminapi.AuthSessionDeleteRes, err error) {
	return c.svc.Logout(ctx, authctx.MustPrincipalFromCtx(ctx), authctx.MustUserFromCtx(ctx))
}
