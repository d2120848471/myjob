package admincontroller

import (
	configapi "myjob/api/config"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type ConfigController struct{ svc service.SMSConfigService }

func NewConfig(svc service.SMSConfigService) *ConfigController { return &ConfigController{svc: svc} }
func (c *ConfigController) GetSMS(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	data, apiErr := c.svc.Get(r.Context())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *ConfigController) SaveSMS(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req configapi.SMSConfigSaveReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Save(r.Context(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
