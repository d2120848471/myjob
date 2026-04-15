package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	"myjob/internal/library/region"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
)

const smsConfigCacheVersion = 2

// ParsePagination 将分页参数归一化到可用范围内。
//
// - page <= 0 视为 1
// - pageSize <= 0 视为 20
// - pageSize 最大为 100
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

// AppendTimeRangeFilters 根据 startTime/endTime 追加 created_at 的筛选条件与参数。
//
// 入参时间格式要求为 "2006-01-02 15:04:05"（与 ParseQueryTime 一致）。
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

// ParseQueryTime 解析查询参数中的时间字符串（本地时区）。
func ParseQueryTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
}

// ContainsString 判断切片 items 是否包含目标字符串。
func ContainsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// MaskPhone 对手机号进行脱敏展示（保留前后若干位）。
func MaskPhone(phone string) string { return region.MaskPhone(phone) }

// MaskSecret 对密钥/口令类敏感信息进行脱敏展示。
func MaskSecret(value string) string { return region.MaskSecret(value) }

// MaskAccessKey 对 access_key 类敏感信息进行脱敏展示。
func MaskAccessKey(value string) string { return region.MaskAccessKey(value) }

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

// BuildMenuTree 将扁平菜单列表构建为树结构（按 sort/id 排序）。
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

// ResolveRegion 解析 IP 的归属地文本（用于日志/审计展示）。
func (c *Core) ResolveRegion(ip string) string {
	if c.regionResolver == nil {
		return ""
	}
	return c.regionResolver.Resolve(ip)
}

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

// LoadSMSConfig 加载短信配置（只返回配置结构体，不包含配置是否完整的标记）。
func (c *Core) LoadSMSConfig(ctx context.Context) (modelruntime.SMSConfig, error) {
	state, err := c.LoadSMSConfigState(ctx)
	if err != nil {
		return modelruntime.SMSConfig{}, err
	}
	return state.Config, nil
}

// LoadSMSConfigState 加载短信配置并返回带“是否已配置”的状态信息（带缓存）。
func (c *Core) LoadSMSConfigState(ctx context.Context) (smsConfigState, error) {
	if cached, err := c.RedisGetString(ctx, authlib.SMSConfigCacheKey()); err == nil {
		var state smsConfigState
		if json.Unmarshal([]byte(cached), &state) == nil && state.Version == smsConfigCacheVersion {
			return state, nil
		}
	}
	rows, err := c.DB().GetCore().GetAll(ctx, `SELECT config_key, config_value, updated_at FROM system_config WHERE config_key LIKE 'sms_%'`)
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
	_ = c.RedisSetString(ctx, authlib.SMSConfigCacheKey(), string(data), 30*time.Minute)
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
	case interface{ String() string }:
		return parseConfigUpdatedAtString(value.String())
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

// UpdateLoginState 更新用户最后登录 IP/时间。
func (c *Core) UpdateLoginState(ctx context.Context, userID int64, ip string) error {
	_, err := c.DB().Exec(ctx, `UPDATE admin_user SET last_login_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`, ip, c.now(), c.now(), userID)
	return err
}

// InsertLoginLog 写入一条登录日志（包含 IP 归属地）。
func (c *Core) InsertLoginLog(ctx context.Context, userID int64, adminName, ip string) error {
	_, err := c.DB().Exec(ctx, `INSERT INTO admin_login_log (admin_id, admin_name, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?)`, userID, adminName, ip, c.ResolveRegion(ip), c.now())
	return err
}

func (c *Core) insertOperationLog(ctx context.Context, evt modelruntime.OperationEvent) error {
	_, err := c.DB().Exec(ctx, `INSERT INTO admin_operation_log (admin_id, admin_name, description, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?, ?)`, evt.AdminID, evt.AdminName, evt.Description, evt.IP, evt.IPRegion, c.now())
	return err
}

// WriteOperation 写入一条操作审计事件（优先写入 auditWriter，以支持异步/缓冲）。
func (c *Core) WriteOperation(ctx context.Context, actor AdminUser, desc, ip string) {
	evt := modelruntime.OperationEvent{AdminID: actor.ID, AdminName: actor.RealName, Description: desc, IP: ip, IPRegion: c.ResolveRegion(ip)}
	if c.auditWriter != nil {
		c.auditWriter.Write(ctx, evt)
		return
	}
	_ = c.insertOperationLog(ctx, evt)
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

// RedisSetString 将字符串写入 Redis，并设置过期时间。
//
// ttl <= 0 时会强制设置为 1s，避免出现永久 key。
func (c *Core) RedisSetString(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second
	}
	seconds := int64(math.Ceil(ttl.Seconds()))
	_, err := c.Redis().GroupString().Set(ctx, key, value, gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
	})
	return err
}

// RedisGetString 从 Redis 读取字符串；当 key 不存在时返回 sql.ErrNoRows。
func (c *Core) RedisGetString(ctx context.Context, key string) (string, error) {
	value, err := c.Redis().GroupString().Get(ctx, key)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() {
		return "", sql.ErrNoRows
	}
	return value.String(), nil
}

// RedisTTL 读取 Redis key 的剩余 TTL。
func (c *Core) RedisTTL(ctx context.Context, key string) (time.Duration, error) {
	seconds, err := c.Redis().GroupGeneric().TTL(ctx, key)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

// RedisSMembers 读取 Redis set 中的所有成员并转换为字符串切片。
func (c *Core) RedisSMembers(ctx context.Context, key string) ([]string, error) {
	values, err := c.Redis().GroupSet().SMembers(ctx, key)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result, nil
}
