package app

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (a *Application) buildLoginUser(ctx context.Context, user AdminUser) map[string]interface{} {
	groupName := "超级管理员"
	if user.GroupID != 0 {
		if group, err := a.getGroupByID(ctx, user.GroupID); err == nil {
			groupName = group.Name
		}
	}
	return map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"real_name":   user.RealName,
		"group_id":    user.GroupID,
		"group_name":  groupName,
		"is_business": user.IsBusiness,
	}
}

func (a *Application) issueSession(ctx context.Context, user AdminUser) (string, []string, error) {
	perms, err := a.loadPermissions(ctx, user.GroupID)
	if err != nil {
		return "", nil, err
	}
	jti := uuid.NewString()
	expiresAt := a.now().Add(time.Duration(a.cfg.Auth.AccessTokenTTLMin) * time.Minute)
	claims := jwtClaims{
		UserID:       user.ID,
		GroupID:      user.GroupID,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(a.now()),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.cfg.Auth.JWTSecret))
	if err != nil {
		return "", nil, err
	}
	payload, _ := json.Marshal(sessionPayload{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: jti, ExpiresAt: expiresAt})
	if err = a.redis.Set(ctx, sessionKey(jti), payload, time.Until(expiresAt)).Err(); err != nil {
		return "", nil, err
	}
	if err = a.redis.SAdd(ctx, userSessionsKey(user.ID), jti).Err(); err != nil {
		return "", nil, err
	}
	_ = a.redis.Expire(ctx, userSessionsKey(user.ID), time.Until(expiresAt)).Err()
	return tokenString, perms, nil
}

func smsVerifyReason(user AdminUser, ip string) string {
	if strings.TrimSpace(user.LastLoginIP) == "" {
		return "first_login"
	}
	if strings.TrimSpace(user.LastLoginIP) != strings.TrimSpace(ip) {
		return "ip_changed"
	}
	return ""
}

func (a *Application) saveTempLogin(ctx context.Context, loginToken string, payload tempLoginPayload) error {
	data, _ := json.Marshal(payload)
	return a.redis.Set(ctx, tempLoginKey(loginToken), data, time.Duration(a.cfg.Auth.TempLoginTTLMin)*time.Minute).Err()
}

func (a *Application) getTempLogin(ctx context.Context, loginToken string) (tempLoginPayload, error) {
	raw, err := a.redis.Get(ctx, tempLoginKey(loginToken)).Result()
	if err != nil {
		return tempLoginPayload{}, err
	}
	var payload tempLoginPayload
	err = json.Unmarshal([]byte(raw), &payload)
	return payload, err
}

func (a *Application) loadPermissions(ctx context.Context, groupID int64) ([]string, error) {
	if groupID == 0 {
		codes := make([]string, 0)
		if err := a.db.SelectContext(ctx, &codes, `SELECT code FROM admin_menu WHERE code <> '' AND status = 1 ORDER BY sort ASC, id ASC`); err != nil {
			return nil, err
		}
		return codes, nil
	}
	if cached, err := a.redis.Get(ctx, permissionCacheKey(groupID)).Result(); err == nil {
		perms := make([]string, 0)
		if json.Unmarshal([]byte(cached), &perms) == nil {
			return perms, nil
		}
	}
	perms := make([]string, 0)
	if err := a.db.SelectContext(ctx, &perms, `
SELECT m.code
FROM admin_group_menu gm
JOIN admin_menu m ON m.id = gm.menu_id
WHERE gm.group_id = ? AND m.code <> '' AND m.status = 1 AND m.super_only = 0
ORDER BY m.sort ASC, m.id ASC
`, groupID); err != nil {
		return nil, err
	}
	data, _ := json.Marshal(perms)
	_ = a.redis.Set(ctx, permissionCacheKey(groupID), data, 30*time.Minute).Err()
	return perms, nil
}

func (a *Application) loadSMSConfig(ctx context.Context) (SMSConfig, error) {
	if cached, err := a.redis.Get(ctx, smsConfigCacheKey()).Result(); err == nil {
		var cfg SMSConfig
		if json.Unmarshal([]byte(cached), &cfg) == nil {
			return cfg, nil
		}
	}
	rows := make([]struct {
		ConfigKey   string `db:"config_key"`
		ConfigValue string `db:"config_value"`
	}, 0)
	if err := a.db.SelectContext(ctx, &rows, `SELECT config_key, config_value FROM system_config WHERE config_key LIKE 'sms_%'`); err != nil {
		return SMSConfig{}, err
	}
	cfg := SMSConfig{}
	for _, row := range rows {
		switch row.ConfigKey {
		case "sms_access_key":
			cfg.AccessKey = row.ConfigValue
		case "sms_access_key_secret":
			cfg.AccessKeySecret = row.ConfigValue
		case "sms_sign_name":
			cfg.SignName = row.ConfigValue
		case "sms_template_code":
			cfg.TemplateCode = row.ConfigValue
		case "sms_expire_minutes":
			cfg.ExpireMinutes, _ = strconv.Atoi(row.ConfigValue)
		case "sms_interval_minutes":
			cfg.IntervalMinutes, _ = strconv.Atoi(row.ConfigValue)
		}
	}
	if cfg.ExpireMinutes == 0 {
		cfg.ExpireMinutes = 30
	}
	if cfg.IntervalMinutes == 0 {
		cfg.IntervalMinutes = 1
	}
	data, _ := json.Marshal(cfg)
	_ = a.redis.Set(ctx, smsConfigCacheKey(), data, 30*time.Minute).Err()
	return cfg, nil
}

func (a *Application) updateLoginState(ctx context.Context, userID int64, ip string) error {
	_, err := a.db.ExecContext(ctx, `UPDATE admin_user SET last_login_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`, ip, a.now(), a.now(), userID)
	return err
}

func (a *Application) insertLoginLog(ctx context.Context, userID int64, adminName, ip string) error {
	_, err := a.db.ExecContext(ctx, `
INSERT INTO admin_login_log (admin_id, admin_name, ip, ip_region, created_at)
VALUES (?, ?, ?, ?, ?)
`, userID, adminName, ip, a.resolveRegion(ip), a.now())
	return err
}

func (a *Application) insertOperationLog(ctx context.Context, evt operationEvent) error {
	_, err := a.db.ExecContext(ctx, `
INSERT INTO admin_operation_log (admin_id, admin_name, description, ip, ip_region, created_at)
VALUES (?, ?, ?, ?, ?, ?)
`, evt.AdminID, evt.AdminName, evt.Description, evt.IP, evt.IPRegion, a.now())
	return err
}

func (a *Application) writeOperation(ctx context.Context, actor AdminUser, desc, ip string) {
	evt := operationEvent{AdminID: actor.ID, AdminName: actor.RealName, Description: desc, IP: ip, IPRegion: a.resolveRegion(ip)}
	if a.syncAudit {
		_ = a.insertOperationLog(ctx, evt)
		return
	}
	select {
	case a.auditCh <- evt:
	default:
		_ = a.insertOperationLog(ctx, evt)
	}
}

func (a *Application) removeSession(ctx context.Context, jti string, userID int64) error {
	if err := a.redis.Del(ctx, sessionKey(jti)).Err(); err != nil {
		return err
	}
	return a.redis.SRem(ctx, userSessionsKey(userID), jti).Err()
}

func (a *Application) removeAllUserSessions(ctx context.Context, userID int64) error {
	sessions, err := a.redis.SMembers(ctx, userSessionsKey(userID)).Result()
	if err == nil {
		for _, jti := range sessions {
			_ = a.redis.Del(ctx, sessionKey(jti)).Err()
		}
	}
	return a.redis.Del(ctx, userSessionsKey(userID)).Err()
}

func (a *Application) ensureGroupActive(ctx context.Context, groupID int64) error {
	if groupID <= 0 {
		return errors.New("用户组错误")
	}
	group, err := a.getGroupByID(ctx, groupID)
	if err != nil {
		return errors.New("用户组不存在")
	}
	if group.Status != statusEnabled {
		return errors.New("用户组不存在或已禁用")
	}
	return nil
}

func (a *Application) activeUsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	var count int
	err := a.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM admin_user WHERE username = ? AND is_deleted = 0 AND id <> ?`, username, excludeID)
	return count > 0, err
}

func (a *Application) activePhoneExists(ctx context.Context, phone string, excludeID int64) (bool, error) {
	var count int
	err := a.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM admin_user WHERE phone = ? AND is_deleted = 0 AND id <> ?`, phone, excludeID)
	return count > 0, err
}

func (a *Application) getUserByUsername(ctx context.Context, username string) (AdminUser, error) {
	var user AdminUser
	err := a.db.GetContext(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE username = ? LIMIT 1
`, username)
	return user, err
}

func (a *Application) getUserByID(ctx context.Context, id int64) (AdminUser, error) {
	var user AdminUser
	err := a.db.GetContext(ctx, &user, `
SELECT id, username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
       COALESCE(last_login_ip, '') AS last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM admin_user WHERE id = ? LIMIT 1
`, id)
	return user, err
}

func (a *Application) getGroupByID(ctx context.Context, id int64) (AdminGroup, error) {
	var group AdminGroup
	err := a.db.GetContext(ctx, &group, `SELECT id, name, description, status, created_at, updated_at FROM admin_group WHERE id = ? LIMIT 1`, id)
	return group, err
}
