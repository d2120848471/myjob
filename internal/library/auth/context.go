package auth

import (
	"context"

	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type contextKey string

const (
	principalContextKey contextKey = "admin-principal"
	userContextKey      contextKey = "admin-user"
)

// WithPrincipal 将鉴权后的 principal 写入 context（用于请求链路透传）。
func WithPrincipal(ctx context.Context, principal modelruntime.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

// PrincipalFromCtx 从 context 中读取 principal。
func PrincipalFromCtx(ctx context.Context) (modelruntime.Principal, bool) {
	value, ok := ctx.Value(principalContextKey).(modelruntime.Principal)
	return value, ok
}

// MustPrincipalFromCtx 从 context 中读取 principal（读取失败时返回零值）。
func MustPrincipalFromCtx(ctx context.Context) modelruntime.Principal {
	value, _ := PrincipalFromCtx(ctx)
	return value
}

// WithUser 将鉴权后的用户信息写入 context（用于请求链路透传）。
func WithUser(ctx context.Context, user entity.AdminUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromCtx 从 context 中读取用户信息。
func UserFromCtx(ctx context.Context) (entity.AdminUser, bool) {
	value, ok := ctx.Value(userContextKey).(entity.AdminUser)
	return value, ok
}

// MustUserFromCtx 从 context 中读取用户信息（读取失败时返回零值）。
func MustUserFromCtx(ctx context.Context) entity.AdminUser {
	value, _ := UserFromCtx(ctx)
	return value
}
