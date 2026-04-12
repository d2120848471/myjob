package response

import (
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/net/ghttp"
)

func Success(r *ghttp.Request, data any) {
	r.Response.WriteJson(modelruntime.ResponseEnvelope{Code: 0, Msg: "success", Data: data})
}

func Error(r *ghttp.Request, err *modelruntime.APIError) {
	if err == nil {
		err = &modelruntime.APIError{HTTPStatus: 500, Code: 500, Message: "internal error"}
	}
	// GoFrame 的 WriteStatus 在无附带内容时会先写入 HTTP 状态文本，
	// 会把 `Unauthorized`/`Internal Server Error` 直接拼进响应体，破坏统一 JSON 包裹。
	r.Response.Status = err.HTTPStatus
	r.Response.WriteJson(modelruntime.ResponseEnvelope{Code: err.Code, Msg: err.Message, Data: nil})
}
