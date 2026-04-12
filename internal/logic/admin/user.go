package adminlogic

import (
	"context"

	"fmt"
	"myjob/internal/kernel"
	"net/http"
	"strings"

	userv1 "myjob/api/admin/user/v1"
	"myjob/internal/model/dto/admin"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gdb"
	"golang.org/x/crypto/bcrypt"
)

type UserLogic struct{ core *kernel.Core }

func (l *UserLogic) List(ctx context.Context, req userv1.ListReq) (map[string]any, *modelruntime.APIError) {
	page, pageSize := kernel.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 0`)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "员工列表查询失败")
	}
	items := make([]kernel.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 0
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "员工列表查询失败")
	}
	return map[string]any{"list": items, "pagination": admin.Pagination{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

func (l *UserLogic) Trash(ctx context.Context, req userv1.ListReq) (map[string]any, *modelruntime.APIError) {
	page, pageSize := kernel.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 1`)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "回收站查询失败")
	}
	items := make([]kernel.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 1
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "回收站查询失败")
	}
	return map[string]any{"list": items, "pagination": admin.Pagination{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

func (l *UserLogic) Add(ctx context.Context, req userv1.AddReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	req.Username = strings.TrimSpace(req.Username)
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if !kernel.UsernameRegexp().MatchString(req.Username) || req.Username != req.ConfirmUsername {
		return nil, apiErr(http.StatusBadRequest, 400, "用户名格式错误")
	}
	if !kernel.PasswordRegexp().MatchString(req.Password) || req.Password != req.ConfirmPassword {
		return nil, apiErr(http.StatusBadRequest, 400, "密码格式错误")
	}
	if req.RealName == "" || !kernel.PhoneRegexp().MatchString(req.Phone) {
		return nil, apiErr(http.StatusBadRequest, 400, "手机号格式错误")
	}
	if err := l.core.EnsureGroupActive(ctx, req.GroupID); err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, err.Error())
	}
	if exists, _ := l.core.ActiveUsernameExists(ctx, req.Username, 0); exists {
		return nil, apiErr(http.StatusConflict, 409, "用户名已存在")
	}
	if exists, _ := l.core.ActivePhoneExists(ctx, req.Phone, 0); exists {
		return nil, apiErr(http.StatusConflict, 409, "手机号已存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "密码加密失败")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_user (username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted, token_version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, 0, 0, 0, 0, ?, ?)`, req.Username, string(hash), req.RealName, req.Phone, req.GroupID, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "员工新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加员工：%s，用户组：%d", req.Username, req.GroupID), ip)
	return map[string]any{"id": id}, nil
}

func (l *UserLogic) Edit(ctx context.Context, id int64, req userv1.EditReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	user, err := l.core.GetUserByID(ctx, id)
	if err != nil {
		return nil, apiErr(http.StatusNotFound, 404, "员工不存在")
	}
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if req.RealName == "" || !kernel.PhoneRegexp().MatchString(req.Phone) {
		return nil, apiErr(http.StatusBadRequest, 400, "手机号格式错误")
	}
	if err = l.core.EnsureGroupActive(ctx, req.GroupID); err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, err.Error())
	}
	if exists, _ := l.core.ActivePhoneExists(ctx, req.Phone, id); exists {
		return nil, apiErr(http.StatusConflict, 409, "手机号已存在")
	}
	newVersion := user.TokenVersion
	if req.Password != "" {
		if !kernel.PasswordRegexp().MatchString(req.Password) || req.Password != req.ConfirmPassword {
			return nil, apiErr(http.StatusBadRequest, 400, "密码格式错误")
		}
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, apiErr(http.StatusInternalServerError, 500, "密码加密失败")
		}
		user.PasswordHash = string(hash)
		newVersion++
	}
	if user.GroupID != req.GroupID {
		newVersion++
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, txErr := tx.Exec(`UPDATE admin_user SET password_hash = ?, real_name = ?, phone = ?, group_id = ?, token_version = ?, updated_at = ? WHERE id = ?`, user.PasswordHash, req.RealName, req.Phone, req.GroupID, newVersion, l.core.Now(), id)
		return txErr
	}); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "员工编辑失败")
	}
	if newVersion != user.TokenVersion {
		_ = l.core.RemoveAllUserSessions(ctx, id)
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑员工：%s", user.Username), ip)
	return map[string]any{}, nil
}

func (l *UserLogic) Delete(ctx context.Context, id int64, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	user, err := l.core.GetUserByID(ctx, id)
	if err != nil {
		return nil, apiErr(http.StatusNotFound, 404, "员工不存在")
	}
	if user.GroupID == 0 {
		return nil, apiErr(http.StatusForbidden, 403, "超级管理员不允许删除")
	}
	now := l.core.Now()
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET username = ?, status = 0, is_deleted = 1, deleted_at = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, kernel.DeletedUsername(user.Username, now), now, now, id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "删除员工失败")
	}
	_ = l.core.RemoveAllUserSessions(ctx, id)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除员工：%s", kernel.RestoreUsername(user.Username)), ip)
	return map[string]any{}, nil
}

func (l *UserLogic) Restore(ctx context.Context, id int64, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	user, err := l.core.GetUserByID(ctx, id)
	if err != nil {
		return nil, apiErr(http.StatusNotFound, 404, "员工不存在")
	}
	if user.IsDeleted != 1 {
		return nil, apiErr(http.StatusConflict, 409, "员工不在回收站")
	}
	original := kernel.RestoreUsername(user.Username)
	if exists, _ := l.core.ActiveUsernameExists(ctx, original, id); exists {
		return nil, apiErr(http.StatusConflict, 409, "用户名已被占用，无法恢复")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET username = ?, is_deleted = 0, status = 0, deleted_at = NULL, token_version = token_version + 1, updated_at = ? WHERE id = ?`, original, l.core.Now(), id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "恢复员工失败")
	}
	_ = l.core.RemoveAllUserSessions(ctx, id)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("恢复员工：%s", original), ip)
	return map[string]any{}, nil
}

func (l *UserLogic) Status(ctx context.Context, id int64, req userv1.StatusReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	user, err := l.core.GetUserByID(ctx, id)
	if err != nil {
		return nil, apiErr(http.StatusNotFound, 404, "员工不存在")
	}
	if user.GroupID == 0 {
		return nil, apiErr(http.StatusForbidden, 403, "超级管理员不允许禁用")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(http.StatusBadRequest, 400, "状态错误")
	}
	if req.Status == 0 {
		_, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET status = 0, token_version = token_version + 1, updated_at = ? WHERE id = ?`, l.core.Now(), id)
		_ = l.core.RemoveAllUserSessions(ctx, id)
	} else {
		_, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET status = 1, updated_at = ? WHERE id = ?`, l.core.Now(), id)
	}
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "状态更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换员工状态：%s -> %d", user.Username, req.Status), ip)
	return map[string]any{}, nil
}

func (l *UserLogic) Notify(ctx context.Context, id int64, req userv1.NotifyReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	if req.BalanceNotify != 0 && req.BalanceNotify != 1 {
		return nil, apiErr(http.StatusBadRequest, 400, "余额通知值错误")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE admin_user SET balance_notify = ?, updated_at = ? WHERE id = ?`, req.BalanceNotify, l.core.Now(), id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "余额通知更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换余额通知：%d", id), ip)
	return map[string]any{}, nil
}

func (l *UserLogic) SetBusiness(ctx context.Context, req userv1.BusinessReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	return l.handleBusiness(ctx, req, actor, ip, 1)
}

func (l *UserLogic) CancelBusiness(ctx context.Context, req userv1.BusinessReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	return l.handleBusiness(ctx, req, actor, ip, 0)
}

func (l *UserLogic) handleBusiness(ctx context.Context, req userv1.BusinessReq, actor kernel.AdminUser, ip string, flag int) (map[string]any, *modelruntime.APIError) {
	if len(req.IDs) == 0 {
		return nil, apiErr(http.StatusBadRequest, 400, "ID列表不能为空")
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(req.IDs)), ",")
	args := make([]any, 0, len(req.IDs)+2)
	args = append(args, flag, l.core.Now())
	for _, id := range req.IDs {
		args = append(args, id)
	}
	query := `UPDATE admin_user SET is_business = ?, updated_at = ? WHERE id IN (` + placeholders + `)`
	if _, err := l.core.DB().Exec(ctx, query, args...); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "批量更新失败")
	}
	action := "批量取消商务"
	if flag == 1 {
		action = "批量设置商务"
	}
	l.core.WriteOperation(ctx, actor, action, ip)
	return map[string]any{}, nil
}
