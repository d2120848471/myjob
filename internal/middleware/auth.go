package middleware

import (
	"myjob/internal/app"
	"myjob/internal/consts"
	authctx "myjob/internal/library/auth"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthGuard struct{ core *app.Core }

func NewAuthGuard(core *app.Core) *AuthGuard { return &AuthGuard{core: core} }

func (g *AuthGuard) Require(permission string, superOnly bool) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		principal, user, err := g.core.AuthenticateRequest(r.Context(), r.GetHeader("Authorization"))
		if err != nil {
			r.SetError(err)
			return
		}
		if superOnly && user.GroupID != 0 {
			r.SetError(gerror.NewCode(consts.CodeForbidden, "仅超级管理员可访问"))
			return
		}
		if permission != "" && user.GroupID != 0 {
			perms, loadErr := g.core.LoadPermissions(r.Context(), user.GroupID)
			if loadErr != nil {
				r.SetError(gerror.NewCode(consts.CodeInternalError, "权限加载失败"))
				return
			}
			if !app.ContainsString(perms, permission) {
				r.SetError(gerror.NewCode(consts.CodeForbidden, "无权限访问"))
				return
			}
		}
		ctx := authctx.WithPrincipal(r.Context(), principal)
		ctx = authctx.WithUser(ctx, user)
		r.SetCtx(ctx)
		r.Middleware.Next()
	}
}
