package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"
)

// BuildLoginUser 构建登录返回用的用户信息（补齐用户组名称等展示字段）。
func (c *Core) BuildLoginUser(ctx context.Context, user AdminUser) modelruntime.LoginUser {
	groupName := "超级管理员"
	if user.GroupID != 0 {
		if group, err := c.GetGroupByID(ctx, user.GroupID); err == nil {
			groupName = group.Name
		}
	}
	return modelruntime.LoginUser{
		ID:         user.ID,
		Username:   user.Username,
		RealName:   user.RealName,
		GroupID:    user.GroupID,
		GroupName:  groupName,
		IsBusiness: user.IsBusiness,
	}
}

// IssueSession 为用户签发访问令牌并创建服务端 Session。
//
// 返回值包含：JWT token、当前用户可用的权限码列表（便于登录后一次性下发）。
func (c *Core) IssueSession(ctx context.Context, user AdminUser) (string, []string, error) {
	perms, err := c.LoadPermissions(ctx, user.GroupID)
	if err != nil {
		return "", nil, err
	}
	ttl := time.Duration(c.cfg.Auth.AccessTokenTTLMin) * time.Minute
	now := c.now()
	tokenString, jti, err := authlib.IssueToken(c.cfg.Auth.JWTSecret, modelruntime.SessionPayload{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion}, now, ttl)
	if err != nil {
		return "", nil, err
	}
	// SessionPayload 额外落 Redis，支持服务端主动失效与版本校验。
	payload := modelruntime.SessionPayload{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: jti, ExpiresAt: now.Add(ttl)}
	if err = authlib.SaveSession(ctx, c.Redis(), payload, ttl); err != nil {
		return "", nil, err
	}
	return tokenString, perms, nil
}

// SMSVerifyReason 判断该用户本次登录是否需要短信二次验证，并返回原因码。
//
// 目前规则：首次登录或登录 IP 变化触发二次验证。
func SMSVerifyReason(user AdminUser, ip string) string {
	if strings.TrimSpace(user.LastLoginIP) == "" {
		return "first_login"
	}
	if strings.TrimSpace(user.LastLoginIP) != strings.TrimSpace(ip) {
		return "ip_changed"
	}
	return ""
}

// SaveTempLogin 保存短信二次验证的临时登录态（带 TTL）。
func (c *Core) SaveTempLogin(ctx context.Context, loginToken string, payload modelruntime.TempLoginPayload) error {
	data, _ := json.Marshal(payload)
	return c.RedisSetString(ctx, authlib.TempLoginKey(loginToken), string(data), time.Duration(c.cfg.Auth.TempLoginTTLMin)*time.Minute)
}

// GetTempLogin 读取短信二次验证的临时登录态。
func (c *Core) GetTempLogin(ctx context.Context, loginToken string) (modelruntime.TempLoginPayload, error) {
	raw, err := c.RedisGetString(ctx, authlib.TempLoginKey(loginToken))
	if err != nil {
		return modelruntime.TempLoginPayload{}, err
	}
	var payload modelruntime.TempLoginPayload
	err = json.Unmarshal([]byte(raw), &payload)
	return payload, err
}

// LoadPermissions 加载用户组的权限码列表（带缓存）。
//
// - groupID=0（超级管理员）返回所有启用菜单的 code（包含 super_only）
// - 普通用户仅返回授权菜单的 code（过滤 super_only）
func (c *Core) LoadPermissions(ctx context.Context, groupID int64) ([]string, error) {
	if groupID == 0 {
		arr, err := c.DB().GetCore().GetArray(ctx, `SELECT code FROM admin_menu WHERE code <> '' AND status = 1 ORDER BY sort ASC, id ASC`)
		if err != nil {
			return nil, err
		}
		perms := make([]string, 0, len(arr))
		for _, item := range arr {
			perms = append(perms, item.String())
		}
		return perms, nil
	}
	// 普通用户权限结果会缓存到 Redis，减少每次鉴权的 JOIN 压力。
	if cached, err := c.RedisGetString(ctx, authlib.PermissionCacheKey(groupID)); err == nil {
		perms := make([]string, 0)
		if json.Unmarshal([]byte(cached), &perms) == nil {
			return perms, nil
		}
	}
	arr, err := c.DB().GetCore().GetArray(ctx, `
SELECT m.code
FROM admin_group_menu gm
JOIN admin_menu m ON m.id = gm.menu_id
WHERE gm.group_id = ? AND m.code <> '' AND m.status = 1 AND m.super_only = 0
ORDER BY m.sort ASC, m.id ASC
`, groupID)
	if err != nil {
		return nil, err
	}
	perms := make([]string, 0, len(arr))
	for _, item := range arr {
		perms = append(perms, item.String())
	}
	data, _ := json.Marshal(perms)
	_ = c.RedisSetString(ctx, authlib.PermissionCacheKey(groupID), string(data), 30*time.Minute)
	return perms, nil
}

// UpdateLoginState 更新用户最后登录 IP/时间。
func (c *Core) UpdateLoginState(ctx context.Context, userID int64, ip string) error {
	_, err := c.DB().Exec(ctx, `UPDATE admin_user SET last_login_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`, ip, c.now(), c.now(), userID)
	return err
}

// RemoveSession 删除单个会话（删除 session key，并从用户会话集合中移除 jti）。
func (c *Core) RemoveSession(ctx context.Context, jti string, userID int64) error {
	if _, err := c.Redis().GroupGeneric().Del(ctx, authlib.SessionKey(jti)); err != nil {
		return err
	}
	_, err := c.Redis().GroupSet().SRem(ctx, authlib.UserSessionsKey(userID), jti)
	return err
}

// RemoveAllUserSessions 删除用户的所有会话（用于改密码/改组等需要踢下线的场景）。
func (c *Core) RemoveAllUserSessions(ctx context.Context, userID int64) error {
	sessions, err := c.RedisSMembers(ctx, authlib.UserSessionsKey(userID))
	if err == nil {
		for _, jti := range sessions {
			_, _ = c.Redis().GroupGeneric().Del(ctx, authlib.SessionKey(jti))
		}
	}
	_, err = c.Redis().GroupGeneric().Del(ctx, authlib.UserSessionsKey(userID))
	return err
}
