package admincontroller

import (
	subjectv1 "myjob/api/admin/subject/v1"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type SubjectController struct{ svc service.SubjectService }

func NewSubject(svc service.SubjectService) *SubjectController { return &SubjectController{svc: svc} }
func (c *SubjectController) List(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	data, apiErr := c.svc.List(r.Context())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *SubjectController) Add(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req subjectv1.AddReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Add(r.Context(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *SubjectController) Edit(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req subjectv1.EditReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Edit(r.Context(), r.GetRouter("id").Int64(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
