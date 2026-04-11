package app

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type smsConfigGetResponse struct {
	AccessKeyMasked           string `json:"access_key_masked"`
	AccessKeySecretMasked     string `json:"access_key_secret_masked"`
	AccessKeyConfigured       bool   `json:"access_key_configured"`
	AccessKeySecretConfigured bool   `json:"access_key_secret_configured"`
	SignName                  string `json:"sign_name"`
	TemplateCode              string `json:"template_code"`
	ExpireMinutes             int    `json:"expire_minutes"`
	IntervalMinutes           int    `json:"interval_minutes"`
	UpdatedAt                 string `json:"updated_at"`
}

type smsConfigSeed struct {
	AccessKey       string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	ExpireMinutes   int
	IntervalMinutes int
}

func TestSMSConfigContract_GetReturnsConfiguredFlagsAndMaskedPreview(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)

	res := h.getJSON("/api/admin/config/sms", token)
	require.Equal(t, 0, res.Code)

	data := decodeSMSConfigGetResponse(t, res.Data)
	require.False(t, data.AccessKeyConfigured)
	require.False(t, data.AccessKeySecretConfigured)
	require.Equal(t, "", data.AccessKeyMasked)
	require.Equal(t, "", data.AccessKeySecretMasked)
	require.Equal(t, "玖权益", data.SignName)
	require.Equal(t, "SMS_000001", data.TemplateCode)
	require.Equal(t, 30, data.ExpireMinutes)
	require.Equal(t, 1, data.IntervalMinutes)
	require.NotEmpty(t, data.UpdatedAt)

	raw := decodeRawMap(t, res.Data)
	require.NotContains(t, raw, "access_key")
	require.NotContains(t, raw, "access_key_secret")
}

func TestSMSConfigContract_FirstSaveAndGetPreview(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)
	at := time.Date(2026, 4, 11, 10, 0, 0, 0, time.Local)
	h.app.now = func() time.Time { return at }

	res := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":             "  LTAI-first-key-1234  ",
		"access_key_secret":      "  secret-first-value-5678  ",
		"sign_name":              "  玖权益签名  ",
		"template_code":          "  SMS_FIRST_001  ",
		"expire_minutes":         15,
		"interval_minutes":       2,
		"keep_access_key":        false,
		"keep_access_key_secret": false,
	}, token)
	require.Equal(t, 0, res.Code)

	cfg := h.mustLoadSMSConfig(t)
	require.Equal(t, "LTAI-first-key-1234", cfg.AccessKey)
	require.Equal(t, "secret-first-value-5678", cfg.AccessKeySecret)
	require.Equal(t, "玖权益签名", cfg.SignName)
	require.Equal(t, "SMS_FIRST_001", cfg.TemplateCode)
	require.Equal(t, 15, cfg.ExpireMinutes)
	require.Equal(t, 2, cfg.IntervalMinutes)

	getRes := h.getJSON("/api/admin/config/sms", token)
	require.Equal(t, 0, getRes.Code)
	data := decodeSMSConfigGetResponse(t, getRes.Data)
	require.True(t, data.AccessKeyConfigured)
	require.True(t, data.AccessKeySecretConfigured)
	require.NotEmpty(t, data.AccessKeyMasked)
	require.NotEmpty(t, data.AccessKeySecretMasked)
	require.NotContains(t, data.AccessKeyMasked, "LTAI-first-key-1234")
	require.NotContains(t, data.AccessKeySecretMasked, "secret-first-value-5678")
	require.Equal(t, "2026-04-11 10:00:00", data.UpdatedAt)

	desc := h.latestOperationDescription(t)
	require.Contains(t, desc, "更新短信配置")
	require.NotContains(t, desc, "LTAI-first-key-1234")
	require.NotContains(t, desc, "secret-first-value-5678")
}

func TestSMSConfigContract_SaveBusinessFieldsWithoutOverwritingKeys(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)
	h.seedSMSConfig(t, smsConfigSeed{
		AccessKey:       "LTAI-existing-ak",
		AccessKeySecret: "existing-secret-sk",
		SignName:        "旧签名",
		TemplateCode:    "SMS_OLD_001",
		ExpireMinutes:   30,
		IntervalMinutes: 1,
	})
	at := time.Date(2026, 4, 11, 11, 0, 0, 0, time.Local)
	h.app.now = func() time.Time { return at }

	res := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":             "masked-value-should-be-ignored",
		"access_key_secret":      "masked-secret-should-be-ignored",
		"sign_name":              "新签名",
		"template_code":          "SMS_NEW_002",
		"expire_minutes":         20,
		"interval_minutes":       3,
		"keep_access_key":        true,
		"keep_access_key_secret": true,
	}, token)
	require.Equal(t, 0, res.Code)

	cfg := h.mustLoadSMSConfig(t)
	require.Equal(t, "LTAI-existing-ak", cfg.AccessKey)
	require.Equal(t, "existing-secret-sk", cfg.AccessKeySecret)
	require.Equal(t, "新签名", cfg.SignName)
	require.Equal(t, "SMS_NEW_002", cfg.TemplateCode)
	require.Equal(t, 20, cfg.ExpireMinutes)
	require.Equal(t, 3, cfg.IntervalMinutes)

	desc := h.latestOperationDescription(t)
	require.Contains(t, desc, "更新短信配置")
	require.NotContains(t, desc, "AccessKey")
	require.NotContains(t, desc, "existing-secret-sk")
}

func TestSMSConfigContract_SaveOnlyAccessKey(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)
	h.seedSMSConfig(t, smsConfigSeed{
		AccessKey:       "LTAI-old-ak",
		AccessKeySecret: "old-secret-sk",
		SignName:        "旧签名",
		TemplateCode:    "SMS_OLD_001",
		ExpireMinutes:   30,
		IntervalMinutes: 1,
	})
	at := time.Date(2026, 4, 11, 12, 0, 0, 0, time.Local)
	h.app.now = func() time.Time { return at }

	res := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":             "LTAI-new-ak",
		"access_key_secret":      "ignored-secret",
		"sign_name":              "旧签名",
		"template_code":          "SMS_OLD_001",
		"expire_minutes":         30,
		"interval_minutes":       1,
		"keep_access_key":        false,
		"keep_access_key_secret": true,
	}, token)
	require.Equal(t, 0, res.Code)

	cfg := h.mustLoadSMSConfig(t)
	require.Equal(t, "LTAI-new-ak", cfg.AccessKey)
	require.Equal(t, "old-secret-sk", cfg.AccessKeySecret)
	require.Contains(t, h.latestOperationDescription(t), "AccessKey")
}

func TestSMSConfigContract_SaveOnlyAccessKeySecret(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)
	h.seedSMSConfig(t, smsConfigSeed{
		AccessKey:       "LTAI-old-ak",
		AccessKeySecret: "old-secret-sk",
		SignName:        "旧签名",
		TemplateCode:    "SMS_OLD_001",
		ExpireMinutes:   30,
		IntervalMinutes: 1,
	})
	at := time.Date(2026, 4, 11, 13, 0, 0, 0, time.Local)
	h.app.now = func() time.Time { return at }

	res := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":             "ignored-ak",
		"access_key_secret":      "new-secret-sk",
		"sign_name":              "旧签名",
		"template_code":          "SMS_OLD_001",
		"expire_minutes":         30,
		"interval_minutes":       1,
		"keep_access_key":        true,
		"keep_access_key_secret": false,
	}, token)
	require.Equal(t, 0, res.Code)

	cfg := h.mustLoadSMSConfig(t)
	require.Equal(t, "LTAI-old-ak", cfg.AccessKey)
	require.Equal(t, "new-secret-sk", cfg.AccessKeySecret)
	require.Contains(t, h.latestOperationDescription(t), "AccessKeySecret")
}

func TestSMSConfigContract_SaveBothKeys(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)
	h.seedSMSConfig(t, smsConfigSeed{
		AccessKey:       "LTAI-old-ak",
		AccessKeySecret: "old-secret-sk",
		SignName:        "旧签名",
		TemplateCode:    "SMS_OLD_001",
		ExpireMinutes:   30,
		IntervalMinutes: 1,
	})
	at := time.Date(2026, 4, 11, 14, 0, 0, 0, time.Local)
	h.app.now = func() time.Time { return at }

	res := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":             "LTAI-both-ak",
		"access_key_secret":      "both-secret-sk",
		"sign_name":              "新签名",
		"template_code":          "SMS_BOTH_003",
		"expire_minutes":         25,
		"interval_minutes":       4,
		"keep_access_key":        false,
		"keep_access_key_secret": false,
	}, token)
	require.Equal(t, 0, res.Code)

	cfg := h.mustLoadSMSConfig(t)
	require.Equal(t, "LTAI-both-ak", cfg.AccessKey)
	require.Equal(t, "both-secret-sk", cfg.AccessKeySecret)
	desc := h.latestOperationDescription(t)
	require.Contains(t, desc, "AccessKey")
	require.Contains(t, desc, "AccessKeySecret")
}

func TestSMSConfigContract_ValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("首次配置不能保留旧密钥", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)

		res := h.putJSON("/api/admin/config/sms", map[string]any{
			"sign_name":              "玖权益",
			"template_code":          "SMS_001",
			"expire_minutes":         30,
			"interval_minutes":       1,
			"keep_access_key":        true,
			"keep_access_key_secret": true,
		}, token)
		require.Equal(t, 400, res.Code)
	})

	t.Run("更新 access key 时必须提供新值", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)
		h.seedSMSConfig(t, smsConfigSeed{AccessKey: "ak", AccessKeySecret: "sk", SignName: "签名", TemplateCode: "SMS", ExpireMinutes: 30, IntervalMinutes: 1})

		res := h.putJSON("/api/admin/config/sms", map[string]any{
			"sign_name":              "玖权益",
			"template_code":          "SMS_001",
			"expire_minutes":         30,
			"interval_minutes":       1,
			"keep_access_key":        false,
			"keep_access_key_secret": true,
		}, token)
		require.Equal(t, 400, res.Code)
	})

	t.Run("保留 secret 但旧 secret 不存在时拒绝", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)
		h.seedSMSConfig(t, smsConfigSeed{AccessKey: "ak", AccessKeySecret: "", SignName: "签名", TemplateCode: "SMS", ExpireMinutes: 30, IntervalMinutes: 1})

		res := h.putJSON("/api/admin/config/sms", map[string]any{
			"access_key":             "new-ak",
			"sign_name":              "玖权益",
			"template_code":          "SMS_001",
			"expire_minutes":         30,
			"interval_minutes":       1,
			"keep_access_key":        false,
			"keep_access_key_secret": true,
		}, token)
		require.Equal(t, 400, res.Code)
	})

	t.Run("范围非法时拒绝", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)
		h.seedSMSConfig(t, smsConfigSeed{AccessKey: "ak", AccessKeySecret: "sk", SignName: "签名", TemplateCode: "SMS", ExpireMinutes: 30, IntervalMinutes: 1})

		res := h.putJSON("/api/admin/config/sms", map[string]any{
			"sign_name":              "玖权益",
			"template_code":          "SMS_001",
			"expire_minutes":         0,
			"interval_minutes":       11,
			"keep_access_key":        true,
			"keep_access_key_secret": true,
		}, token)
		require.Equal(t, 400, res.Code)
	})
}

func TestSMSConfigContract_NonSuperUsersCannotReadOrWrite(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	groupID := h.createGroupWithMenus(t, "授权测试组", []int64{1, 2, 6})
	userID := h.createUserForSMSFlow(context.Background(), "permSms01", "Perm_123", "13800005555")
	require.NotZero(t, userID)
	h.execSQL(t, `UPDATE admin_user SET group_id = ?, last_login_ip = ? WHERE id = ?`, groupID, "127.0.0.1", userID)

	login := h.postJSON("/api/admin/login", map[string]any{
		"username": "permSms01",
		"password": "Perm_123",
	}, "")
	require.Equal(t, 0, login.Code)
	token := decodeRawMap(t, login.Data)["token"].(string)

	getRes := h.getJSON("/api/admin/config/sms", token)
	require.Equal(t, 403, getRes.Code)

	putRes := h.putJSON("/api/admin/config/sms", map[string]any{
		"sign_name":              "玖权益",
		"template_code":          "SMS_001",
		"expire_minutes":         30,
		"interval_minutes":       1,
		"keep_access_key":        true,
		"keep_access_key_secret": true,
	}, token)
	require.Equal(t, 403, putRes.Code)
}

func decodeSMSConfigGetResponse(t *testing.T, raw json.RawMessage) smsConfigGetResponse {
	t.Helper()
	var out smsConfigGetResponse
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func (h *testHarness) mustLoadSMSConfig(t *testing.T) SMSConfig {
	t.Helper()
	cfg, err := h.app.loadSMSConfig(context.Background())
	require.NoError(t, err)
	return cfg
}

func (h *testHarness) latestOperationDescription(t *testing.T) string {
	t.Helper()
	var desc string
	err := h.app.db.GetContext(context.Background(), &desc, `SELECT description FROM admin_operation_log ORDER BY id DESC LIMIT 1`)
	require.NoError(t, err)
	return desc
}

func (h *testHarness) seedSMSConfig(t *testing.T, cfg smsConfigSeed) {
	t.Helper()
	now := h.app.now()
	values := map[string]string{
		"sms_access_key":        cfg.AccessKey,
		"sms_access_key_secret": cfg.AccessKeySecret,
		"sms_sign_name":         cfg.SignName,
		"sms_template_code":     cfg.TemplateCode,
		"sms_expire_minutes":    intToString(cfg.ExpireMinutes),
		"sms_interval_minutes":  intToString(cfg.IntervalMinutes),
	}
	for key, value := range values {
		_, err := h.app.db.ExecContext(context.Background(), `UPDATE system_config SET config_value = ?, updated_at = ? WHERE config_key = ?`, value, now, key)
		require.NoError(t, err)
	}
	require.NoError(t, h.app.redis.Del(context.Background(), smsConfigCacheKey()).Err())
}

func intToString(v int) string {
	return strconv.Itoa(v)
}
