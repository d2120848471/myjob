package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"myjob/internal/bootstrap"

	"github.com/stretchr/testify/require"
)

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func TestRuntime_LoginSmoke(t *testing.T) {
	if os.Getenv("MYJOB_RUN_INTEGRATION") != "1" {
		t.Skip("set MYJOB_RUN_INTEGRATION=1 to run integration smoke tests")
	}
	if os.Getenv("SUPER_ADMIN_PHONE") == "" || os.Getenv("SUPER_ADMIN_PASSWORD") == "" {
		t.Skip("SUPER_ADMIN_PHONE and SUPER_ADMIN_PASSWORD are required")
	}

	app, err := bootstrap.NewApplicationFromEnv()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	// 集成测试显式启动真实 server，覆盖 GoFrame 运行时初始化链路。
	app.Server().SetAddr("127.0.0.1:0")
	require.NoError(t, app.Server().Start())

	body, err := json.Marshal(map[string]any{
		"username": "admin",
		"password": os.Getenv("SUPER_ADMIN_PASSWORD"),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var env apiEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, 0, env.Code)
}
