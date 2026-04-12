package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

type AuditLogLogic struct{ core *app.Core }

func (l *AuditLogLogic) OperationList(ctx context.Context, req *adminapi.OperationLogListReq) (*adminapi.OperationLogListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	args := []any{}
	conditions := []string{"1=1"}
	if adminID := strings.TrimSpace(req.AdminID); adminID != "" {
		conditions = append(conditions, "admin_id = ?")
		args = append(args, adminID)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		conditions = append(conditions, "description LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if err := app.AppendTimeRangeFilters(req.StartTime, req.EndTime, &conditions, &args); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "时间范围格式错误")
	}
	where := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_operation_log WHERE `+where, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "操作日志查询失败")
	}
	items := make([]app.OperationLog, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, admin_id, admin_name, description, ip, ip_region, created_at FROM admin_operation_log WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "操作日志查询失败")
	}
	return &adminapi.OperationLogListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()}}, nil
}

func (l *AuditLogLogic) LoginList(ctx context.Context, req *adminapi.LoginLogListReq) (*adminapi.LoginLogListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	args := []any{}
	conditions := []string{"1=1"}
	if adminID := strings.TrimSpace(req.AdminID); adminID != "" {
		conditions = append(conditions, "admin_id = ?")
		args = append(args, adminID)
	}
	if err := app.AppendTimeRangeFilters(req.StartTime, req.EndTime, &conditions, &args); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "时间范围格式错误")
	}
	where := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_login_log WHERE `+where, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录日志查询失败")
	}
	items := make([]app.LoginLog, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, admin_id, admin_name, ip, ip_region, created_at FROM admin_login_log WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录日志查询失败")
	}
	return &adminapi.LoginLogListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()}}, nil
}
