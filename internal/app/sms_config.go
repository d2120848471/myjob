package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"
)

const smsConfigCacheVersion = 2

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
