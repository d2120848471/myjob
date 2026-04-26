package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/model/entity"

	"github.com/stretchr/testify/require"
)

func TestAutoSubscribeKakayunBindingRecordsSuccess(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write([]byte(`{"code":1,"msg":"success","data":[]}`))
		case "/dockapiv3/user/seturl", "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	callbackURL := "https://public.example.com/api/open/supplier-platforms/kakayun/" + int64ToStringForAdminTest(binding.PlatformAccountID) + "/product-change-callback"

	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, callbackURL, supplierProductSubscriptionActionSubscribe)
	require.NoError(t, err)

	require.Contains(t, requests, "/dockapiv3/user/geturl")
	require.Contains(t, requests, "/dockapiv3/user/seturl")
	require.Contains(t, requests, "/dockapiv3/goods/subscribe")

	status, err := core.DB().GetCore().GetValue(context.Background(), `SELECT status FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	require.Equal(t, supplierProductSubscriptionStatusSubscribed, status.String())
}

func TestAutoSubscribeKakayunBindingParsesRawResponseWhenSnapshotIsTruncated(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	logic := NewProductGoodsLogic(core)
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	callbackURL := "https://public.example.com/api/open/supplier-platforms/kakayun/" + int64ToStringForAdminTest(binding.PlatformAccountID) + "/product-change-callback"
	items := []map[string]any{{"url": callbackURL, "createtime": 1735002160}}
	for i := 0; i < 160; i++ {
		items = append(items, map[string]any{
			"url":        "https://public.example.com/very-long-callback-path/" + strconv.Itoa(i) + "/abcdefghijklmnopqrstuvwxyz",
			"createtime": 1735002160 + i,
		})
	}
	rawGetURLsResponse, err := json.Marshal(map[string]any{"code": 1, "msg": "success", "data": items})
	require.NoError(t, err)
	require.Greater(t, len(rawGetURLsResponse), 4096)

	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write(rawGetURLsResponse)
		case "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL

	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, callbackURL, supplierProductSubscriptionActionSubscribe)
	require.NoError(t, err)

	require.Contains(t, requests, "/dockapiv3/user/geturl")
	require.NotContains(t, requests, "/dockapiv3/user/seturl")
	require.Contains(t, requests, "/dockapiv3/goods/subscribe")
	status, err := core.DB().GetCore().GetValue(context.Background(), `SELECT status FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	require.Equal(t, supplierProductSubscriptionStatusSubscribed, status.String())
}

func TestAutoSubscribeKakayunBindingAddsReceiveURLWithoutOldReceiveURL(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	var setURLBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write([]byte(`{"code":1,"msg":"success","data":[{"url":"https://other.example.com/callback","createtime":1735002160}]}`))
		case "/dockapiv3/user/seturl":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&setURLBody))
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		case "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	callbackURL := "https://public.example.com/api/open/supplier-platforms/kakayun/" + int64ToStringForAdminTest(binding.PlatformAccountID) + "/product-change-callback"

	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, callbackURL, supplierProductSubscriptionActionSubscribe)
	require.NoError(t, err)
	require.Equal(t, callbackURL, setURLBody["receiveurl"])
	require.NotContains(t, setURLBody, "oldreceiveurl")
}

func TestAutoSubscribeKakayunBindingRecordsFailureWithoutReturningError(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"msg":"签名错误"}`))
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	callbackURL := "https://public.example.com/api/open/supplier-platforms/kakayun/" + int64ToStringForAdminTest(binding.PlatformAccountID) + "/product-change-callback"

	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, callbackURL, supplierProductSubscriptionActionSubscribe)
	require.NoError(t, err)

	var row struct {
		Status    string `db:"status"`
		LastError string `db:"last_error"`
	}
	err = core.DB().GetCore().GetScan(context.Background(), &row, `SELECT status, last_error FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	require.Equal(t, supplierProductSubscriptionStatusFailed, row.Status)
	require.Contains(t, row.LastError, "签名错误")
}

func TestAutoSubscribeKakayunBindingFailurePreservesHistoricalTimes(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"msg":"签名错误"}`))
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	callbackURL := "https://public.example.com/api/open/supplier-platforms/kakayun/" + int64ToStringForAdminTest(binding.PlatformAccountID) + "/product-change-callback"
	subscribedAt := time.Date(2026, 4, 26, 8, 0, 0, 0, time.Local)
	canceledAt := time.Date(2026, 4, 26, 9, 0, 0, 0, time.Local)
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusCanceled, supplierProductSubscriptionActionCancel, callbackURL, "", "{}", "{}", subscribedAt, canceledAt)
	require.NoError(t, err)

	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, callbackURL, supplierProductSubscriptionActionSubscribe)
	require.NoError(t, err)

	row := loadSubscriptionStatusForTest(t, core, binding.BindingID)
	require.Equal(t, supplierProductSubscriptionStatusFailed, row.Status)
	require.Contains(t, row.LastError, "签名错误")
	requireSubscriptionTimeEqual(t, subscribedAt, row.SubscribedAt)
	requireSubscriptionTimeEqual(t, canceledAt, row.CanceledAt)
}

func TestCancelSubscriptionPreservesSubscribedAt(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	logic := NewProductGoodsLogic(core)
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	now := core.Now()
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusSubscribed, supplierProductSubscriptionActionSubscribe, "https://public.example.com/callback", "", "{}", "{}", now, nil)
	require.NoError(t, err)

	idValue, err := core.DB().GetCore().GetValue(context.Background(), `SELECT id FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	id := idValue.Int64()

	_, err = logic.CancelSupplierProductSubscription(context.Background(), &adminapi.SupplierProductSubscriptionCancelReq{ID: id}, entity.AdminUser{}, "127.0.0.1")
	require.NoError(t, err)

	count, err := core.DB().GetCore().GetValue(context.Background(), `SELECT COUNT(*) FROM supplier_product_subscription WHERE id = ? AND status = ? AND subscribed_at IS NOT NULL AND canceled_at IS NOT NULL`, id, supplierProductSubscriptionStatusCanceled)
	require.NoError(t, err)
	require.Equal(t, 1, count.Int())
}

func TestCancelSubscriptionFailureKeepsOriginalStatusAndCanceledAt(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/goods/cancelsubscribe", r.URL.Path)
		_, _ = w.Write([]byte(`{"code":0,"msg":"签名错误"}`))
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	subscribedAt := time.Date(2026, 4, 26, 8, 0, 0, 0, time.Local)
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusSubscribed, supplierProductSubscriptionActionSubscribe, "https://public.example.com/callback", "", "{}", "{}", subscribedAt, nil)
	require.NoError(t, err)

	idValue, err := core.DB().GetCore().GetValue(context.Background(), `SELECT id FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	id := idValue.Int64()

	_, err = logic.CancelSupplierProductSubscription(context.Background(), &adminapi.SupplierProductSubscriptionCancelReq{ID: id}, entity.AdminUser{}, "127.0.0.1")
	require.Error(t, err)

	row := loadSubscriptionStatusForTest(t, core, binding.BindingID)
	require.Equal(t, supplierProductSubscriptionStatusSubscribed, row.Status)
	require.Equal(t, supplierProductSubscriptionActionCancel, row.LastAction)
	require.Contains(t, row.LastError, "签名错误")
	require.False(t, row.CanceledAt.Valid)
	requireSubscriptionTimeEqual(t, subscribedAt, row.SubscribedAt)
}

func TestResubscribeSubscriptionRecordsResubscribeAction(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write([]byte(`{"code":1,"msg":"success","data":[]}`))
		case "/dockapiv3/user/seturl", "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusCanceled, supplierProductSubscriptionActionCancel, "https://public.example.com/old-callback", "", "{}", "{}", nil, core.Now())
	require.NoError(t, err)

	idValue, err := core.DB().GetCore().GetValue(context.Background(), `SELECT id FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	id := idValue.Int64()

	_, err = logic.ResubscribeSupplierProductSubscription(context.Background(), &adminapi.SupplierProductSubscriptionResubscribeReq{ID: id}, entity.AdminUser{}, "127.0.0.1")
	require.NoError(t, err)

	var row struct {
		LastAction string `db:"last_action"`
	}
	err = core.DB().GetCore().GetScan(context.Background(), &row, `SELECT last_action FROM supplier_product_subscription WHERE id = ?`, id)
	require.NoError(t, err)
	require.Equal(t, supplierProductSubscriptionActionResubscribe, row.LastAction)
}

func TestResubscribeSubscriptionFailureReturnsErrorAndPreservesHistoricalTimes(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write([]byte(`{"code":1,"msg":"success","data":[{"url":"https://public.example.com/old-callback","createtime":1735002160}]}`))
		case "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":0,"msg":"签名错误"}`))
		default:
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		}
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	subscribedAt := time.Date(2026, 4, 26, 8, 0, 0, 0, time.Local)
	canceledAt := time.Date(2026, 4, 26, 9, 0, 0, 0, time.Local)
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusCanceled, supplierProductSubscriptionActionCancel, "https://public.example.com/old-callback", "", "{}", "{}", subscribedAt, canceledAt)
	require.NoError(t, err)

	idValue, err := core.DB().GetCore().GetValue(context.Background(), `SELECT id FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	id := idValue.Int64()

	_, err = logic.ResubscribeSupplierProductSubscription(context.Background(), &adminapi.SupplierProductSubscriptionResubscribeReq{ID: id}, entity.AdminUser{}, "127.0.0.1")
	require.Error(t, err)

	row := loadSubscriptionStatusForTest(t, core, binding.BindingID)
	require.Equal(t, supplierProductSubscriptionStatusFailed, row.Status)
	require.Equal(t, supplierProductSubscriptionActionResubscribe, row.LastAction)
	require.Contains(t, row.LastError, "签名错误")
	requireSubscriptionTimeEqual(t, subscribedAt, row.SubscribedAt)
	requireSubscriptionTimeEqual(t, canceledAt, row.CanceledAt)
}

func TestResubscribeSubscriptionSuccessSetsSubscribedStatus(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dockapiv3/user/geturl":
			_, _ = w.Write([]byte(`{"code":1,"msg":"success","data":[{"url":"https://public.example.com/old-callback","createtime":1735002160}]}`))
		case "/dockapiv3/goods/subscribe":
			_, _ = w.Write([]byte(`{"code":1,"msg":"成功"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL
	binding := seedKakayunSubscriptionBinding(t, logic, core)
	err = logic.upsertSupplierProductSubscription(context.Background(), binding, supplierProductSubscriptionStatusCanceled, supplierProductSubscriptionActionCancel, "https://public.example.com/old-callback", "历史错误", "{}", "{}", nil, core.Now())
	require.NoError(t, err)

	idValue, err := core.DB().GetCore().GetValue(context.Background(), `SELECT id FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	id := idValue.Int64()

	_, err = logic.ResubscribeSupplierProductSubscription(context.Background(), &adminapi.SupplierProductSubscriptionResubscribeReq{ID: id}, entity.AdminUser{}, "127.0.0.1")
	require.NoError(t, err)

	row := loadSubscriptionStatusForTest(t, core, binding.BindingID)
	require.Equal(t, supplierProductSubscriptionStatusSubscribed, row.Status)
	require.Equal(t, supplierProductSubscriptionActionResubscribe, row.LastAction)
	require.Empty(t, row.LastError)
	require.True(t, row.SubscribedAt.Valid)
}

func seedKakayunSubscriptionBinding(t *testing.T, logic *ProductGoodsLogic, core *app.Core) productGoodsChannelSubscriptionTarget {
	t.Helper()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, "qqlogin.yxp8.cn", 1, 1)
	candidate := loadSinglePriceChangeCandidate(t, logic, goodsID)
	target, err := logic.loadProductGoodsChannelSubscriptionTarget(context.Background(), candidate.BindingID)
	require.NoError(t, err)
	return target
}

func int64ToStringForAdminTest(value int64) string {
	return strconv.FormatInt(value, 10)
}

type subscriptionStatusForTest struct {
	Status       string       `db:"status"`
	LastAction   string       `db:"last_action"`
	LastError    string       `db:"last_error"`
	SubscribedAt sql.NullTime `db:"subscribed_at"`
	CanceledAt   sql.NullTime `db:"canceled_at"`
}

func loadSubscriptionStatusForTest(t *testing.T, core *app.Core, bindingID int64) subscriptionStatusForTest {
	t.Helper()
	row, err := core.DB().GetCore().GetOne(context.Background(), `
SELECT status, last_action, last_error, subscribed_at, canceled_at
FROM supplier_product_subscription
WHERE binding_id = ?
`, bindingID)
	require.NoError(t, err)
	require.NotNil(t, row)
	return subscriptionStatusForTest{
		Status:       row["status"].String(),
		LastAction:   row["last_action"].String(),
		LastError:    row["last_error"].String(),
		SubscribedAt: nullableTimeFromRecord(row, "subscribed_at"),
		CanceledAt:   nullableTimeFromRecord(row, "canceled_at"),
	}
}

func requireSubscriptionTimeEqual(t *testing.T, expected time.Time, actual sql.NullTime) {
	t.Helper()
	require.True(t, actual.Valid)
	require.Equal(t, expected.Format("2006-01-02 15:04:05"), actual.Time.Format("2006-01-02 15:04:05"))
}
