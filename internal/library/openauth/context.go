package openauth

import (
	"context"

	"myjob/internal/model/entity"
)

type contextKey string

const callerContextKey contextKey = "open-caller"

// WithCaller 将开放接口鉴权后的调用方信息写入 context（用于请求链路透传）。
func WithCaller(ctx context.Context, caller entity.OpenCaller) context.Context {
	return context.WithValue(ctx, callerContextKey, caller)
}

// CallerFromCtx 从 context 中读取调用方信息。
func CallerFromCtx(ctx context.Context) (entity.OpenCaller, bool) {
	value, ok := ctx.Value(callerContextKey).(entity.OpenCaller)
	return value, ok
}

// MustCallerFromCtx 从 context 中读取调用方信息（读取失败时返回零值）。
func MustCallerFromCtx(ctx context.Context) entity.OpenCaller {
	value, _ := CallerFromCtx(ctx)
	return value
}

