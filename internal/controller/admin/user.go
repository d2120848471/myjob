package admincontroller

import (
	userapi "myjob/api/user"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type UserController struct{ svc service.UserService }

func NewUser(svc service.UserService) *UserController { return &UserController{svc: svc} }
func (c *UserController) List(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	var req userapi.ListReq
	_ = r.Parse(&req)
	data, apiErr := c.svc.List(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) Trash(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	var req userapi.ListReq
	_ = r.Parse(&req)
	data, apiErr := c.svc.Trash(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) Add(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.AddReq
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
func (c *UserController) Edit(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.EditReq
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
func (c *UserController) Delete(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	data, apiErr := c.svc.Delete(r.Context(), r.GetRouter("id").Int64(), actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) Restore(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	data, apiErr := c.svc.Restore(r.Context(), r.GetRouter("id").Int64(), actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) Status(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.StatusReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Status(r.Context(), r.GetRouter("id").Int64(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) Notify(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.NotifyReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Notify(r.Context(), r.GetRouter("id").Int64(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) SetBusiness(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.BusinessReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.SetBusiness(r.Context(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *UserController) CancelBusiness(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req userapi.BusinessReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.CancelBusiness(r.Context(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
