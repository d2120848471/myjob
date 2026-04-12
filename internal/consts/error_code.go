package consts

import "github.com/gogf/gf/v2/errors/gcode"

var (
	CodeBadRequest      = gcode.New(400, "请求参数错误", nil)
	CodeUnauthorized    = gcode.New(401, "未登录或登录已失效", nil)
	CodeForbidden       = gcode.New(403, "无权限访问", nil)
	CodeConflict        = gcode.New(409, "资源冲突", nil)
	CodeTooManyRequests = gcode.New(429, "请求过于频繁", nil)
	CodeInternalError   = gcode.New(500, "内部错误", nil)
)
