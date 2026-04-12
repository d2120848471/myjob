package adminlogic

import (
	"context"
	"net/http"
	"strings"

	logapi "myjob/api/log"
	"myjob/internal/kernel"
	modelruntime "myjob/internal/model/runtime"
)

type AuditLogLogic struct{ core *kernel.Core }

func (l *AuditLogLogic) OperationList(ctx context.Context, req logapi.ListReq) (map[string]any, *modelruntime.APIError) {
	page, pageSize := kernel.ParsePagination(req.Page, req.PageSize)
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
	if err := kernel.AppendTimeRangeFilters(req.StartTime, req.EndTime, &conditions, &args); err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "时间范围格式错误")
	}
	where := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_operation_log WHERE `+where, args...)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "操作日志查询失败")
	}
	items := make([]kernel.OperationLog, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, admin_id, admin_name, description, ip, ip_region, created_at FROM admin_operation_log WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "操作日志查询失败")
	}
	return map[string]any{"list": items, "pagination": map[string]any{"page": page, "page_size": pageSize, "total": total.Int()}}, nil
}

func (l *AuditLogLogic) LoginList(ctx context.Context, req logapi.ListReq) (map[string]any, *modelruntime.APIError) {
	page, pageSize := kernel.ParsePagination(req.Page, req.PageSize)
	args := []any{}
	conditions := []string{"1=1"}
	if adminID := strings.TrimSpace(req.AdminID); adminID != "" {
		conditions = append(conditions, "admin_id = ?")
		args = append(args, adminID)
	}
	if err := kernel.AppendTimeRangeFilters(req.StartTime, req.EndTime, &conditions, &args); err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "时间范围格式错误")
	}
	where := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_login_log WHERE `+where, args...)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录日志查询失败")
	}
	items := make([]kernel.LoginLog, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, admin_id, admin_name, ip, ip_region, created_at FROM admin_login_log WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录日志查询失败")
	}
	return map[string]any{"list": items, "pagination": map[string]any{"page": page, "page_size": pageSize, "total": total.Int()}}, nil
}
