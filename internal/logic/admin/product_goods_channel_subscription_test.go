package adminlogic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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
