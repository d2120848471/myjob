package admincontroller

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/service"
)

// OperationLogController 提供操作日志查询相关 HTTP handler。
type OperationLogController struct{ svc service.AuditLogService }

// NewOperationLog 创建 OperationLogController。
func NewOperationLog(svc service.AuditLogService) *OperationLogController {
	return &OperationLogController{svc: svc}
}

// List 返回操作日志分页列表。
func (c *OperationLogController) List(ctx context.Context, req *adminapi.OperationLogListReq) (res *adminapi.OperationLogListRes, err error) {
	return c.svc.OperationList(ctx, req)
}
