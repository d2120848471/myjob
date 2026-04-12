package kernel

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	"myjob/internal/library/region"
	modelruntime "myjob/internal/model/runtime"
)

const smsConfigCacheVersion = 2

func ParsePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func AppendTimeRangeFilters(startTime, endTime string, conditions *[]string, args *[]any) error {
	if strings.TrimSpace(startTime) != "" {
		parsed, err := ParseQueryTime(startTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at >= ?")
		*args = append(*args, parsed)
	}
	if strings.TrimSpace(endTime) != "" {
		parsed, err := ParseQueryTime(endTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at <= ?")
		*args = append(*args, parsed)
	}
	return nil
}

func ParseQueryTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
}

func ContainsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func MaskPhone(phone string) string     { return region.MaskPhone(phone) }
func MaskSecret(value string) string    { return region.MaskSecret(value) }
func MaskAccessKey(value string) string { return region.MaskAccessKey(value) }

func DeletedUsername(username string, now time.Time) string {
	return username + "__deleted_" + now.Format("20060102150405")
}

func RestoreUsername(username string) string {
	if idx := strings.Index(username, "__deleted_"); idx > 0 {
		return username[:idx]
	}
	return username
}

func BuildMenuTree(items []AdminMenu, parentID int64) []*AdminMenu {
	grouped := make(map[int64][]*AdminMenu)
	for i := range items {
		item := items[i]
		grouped[item.ParentID] = append(grouped[item.ParentID], &item)
	}
	var walk func(int64) []*AdminMenu
	walk = func(pid int64) []*AdminMenu {
		current := grouped[pid]
		sort.Slice(current, func(i, j int) bool {
			if current[i].Sort == current[j].Sort {
				return current[i].ID < current[j].ID
			}
			return current[i].Sort < current[j].Sort
		})
		for _, node := range current {
			node.Children = walk(node.ID)
		}
		return current
	}
	return walk(parentID)
}

func (c *Core) ResolveRegion(ip string) string {
	if c.regionResolver == nil {
		return ""
	}
	return c.regionResolver.Resolve(ip)
}

func (c *Core) BuildLoginUser(ctx context.Context, user AdminUser) map[string]any {
	groupName := "超级管理员"
	if user.GroupID != 0 {
		if group, err := c.GetGroupByID(ctx, user.GroupID); err == nil {
			groupName = group.Name
		}
	}
	return map[string]any{"id": user.ID, "username": user.Username, "real_name": user.RealName, "group_id": user.GroupID, "group_name": groupName, "is_business": user.IsBusiness}
}

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
	payload := modelruntime.SessionPayload{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: jti, ExpiresAt: now.Add(ttl)}
	if err = authlib.SaveSession(ctx, c.redis, payload, ttl); err != nil {
		return "", nil, err
	}
	return tokenString, perms, nil
}

func SMSVerifyReason(user AdminUser, ip string) string {
	if strings.TrimSpace(user.LastLoginIP) == "" {
		return "first_login"
	}
	if strings.TrimSpace(user.LastLoginIP) != strings.TrimSpace(ip) {
		return "ip_changed"
	}
	return ""
}

func (c *Core) SaveTempLogin(ctx context.Context, loginToken string, payload modelruntime.TempLoginPayload) error {
	data, _ := json.Marshal(payload)
	return c.redis.Set(ctx, authlib.TempLoginKey(loginToken), data, time.Duration(c.cfg.Auth.TempLoginTTLMin)*time.Minute).Err()
}

func (c *Core) GetTempLogin(ctx context.Context, loginToken string) (modelruntime.TempLoginPayload, error) {
	raw, err := c.redis.Get(ctx, authlib.TempLoginKey(loginToken)).Result()
	if err != nil {
		return modelruntime.TempLoginPayload{}, err
	}
	var payload modelruntime.TempLoginPayload
	err = json.Unmarshal([]byte(raw), &payload)
	return payload, err
}

func (c *Core) LoadPermissions(ctx context.Context, groupID int64) ([]string, error) {
	if groupID == 0 {
		arr, err := c.db.GetCore().GetArray(ctx, `SELECT code FROM admin_menu WHERE code <> '' AND status = 1 ORDER BY sort ASC, id ASC`)
		if err != nil {
			return nil, err
		}
		perms := make([]string, 0, len(arr))
		for _, item := range arr {
			perms = append(perms, item.String())
		}
		return perms, nil
	}
	if cached, err := c.redis.Get(ctx, authlib.PermissionCacheKey(groupID)).Result(); err == nil {
		perms := make([]string, 0)
		if json.Unmarshal([]byte(cached), &perms) == nil {
			return perms, nil
		}
	}
	arr, err := c.db.GetCore().GetArray(ctx, `
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
	_ = c.redis.Set(ctx, authlib.PermissionCacheKey(groupID), data, 30*time.Minute).Err()
	return perms, nil
}

func (c *Core) LoadSMSConfig(ctx context.Context) (modelruntime.SMSConfig, error) {
	state, err := c.LoadSMSConfigState(ctx)
	if err != nil {
		return modelruntime.SMSConfig{}, err
	}
	return state.Config, nil
}

func (c *Core) LoadSMSConfigState(ctx context.Context) (smsConfigState, error) {
	if cached, err := c.redis.Get(ctx, authlib.SMSConfigCacheKey()).Result(); err == nil {
		var state smsConfigState
		if json.Unmarshal([]byte(cached), &state) == nil && state.Version == smsConfigCacheVersion {
			return state, nil
		}
	}
	rows, err := c.db.GetCore().GetAll(ctx, `SELECT config_key, config_value, updated_at FROM system_config WHERE config_key LIKE 'sms_%'`)
	if err != nil {
		return smsConfigState{}, err
	}
	state := smsConfigState{Version: smsConfigCacheVersion, Config: modelruntime.SMSConfig{SignName: "玖权益", TemplateCode: "SMS_000001", ExpireMinutes: 30, IntervalMinutes: 1}}
	for _, row := range rows {
		key := row["config_key"].String()
		value := strings.TrimSpace(row["config_value"].String())
		switch key {
		case "sms_access_key":
			state.Config.AccessKey = value
			state.AccessKeyConfigured = value != ""
		case "sms_access_key_secret":
			state.Config.AccessKeySecret = value
			state.AccessKeySecretConfigured = value != ""
		case "sms_sign_name":
			if value != "" {
				state.Config.SignName = value
			}
		case "sms_template_code":
			if value != "" {
				state.Config.TemplateCode = value
			}
		case "sms_expire_minutes":
			if minutes, err := strconv.Atoi(value); err == nil && minutes > 0 {
				state.Config.ExpireMinutes = minutes
			}
		case "sms_interval_minutes":
			if minutes, err := strconv.Atoi(value); err == nil && minutes > 0 {
				state.Config.IntervalMinutes = minutes
			}
		}
		if updatedAt, ok := parseConfigUpdatedAt(row["updated_at"].Val()); ok && (state.UpdatedAt.IsZero() || updatedAt.After(state.UpdatedAt)) {
			state.UpdatedAt = updatedAt
		}
	}
	data, _ := json.Marshal(state)
	_ = c.redis.Set(ctx, authlib.SMSConfigCacheKey(), data, 30*time.Minute).Err()
	return state, nil
}

func parseConfigUpdatedAt(raw any) (time.Time, bool) {
	switch value := raw.(type) {
	case time.Time:
		if value.IsZero() {
			return time.Time{}, false
		}
		return value, true
	case string:
		return parseConfigUpdatedAtString(value)
	case []byte:
		return parseConfigUpdatedAtString(string(value))
	case sql.NullTime:
		if !value.Valid {
			return time.Time{}, false
		}
		return value.Time, true
	default:
		return time.Time{}, false
	}
}

func parseConfigUpdatedAtString(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func (c *Core) UpdateLoginState(ctx context.Context, userID int64, ip string) error {
	_, err := c.db.Exec(ctx, `UPDATE admin_user SET last_login_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`, ip, c.now(), c.now(), userID)
	return err
}

func (c *Core) InsertLoginLog(ctx context.Context, userID int64, adminName, ip string) error {
	_, err := c.db.Exec(ctx, `INSERT INTO admin_login_log (admin_id, admin_name, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?)`, userID, adminName, ip, c.ResolveRegion(ip), c.now())
	return err
}

func (c *Core) insertOperationLog(ctx context.Context, evt modelruntime.OperationEvent) error {
	_, err := c.db.Exec(ctx, `INSERT INTO admin_operation_log (admin_id, admin_name, description, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?, ?)`, evt.AdminID, evt.AdminName, evt.Description, evt.IP, evt.IPRegion, c.now())
	return err
}

func (c *Core) WriteOperation(ctx context.Context, actor AdminUser, desc, ip string) {
	evt := modelruntime.OperationEvent{AdminID: actor.ID, AdminName: actor.RealName, Description: desc, IP: ip, IPRegion: c.ResolveRegion(ip)}
	if c.auditWriter != nil {
		c.auditWriter.Write(ctx, evt)
		return
	}
	_ = c.insertOperationLog(ctx, evt)
}

func (c *Core) RemoveSession(ctx context.Context, jti string, userID int64) error {
	if err := c.redis.Del(ctx, authlib.SessionKey(jti)).Err(); err != nil {
		return err
	}
	return c.redis.SRem(ctx, authlib.UserSessionsKey(userID), jti).Err()
}

func (c *Core) RemoveAllUserSessions(ctx context.Context, userID int64) error {
	sessions, err := c.redis.SMembers(ctx, authlib.UserSessionsKey(userID)).Result()
	if err == nil {
		for _, jti := range sessions {
			_ = c.redis.Del(ctx, authlib.SessionKey(jti)).Err()
		}
	}
	return c.redis.Del(ctx, authlib.UserSessionsKey(userID)).Err()
}

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

func (c *Core) ActiveUsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	count, err := c.db.GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE username = ? AND is_deleted = 0 AND id <> ?`, username, excludeID)
	return count.Int() > 0, err
}

func (c *Core) ActivePhoneExists(ctx context.Context, phone string, excludeID int64) (bool, error) {
	count, err := c.db.GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE phone = ? AND is_deleted = 0 AND id <> ?`, phone, excludeID)
	return count.Int() > 0, err
}

func (c *Core) GetUserByUsername(ctx context.Context, username string) (AdminUser, error) {
	var user AdminUser
	err := c.db.GetCore().GetScan(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE username = ? LIMIT 1
`, username)
	return user, err
}

func (c *Core) GetUserByID(ctx context.Context, id int64) (AdminUser, error) {
	var user AdminUser
	err := c.db.GetCore().GetScan(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE id = ? LIMIT 1
`, id)
	return user, err
}

func (c *Core) GetGroupByID(ctx context.Context, id int64) (AdminGroup, error) {
	var group AdminGroup
	err := c.db.GetCore().GetScan(ctx, &group, `SELECT id, name, description, status, created_at, updated_at FROM admin_group WHERE id = ? LIMIT 1`, id)
	return group, err
}
