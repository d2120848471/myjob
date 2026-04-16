package admincontroller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// clientIP 从当前请求上下文中提取客户端 IP；若 ctx 中没有 HTTP 请求则返回空字符串。
//
// 注意：该文件仅承载极小的请求元信息辅助，避免继续吸附无关职责。
func clientIP(ctx context.Context) string {
	if request := g.RequestFromCtx(ctx); request != nil {
		return request.GetClientIp()
	}
	return ""
}
