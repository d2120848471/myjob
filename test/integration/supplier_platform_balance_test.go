package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"myjob/internal/bootstrap"

	"github.com/stretchr/testify/require"
)

type supplierAPIEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type supplierIntegrationHarness struct {
	app     *bootstrap.Application
	handler http.Handler
}

func newSupplierIntegrationHarness(t *testing.T) *supplierIntegrationHarness {
	t.Helper()
	app, err := bootstrap.NewTestApplication()
	require.NoError(t, err)
	t.Cleanup(func() { _ = app.Close() })
	return &supplierIntegrationHarness{app: app, handler: app.Handler()}
}

func (h *supplierIntegrationHarness) request(method, path string, body any, token string) supplierAPIEnvelope {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
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

	var env supplierAPIEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	return env
}

func (h *supplierIntegrationHarness) postJSON(path string, body any, token string) supplierAPIEnvelope {
	return h.request(http.MethodPost, path, body, token)
}

func (h *supplierIntegrationHarness) putJSON(path string, body any, token string) supplierAPIEnvelope {
	return h.request(http.MethodPut, path, body, token)
}

func (h *supplierIntegrationHarness) loginAdmin(t *testing.T) string {
	t.Helper()
	res := h.postJSON("/api/admin/auth/login", map[string]any{
		"username": "admin",
		"password": "abc123",
	}, "")
	require.Equal(t, 0, res.Code)

	var data struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotEmpty(t, data.Token)
	return data.Token
}

func (h *supplierIntegrationHarness) createSubject(t *testing.T, token, name string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/subjects", map[string]any{
		"name":    name,
		"has_tax": 1,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *supplierIntegrationHarness) createPlatform(t *testing.T, token string, subjectID int64, domain, backupDomain string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             "集成测试平台",
		"domain":           domain,
		"backup_domain":    backupDomain,
		"type_id":          35,
		"subject_id":       subjectID,
		"has_tax":          1,
		"token_id":         "1008612345",
		"secret_key":       "secret-key-updated",
		"threshold_amount": "5000.0000",
		"sort":             1,
		"crowd_name":       "集成测试群",
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func TestSupplierPlatformRefresh_UsesBackupAndHTTPDowngrade(t *testing.T) {
	h := newSupplierIntegrationHarness(t)
	token := h.loginAdmin(t)
	subjectID := h.createSubject(t, token, "集成主体-主备")

	var hitCount int32
	backupServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hitCount, 1)
		require.Equal(t, "/api/customer", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"ok","msg":"查询成功","data":{"balance":"100.5000"}}`))
	}))
	defer backupServer.Close()

	platformID := h.createPlatform(t, token, subjectID, unusedHost(t), strings.TrimPrefix(backupServer.URL, "http://"))

	refreshRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(platformID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, refreshRes.Code)

	var refreshData struct {
		Balance       string `json:"balance"`
		ConnectStatus int    `json:"connect_status"`
		Message       string `json:"message"`
		TraceID       string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(refreshRes.Data, &refreshData))
	require.Equal(t, "100.5000", refreshData.Balance)
	require.Equal(t, 1, refreshData.ConnectStatus)
	require.Equal(t, "查询成功", refreshData.Message)
	require.NotEmpty(t, refreshData.TraceID)
	require.EqualValues(t, 1, atomic.LoadInt32(&hitCount))

	rows, err := h.app.Core().DB().GetCore().GetAll(context.Background(), `
SELECT request_url, success, trace_id
FROM supplier_platform_balance_log
WHERE platform_id = ?
ORDER BY id DESC
LIMIT 1
`, platformID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "http://"+strings.TrimPrefix(backupServer.URL, "http://")+"/api/customer", rows[0]["request_url"].String())
	require.Equal(t, 1, rows[0]["success"].Int())
	require.Equal(t, refreshData.TraceID, rows[0]["trace_id"].String())
}

func TestSupplierPlatformRefresh_PrefersPrimaryHTTPBeforeBackup(t *testing.T) {
	h := newSupplierIntegrationHarness(t)
	token := h.loginAdmin(t)
	subjectID := h.createSubject(t, token, "集成主体-顺序")

	var primaryHits int32
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&primaryHits, 1)
		require.Equal(t, "/api/customer", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"ok","msg":"查询成功","data":{"balance":"88.6600"}}`))
	}))
	defer primaryServer.Close()

	var backupHits int32
	backupServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backupHits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"ok","msg":"查询成功","data":{"balance":"66.8800"}}`))
	}))
	defer backupServer.Close()

	platformID := h.createPlatform(t, token, subjectID, strings.TrimPrefix(primaryServer.URL, "http://"), strings.TrimPrefix(backupServer.URL, "http://"))

	refreshRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(platformID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, refreshRes.Code)

	var refreshData struct {
		Balance string `json:"balance"`
		TraceID string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(refreshRes.Data, &refreshData))
	require.Equal(t, "88.6600", refreshData.Balance)
	require.NotEmpty(t, refreshData.TraceID)
	require.EqualValues(t, 1, atomic.LoadInt32(&primaryHits))
	require.EqualValues(t, 0, atomic.LoadInt32(&backupHits))

	rows, err := h.app.Core().DB().GetCore().GetAll(context.Background(), `
SELECT request_url, success
FROM supplier_platform_balance_log
WHERE platform_id = ?
ORDER BY id DESC
LIMIT 1
`, platformID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "http://"+strings.TrimPrefix(primaryServer.URL, "http://")+"/api/customer", rows[0]["request_url"].String())
	require.Equal(t, 1, rows[0]["success"].Int())
}

func TestSupplierPlatformRefresh_SplitsBusinessAndTransportFailures(t *testing.T) {
	h := newSupplierIntegrationHarness(t)
	token := h.loginAdmin(t)
	subjectID := h.createSubject(t, token, "集成主体-失败")

	mode := "success"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/customer", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch mode {
		case "success":
			_, _ = w.Write([]byte(`{"code":"ok","msg":"查询成功","data":{"balance":"24588.5010"}}`))
		case "business_fail":
			_, _ = w.Write([]byte(`{"code":"invalid_sign","msg":"签名错误"}`))
		default:
			http.Error(w, "unexpected mode", http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	platformID := h.createPlatform(t, token, subjectID, host, host)

	successRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(platformID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, successRes.Code)

	mode = "business_fail"
	failRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(platformID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, failRes.Code)

	var failData struct {
		Balance       string `json:"balance"`
		ConnectStatus int    `json:"connect_status"`
		Message       string `json:"message"`
		TraceID       string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(failRes.Data, &failData))
	require.Equal(t, "24588.5010", failData.Balance)
	require.Equal(t, 2, failData.ConnectStatus)
	require.Equal(t, "签名错误", failData.Message)
	require.NotEmpty(t, failData.TraceID)

	logRows, err := h.app.Core().DB().GetCore().GetAll(context.Background(), `
SELECT request_snapshot, response_snapshot, trace_id
FROM supplier_platform_balance_log
WHERE platform_id = ? AND success = 0
ORDER BY id DESC
LIMIT 1
`, platformID)
	require.NoError(t, err)
	require.Len(t, logRows, 1)
	require.NotContains(t, logRows[0]["request_snapshot"].String(), "secret-key-updated")
	require.NotContains(t, logRows[0]["request_snapshot"].String(), "1008612345")
	require.Contains(t, logRows[0]["request_snapshot"].String(), "1008")
	require.Contains(t, logRows[0]["request_snapshot"].String(), "2345")
	require.Contains(t, logRows[0]["response_snapshot"].String(), "invalid_sign")
	require.Equal(t, failData.TraceID, logRows[0]["trace_id"].String())

	updateRes := h.putJSON("/api/admin/supplier-platforms/"+int64ToString(platformID), map[string]any{
		"name":             "集成测试平台-传输失败",
		"domain":           unusedHost(t),
		"backup_domain":    unusedHost(t),
		"type_id":          35,
		"subject_id":       subjectID,
		"has_tax":          1,
		"token_id":         "1008612345",
		"secret_key":       "secret-key-updated",
		"threshold_amount": "5000.0000",
		"sort":             1,
		"crowd_name":       "集成测试群",
	}, token)
	require.Equal(t, 0, updateRes.Code)

	transportRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(platformID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, transportRes.Code)

	var transportData struct {
		Balance       string `json:"balance"`
		ConnectStatus int    `json:"connect_status"`
		Message       string `json:"message"`
		TraceID       string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(transportRes.Data, &transportData))
	require.Equal(t, "24588.5010", transportData.Balance)
	require.Equal(t, 2, transportData.ConnectStatus)
	require.NotEmpty(t, transportData.Message)
	require.NotEqual(t, "签名错误", transportData.Message)
	require.NotEmpty(t, transportData.TraceID)

	accountRows, err := h.app.Core().DB().GetCore().GetAll(context.Background(), `
SELECT last_balance, last_balance_status, last_balance_message, last_balance_trace_id
FROM supplier_platform_account
WHERE id = ?
`, platformID)
	require.NoError(t, err)
	require.Len(t, accountRows, 1)
	require.Equal(t, "24588.5010", accountRows[0]["last_balance"].String())
	require.Equal(t, 2, accountRows[0]["last_balance_status"].Int())
	require.Equal(t, transportData.TraceID, accountRows[0]["last_balance_trace_id"].String())
	require.Equal(t, transportData.Message, accountRows[0]["last_balance_message"].String())
}

func TestSupplierPlatformRefresh_LiveProviderBalance(t *testing.T) {
	if os.Getenv("MYJOB_RUN_SUPPLIER_LIVE") != "1" {
		t.Skip("set MYJOB_RUN_SUPPLIER_LIVE=1 to run live supplier balance verification")
	}

	typeID, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SUPPLIER_LIVE_TYPE_ID")))
	require.NoError(t, err)
	domain := strings.TrimSpace(os.Getenv("SUPPLIER_LIVE_DOMAIN"))
	backupDomain := strings.TrimSpace(os.Getenv("SUPPLIER_LIVE_BACKUP_DOMAIN"))
	if backupDomain == "" {
		backupDomain = domain
	}
	tokenID := strings.TrimSpace(os.Getenv("SUPPLIER_LIVE_TOKEN_ID"))
	secretKey := strings.TrimSpace(os.Getenv("SUPPLIER_LIVE_SECRET_KEY"))
	require.NotEmpty(t, domain)
	require.NotEmpty(t, tokenID)
	require.NotEmpty(t, secretKey)

	h := newSupplierIntegrationHarness(t)
	token := h.loginAdmin(t)
	subjectID := h.createSubject(t, token, "Live余额验证主体")
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             strings.TrimSpace(defaultString(os.Getenv("SUPPLIER_LIVE_NAME"), "Live余额验证平台")),
		"domain":           domain,
		"backup_domain":    backupDomain,
		"type_id":          typeID,
		"subject_id":       subjectID,
		"has_tax":          0,
		"token_id":         tokenID,
		"secret_key":       secretKey,
		"threshold_amount": "0.0000",
		"sort":             0,
		"crowd_name":       "Live余额验证",
	}, token)
	require.Equal(t, 0, res.Code)

	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &createData))
	require.NotZero(t, createData.ID)

	refreshRes := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, refreshRes.Code)

	var refreshData struct {
		Balance           string `json:"balance"`
		ConnectStatus     int    `json:"connect_status"`
		ConnectStatusText string `json:"connect_status_text"`
		Message           string `json:"message"`
		TraceID           string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(refreshRes.Data, &refreshData))
	require.Equal(t, 1, refreshData.ConnectStatus, "live provider response: %s", refreshData.Message)
	require.NotEmpty(t, refreshData.Balance)
	require.NotEmpty(t, refreshData.TraceID)
	t.Logf("live supplier balance=%s status=%s message=%s trace_id=%s", refreshData.Balance, refreshData.ConnectStatusText, refreshData.Message, refreshData.TraceID)
}

func unusedHost(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())
	return addr
}

func int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
