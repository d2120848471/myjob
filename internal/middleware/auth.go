package middleware

import (
	"myjob/internal/app"
	"myjob/internal/consts"
	authctx "myjob/internal/library/auth"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuthGuard 提供后台接口鉴权与权限校验中间件。
type AuthGuard struct{ core *app.Core }

// NewAuthGuard 创建鉴权守卫中间件，依赖 core 提供的鉴权与权限加载能力。
func NewAuthGuard(core *app.Core) *AuthGuard { return &AuthGuard{core: core} }

// Require 返回一个中间件：校验登录态、可选校验权限码，并将 principal/user 写入请求上下文。
//
// - permission 为空：仅要求已登录
// - superOnly 为 true：仅允许超级管理员访问（group_id=0）
func (g *AuthGuard) Require(permission string, superOnly bool) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// 先解析并校验登录态（Bearer JWT + Redis Session）。
		principal, user, err := g.core.AuthenticateRequest(r.Context(), r.GetHeader("Authorization"))
		if err != nil {
			r.SetError(err)
			return
		}
		// 超级管理员在系统中约定 group_id=0。
		if superOnly && user.GroupID != 0 {
			r.SetError(gerror.NewCode(consts.CodeForbidden, "仅超级管理员可访问"))
			return
		}
		// 超级管理员默认拥有全部权限；普通用户需要加载其所属用户组的授权菜单。
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
		// 将鉴权后的信息写入 context，供 controller/logic 使用（避免重复查询）。
		ctx := authctx.WithPrincipal(r.Context(), principal)
		ctx = authctx.WithUser(ctx, user)
		r.SetCtx(ctx)
		r.Middleware.Next()
	}
}
