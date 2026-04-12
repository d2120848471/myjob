package admincontroller

import (
	groupv1 "myjob/api/admin/group/v1"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type GroupController struct{ svc service.GroupService }

func NewGroup(svc service.GroupService) *GroupController { return &GroupController{svc: svc} }
func (c *GroupController) List(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	var req groupv1.ListReq
	_ = r.Parse(&req)
	data, apiErr := c.svc.List(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *GroupController) Add(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req groupv1.AddReq
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
func (c *GroupController) Edit(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req groupv1.EditReq
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
func (c *GroupController) Delete(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	data, apiErr := c.svc.Delete(r.Context(), r.GetRouter("id").Int64(), actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *GroupController) Status(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req groupv1.StatusReq
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
func (c *GroupController) AuthGet(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	data, apiErr := c.svc.AuthGet(r.Context(), r.GetRouter("id").Int64())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *GroupController) AuthSave(r *ghttp.Request, _ modelruntime.Principal, actor entity.AdminUser) {
	var req groupv1.AuthSaveReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.AuthSave(r.Context(), r.GetRouter("id").Int64(), req, actor, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *GroupController) MenuTree(r *ghttp.Request, _ modelruntime.Principal, _ entity.AdminUser) {
	data, apiErr := c.svc.MenuTree(r.Context())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
