package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

type LoginLogController struct{ svc service.AuditLogService }

func NewLoginLog(svc service.AuditLogService) *LoginLogController {
	return &LoginLogController{svc: svc}
}

func (c *LoginLogController) List(ctx context.Context, req *adminapi.LoginLogListReq) (res *adminapi.LoginLogListRes, err error) {
	return c.svc.LoginList(ctx, req)
}
