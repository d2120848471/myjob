package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"myjob/internal/bootstrap"
	orderlogic "myjob/internal/logic/order"
	"myjob/internal/model/entity"
	"myjob/internal/service"

	"github.com/stretchr/testify/require"
)

type orderIntegrationHarness struct {
	app          *bootstrap.Application
	handler      http.Handler
	orderService service.OrderService
}

func newOrderIntegrationHarness(t *testing.T) *orderIntegrationHarness {
	t.Helper()
	app, err := bootstrap.NewTestApplication()
	require.NoError(t, err)
	t.Cleanup(func() { _ = app.Close() })
	return &orderIntegrationHarness{
		app:          app,
		handler:      app.Handler(),
		orderService: orderlogic.NewOrderLogic(app.Core()),
	}
}

func TestOrderWorkerSubmitsPendingOrderToKakayun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "2478510", payload["goodsid"])
		require.Equal(t, "13800138000", payload["attach"])
		require.Contains(t, payload["usorderno"], "-T1")
		require.Equal(t, "10.0000", payload["maxmoney"])
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240001","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "processing", order.Status)
	require.Equal(t, 1, order.AttemptCount)
	attempt := h.loadCurrentAttempt(t, order.ID)
	require.Equal(t, "SD202604240001", attempt.SupplierOrderNo)
	require.Equal(t, "submitted", attempt.Status)
}

func TestOrderWorkerPassesKakayunMaxMoneyWithAllowedLoss(t *testing.T) {
	captured := make([]map[string]any, 0, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		captured = append(captured, payload)
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240099","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "允许亏本品牌", "视频会员", "爱奇艺")
	subjectID := h.createSubject(t, token, "允许亏本渠道主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "允许亏本商品", "10.0000")
	platformID := h.createKakayunPlatform(t, token, "允许亏本云发卡", subjectID, 0, strings.TrimPrefix(server.URL, "http://"))
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478599",
		"supplier_goods_name": "允许亏本测试商品",
		"source_cost_price":   "20.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	saveConfig := h.request(http.MethodPut, "/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   0,
		"reorder_timeout_enabled": 0,
		"reorder_timeout_minutes": 0,
		"order_strategy":          "fixed_order",
		"sync_cost_price_enabled": 0,
		"sync_goods_name_enabled": 0,
		"allow_loss_sale_enabled": 1,
		"max_loss_amount":         "2.5000",
		"combo_goods_enabled":     0,
	}, token)
	require.Equal(t, 0, saveConfig.Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	require.Len(t, captured, 1)
	require.Equal(t, "12.5000", captured[0]["maxmoney"])
}

func TestOrderWorkerPollsSuccessAndStops(t *testing.T) {
	state := "processing"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/order/create":
			_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240002","usorderno":"O-T1"}}`))
		case "/dockapiv3/order/get":
			status := 3
			if state == "success" {
				status = 5
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":1,"message":"ok","data":{"orderno":"SD202604240002","usorderno":"O-T1","status":%d,"refundstatus":0,"receipt":"ok"}}`, status)))
		}
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	state = "success"
	h.forceOrderDueForPoll(t, orderNo)
	require.NoError(t, h.orderService.PollDueOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "success", order.Status)
}

func TestOrderWorkerReordersOnlyInsideConfiguredWindow(t *testing.T) {
	failCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/order/create":
			failCount++
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			if failCount == 1 {
				require.Equal(t, "10.0000", payload["maxmoney"])
			} else {
				require.Equal(t, "11.0000", payload["maxmoney"])
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":1,"message":"下单成功","data":{"orderno":"SD20260424%04d","usorderno":"O-T%d"}}`, failCount, failCount)))
		case "/dockapiv3/order/get":
			_, _ = w.Write([]byte(`{"code":1,"message":"ok","data":{"orderno":"SD202604240003","usorderno":"O-T1","status":4,"refundstatus":1,"receipt":"失败"}}`))
		}
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithTwoChannelsAndReorderWindow(t, server.URL, 10)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	h.forceOrderDueForPoll(t, orderNo)
	require.NoError(t, h.orderService.PollDueOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "processing", order.Status)
	require.Equal(t, 2, order.AttemptCount)
}

func TestOrderWorkerReordersWhenCreateExplicitlyFails(t *testing.T) {
	var createCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		current := createCount.Add(1)
		if current == 1 {
			_, _ = w.Write([]byte(`{"code":0,"message":"库存不足"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240004","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithTwoChannelsAndReorderWindow(t, server.URL, 10)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "processing", order.Status)
	require.Equal(t, 2, order.AttemptCount)
	require.EqualValues(t, 2, createCount.Load())
}

func TestOrderWorkerDoesNotSubmitWhenGoodsDisabledAfterCreate(t *testing.T) {
	var createCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createCount.Add(1)
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240005","usorderno":"O-T1"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	order := h.loadOrder(t, orderNo)
	_, err := h.app.Core().DB().Exec(context.Background(), `UPDATE product_goods SET status = 0, updated_at = ? WHERE id = ?`, h.app.Core().Now(), order.GoodsID)
	require.NoError(t, err)

	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order = h.loadOrder(t, orderNo)
	require.Equal(t, "failed", order.Status)
	require.Equal(t, 0, order.AttemptCount)
	require.EqualValues(t, 0, createCount.Load())
}

func TestOrderWorkerUsesSelectedChannelSubjectAndAutoPriceAmounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		payload := map[string]any{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240008","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "订单利润品牌", "视频会员", "网易云")
	subjectID := h.createSubject(t, token, "订单利润渠道主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "订单利润商品", "2.0000")
	platformID := h.createKakayunPlatform(t, token, "订单利润云发卡", subjectID, 0, strings.TrimPrefix(server.URL, "http://"))

	createBinding := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478512",
		"supplier_goods_name": "云发卡利润测试商品",
		"source_cost_price":   "11.0000",
		"dock_status":         1,
		"sort":                10,
	}, token)
	require.Equal(t, 0, createBinding.Code)
	var bindingData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createBinding.Data, &bindingData))

	autoPrice := h.request(http.MethodPatch, "/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(bindingData.ID)+"/auto-price", map[string]any{
		"is_auto_change": 1,
		"add_type":       "fixed",
		"default_price":  "1.0000",
	}, token)
	require.Equal(t, 0, autoPrice.Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	createOrder := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, createOrder.Code)
	var createData struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(createOrder.Data, &createData))

	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order := h.loadOrder(t, createData.OrderNo)
	require.Equal(t, "12.0000", order.UnitPrice)
	require.Equal(t, "12.0000", order.OrderAmount)
	require.Equal(t, "11.0000", order.CostAmount)
	require.Equal(t, "1.0000", order.ProfitAmount)

	attempt := h.loadCurrentAttempt(t, order.ID)
	require.Equal(t, subjectID, attempt.PlatformSubjectID)
	require.Equal(t, "订单利润渠道主体", attempt.PlatformSubjectName)

	listRes := h.getJSON("/api/admin/orders?page=1&page_size=20&keyword="+createData.OrderNo+"&keyword_by=order_no", token)
	require.Equal(t, 0, listRes.Code)
	var listData struct {
		List []struct {
			SalesSubjectName string `json:"sales_subject_name"`
			OrderAmount      string `json:"order_amount"`
			CostAmount       string `json:"cost_amount"`
			ProfitAmount     string `json:"profit_amount"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, "订单利润渠道主体", listData.List[0].SalesSubjectName)
	require.Equal(t, "12.0000", listData.List[0].OrderAmount)
	require.Equal(t, "11.0000", listData.List[0].CostAmount)
	require.Equal(t, "1.0000", listData.List[0].ProfitAmount)
}

func TestSubmitPendingOnceClaimsOrderBeforeCallingSupplier(t *testing.T) {
	var requestCount atomic.Int32
	firstRequest := make(chan struct{})
	releaseFirst := make(chan struct{})
	var firstOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		current := requestCount.Add(1)
		if current == 1 {
			firstOnce.Do(func() { close(firstRequest) })
			<-releaseFirst
		}
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240006","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	done := make(chan error, 2)
	go func() { done <- h.orderService.SubmitPendingOnce(context.Background()) }()
	select {
	case <-firstRequest:
	case <-time.After(5 * time.Second):
		t.Fatal("first supplier request was not started")
	}

	go func() { done <- h.orderService.SubmitPendingOnce(context.Background()) }()
	time.Sleep(150 * time.Millisecond)
	close(releaseFirst)
	require.NoError(t, <-done)
	require.NoError(t, <-done)

	order := h.loadOrder(t, orderNo)
	require.Equal(t, "processing", order.Status)
	require.Equal(t, 1, order.AttemptCount)
	require.EqualValues(t, 1, requestCount.Load())
}

func TestOrderWorkerCreateCode9999BecomesUnknownAndPollsLater(t *testing.T) {
	var queryCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/order/create":
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			_, _ = w.Write([]byte(`{"code":9999,"message":"系统繁忙","data":{"usorderno":"` + payload["usorderno"].(string) + `"}}`))
		case "/dockapiv3/order/get":
			queryCount.Add(1)
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			_, _ = w.Write([]byte(`{"code":1,"message":"ok","data":{"orderno":"SD202604240007","usorderno":"` + payload["usorderno"].(string) + `","status":5,"refundstatus":0,"receipt":"成功"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "unknown", order.Status)
	require.Equal(t, 1, order.AttemptCount)
	require.EqualValues(t, 1, h.scalarInt(t, `SELECT COUNT(*) FROM external_order WHERE order_no = ? AND current_attempt_id IS NOT NULL AND next_poll_at IS NOT NULL`, orderNo))
	attempt := h.loadCurrentAttempt(t, order.ID)
	require.Equal(t, "unknown", attempt.Status)
	require.Contains(t, attempt.SupplierUSOrderNo, "-T1")

	h.forceOrderDueForPoll(t, orderNo)
	require.NoError(t, h.orderService.PollDueOnce(context.Background()))
	order = h.loadOrder(t, orderNo)
	require.Equal(t, "success", order.Status)
	require.EqualValues(t, 1, queryCount.Load())
}

func TestOrderWorkerDoesNotReorderOnUnknownCreate(t *testing.T) {
	var createCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		createCount.Add(1)
		_, _ = w.Write([]byte(`{"code":9999,"message":"状态未知","data":{"usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithTwoChannelsAndReorderWindow(t, server.URL, 10)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "unknown", order.Status)
	require.Equal(t, 1, order.AttemptCount)
	require.EqualValues(t, 1, createCount.Load())
}

func TestOrderWorkerRecoversStuckSubmittingOrder(t *testing.T) {
	var createCount atomic.Int32
	var queryCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/order/create":
			createCount.Add(1)
			_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"unexpected","usorderno":"unexpected"}}`))
		case "/dockapiv3/order/get":
			queryCount.Add(1)
			_, _ = w.Write([]byte(`{"code":9999,"message":"订单仍需确认"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	_, err := h.app.Core().DB().Exec(context.Background(), `
UPDATE external_order
SET status = 'processing', current_attempt_id = NULL, next_poll_at = NULL, updated_at = ?
WHERE order_no = ?
`, h.app.Core().Now(), orderNo)
	require.NoError(t, err)

	require.NoError(t, h.orderService.PollDueOnce(context.Background()))
	order := h.loadOrder(t, orderNo)
	require.Equal(t, "unknown", order.Status)
	require.Equal(t, 1, order.AttemptCount)
	require.EqualValues(t, 1, h.scalarInt(t, `SELECT COUNT(*) FROM external_order WHERE order_no = ? AND current_attempt_id IS NOT NULL AND next_poll_at IS NOT NULL`, orderNo))
	attempt := h.loadCurrentAttempt(t, order.ID)
	require.Equal(t, "unknown", attempt.Status)
	require.Equal(t, orderNo+"-T1", attempt.SupplierUSOrderNo)
	require.EqualValues(t, 0, createCount.Load())
	require.EqualValues(t, 1, queryCount.Load())
}

func TestExternalOrderNullableFieldsScanSafely(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"message":"noop"}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	orderNo := h.createPendingOpenOrderWithKakayunServer(t, server.URL)
	var order entity.ExternalOrder
	require.NoError(t, h.app.Core().DB().GetCore().GetScan(context.Background(), &order, `SELECT * FROM external_order WHERE order_no = ?`, orderNo))
	require.Equal(t, orderNo, order.OrderNo)
	require.Equal(t, "pending_submit", order.Status)
}

func (h *orderIntegrationHarness) request(method, path string, body any, token string) supplierAPIEnvelope {
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

func (h *orderIntegrationHarness) getJSON(path string, token string) supplierAPIEnvelope {
	return h.request(http.MethodGet, path, nil, token)
}

func (h *orderIntegrationHarness) postJSON(path string, body any, token string) supplierAPIEnvelope {
	return h.request(http.MethodPost, path, body, token)
}

func (h *orderIntegrationHarness) loginAdmin(t *testing.T) string {
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

func (h *orderIntegrationHarness) createPendingOpenOrderWithKakayunServer(t *testing.T, serverURL string) string {
	t.Helper()
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "订单集成品牌", "视频会员", "腾讯视频")
	subjectID := h.createSubject(t, token, "订单集成主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "订单集成商品", "20.0000")
	platformID := h.createKakayunPlatform(t, token, "订单集成云发卡", subjectID, 0, strings.TrimPrefix(serverURL, "http://"))
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478510",
		"supplier_goods_name": "云发卡测试直充商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	var createData struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotEmpty(t, createData.OrderNo)
	return createData.OrderNo
}

func (h *orderIntegrationHarness) createPendingOpenOrderWithTwoChannelsAndReorderWindow(t *testing.T, serverURL string, timeoutMinutes int) string {
	t.Helper()
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "订单补单品牌", "视频会员", "腾讯视频")
	subjectID := h.createSubject(t, token, "订单补单主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "订单补单商品", "20.0000")
	host := strings.TrimPrefix(serverURL, "http://")
	firstPlatformID := h.createKakayunPlatformWithToken(t, token, "订单补单云发卡A", subjectID, 0, host, "10052-a")
	secondPlatformID := h.createKakayunPlatformWithToken(t, token, "订单补单云发卡B", subjectID, 0, host, "10052-b")
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": firstPlatformID,
		"supplier_goods_no":   "2478510",
		"supplier_goods_name": "云发卡测试直充商品A",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": secondPlatformID,
		"supplier_goods_no":   "2478511",
		"supplier_goods_name": "云发卡测试直充商品B",
		"source_cost_price":   "11.0000",
		"dock_status":         1,
		"sort":                20,
	}, token).Code)
	saveConfig := h.request(http.MethodPut, "/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"reorder_timeout_enabled": 1,
		"reorder_timeout_minutes": timeoutMinutes,
		"order_strategy":          "fixed_order",
		"sync_cost_price_enabled": 0,
		"sync_goods_name_enabled": 0,
		"allow_loss_sale_enabled": 0,
		"max_loss_amount":         "0.0000",
		"combo_goods_enabled":     0,
	}, token)
	require.Equal(t, 0, saveConfig.Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	var createData struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotEmpty(t, createData.OrderNo)
	return createData.OrderNo
}

func (h *orderIntegrationHarness) createBrandPath(t *testing.T, token, topName, childName, leafName string) int64 {
	t.Helper()
	top := h.postJSON("/api/admin/brands", map[string]any{"name": topName, "is_visible": 1}, token)
	require.Equal(t, 0, top.Code)
	var topData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(top.Data, &topData))

	child := h.postJSON("/api/admin/brands", map[string]any{"parent_id": topData.ID, "name": childName, "is_visible": 1}, token)
	require.Equal(t, 0, child.Code)
	var childData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(child.Data, &childData))

	leaf := h.postJSON("/api/admin/brands", map[string]any{"parent_id": childData.ID, "name": leafName, "is_visible": 1}, token)
	require.Equal(t, 0, leaf.Code)
	var leafData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(leaf.Data, &leafData))
	return leafData.ID
}

func (h *orderIntegrationHarness) createSubject(t *testing.T, token, name string, hasTax int) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/subjects", map[string]any{"name": name, "has_tax": hasTax}, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}

func (h *orderIntegrationHarness) createDirectRechargeGoods(t *testing.T, token string, brandID int64, name, defaultSellPrice string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":           brandID,
		"name":               name,
		"goods_type":         "direct_recharge",
		"supply_type":        "channel",
		"is_export":          1,
		"is_douyin":          0,
		"has_tax":            0,
		"exception_notify":   1,
		"balance_limit":      "0",
		"default_sell_price": defaultSellPrice,
		"min_purchase_qty":   1,
		"max_purchase_qty":   1,
		"status":             1,
	}, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}

func (h *orderIntegrationHarness) createKakayunPlatform(t *testing.T, token, name string, subjectID int64, hasTax int, host string) int64 {
	t.Helper()
	return h.createKakayunPlatformWithToken(t, token, name, subjectID, hasTax, host, "10052")
}

func (h *orderIntegrationHarness) createKakayunPlatformWithToken(t *testing.T, token, name string, subjectID int64, hasTax int, host, tokenID string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           host,
		"backup_domain":    host,
		"type_id":          6,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             1,
		"crowd_name":       "订单群",
	}, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}

func (h *orderIntegrationHarness) forceOrderDueForPoll(t *testing.T, orderNo string) {
	t.Helper()
	_, err := h.app.Core().DB().Exec(context.Background(), `UPDATE external_order SET next_poll_at = ? WHERE order_no = ?`, h.app.Core().Now().Add(-time.Second), orderNo)
	require.NoError(t, err)
}

func (h *orderIntegrationHarness) loadOrder(t *testing.T, orderNo string) entity.ExternalOrder {
	t.Helper()
	order := entity.ExternalOrder{}
	require.NoError(t, h.app.Core().DB().GetCore().GetScan(context.Background(), &order, `SELECT * FROM external_order WHERE order_no = ?`, orderNo))
	return order
}

func (h *orderIntegrationHarness) loadCurrentAttempt(t *testing.T, orderID int64) entity.ExternalOrderAttempt {
	t.Helper()
	attempt := entity.ExternalOrderAttempt{}
	require.NoError(t, h.app.Core().DB().GetCore().GetScan(context.Background(), &attempt, `SELECT * FROM external_order_attempt WHERE order_id = ? ORDER BY id DESC LIMIT 1`, orderID))
	return attempt
}

func (h *orderIntegrationHarness) scalarInt(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	value, err := h.app.Core().DB().GetCore().GetValue(context.Background(), query, args...)
	require.NoError(t, err)
	return value.Int64()
}
