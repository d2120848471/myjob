package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	"myjob/internal/service"
)

type OperationLogController struct{ svc service.AuditLogService }

func NewOperationLog(svc service.AuditLogService) *OperationLogController {
	return &OperationLogController{svc: svc}
}

func (c *OperationLogController) List(ctx context.Context, req *v1.OperationLogListReq) (res *v1.OperationLogListRes, err error) {
	return c.svc.OperationList(ctx, req)
}
