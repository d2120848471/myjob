package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"myjob/internal/consts"
)

// DeletedUsername 返回软删除时写回数据库的用户名（追加后缀避免唯一索引冲突）。
func DeletedUsername(username string, now time.Time) string {
	return username + "__deleted_" + now.Format("20060102150405")
}

// RestoreUsername 将软删除的用户名还原为原始用户名（去掉 "__deleted_" 后缀）。
func RestoreUsername(username string) string {
	if idx := strings.Index(username, "__deleted_"); idx > 0 {
		return username[:idx]
	}
	return username
}

// EnsureGroupActive 校验用户组是否存在且启用（用于创建/编辑用户等入口校验）。
func (c *Core) EnsureGroupActive(ctx context.Context, groupID int64) error {
	if groupID <= 0 {
		return errors.New("用户组错误")
	}
	group, err := c.GetGroupByID(ctx, groupID)
	if err != nil {
		return errors.New("用户组不存在")
	}
	if group.Status != consts.StatusEnabled {
		return errors.New("用户组不存在或已禁用")
	}
	return nil
}

// ActiveUsernameExists 判断用户名在未删除用户中是否已存在（可排除指定 id）。
func (c *Core) ActiveUsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	count, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE username = ? AND is_deleted = 0 AND id <> ?`, username, excludeID)
	return count.Int() > 0, err
}

// ActivePhoneExists 判断手机号在未删除用户中是否已存在（可排除指定 id）。
func (c *Core) ActivePhoneExists(ctx context.Context, phone string, excludeID int64) (bool, error) {
	count, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE phone = ? AND is_deleted = 0 AND id <> ?`, phone, excludeID)
	return count.Int() > 0, err
}

// GetUserByUsername 根据用户名查询用户记录。
func (c *Core) GetUserByUsername(ctx context.Context, username string) (AdminUser, error) {
	var user AdminUser
	err := c.DB().GetCore().GetScan(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE username = ? LIMIT 1
`, username)
	return user, err
}

// GetUserByID 根据用户 id 查询用户记录。
func (c *Core) GetUserByID(ctx context.Context, id int64) (AdminUser, error) {
	var user AdminUser
	err := c.DB().GetCore().GetScan(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE id = ? LIMIT 1
`, id)
	return user, err
}

// GetGroupByID 根据用户组 id 查询用户组记录。
func (c *Core) GetGroupByID(ctx context.Context, id int64) (AdminGroup, error) {
	var group AdminGroup
	err := c.DB().GetCore().GetScan(ctx, &group, `SELECT id, name, description, status, created_at, updated_at FROM admin_group WHERE id = ? LIMIT 1`, id)
	return group, err
}
