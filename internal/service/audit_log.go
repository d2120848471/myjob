package service

import (
	"context"

	adminapi "myjob/api"
)

// AuditLogService 定义操作日志与登录日志查询能力。
type AuditLogService interface {
	OperationList(ctx context.Context, req *adminapi.OperationLogListReq) (*adminapi.OperationLogListRes, error)
	LoginList(ctx context.Context, req *adminapi.LoginLogListReq) (*adminapi.LoginLogListRes, error)
}
