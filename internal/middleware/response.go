package middleware

import (
	"mime"

	"myjob/internal/consts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// JSONResponse 是统一的 HTTP JSON 响应包装。
//
// 约定结构：{code, message, data}，便于前端统一处理。
type JSONResponse struct {
	Code    int    `json:"code" dc:"业务状态码"`
	Message string `json:"message" dc:"业务消息"`
	Data    any    `json:"data" dc:"业务数据"`
}

var streamContentTypes = map[string]struct{}{
	"text/event-stream":         {},
	"application/octet-stream":  {},
	"multipart/x-mixed-replace": {},
}

// Response 在 handler 执行后统一将结果包装为 JSONResponse。
//
// 若 handler 已自行写入响应，或响应为流式类型（如 SSE），则跳过包装。
func Response(r *ghttp.Request) {
	r.Middleware.Next()

	// handler 已经写入响应（例如自定义输出/文件下载），则不再二次包装 JSON。
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	// SSE/流式等场景不应强制包成 JSON，否则会破坏上游协议。
	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	if _, ok := streamContentTypes[mediaType]; ok {
		return
	}

	var (
		err     = r.GetError()
		res     = r.GetHandlerResponse()
		code    = gerror.Code(err)
		message string
	)

	if err != nil {
		// 业务错误：优先使用携带的 code；若没有则统一为内部错误。
		if code == gcode.CodeNil {
			code = consts.CodeInternalError
		}
		message = err.Error()
		if message == "" {
			message = code.Message()
		}
	} else {
		// 正常返回：统一 code=0，message 使用框架默认 OK 文案。
		code = gcode.CodeOK
		message = code.Message()
	}

	r.Response.WriteJson(JSONResponse{
		Code:    code.Code(),
		Message: message,
		Data:    res,
	})
}
