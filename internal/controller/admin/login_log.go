package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	"myjob/internal/service"
)

type LoginLogController struct{ svc service.AuditLogService }

func NewLoginLog(svc service.AuditLogService) *LoginLogController {
	return &LoginLogController{svc: svc}
}

func (c *LoginLogController) List(ctx context.Context, req *v1.LoginLogListReq) (res *v1.LoginLogListRes, err error) {
	return c.svc.LoginList(ctx, req)
}
