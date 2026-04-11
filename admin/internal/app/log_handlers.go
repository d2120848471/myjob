package app

import (
	"net/http"
	"strings"
)

func (a *Application) handleOperationLogList(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	page, pageSize := parsePagination(r)
	args := []interface{}{}
	conditions := []string{"1=1"}
	if adminID := strings.TrimSpace(r.URL.Query().Get("admin_id")); adminID != "" {
		conditions = append(conditions, "admin_id = ?")
		args = append(args, adminID)
	}
	if keyword := strings.TrimSpace(r.URL.Query().Get("keyword")); keyword != "" {
		conditions = append(conditions, "description LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if err := appendTimeRangeFilters(r.URL.Query().Get("start_time"), r.URL.Query().Get("end_time"), &conditions, &args); err != nil {
		writeError(w, http.StatusBadRequest, 400, "时间范围格式错误")
		return
	}
	where := strings.Join(conditions, " AND ")
	var total int
	if err := a.db.GetContext(r.Context(), &total, `SELECT COUNT(*) FROM admin_operation_log WHERE `+where, args...); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "操作日志查询失败")
		return
	}
	items := make([]operationLog, 0)
	queryArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT id, admin_id, admin_name, description, ip, ip_region, created_at
FROM admin_operation_log
WHERE `+where+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "操作日志查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"list":       items,
		"pagination": map[string]interface{}{"page": page, "page_size": pageSize, "total": total},
	})
}

func (a *Application) handleLoginLogList(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	page, pageSize := parsePagination(r)
	args := []interface{}{}
	conditions := []string{"1=1"}
	if adminID := strings.TrimSpace(r.URL.Query().Get("admin_id")); adminID != "" {
		conditions = append(conditions, "admin_id = ?")
		args = append(args, adminID)
	}
	if err := appendTimeRangeFilters(r.URL.Query().Get("start_time"), r.URL.Query().Get("end_time"), &conditions, &args); err != nil {
		writeError(w, http.StatusBadRequest, 400, "时间范围格式错误")
		return
	}
	where := strings.Join(conditions, " AND ")
	var total int
	if err := a.db.GetContext(r.Context(), &total, `SELECT COUNT(*) FROM admin_login_log WHERE `+where, args...); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录日志查询失败")
		return
	}
	items := make([]loginLog, 0)
	queryArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT id, admin_id, admin_name, ip, ip_region, created_at
FROM admin_login_log
WHERE `+where+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录日志查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"list":       items,
		"pagination": map[string]interface{}{"page": page, "page_size": pageSize, "total": total},
	})
}
