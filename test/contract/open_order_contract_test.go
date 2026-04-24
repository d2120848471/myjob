package contract_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_OpenOrderPathsExposed(t *testing.T) {
	h := newTestHarness(t)
	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/open/orders")
	require.Contains(t, res.body, "/api/open/orders/{orderNo}")
}

func TestOpenOrderRejectsInvalidToken(t *testing.T) {
	h := newTestHarness(t)
	res := h.postJSON("/api/open/orders", map[string]any{
		"token":    "wrong",
		"goods_id": "G0001",
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 401, res.Code)
}

func TestOpenOrderQueryRejectsInvalidToken(t *testing.T) {
	h := newTestHarness(t)
	res := h.getJSON("/api/open/orders/O20260424153000123456?token=wrong", "")
	require.Equal(t, 401, res.Code)
}

func TestOpenOrderQueryByQueryParamRejectsInvalidToken(t *testing.T) {
	h := newTestHarness(t)
	res := h.getJSON("/api/open/orders?token=wrong&order_no=O20260424153000123456", "")
	require.Equal(t, 401, res.Code)
}

func TestOpenOrderCreateAndQueryMinimumFlow(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	_, _, leafBrandID := h.createBrandPath(t, token, "开放订单品牌", "视频会员", "腾讯视频月卡")
	platformSubjectID := h.createSubject(t, token, "开放订单主体", 0)
	goodsID := h.createDirectRechargeChannelProductGoods(t, token, leafBrandID, "开放订单商品", "20.0000")
	platformID := h.createKakayunPlatformAccount(t, token, "云发卡开放订单账号", platformSubjectID, 0, "open-order-token-id")
	createBinding := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478510",
		"supplier_goods_name": "云发卡测试直充商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token)
	require.Equal(t, 0, createBinding.Code)

	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	var createData struct {
		OrderNo    string `json:"order_no"`
		StatusCode string `json:"status_code"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotEmpty(t, createData.OrderNo)
	require.Equal(t, "pending_submit", createData.StatusCode)

	query := h.getJSON("/api/open/orders/"+createData.OrderNo+"?token=test-open-order-token", "")
	require.Equal(t, 0, query.Code)
	var queryData struct {
		OrderNo    string `json:"order_no"`
		StatusCode string `json:"status_code"`
		GoodsID    string `json:"goods_id"`
		Account    string `json:"account"`
		Quantity   int    `json:"quantity"`
	}
	require.NoError(t, json.Unmarshal(query.Data, &queryData))
	require.Equal(t, createData.OrderNo, queryData.OrderNo)
	require.Equal(t, "pending_submit", queryData.StatusCode)
	require.Equal(t, goodsDetail.GoodsCode, queryData.GoodsID)
	require.Equal(t, "13800138000", queryData.Account)
	require.Equal(t, 1, queryData.Quantity)
}

func TestOpenOrderDuplicateRequestsCreateDifferentOrders(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	_, _, leafBrandID := h.createBrandPath(t, token, "重复开放订单品牌", "视频会员", "腾讯视频周卡")
	platformSubjectID := h.createSubject(t, token, "重复开放订单主体", 0)
	goodsID := h.createDirectRechargeChannelProductGoods(t, token, leafBrandID, "重复开放订单商品", "9.0000")
	platformID := h.createKakayunPlatformAccount(t, token, "重复云发卡账号", platformSubjectID, 0, "duplicate-open-order")
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478510",
		"supplier_goods_name": "云发卡测试直充商品",
		"source_cost_price":   "5.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	body := map[string]any{"token": "test-open-order-token", "goods_id": goodsDetail.GoodsCode, "account": "13800138000", "quantity": 1}
	first := h.postJSON("/api/open/orders", body, "")
	second := h.postJSON("/api/open/orders", body, "")
	require.Equal(t, 0, first.Code)
	require.Equal(t, 0, second.Code)
	var firstData, secondData struct {
		OrderNo string `json:"order_no"`
	}
	require.NoError(t, json.Unmarshal(first.Data, &firstData))
	require.NoError(t, json.Unmarshal(second.Data, &secondData))
	require.NotEqual(t, firstData.OrderNo, secondData.OrderNo)
}

func (h *testHarness) createDirectRechargeChannelProductGoods(t *testing.T, token string, brandID int64, name, defaultSellPrice string) int64 {
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
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) createKakayunPlatformAccount(t *testing.T, token, name string, subjectID int64, hasTax int, tokenID string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           tokenID + ".test.local",
		"backup_domain":    tokenID + ".backup.local",
		"type_id":          6,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}
