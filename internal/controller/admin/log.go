package admincontroller

import (
	logapi "myjob/api/log"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type LogController struct{ svc service.AuditLogService }

func NewLog(svc service.AuditLogService) *LogController { return &LogController{svc: svc} }
func (c *LogController) Operation(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	var req logapi.ListReq
	_ = r.Parse(&req)
	data, apiErr := c.svc.OperationList(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *LogController) Login(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	var req logapi.ListReq
	_ = r.Parse(&req)
	data, apiErr := c.svc.LoginList(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
