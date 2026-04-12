package admincontroller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func clientIP(ctx context.Context) string {
	if request := g.RequestFromCtx(ctx); request != nil {
		return request.GetClientIp()
	}
	return ""
}
