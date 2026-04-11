package app

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultAccessTokenTTLIsSevenDays(t *testing.T) {
	t.Parallel()

	require.Equal(t, 10080, defaultConfig().Auth.AccessTokenTTLMin)

	exampleCfg, err := LoadConfig(filepath.Join("..", "..", "manifest", "config", "config.example.yaml"))
	require.NoError(t, err)
	require.Equal(t, 10080, exampleCfg.Auth.AccessTokenTTLMin)

	localCfg, err := LoadConfig(filepath.Join("..", "..", "manifest", "config", "config.local.yaml"))
	require.NoError(t, err)
	require.Equal(t, 10080, localCfg.Auth.AccessTokenTTLMin)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("auth:\n  access_token_ttl_minutes: 0\n"), 0o644))

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, 10080, cfg.Auth.AccessTokenTTLMin)
}

func TestLoginContract_ExplicitNeedSMSVerifyAndEnumReason(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)

	directLogin := h.postJSON("/api/admin/login", map[string]any{
		"username": "admin",
		"password": "Admin_123",
	}, "")
	require.Equal(t, 0, directLogin.Code)

	directData := decodeRawMap(t, directLogin.Data)
	require.Contains(t, directData, "need_sms_verify")
	require.Equal(t, false, directData["need_sms_verify"])

	userID := h.createUserForSMSFlow(context.Background(), "need001", "Need_123", "13800001111")
	require.NotZero(t, userID)

	firstLogin := h.postJSON("/api/admin/login", map[string]any{
		"username": "need001",
		"password": "Need_123",
	}, "")
	require.Equal(t, 0, firstLogin.Code)

	firstData := decodeRawMap(t, firstLogin.Data)
	require.Equal(t, true, firstData["need_sms_verify"])
	require.Equal(t, "first_login", firstData["reason"])

	h.execSQL(t, `UPDATE admin_user SET last_login_ip = ? WHERE id = ?`, "1.1.1.1", userID)
	changedLogin := h.postJSONWithIP("/api/admin/login", map[string]any{
		"username": "need001",
		"password": "Need_123",
	}, "", "8.8.8.8")
	require.Equal(t, 0, changedLogin.Code)

	changedData := decodeRawMap(t, changedLogin.Data)
	require.Equal(t, true, changedData["need_sms_verify"])
	require.Equal(t, "ip_changed", changedData["reason"])
}

func TestGroupListReturnsPagination(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	token := h.loginAdmin(t)

	for _, name := range []string{"客服组", "风控组"} {
		res := h.postJSON("/api/admin/group/add", map[string]any{
			"name":        name,
			"description": name + "描述",
		}, token)
		require.Equal(t, 0, res.Code)
	}

	listRes := h.getJSON("/api/admin/group/list?page=1&page_size=1", token)
	require.Equal(t, 0, listRes.Code)

	data := decodeRawMap(t, listRes.Data)
	require.Contains(t, data, "pagination")
	require.Len(t, decodeRawList(t, data["list"]), 1)

	pagination := decodeMapValue(t, data["pagination"])
	require.Equal(t, float64(1), pagination["page"])
	require.Equal(t, float64(1), pagination["page_size"])
	require.GreaterOrEqual(t, pagination["total"].(float64), float64(3))
}

func TestLogFiltersSupportTimeWindowAndEmptyFallbackRegion(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	firstLoginAt := time.Date(2026, 4, 11, 10, 0, 0, 0, time.Local)
	secondLoginAt := firstLoginAt.Add(2 * time.Hour)
	opFirstAt := secondLoginAt.Add(time.Hour)
	opSecondAt := opFirstAt.Add(time.Hour)

	h.app.now = func() time.Time { return firstLoginAt }
	_ = h.loginAdmin(t)

	h.app.now = func() time.Time { return secondLoginAt }
	token := h.loginAdmin(t)

	h.app.now = func() time.Time { return opFirstAt }
	addGroup := h.postJSON("/api/admin/group/add", map[string]any{
		"name":        "审计组A",
		"description": "第一次写入",
	}, token)
	require.Equal(t, 0, addGroup.Code)

	h.app.now = func() time.Time { return opSecondAt }
	addSubject := h.postJSON("/api/admin/subject/add", map[string]any{
		"name":    "主体B",
		"has_tax": 1,
	}, token)
	require.Equal(t, 0, addSubject.Code)

	loginPath := "/api/admin/log/login?page=1&page_size=20&start_time=" + url.QueryEscape(secondLoginAt.Add(-time.Minute).Format("2006-01-02 15:04:05")) + "&end_time=" + url.QueryEscape(secondLoginAt.Add(time.Minute).Format("2006-01-02 15:04:05"))
	loginLogs := h.getJSON(loginPath, token)
	require.Equal(t, 0, loginLogs.Code)
	loginData := decodeRawMap(t, loginLogs.Data)
	loginList := decodeRawList(t, loginData["list"])
	require.Len(t, loginList, 1)
	require.Equal(t, "", decodeMapValue(t, loginList[0])["ip_region"])

	opPath := "/api/admin/log/operation?page=1&page_size=20&start_time=" + url.QueryEscape(opSecondAt.Add(-time.Minute).Format("2006-01-02 15:04:05")) + "&end_time=" + url.QueryEscape(opSecondAt.Add(time.Minute).Format("2006-01-02 15:04:05"))
	opLogs := h.getJSON(opPath, token)
	require.Equal(t, 0, opLogs.Code)
	opData := decodeRawMap(t, opLogs.Data)
	opList := decodeRawList(t, opData["list"])
	require.Len(t, opList, 1)
	require.Contains(t, decodeMapValue(t, opList[0])["description"], "主体B")
	require.Equal(t, "", decodeMapValue(t, opList[0])["ip_region"])
}

func TestPermissionSemantics_SubjectRequiresOwnPermissionAndMenuTreeHidesRestrictedNodes(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	groupID := h.createGroupWithMenus(t, "授权测试组", []int64{1, 2, 6})
	userID := h.createUserForSMSFlow(context.Background(), "perm001", "Perm_123", "13800003333")
	require.NotZero(t, userID)

	h.execSQL(t, `UPDATE admin_user SET group_id = ?, last_login_ip = ? WHERE id = ?`, groupID, "127.0.0.1", userID)
	h.execSQL(t, `UPDATE admin_menu SET status = 0 WHERE code = ?`, "subject.manage")

	login := h.postJSON("/api/admin/login", map[string]any{
		"username": "perm001",
		"password": "Perm_123",
	}, "")
	require.Equal(t, 0, login.Code)
	token := decodeRawMap(t, login.Data)["token"].(string)

	subjectRes := h.getJSON("/api/admin/subject/list", token)
	require.Equal(t, 403, subjectRes.Code)

	smsConfigRes := h.getJSON("/api/admin/config/sms", token)
	require.Equal(t, 403, smsConfigRes.Code)

	menuTreeRes := h.getJSON("/api/admin/menu/tree", token)
	require.Equal(t, 0, menuTreeRes.Code)
	codes := flattenMenuCodes(decodeTreeList(t, menuTreeRes.Data))
	require.NotContains(t, codes, "subject.manage")
	require.NotContains(t, codes, "config.sms")
}

func (h *testHarness) postJSONWithIP(path string, body any, token, remoteIP string) apiEnvelope {
	return h.requestWithIP("POST", path, body, token, remoteIP)
}

func (h *testHarness) requestWithIP(method, path string, body any, token, remoteIP string) apiEnvelope {
	var env apiEnvelope
	var payload any = body
	resp := h.requestWithCustomizer(method, path, payload, token, func(req *httpRequestConfig) {
		req.remoteAddr = net.JoinHostPort(remoteIP, "12345")
	})
	env = resp
	return env
}

type httpRequestConfig struct {
	remoteAddr string
}

func (h *testHarness) requestWithCustomizer(method, path string, body any, token string, apply func(*httpRequestConfig)) apiEnvelope {
	cfg := &httpRequestConfig{remoteAddr: "127.0.0.1:12345"}
	if apply != nil {
		apply(cfg)
	}
	return h.requestWithRemoteAddr(method, path, body, token, cfg.remoteAddr)
}

func (h *testHarness) requestWithRemoteAddr(method, path string, body any, token, remoteAddr string) apiEnvelope {
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		data, _ := json.Marshal(body)
		reader = strings.NewReader(string(data))
	}
	req := newTestRequest(method, path, reader, token, remoteAddr)
	return h.serve(req)
}

func newTestRequest(method, path string, reader *strings.Reader, token, remoteAddr string) *testRequest {
	return &testRequest{method: method, path: path, body: reader, token: token, remoteAddr: remoteAddr}
}

type testRequest struct {
	method     string
	path       string
	body       *strings.Reader
	token      string
	remoteAddr string
}

func (h *testHarness) serve(cfg *testRequest) apiEnvelope {
	req := newHTTPRequest(cfg.method, cfg.path, cfg.body, cfg.token, cfg.remoteAddr)
	return h.do(req)
}

func newHTTPRequest(method, path string, body *strings.Reader, token, remoteAddr string) *requestEnvelope {
	return &requestEnvelope{method: method, path: path, body: body, token: token, remoteAddr: remoteAddr}
}

type requestEnvelope struct {
	method     string
	path       string
	body       *strings.Reader
	token      string
	remoteAddr string
}

func (h *testHarness) do(cfg *requestEnvelope) apiEnvelope {
	req := buildTestHTTPRequest(cfg.method, cfg.path, cfg.body, cfg.token, cfg.remoteAddr)
	return h.doRaw(req)
}

func buildTestHTTPRequest(method, path string, body *strings.Reader, token, remoteAddr string) *http.Request {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.RemoteAddr = remoteAddr
	return req
}

func (h *testHarness) doRaw(req *http.Request) apiEnvelope {
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	var env apiEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	return env
}

func (h *testHarness) execSQL(t *testing.T, query string, args ...any) {
	t.Helper()
	_, err := h.app.db.ExecContext(context.Background(), query, args...)
	require.NoError(t, err)
}

func (h *testHarness) createGroupWithMenus(t *testing.T, name string, menuIDs []int64) int64 {
	t.Helper()
	result, err := h.app.db.ExecContext(context.Background(), `
INSERT INTO admin_group (name, description, status, created_at, updated_at)
VALUES (?, ?, 1, ?, ?)
`, name, name+"描述", h.app.now(), h.app.now())
	require.NoError(t, err)
	groupID, err := result.LastInsertId()
	require.NoError(t, err)
	for _, menuID := range menuIDs {
		_, err = h.app.db.ExecContext(context.Background(), `INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (?, ?, ?)`, groupID, menuID, h.app.now())
		require.NoError(t, err)
	}
	return groupID
}

func decodeRawMap(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func decodeMapValue(t *testing.T, value any) map[string]any {
	t.Helper()
	bytes, err := json.Marshal(value)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(bytes, &out))
	return out
}

func decodeRawList(t *testing.T, value any) []any {
	t.Helper()
	bytes, err := json.Marshal(value)
	require.NoError(t, err)
	var out []any
	require.NoError(t, json.Unmarshal(bytes, &out))
	return out
}

func decodeTreeList(t *testing.T, raw json.RawMessage) []map[string]any {
	t.Helper()
	var out []map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func flattenMenuCodes(nodes []map[string]any) []string {
	codes := make([]string, 0)
	var walk func([]map[string]any)
	walk = func(items []map[string]any) {
		for _, item := range items {
			if code, ok := item["code"].(string); ok && code != "" {
				codes = append(codes, code)
			}
			children, ok := item["children"]
			if !ok {
				continue
			}
			childBytes, err := json.Marshal(children)
			if err != nil {
				continue
			}
			var childNodes []map[string]any
			if json.Unmarshal(childBytes, &childNodes) == nil {
				walk(childNodes)
			}
		}
	}
	walk(nodes)
	return codes
}
