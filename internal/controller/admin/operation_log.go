package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

type OperationLogController struct{ svc service.AuditLogService }

func NewOperationLog(svc service.AuditLogService) *OperationLogController {
	return &OperationLogController{svc: svc}
}

func (c *OperationLogController) List(ctx context.Context, req *adminapi.OperationLogListReq) (res *adminapi.OperationLogListRes, err error) {
	return c.svc.OperationList(ctx, req)
}
