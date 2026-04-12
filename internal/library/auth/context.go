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

func WithPrincipal(ctx context.Context, principal modelruntime.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func PrincipalFromCtx(ctx context.Context) (modelruntime.Principal, bool) {
	value, ok := ctx.Value(principalContextKey).(modelruntime.Principal)
	return value, ok
}

func MustPrincipalFromCtx(ctx context.Context) modelruntime.Principal {
	value, _ := PrincipalFromCtx(ctx)
	return value
}

func WithUser(ctx context.Context, user entity.AdminUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromCtx(ctx context.Context) (entity.AdminUser, bool) {
	value, ok := ctx.Value(userContextKey).(entity.AdminUser)
	return value, ok
}

func MustUserFromCtx(ctx context.Context) entity.AdminUser {
	value, _ := UserFromCtx(ctx)
	return value
}
