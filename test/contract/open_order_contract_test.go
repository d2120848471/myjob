package contract_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func signOpenV1(appSecret, method, path, timestamp, nonce string, body []byte) string {
	canonical := strings.ToUpper(method) + "\n" +
		path + "\n" +
		timestamp + "\n" +
		nonce + "\n" +
		sha256Hex(body)
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestOpenOrder_CreateAndQuery_SignatureAndNonce(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/buygoods":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH-OPEN-001"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstream.Close()

	h := newTestHarness(t)
	ctx := context.Background()
	now := h.app.Core().Now()

	// 交易基础数据：主体、模板、商品、渠道账号、绑定。
	subjectResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	templateResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO product_template (title, template_type, is_shared, account_name, validate_type, created_at, updated_at)
VALUES ('手机号模板', 'local', 0, '手机号', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	templateID, _ := templateResult.LastInsertId()

	goodsCode := "P-OPEN-001"
	goodsResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '开放商品', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	accountResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('优卡云账号', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = h.app.Core().DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '100', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	// Open 调用方。
	appKey := "app-key-001"
	appSecret := "app-secret-001"
	_, err = h.app.Core().DB().Exec(ctx, `
INSERT INTO open_caller (name, app_key, app_secret, status, allowed_ip_list, created_at, updated_at)
VALUES ('调用方A', ?, ?, 'enabled', '["127.0.0.1"]', ?, ?)
`, appKey, appSecret, now, now)
	require.NoError(t, err)

	body := map[string]any{
		"client_order_no": "C-OPEN-001",
		"goods_code":      goodsCode,
		"quantity":        1,
		"payload": map[string]any{
			"mobile": "13800138000",
		},
	}
	bodyRaw, _ := json.Marshal(body)
	timestamp := strconv.FormatInt(now.Unix(), 10)
	nonce := "n-open-001"
	path := "/api/open/orders"

	signature := signOpenV1(appSecret, http.MethodPost, path, timestamp, nonce, bodyRaw)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Key", appKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", signature)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)

	var env apiEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, 0, env.Code)

	var data struct {
		OrderNo string `json:"order_no"`
		Status  string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(env.Data, &data))
	require.NotEmpty(t, data.OrderNo)
	require.Equal(t, "processing", data.Status)

	// nonce 重放：同 nonce 第二次应被拒绝（即便 client_order_no 幂等）。
	req2 := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-App-Key", appKey)
	req2.Header.Set("X-Timestamp", timestamp)
	req2.Header.Set("X-Nonce", nonce)
	req2.Header.Set("X-Signature", signature)
	req2.RemoteAddr = "127.0.0.1:12345"
	rec2 := httptest.NewRecorder()
	h.handler.ServeHTTP(rec2, req2)
	require.Equal(t, 200, rec2.Code)

	var env2 apiEnvelope
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &env2))
	require.NotEqual(t, 0, env2.Code)

	// 幂等：同 client_order_no，但更换 nonce，应返回同一 order_no。
	nonce3 := "n-open-001-3"
	signature3 := signOpenV1(appSecret, http.MethodPost, path, timestamp, nonce3, bodyRaw)
	req3 := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-App-Key", appKey)
	req3.Header.Set("X-Timestamp", timestamp)
	req3.Header.Set("X-Nonce", nonce3)
	req3.Header.Set("X-Signature", signature3)
	req3.RemoteAddr = "127.0.0.1:12345"
	rec3 := httptest.NewRecorder()
	h.handler.ServeHTTP(rec3, req3)
	require.Equal(t, 200, rec3.Code)

	var env3 apiEnvelope
	require.NoError(t, json.Unmarshal(rec3.Body.Bytes(), &env3))
	require.Equal(t, 0, env3.Code)
	var data3 struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(env3.Data, &data3))
	require.Equal(t, data.OrderNo, data3.OrderNo)

	// 按内部订单号查单。
	getPath := "/api/open/orders/" + data.OrderNo
	getTimestamp := strconv.FormatInt(now.Unix(), 10)
	getNonce := "n-open-get-001"
	getSig := signOpenV1(appSecret, http.MethodGet, getPath, getTimestamp, getNonce, nil)

	getReq := httptest.NewRequest(http.MethodGet, getPath, nil)
	getReq.Header.Set("X-App-Key", appKey)
	getReq.Header.Set("X-Timestamp", getTimestamp)
	getReq.Header.Set("X-Nonce", getNonce)
	getReq.Header.Set("X-Signature", getSig)
	getReq.RemoteAddr = "127.0.0.1:12345"
	getRec := httptest.NewRecorder()
	h.handler.ServeHTTP(getRec, getReq)
	require.Equal(t, 200, getRec.Code)

	var getEnv apiEnvelope
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &getEnv))
	require.Equal(t, 0, getEnv.Code)

	var getData struct {
		OrderNo        string `json:"order_no"`
		ClientOrderNo  string `json:"client_order_no"`
		Status         string `json:"status"`
		UpstreamOrders []any  `json:"upstream_orders"`
	}
	require.NoError(t, json.Unmarshal(getEnv.Data, &getData))
	require.Equal(t, data.OrderNo, getData.OrderNo)
	require.Equal(t, "C-OPEN-001", getData.ClientOrderNo)
	require.Equal(t, "processing", getData.Status)
	require.NotEmpty(t, getData.UpstreamOrders)

	// 按调用方订单号查单。
	getByClientPath := "/api/open/orders/by-client/C-OPEN-001"
	getNonce2 := "n-open-get-002"
	getSig2 := signOpenV1(appSecret, http.MethodGet, getByClientPath, getTimestamp, getNonce2, nil)

	getReq2 := httptest.NewRequest(http.MethodGet, getByClientPath, nil)
	getReq2.Header.Set("X-App-Key", appKey)
	getReq2.Header.Set("X-Timestamp", getTimestamp)
	getReq2.Header.Set("X-Nonce", getNonce2)
	getReq2.Header.Set("X-Signature", getSig2)
	getReq2.RemoteAddr = "127.0.0.1:12345"
	getRec2 := httptest.NewRecorder()
	h.handler.ServeHTTP(getRec2, getReq2)
	require.Equal(t, 200, getRec2.Code)

	var getEnv2 apiEnvelope
	require.NoError(t, json.Unmarshal(getRec2.Body.Bytes(), &getEnv2))
	require.Equal(t, 0, getEnv2.Code)

	var getData2 struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(getEnv2.Data, &getData2))
	require.Equal(t, data.OrderNo, getData2.OrderNo)
}

func TestOpenSignature_RejectsInvalidSignatureAndTimestamp(t *testing.T) {
	h := newTestHarness(t)
	ctx := context.Background()
	now := h.app.Core().Now()

	appKey := "app-key-002"
	appSecret := "app-secret-002"
	_, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO open_caller (name, app_key, app_secret, status, allowed_ip_list, created_at, updated_at)
VALUES ('调用方B', ?, ?, 'enabled', '["127.0.0.1"]', ?, ?)
`, appKey, appSecret, now, now)
	require.NoError(t, err)

	body := map[string]any{
		"client_order_no": "C-OPEN-002",
		"goods_code":      "P-NOT-EXIST",
		"quantity":        1,
		"payload":         map[string]any{"mobile": "13800138000"},
	}
	bodyRaw, _ := json.Marshal(body)
	path := "/api/open/orders"

	// 签名错误。
	timestamp := strconv.FormatInt(now.Unix(), 10)
	nonce := "n-open-002"
	badSig := "bad"

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Key", appKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", badSig)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)

	var env apiEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.NotEqual(t, 0, env.Code)

	// 时间戳过期。
	oldTimestamp := strconv.FormatInt(now.Add(-10*time.Minute).Unix(), 10)
	nonce2 := "n-open-003"
	sig := signOpenV1(appSecret, http.MethodPost, path, oldTimestamp, nonce2, bodyRaw)
	req2 := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-App-Key", appKey)
	req2.Header.Set("X-Timestamp", oldTimestamp)
	req2.Header.Set("X-Nonce", nonce2)
	req2.Header.Set("X-Signature", sig)
	req2.RemoteAddr = "127.0.0.1:12345"
	rec2 := httptest.NewRecorder()
	h.handler.ServeHTTP(rec2, req2)
	require.Equal(t, 200, rec2.Code)

	var env2 apiEnvelope
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &env2))
	require.NotEqual(t, 0, env2.Code)
}

func TestOpenSignature_RejectsIPNotAllowed(t *testing.T) {
	h := newTestHarness(t)
	ctx := context.Background()
	now := h.app.Core().Now()

	appKey := "app-key-003"
	appSecret := "app-secret-003"
	_, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO open_caller (name, app_key, app_secret, status, allowed_ip_list, created_at, updated_at)
VALUES ('调用方C', ?, ?, 'enabled', '["10.0.0.1"]', ?, ?)
`, appKey, appSecret, now, now)
	require.NoError(t, err)

	body := map[string]any{
		"client_order_no": "C-OPEN-003",
		"goods_code":      "P-NOT-EXIST",
		"quantity":        1,
		"payload":         map[string]any{"mobile": "13800138000"},
	}
	bodyRaw, _ := json.Marshal(body)
	path := "/api/open/orders"

	timestamp := strconv.FormatInt(now.Unix(), 10)
	nonce := "n-open-004"
	sig := signOpenV1(appSecret, http.MethodPost, path, timestamp, nonce, bodyRaw)

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(bodyRaw)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Key", appKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", sig)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)

	var env apiEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.NotEqual(t, 0, env.Code)
	require.Equal(t, "ip_not_allowed", env.Message)
}
