package middleware

import (
	"myjob/internal/kernel"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthenticatedHandler func(r *ghttp.Request, principal modelruntime.Principal, user entity.AdminUser)

type AuthGuard struct{ core *kernel.Core }

func NewAuthGuard(core *kernel.Core) *AuthGuard { return &AuthGuard{core: core} }

func (g *AuthGuard) Wrap(permission string, superOnly bool, next AuthenticatedHandler) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		// 先校验登录态，再按“超级管理员专属”和“普通权限码”两层做兜底。
		principal, user, err := g.core.AuthenticateRequest(r.Context(), r.GetHeader("Authorization"))
		if err != nil {
			response.Error(r, err)
			return
		}
		if superOnly && user.GroupID != 0 {
			response.Error(r, &modelruntime.APIError{HTTPStatus: 403, Code: 403, Message: "仅超级管理员可访问"})
			return
		}
		if permission != "" && user.GroupID != 0 {
			perms, loadErr := g.core.LoadPermissions(r.Context(), user.GroupID)
			if loadErr != nil {
				response.Error(r, &modelruntime.APIError{HTTPStatus: 500, Code: 500, Message: "权限加载失败"})
				return
			}
			if !kernel.ContainsString(perms, permission) {
				response.Error(r, &modelruntime.APIError{HTTPStatus: 403, Code: 403, Message: "无权限访问"})
				return
			}
		}
		next(r, principal, user)
	}
}
