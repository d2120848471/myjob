package app

import (
	admindto "admin/internal/model/dto/admin"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func (a *Application) handleUserList(w http.ResponseWriter, r *http.Request, p principal, user AdminUser) {
	_ = p
	_ = user
	page, pageSize := parsePagination(r)
	total := 0
	if err := a.db.GetContext(r.Context(), &total, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 0`); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "员工列表查询失败")
		return
	}
	items := make([]userListItem, 0)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 0
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "员工列表查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"list":       items,
		"pagination": admindto.Pagination{Page: page, PageSize: pageSize, Total: total},
	})
}

func (a *Application) handleUserTrash(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	page, pageSize := parsePagination(r)
	total := 0
	if err := a.db.GetContext(r.Context(), &total, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 1`); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "回收站查询失败")
		return
	}
	items := make([]userListItem, 0)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 1
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "回收站查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"list":       items,
		"pagination": admindto.Pagination{Page: page, PageSize: pageSize, Total: total},
	})
}

func (a *Application) handleUserAdd(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	var req struct {
		Username        string `json:"username"`
		ConfirmUsername string `json:"confirm_username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		RealName        string `json:"real_name"`
		Phone           string `json:"phone"`
		GroupID         int64  `json:"group_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if !usernameRegexp.MatchString(req.Username) || req.Username != req.ConfirmUsername {
		writeError(w, http.StatusBadRequest, 400, "用户名格式错误")
		return
	}
	if !passwordRegexp.MatchString(req.Password) || req.Password != req.ConfirmPassword {
		writeError(w, http.StatusBadRequest, 400, "密码格式错误")
		return
	}
	if req.RealName == "" || !phoneRegexp.MatchString(req.Phone) {
		writeError(w, http.StatusBadRequest, 400, "手机号格式错误")
		return
	}
	if err := a.ensureGroupActive(r.Context(), req.GroupID); err != nil {
		writeError(w, http.StatusBadRequest, 400, err.Error())
		return
	}
	if exists, _ := a.activeUsernameExists(r.Context(), req.Username, 0); exists {
		writeError(w, http.StatusConflict, 409, "用户名已存在")
		return
	}
	if exists, _ := a.activePhoneExists(r.Context(), req.Phone, 0); exists {
		writeError(w, http.StatusConflict, 409, "手机号已存在")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "密码加密失败")
		return
	}
	result, err := a.db.ExecContext(r.Context(), `
INSERT INTO admin_user (
    username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, 1, 0, 0, 0, 0, ?, ?)
`, req.Username, string(hash), req.RealName, req.Phone, req.GroupID, a.now(), a.now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "员工新增失败")
		return
	}
	id, _ := result.LastInsertId()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("添加员工：%s，用户组：%d", req.Username, req.GroupID), requestIP(r))
	writeSuccess(w, map[string]interface{}{"id": id})
}

func (a *Application) handleUserEdit(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "员工ID错误")
		return
	}
	user, err := a.getUserByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, 404, "员工不存在")
		return
	}
	var req struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		RealName        string `json:"real_name"`
		Phone           string `json:"phone"`
		GroupID         int64  `json:"group_id"`
	}
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if req.RealName == "" || !phoneRegexp.MatchString(req.Phone) {
		writeError(w, http.StatusBadRequest, 400, "手机号格式错误")
		return
	}
	if err = a.ensureGroupActive(r.Context(), req.GroupID); err != nil {
		writeError(w, http.StatusBadRequest, 400, err.Error())
		return
	}
	if exists, _ := a.activePhoneExists(r.Context(), req.Phone, id); exists {
		writeError(w, http.StatusConflict, 409, "手机号已存在")
		return
	}
	newVersion := user.TokenVersion
	tx, err := a.db.BeginTxx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "事务开启失败")
		return
	}
	defer tx.Rollback()
	if req.Password != "" {
		if !passwordRegexp.MatchString(req.Password) || req.Password != req.ConfirmPassword {
			writeError(w, http.StatusBadRequest, 400, "密码格式错误")
			return
		}
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			writeError(w, http.StatusInternalServerError, 500, "密码加密失败")
			return
		}
		user.PasswordHash = string(hash)
		newVersion++
	}
	if user.GroupID != req.GroupID {
		newVersion++
	}
	if _, err = tx.ExecContext(r.Context(), `
UPDATE admin_user
SET password_hash = ?, real_name = ?, phone = ?, group_id = ?, token_version = ?, updated_at = ?
WHERE id = ?
`, user.PasswordHash, req.RealName, req.Phone, req.GroupID, newVersion, a.now(), id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "员工编辑失败")
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "员工编辑失败")
		return
	}
	if newVersion != user.TokenVersion {
		_ = a.removeAllUserSessions(r.Context(), id)
	}
	a.writeOperation(r.Context(), actor, fmt.Sprintf("编辑员工：%s", user.Username), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleUserDelete(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "员工ID错误")
		return
	}
	user, err := a.getUserByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, 404, "员工不存在")
		return
	}
	if user.GroupID == 0 {
		writeError(w, http.StatusForbidden, 403, "超级管理员不允许删除")
		return
	}
	now := a.now()
	_, err = a.db.ExecContext(r.Context(), `
UPDATE admin_user
SET username = ?, status = 0, is_deleted = 1, deleted_at = ?, token_version = token_version + 1, updated_at = ?
WHERE id = ?
`, deletedUsername(user.Username, now), now, now, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "删除员工失败")
		return
	}
	_ = a.removeAllUserSessions(r.Context(), id)
	a.writeOperation(r.Context(), actor, fmt.Sprintf("删除员工：%s", restoreUsername(user.Username)), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleUserRestore(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "员工ID错误")
		return
	}
	user, err := a.getUserByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, 404, "员工不存在")
		return
	}
	if user.IsDeleted != 1 {
		writeError(w, http.StatusConflict, 409, "员工不在回收站")
		return
	}
	original := restoreUsername(user.Username)
	if exists, _ := a.activeUsernameExists(r.Context(), original, id); exists {
		writeError(w, http.StatusConflict, 409, "用户名已被占用，无法恢复")
		return
	}
	_, err = a.db.ExecContext(r.Context(), `
UPDATE admin_user
SET username = ?, is_deleted = 0, status = 0, deleted_at = NULL, token_version = token_version + 1, updated_at = ?
WHERE id = ?
`, original, a.now(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "恢复员工失败")
		return
	}
	_ = a.removeAllUserSessions(r.Context(), id)
	a.writeOperation(r.Context(), actor, fmt.Sprintf("恢复员工：%s", original), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleUserStatus(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "员工ID错误")
		return
	}
	user, err := a.getUserByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, 404, "员工不存在")
		return
	}
	if user.GroupID == 0 {
		writeError(w, http.StatusForbidden, 403, "超级管理员不允许禁用")
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err = decodeJSON(r, &req); err != nil || (req.Status != 0 && req.Status != 1) {
		writeError(w, http.StatusBadRequest, 400, "状态错误")
		return
	}
	if req.Status == 0 {
		_, err = a.db.ExecContext(r.Context(), `UPDATE admin_user SET status = 0, token_version = token_version + 1, updated_at = ? WHERE id = ?`, a.now(), id)
		_ = a.removeAllUserSessions(r.Context(), id)
	} else {
		_, err = a.db.ExecContext(r.Context(), `UPDATE admin_user SET status = 1, updated_at = ? WHERE id = ?`, a.now(), id)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "状态更新失败")
		return
	}
	a.writeOperation(r.Context(), actor, fmt.Sprintf("切换员工状态：%s -> %d", user.Username, req.Status), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleUserNotify(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "员工ID错误")
		return
	}
	var req struct {
		BalanceNotify int `json:"balance_notify"`
	}
	if err = decodeJSON(r, &req); err != nil || (req.BalanceNotify != 0 && req.BalanceNotify != 1) {
		writeError(w, http.StatusBadRequest, 400, "余额通知值错误")
		return
	}
	if _, err = a.db.ExecContext(r.Context(), `UPDATE admin_user SET balance_notify = ?, updated_at = ? WHERE id = ?`, req.BalanceNotify, a.now(), id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "余额通知更新失败")
		return
	}
	a.writeOperation(r.Context(), actor, fmt.Sprintf("切换余额通知：%d", id), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleUserSetBusiness(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	a.handleUserBusiness(w, r, actor, 1)
}

func (a *Application) handleUserCancelBusiness(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	a.handleUserBusiness(w, r, actor, 0)
}

func (a *Application) handleUserBusiness(w http.ResponseWriter, r *http.Request, actor AdminUser, flag int) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil || len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, 400, "ID列表不能为空")
		return
	}
	query, args, err := sqlx.In(`UPDATE admin_user SET is_business = ?, updated_at = ? WHERE id IN (?)`, flag, a.now(), req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "批量更新失败")
		return
	}
	query = a.db.Rebind(query)
	if _, err = a.db.ExecContext(r.Context(), query, args...); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "批量更新失败")
		return
	}
	action := "批量取消商务"
	if flag == 1 {
		action = "批量设置商务"
	}
	a.writeOperation(r.Context(), actor, action, requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}
