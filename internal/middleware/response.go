package middleware

import (
	"mime"

	"myjob/internal/consts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

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

func Response(r *ghttp.Request) {
	r.Middleware.Next()

	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

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
		if code == gcode.CodeNil {
			code = consts.CodeInternalError
		}
		message = err.Error()
		if message == "" {
			message = code.Message()
		}
	} else {
		code = gcode.CodeOK
		message = code.Message()
	}

	r.Response.WriteJson(JSONResponse{
		Code:    code.Code(),
		Message: message,
		Data:    res,
	})
}
