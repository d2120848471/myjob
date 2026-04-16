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

// Add 新增员工账号，并写入操作日志。
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

// Edit 编辑员工信息；当密码或用户组变更时会提升 token_version 并踢下线历史会话。
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
	// token_version 变化表示旧 token 全量失效，需要删除该用户的所有 session。
	if newVersion != user.TokenVersion {
		_ = l.core.RemoveAllUserSessions(ctx, req.ID)
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑员工：%s", user.Username), ip)
	return &adminapi.UserUpdateRes{}, nil
}

// Delete 软删除员工（写入删除后缀用户名以规避唯一索引），并踢下线其所有会话。
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

// Restore 将员工从回收站恢复，并踢下线其所有会话（避免旧 token 继续使用）。
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

// Status 切换员工启用/禁用状态；禁用时会提升 token_version 并踢下线该用户所有会话。
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
		// 禁用时提升 token_version，确保旧 token 立即失效。
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
