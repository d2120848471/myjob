package contract_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenSupplierProductChangeCallbackReturnsPlainOK(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "推送商品", "视频充值", "会员月卡")
	subjectID := h.createSubject(t, token, "推送主体", 1)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "推送商品A", "18.8000")
	platformID := h.createKakayunSupplierPlatformAccount(t, token, "卡卡云推送账号", subjectID, 0, "merchant-push", "secret-key")
	createBinding := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2582531",
		"supplier_goods_name": "推送前商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token)
	require.Equal(t, 0, createBinding.Code)

	enableProductGoodsChannelCostSync(t, h, goodsID)

	timestamp := time.Now().Unix()
	body := map[string]any{
		"goodsid":     "2582531",
		"goodsprice":  "12.0000",
		"goodsstock":  985,
		"goodsstatus": 1,
		"goodstype":   1,
		"goodsname":   "推送后商品",
		"update_time": timestamp,
		"timestamp":   timestamp,
	}
	body["sign"] = kakayunContractSign(body, "secret-key")

	res := h.rawRequest(http.MethodPost, "/api/open/supplier-platforms/kakayun/"+int64ToString(platformID)+"/product-change-callback", body, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Equal(t, "ok", res.body)
	require.NotContains(t, res.body, `"code"`)

	var binding struct {
		SourceCostPrice string `db:"source_cost_price"`
		CostPrice       string `db:"cost_price"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &binding, `
SELECT source_cost_price, cost_price
FROM product_goods_channel_binding
WHERE goods_id = ? AND platform_account_id = ? AND supplier_goods_no = ?
`, goodsID, platformID, "2582531")
	require.NoError(t, err)
	require.Equal(t, "12.0000", binding.SourceCostPrice)
	require.Equal(t, "12.0000", binding.CostPrice)

	count, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM product_goods_channel_price_change_log
WHERE goods_id = ? AND platform_account_id = ? AND supplier_goods_no = ? AND source = 'push'
`, goodsID, platformID, "2582531")
	require.NoError(t, err)
	require.Equal(t, 1, count.Int())
}

func TestSupplierProductSubscriptionListCancelAndResubscribe(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	subscriptionID := h.seedSupplierProductSubscription(t, "kakayun", "2582531", "subscribed")

	list := h.getJSON("/api/admin/supplier-product-subscriptions?page=1&page_size=20&supplier_goods_no=2582531", token)
	require.Equal(t, 0, list.Code)
	require.Contains(t, string(list.Data), "2582531")
	require.Contains(t, string(list.Data), "subscribed")

	cancel := h.postJSON("/api/admin/supplier-product-subscriptions/"+int64ToString(subscriptionID)+"/cancel", map[string]any{}, token)
	require.Equal(t, 0, cancel.Code)

	afterCancel := h.getJSON("/api/admin/supplier-product-subscriptions?page=1&page_size=20&status=canceled", token)
	require.Equal(t, 0, afterCancel.Code)
	require.Contains(t, string(afterCancel.Data), "canceled")
	retainedTimes, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM supplier_product_subscription
WHERE id = ? AND subscribed_at IS NOT NULL AND canceled_at IS NOT NULL
`, subscriptionID)
	require.NoError(t, err)
	require.Equal(t, 1, retainedTimes.Int())

	resubscribe := h.postJSON("/api/admin/supplier-product-subscriptions/"+int64ToString(subscriptionID)+"/resubscribe", map[string]any{}, token)
	require.Equal(t, 0, resubscribe.Code)
}

func (h *testHarness) createKakayunSupplierPlatformAccount(t *testing.T, token, name string, subjectID int64, hasTax int, tokenID, secretKey string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           "qqlogin.yxp8.cn",
		"backup_domain":    "qqlogin-backup.yxp8.cn",
		"type_id":          6,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       secretKey,
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
	}, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}

func (h *testHarness) seedSupplierProductSubscription(t *testing.T, providerCode, supplierGoodsNo, status string) int64 {
	t.Helper()
	now := h.app.Core().Now()
	result, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO supplier_product_subscription (
    provider_code, platform_account_id, platform_account_name, goods_id, binding_id,
    supplier_goods_no, supplier_goods_name, callback_url, status, last_action, last_error,
    request_snapshot, response_snapshot, subscribed_at, canceled_at, created_at, updated_at
) VALUES (?, 1, '卡卡云测试账号', 0, 0, ?, '测试订阅商品', 'https://public.example.com/api/open/supplier-platforms/kakayun/1/product-change-callback', ?, 'subscribe', '', '{}', '{}', ?, NULL, ?, ?)
`, providerCode, supplierGoodsNo, status, now, now, now)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}

func enableProductGoodsChannelCostSync(t *testing.T, h *testHarness, goodsID int64) {
	t.Helper()
	now := h.app.Core().Now()
	_, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO product_goods_channel_config (
    goods_id, smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy,
    sync_cost_price_enabled, sync_goods_name_enabled, allow_loss_sale_enabled, max_loss_amount, combo_goods_enabled,
    created_at, updated_at
) VALUES (?, 0, 0, 0, 'fixed_order', 1, 0, 0, '0.0000', 0, ?, ?)
ON DUPLICATE KEY UPDATE
    sync_cost_price_enabled = VALUES(sync_cost_price_enabled),
    sync_goods_name_enabled = VALUES(sync_goods_name_enabled),
    updated_at = VALUES(updated_at)
`, goodsID, now, now)
	require.NoError(t, err)
}

func kakayunContractSign(payload map[string]any, secretKey string) string {
	params := make(map[string]string, len(payload))
	for key, value := range payload {
		if key == "sign" || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" || text == "<nil>" {
			continue
		}
		params[key] = text
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + strings.TrimSpace(secretKey)))
	return hex.EncodeToString(sum[:])
}
