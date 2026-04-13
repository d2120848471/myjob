package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type systemSettingItem struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Value      string `json:"value"`
	ValueType  string `json:"value_type"`
	Unit       string `json:"unit"`
	Required   bool   `json:"required"`
	Configured bool   `json:"configured"`
	UpdatedAt  string `json:"updated_at"`
}

func TestSystemSettings_SaveAndGetFinanceConfig(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	saveRaw := h.rawRequest(http.MethodPut, "/api/admin/settings/system", map[string]any{
		"group": "finance",
		"items": []map[string]any{
			{"key": "tax_exclusive_rate", "value": "4.5"},
			{"key": "tax_inclusive_rate", "value": "3.8"},
		},
	}, token)
	require.Equal(t, http.StatusOK, saveRaw.status)
	require.Equal(t, 0, saveRaw.env.Code)

	getRaw := h.rawRequest(http.MethodGet, "/api/admin/settings/system?group=finance", nil, token)
	require.Equal(t, http.StatusOK, getRaw.status)
	require.Equal(t, 0, getRaw.env.Code)

	var data struct {
		Group string              `json:"group"`
		Items []systemSettingItem `json:"items"`
	}
	require.NoError(t, json.Unmarshal(getRaw.env.Data, &data))
	require.Equal(t, "finance", data.Group)
	require.Len(t, data.Items, 2)

	items := make(map[string]systemSettingItem, len(data.Items))
	for _, item := range data.Items {
		items[item.Key] = item
	}

	exclusive := items["tax_exclusive_rate"]
	require.Equal(t, "未税->含税税率", exclusive.Label)
	require.Equal(t, "4.5", exclusive.Value)
	require.Equal(t, "decimal", exclusive.ValueType)
	require.Equal(t, "%", exclusive.Unit)
	require.True(t, exclusive.Required)
	require.True(t, exclusive.Configured)
	require.NotEmpty(t, exclusive.UpdatedAt)

	inclusive := items["tax_inclusive_rate"]
	require.Equal(t, "含税->未税税率", inclusive.Label)
	require.Equal(t, "3.8", inclusive.Value)
	require.Equal(t, "decimal", inclusive.ValueType)
	require.Equal(t, "%", inclusive.Unit)
	require.True(t, inclusive.Required)
	require.True(t, inclusive.Configured)
	require.NotEmpty(t, inclusive.UpdatedAt)
}

func TestSystemSettings_SaveRejectsUnknownKeyWithoutPersisting(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	saveRaw := h.rawRequest(http.MethodPut, "/api/admin/settings/system", map[string]any{
		"group": "finance",
		"items": []map[string]any{
			{"key": "tax_exclusive_rate", "value": "4.5"},
			{"key": "unknown_key", "value": "bad"},
		},
	}, token)
	require.Equal(t, http.StatusOK, saveRaw.status)
	require.NotEqual(t, 0, saveRaw.env.Code)

	getRaw := h.rawRequest(http.MethodGet, "/api/admin/settings/system?group=finance", nil, token)
	require.Equal(t, http.StatusOK, getRaw.status)
	require.Equal(t, 0, getRaw.env.Code)

	var data struct {
		Items []systemSettingItem `json:"items"`
	}
	require.NoError(t, json.Unmarshal(getRaw.env.Data, &data))
	require.Len(t, data.Items, 2)

	items := make(map[string]systemSettingItem, len(data.Items))
	for _, item := range data.Items {
		items[item.Key] = item
	}
	require.False(t, items["tax_exclusive_rate"].Configured)
	require.Empty(t, items["tax_exclusive_rate"].Value)
}

func TestSystemSettings_RequiresSuperAdmin(t *testing.T) {
	h := newTestHarness(t)
	phone := "13800005555"
	_ = h.createUserForSMSFlow(context.Background(), "staff01", "Staff_123", phone)
	token := loginTestUser(t, h, "staff01", "Staff_123", phone)

	getRaw := h.rawRequest(http.MethodGet, "/api/admin/settings/system?group=finance", nil, token)
	require.Equal(t, http.StatusOK, getRaw.status)
	require.NotEqual(t, 0, getRaw.env.Code)
	require.Equal(t, "仅超级管理员可访问", getRaw.env.Message)
}

func TestSystemSettings_SaveRejectsInvalidFinanceRate(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	saveRaw := h.rawRequest(http.MethodPut, "/api/admin/settings/system", map[string]any{
		"group": "finance",
		"items": []map[string]any{
			{"key": "tax_exclusive_rate", "value": "100"},
			{"key": "tax_inclusive_rate", "value": "3.8"},
		},
	}, token)
	require.Equal(t, http.StatusOK, saveRaw.status)
	require.NotEqual(t, 0, saveRaw.env.Code)
}

func TestSystemSettings_SaveRejectsInvalidRobotWebhookURL(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	saveRaw := h.rawRequest(http.MethodPut, "/api/admin/settings/system", map[string]any{
		"group": "integration",
		"items": []map[string]any{
			{"key": "robot_webhook_url", "value": "ftp://invalid.example.com"},
		},
	}, token)
	require.Equal(t, http.StatusOK, saveRaw.status)
	require.NotEqual(t, 0, saveRaw.env.Code)
}

func loginTestUser(t *testing.T, h *testHarness, username, password, phone string) string {
	t.Helper()
	res := h.postJSON("/api/admin/auth/login", map[string]any{
		"username": username,
		"password": password,
	}, "")
	require.Equal(t, 0, res.Code)

	var data struct {
		Token         string `json:"token"`
		NeedSMSVerify bool   `json:"need_sms_verify"`
		LoginToken    string `json:"login_token"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	if data.Token != "" {
		return data.Token
	}
	require.True(t, data.NeedSMSVerify)
	require.NotEmpty(t, data.LoginToken)

	sendRes := h.postJSON("/api/admin/auth/sms/send", map[string]any{
		"login_token": data.LoginToken,
	}, "")
	require.Equal(t, 0, sendRes.Code)

	verifyRes := h.postJSON("/api/admin/auth/sms/verify", map[string]any{
		"login_token": data.LoginToken,
		"sms_code":    h.lastSMSCode(t, phone),
	}, "")
	require.Equal(t, 0, verifyRes.Code)

	var verifyData struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(verifyRes.Data, &verifyData))
	data.Token = verifyData.Token
	require.NotEmpty(t, data.Token)
	return data.Token
}
