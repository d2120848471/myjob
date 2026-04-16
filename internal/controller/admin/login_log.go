package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// LoginLogController 提供登录日志查询相关 HTTP handler。
type LoginLogController struct{ svc service.AuditLogService }

// NewLoginLog 创建 LoginLogController。
func NewLoginLog(svc service.AuditLogService) *LoginLogController {
	return &LoginLogController{svc: svc}
}

// List 返回登录日志分页列表。
func (c *LoginLogController) List(ctx context.Context, req *adminapi.LoginLogListReq) (res *adminapi.LoginLogListRes, err error) {
	return c.svc.LoginList(ctx, req)
}
