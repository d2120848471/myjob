package app

import (
	admindto "admin/internal/model/dto/admin"
	"fmt"
	"net/http"
	"strings"
)

func (a *Application) handleGroupList(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	page, pageSize := parsePagination(r)
	var total int
	if err := a.db.GetContext(r.Context(), &total, `SELECT COUNT(*) FROM admin_group`); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组列表查询失败")
		return
	}
	items := make([]groupListItem, 0)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT g.id, g.name, g.description, g.status,
       (SELECT COUNT(*) FROM admin_user u WHERE u.group_id = g.id) AS user_count
FROM admin_group g
ORDER BY g.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组列表查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"list":       items,
		"pagination": admindto.Pagination{Page: page, PageSize: pageSize, Total: total},
	})
}

func (a *Application) handleGroupAdd(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, 400, "请输入用户组名称")
		return
	}
	var exists int
	if err := a.db.GetContext(r.Context(), &exists, `SELECT COUNT(*) FROM admin_group WHERE name = ?`, req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组查询失败")
		return
	}
	if exists > 0 {
		writeError(w, http.StatusConflict, 409, "用户组名称已存在")
		return
	}
	result, err := a.db.ExecContext(r.Context(), `INSERT INTO admin_group (name, description, status, created_at, updated_at) VALUES (?, ?, 1, ?, ?)`, req.Name, strings.TrimSpace(req.Description), a.now(), a.now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组新增失败")
		return
	}
	id, _ := result.LastInsertId()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("添加用户组：%s", req.Name), requestIP(r))
	writeSuccess(w, map[string]interface{}{"id": id})
}

func (a *Application) handleGroupEdit(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户组ID错误")
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, 400, "请输入用户组名称")
		return
	}
	var exists int
	if err = a.db.GetContext(r.Context(), &exists, `SELECT COUNT(*) FROM admin_group WHERE name = ? AND id <> ?`, req.Name, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组查询失败")
		return
	}
	if exists > 0 {
		writeError(w, http.StatusConflict, 409, "用户组名称已存在")
		return
	}
	if _, err = a.db.ExecContext(r.Context(), `UPDATE admin_group SET name = ?, description = ?, updated_at = ? WHERE id = ?`, req.Name, strings.TrimSpace(req.Description), a.now(), id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组编辑失败")
		return
	}
	a.writeOperation(r.Context(), actor, fmt.Sprintf("编辑用户组：%s", req.Name), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleGroupDelete(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户组ID错误")
		return
	}
	var userCount int
	if err = a.db.GetContext(r.Context(), &userCount, `SELECT COUNT(*) FROM admin_user WHERE group_id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组校验失败")
		return
	}
	if userCount > 0 {
		writeError(w, http.StatusConflict, 409, fmt.Sprintf("该用户组下还有 %d 名员工，请先转移", userCount))
		return
	}
	tx, txErr := a.db.BeginTxx(r.Context(), nil)
	if txErr != nil {
		writeError(w, http.StatusInternalServerError, 500, "事务开启失败")
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM admin_group_menu WHERE group_id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组删除失败")
		return
	}
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM admin_group WHERE id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组删除失败")
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组删除失败")
		return
	}
	_ = a.redis.Del(r.Context(), permissionCacheKey(id)).Err()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("删除用户组：%d", id), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleGroupStatus(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户组ID错误")
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err = decodeJSON(r, &req); err != nil || (req.Status != 0 && req.Status != 1) {
		writeError(w, http.StatusBadRequest, 400, "状态错误")
		return
	}
	if _, err = a.db.ExecContext(r.Context(), `UPDATE admin_group SET status = ?, updated_at = ? WHERE id = ?`, req.Status, a.now(), id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "用户组状态更新失败")
		return
	}
	_ = a.redis.Del(r.Context(), permissionCacheKey(id)).Err()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("切换用户组状态：%d -> %d", id, req.Status), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleGroupAuthGet(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户组ID错误")
		return
	}
	ids := make([]int64, 0)
	if err = a.db.SelectContext(r.Context(), &ids, `
SELECT gm.menu_id
FROM admin_group_menu gm
JOIN admin_menu m ON m.id = gm.menu_id
WHERE gm.group_id = ? AND m.status = 1 AND m.super_only = 0
ORDER BY gm.menu_id ASC
`, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "授权查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{"menu_ids": ids})
}

func (a *Application) handleGroupAuthSave(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户组ID错误")
		return
	}
	var req struct {
		MenuIDs []int64 `json:"menu_ids"`
	}
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	tx, txErr := a.db.BeginTxx(r.Context(), nil)
	if txErr != nil {
		writeError(w, http.StatusInternalServerError, 500, "事务开启失败")
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM admin_group_menu WHERE group_id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "授权保存失败")
		return
	}
	allowed := make([]int64, 0, len(req.MenuIDs))
	if len(req.MenuIDs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(req.MenuIDs)), ",")
		queryArgs := make([]interface{}, 0, len(req.MenuIDs))
		for _, menuID := range req.MenuIDs {
			queryArgs = append(queryArgs, menuID)
		}
		if err = tx.SelectContext(r.Context(), &allowed, `
SELECT id
FROM admin_menu
WHERE status = 1 AND super_only = 0 AND id IN (`+placeholders+`)
ORDER BY sort ASC, id ASC
`, queryArgs...); err != nil {
			writeError(w, http.StatusInternalServerError, 500, "授权保存失败")
			return
		}
	}
	for _, menuID := range allowed {
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (?, ?, ?)`, id, menuID, a.now()); err != nil {
			writeError(w, http.StatusInternalServerError, 500, "授权保存失败")
			return
		}
	}
	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "授权保存失败")
		return
	}
	_ = a.redis.Del(r.Context(), permissionCacheKey(id)).Err()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("保存用户组授权：%d，共 %d 个权限", id, len(allowed)), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}
