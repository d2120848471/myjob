package admincontroller

import (
	authapi "myjob/api/auth"
	"myjob/internal/library/response"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthController struct{ svc service.AuthService }

func NewAuth(svc service.AuthService) *AuthController { return &AuthController{svc: svc} }
func (c *AuthController) Login(r *ghttp.Request) {
	var req authapi.LoginReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.Login(r.Context(), req, r.GetClientIp())
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *AuthController) LoginSMSSend(r *ghttp.Request) {
	var req authapi.LoginSMSSendReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.LoginSMSSend(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *AuthController) LoginSMSVerify(r *ghttp.Request) {
	var req authapi.LoginSMSVerifyReq
	if err := r.Parse(&req); err != nil {
		response.Error(r, &modelruntime.APIError{HTTPStatus: 400, Code: 400, Message: "参数错误"})
		return
	}
	data, apiErr := c.svc.LoginSMSVerify(r.Context(), req)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *AuthController) Me(r *ghttp.Request, principal modelruntime.Principal, user entity.AdminUser) {
	data, apiErr := c.svc.Me(r.Context(), principal, user)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
func (c *AuthController) Logout(r *ghttp.Request, principal modelruntime.Principal, user entity.AdminUser) {
	data, apiErr := c.svc.Logout(r.Context(), principal, user)
	if apiErr != nil {
		response.Error(r, apiErr)
		return
	}
	response.Success(r, data)
}
