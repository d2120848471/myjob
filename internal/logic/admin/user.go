package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
	"golang.org/x/crypto/bcrypt"
)

type UserLogic struct{ core *app.Core }

func (l *UserLogic) List(ctx context.Context, req *adminapi.UserListReq) (*adminapi.UserListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 0`)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工列表查询失败")
	}
	items := make([]app.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 0
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工列表查询失败")
	}
	return &adminapi.UserListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

func (l *UserLogic) Trash(ctx context.Context, req *adminapi.UserTrashReq) (*adminapi.UserTrashRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 1`)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "回收站查询失败")
	}
	items := make([]app.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 1
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "回收站查询失败")
	}
	return &adminapi.UserTrashRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

func (l *UserLogic) Add(ctx context.Context, req *adminapi.UserCreateReq, actor app.AdminUser, ip string) (*adminapi.UserCreateRes, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if !app.UsernameRegexp().MatchString(req.Username) || req.Username != req.ConfirmUsername {
		return nil, apiErr(consts.CodeBadRequest, "用户名格式错误")
	}
	if !app.PasswordRegexp().MatchString(req.Password) || req.Password != req.ConfirmPassword {
		return nil, apiErr(consts.CodeBadRequest, "密码格式错误")
	}
	if req.RealName == "" || !app.PhoneRegexp().MatchString(req.Phone) {
		return nil, apiErr(consts.CodeBadRequest, "手机号格式错误")
	}
	if err := l.core.EnsureGroupActive(ctx, req.GroupID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if exists, _ := l.core.ActiveUsernameExists(ctx, req.Username, 0); exists {
		return nil, apiErr(consts.CodeConflict, "用户名已存在")
	}
	if exists, _ := l.core.ActivePhoneExists(ctx, req.Phone, 0); exists {
		return nil, apiErr(consts.CodeConflict, "手机号已存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "密码加密失败")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_user (username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted, token_version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, 0, 0, 0, 0, ?, ?)`, req.Username, string(hash), req.RealName, req.Phone, req.GroupID, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加员工：%s，用户组：%d", req.Username, req.GroupID), ip)
	return &adminapi.UserCreateRes{ID: id}, nil
}

func (l *UserLogic) Edit(ctx context.Context, req *adminapi.UserUpdateReq, actor app.AdminUser, ip string) (*adminapi.UserUpdateRes, error) {
	user, err := l.core.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "员工不存在")
	}
	req.RealName = strings.TrimSpace(req.RealName)
	req.Phone = strings.TrimSpace(req.Phone)
	if req.RealName == "" || !app.PhoneRegexp().MatchString(req.Phone) {
		return nil, apiErr(consts.CodeBadRequest, "手机号格式错误")
	}
	if err = l.core.EnsureGroupActive(ctx, req.GroupID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if exists, _ := l.core.ActivePhoneExists(ctx, req.Phone, req.ID); exists {
		return nil, apiErr(consts.CodeConflict, "手机号已存在")
	}
	newVersion := user.TokenVersion
	if req.Password != "" {
		if !app.PasswordRegexp().MatchString(req.Password) || req.Password != req.ConfirmPassword {
			return nil, apiErr(consts.CodeBadRequest, "密码格式错误")
		}
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, apiErr(consts.CodeInternalError, "密码加密失败")
		}
		user.PasswordHash = string(hash)
		newVersion++
	}
	if user.GroupID != req.GroupID {
		newVersion++
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, txErr := tx.Exec(`UPDATE admin_user SET password_hash = ?, real_name = ?, phone = ?, group_id = ?, token_version = ?, updated_at = ? WHERE id = ?`, user.PasswordHash, req.RealName, req.Phone, req.GroupID, newVersion, l.core.Now(), req.ID)
		return txErr
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工编辑失败")
	}
	if newVersion != user.TokenVersion {
		_ = l.core.RemoveAllUserSessions(ctx, req.ID)
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑员工：%s", user.Username), ip)
	return &adminapi.UserUpdateRes{}, nil
}

func (l *UserLogic) Delete(ctx context.Context, req *adminapi.UserDeleteReq, actor app.AdminUser, ip string) (*adminapi.UserDeleteRes, error) {
	user, err := l.core.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "员工不存在")
	}
	if user.GroupID == 0 {
		return nil, apiErr(consts.CodeForbidden, "超级管理员不允许删除")
	}
	now := l.core.Now()
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET username = ?, status = 0, is_deleted = 1, deleted_at = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, app.DeletedUsername(user.Username, now), now, now, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "删除员工失败")
	}
	_ = l.core.RemoveAllUserSessions(ctx, req.ID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除员工：%s", app.RestoreUsername(user.Username)), ip)
	return &adminapi.UserDeleteRes{}, nil
}

func (l *UserLogic) Restore(ctx context.Context, req *adminapi.UserRestoreReq, actor app.AdminUser, ip string) (*adminapi.UserRestoreRes, error) {
	user, err := l.core.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "员工不存在")
	}
	if user.IsDeleted != 1 {
		return nil, apiErr(consts.CodeConflict, "员工不在回收站")
	}
	original := app.RestoreUsername(user.Username)
	if exists, _ := l.core.ActiveUsernameExists(ctx, original, req.ID); exists {
		return nil, apiErr(consts.CodeConflict, "用户名已被占用，无法恢复")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET username = ?, is_deleted = 0, status = 0, deleted_at = NULL, token_version = token_version + 1, updated_at = ? WHERE id = ?`, original, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "恢复员工失败")
	}
	_ = l.core.RemoveAllUserSessions(ctx, req.ID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("恢复员工：%s", original), ip)
	return &adminapi.UserRestoreRes{}, nil
}

func (l *UserLogic) Status(ctx context.Context, req *adminapi.UserStatusReq, actor app.AdminUser, ip string) (*adminapi.UserStatusRes, error) {
	user, err := l.core.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "员工不存在")
	}
	if user.GroupID == 0 {
		return nil, apiErr(consts.CodeForbidden, "超级管理员不允许禁用")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if req.Status == 0 {
		_, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET status = 0, token_version = token_version + 1, updated_at = ? WHERE id = ?`, l.core.Now(), req.ID)
		_ = l.core.RemoveAllUserSessions(ctx, req.ID)
	} else {
		_, err = l.core.DB().Exec(ctx, `UPDATE admin_user SET status = 1, updated_at = ? WHERE id = ?`, l.core.Now(), req.ID)
	}
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "状态更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换员工状态：%s -> %d", user.Username, req.Status), ip)
	return &adminapi.UserStatusRes{}, nil
}

func (l *UserLogic) Notify(ctx context.Context, req *adminapi.UserNotifyReq, actor app.AdminUser, ip string) (*adminapi.UserNotifyRes, error) {
	if req.BalanceNotify != 0 && req.BalanceNotify != 1 {
		return nil, apiErr(consts.CodeBadRequest, "余额通知值错误")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE admin_user SET balance_notify = ?, updated_at = ? WHERE id = ?`, req.BalanceNotify, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "余额通知更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换余额通知：%d", req.ID), ip)
	return &adminapi.UserNotifyRes{}, nil
}

func (l *UserLogic) SetBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq, actor app.AdminUser, ip string) (*adminapi.UserBusinessAssignRes, error) {
	if err := l.handleBusiness(ctx, req.IDs, actor, ip, 1); err != nil {
		return nil, err
	}
	return &adminapi.UserBusinessAssignRes{}, nil
}

func (l *UserLogic) CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq, actor app.AdminUser, ip string) (*adminapi.UserBusinessCancelRes, error) {
	if err := l.handleBusiness(ctx, req.IDs, actor, ip, 0); err != nil {
		return nil, err
	}
	return &adminapi.UserBusinessCancelRes{}, nil
}

func (l *UserLogic) handleBusiness(ctx context.Context, ids []int64, actor app.AdminUser, ip string, flag int) error {
	if len(ids) == 0 {
		return apiErr(consts.CodeBadRequest, "ID列表不能为空")
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+2)
	args = append(args, flag, l.core.Now())
	for _, id := range ids {
		args = append(args, id)
	}
	query := `UPDATE admin_user SET is_business = ?, updated_at = ? WHERE id IN (` + placeholders + `)`
	if _, err := l.core.DB().Exec(ctx, query, args...); err != nil {
		return apiErr(consts.CodeInternalError, "批量更新失败")
	}
	action := "批量取消商务"
	if flag == 1 {
		action = "批量设置商务"
	}
	l.core.WriteOperation(ctx, actor, action, ip)
	return nil
}
