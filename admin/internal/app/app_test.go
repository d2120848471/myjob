package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func TestAuthFlow_LoginSMSVerifyMeLogout(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)

	loginRes := h.postJSON("/api/admin/login", map[string]any{
		"username": "admin",
		"password": "Admin_123",
	}, "")
	require.Equal(t, 0, loginRes.Code)

	var loginData struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(loginRes.Data, &loginData))
	require.NotEmpty(t, loginData.Token)

	meRes := h.postJSON("/api/admin/me", map[string]any{}, loginData.Token)
	require.Equal(t, 0, meRes.Code)

	smsUserID := h.createUserForSMSFlow(context.Background(), "need001", "Need_123", "13800001111")
	require.NotZero(t, smsUserID)

	smsLogin := h.postJSON("/api/admin/login", map[string]any{
		"username": "need001",
		"password": "Need_123",
	}, "")
	require.Equal(t, 0, smsLogin.Code)
	var smsLoginData struct {
		NeedSMSVerify bool   `json:"need_sms_verify"`
		LoginToken    string `json:"login_token"`
	}
	require.NoError(t, json.Unmarshal(smsLogin.Data, &smsLoginData))
	require.True(t, smsLoginData.NeedSMSVerify)
	require.NotEmpty(t, smsLoginData.LoginToken)

	sendRes := h.postJSON("/api/admin/login/sms/send", map[string]any{
		"login_token": smsLoginData.LoginToken,
	}, "")
	require.Equal(t, 0, sendRes.Code)

	code := h.lastSMSCode(t, "13800001111")
	verifyRes := h.postJSON("/api/admin/login/sms/verify", map[string]any{
		"login_token": smsLoginData.LoginToken,
		"sms_code":    code,
	}, "")
	require.Equal(t, 0, verifyRes.Code)

	var verifyData struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(verifyRes.Data, &verifyData))
	require.NotEmpty(t, verifyData.Token)

	logoutRes := h.postJSON("/api/admin/logout", map[string]any{}, loginData.Token)
	require.Equal(t, 0, logoutRes.Code)

	meAfterLogout := h.postJSON("/api/admin/me", map[string]any{}, loginData.Token)
	require.NotEqual(t, 0, meAfterLogout.Code)
}

func TestAdminManagementFlows(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)

	addGroup := h.postJSON("/api/admin/group/add", map[string]any{
		"name":        "客服组",
		"description": "客服权限组",
	}, token)
	require.Equal(t, 0, addGroup.Code)

	var addGroupData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(addGroup.Data, &addGroupData))
	require.NotZero(t, addGroupData.ID)

	saveAuth := h.putJSON("/api/admin/group/1/auth", map[string]any{
		"menu_ids": []int64{1, 2, 3, 4},
	}, token)
	require.Equal(t, 0, saveAuth.Code)

	addUser := h.postJSON("/api/admin/user/add", map[string]any{
		"username":         "alice01",
		"confirm_username": "alice01",
		"password":         "Alice_123",
		"confirm_password": "Alice_123",
		"real_name":        "Alice",
		"phone":            "13800002222",
		"group_id":         addGroupData.ID,
	}, token)
	require.Equal(t, 0, addUser.Code)

	listUsers := h.getJSON("/api/admin/user/list?page=1&page_size=20", token)
	require.Equal(t, 0, listUsers.Code)

	trashBefore := h.getJSON("/api/admin/user/trash?page=1&page_size=20", token)
	require.Equal(t, 0, trashBefore.Code)

	deleteUser := h.deleteJSON("/api/admin/user/2", token)
	require.Equal(t, 0, deleteUser.Code)

	trashAfter := h.getJSON("/api/admin/user/trash?page=1&page_size=20", token)
	require.Equal(t, 0, trashAfter.Code)

	restoreUser := h.putJSON("/api/admin/user/2/restore", map[string]any{}, token)
	require.Equal(t, 0, restoreUser.Code)

	addSubject := h.postJSON("/api/admin/subject/add", map[string]any{
		"name":    "主体A",
		"has_tax": 1,
	}, token)
	require.Equal(t, 0, addSubject.Code)

	smsSave := h.putJSON("/api/admin/config/sms", map[string]any{
		"access_key":        "LTAI-test",
		"access_key_secret": "secret-test-value",
		"sign_name":         "玖权益",
		"template_code":     "SMS_000001",
		"expire_minutes":    15,
		"interval_minutes":  2,
	}, token)
	require.Equal(t, 0, smsSave.Code)

	smsGet := h.getJSON("/api/admin/config/sms", token)
	require.Equal(t, 0, smsGet.Code)

	opLogs := h.getJSON("/api/admin/log/operation?page=1&page_size=20", token)
	require.Equal(t, 0, opLogs.Code)

	loginLogs := h.getJSON("/api/admin/log/login?page=1&page_size=20", token)
	require.Equal(t, 0, loginLogs.Code)
}

type testHarness struct {
	app     *Application
	handler http.Handler
}

func newTestHarness(t *testing.T) *testHarness {
	t.Helper()
	app, err := NewTestApplication()
	require.NoError(t, err)
	return &testHarness{app: app, handler: app.Handler()}
}

func (h *testHarness) request(method, path string, body any, token string) apiEnvelope {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	var env apiEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	return env
}

func (h *testHarness) postJSON(path string, body any, token string) apiEnvelope {
	return h.request(http.MethodPost, path, body, token)
}

func (h *testHarness) putJSON(path string, body any, token string) apiEnvelope {
	return h.request(http.MethodPut, path, body, token)
}

func (h *testHarness) getJSON(path string, token string) apiEnvelope {
	return h.request(http.MethodGet, path, nil, token)
}

func (h *testHarness) deleteJSON(path string, token string) apiEnvelope {
	return h.request(http.MethodDelete, path, nil, token)
}

func (h *testHarness) loginAdmin(t *testing.T) string {
	t.Helper()
	res := h.postJSON("/api/admin/login", map[string]any{
		"username": "admin",
		"password": "Admin_123",
	}, "")
	require.Equal(t, 0, res.Code)
	var data struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotEmpty(t, data.Token)
	return data.Token
}

func (h *testHarness) lastSMSCode(t *testing.T, phone string) string {
	t.Helper()
	code, err := h.app.LastMockSMSCode(phone)
	require.NoError(t, err)
	return code
}

func (h *testHarness) createUserForSMSFlow(ctx context.Context, username, password, phone string) int64 {
	userID, _ := h.app.CreateTestUser(ctx, username, password, phone)
	return userID
}
