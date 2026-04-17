package tradelogic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"

	"myjob/internal/app"

	"github.com/stretchr/testify/require"
)

func xingquanyiSign(secret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	builder.WriteString(secret)
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(params[key])
	}
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func TestTradeOrderLogic_HandleProviderPriceNotify_UpdatesBindingCost(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES ('P-PRICE-001', 1, '交易商品P', 'card_secret', 'channel', 1, ?, '29.9000', 1, 5, 1, ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('星权益账号', 'xingquanyi', '星权益', 35, ?, 1, 'xqy.example.com', '6', 'secretXYZ', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, sync_cost_enabled, route_mode, created_at, updated_at)
VALUES (?, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '2', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	params := map[string]string{
		"customer_id":  "6",
		"timestamp":    strconv.FormatInt(now.Unix(), 10),
		"event_type":   "price_changed",
		"event_data":   `{"price":"11.000"}`,
		"product_id":   "2",
		"product_name": "爱奇艺黄金会员年卡",
		"product_type": "1",
	}
	payload := map[string]any{
		"customer_id":  6,
		"timestamp":    now.Unix(),
		"event_type":   params["event_type"],
		"event_data":   params["event_data"],
		"product_id":   2,
		"product_name": params["product_name"],
		"product_type": 1,
	}
	payload["sign"] = xingquanyiSign("secretXYZ", params)
	body, _ := json.Marshal(payload)

	logic := NewTradeOrderLogic(core, nil, nil)
	_, _, err = logic.HandleProviderPriceNotify(ctx, "xingquanyi", http.Header{"Content-Type": []string{"application/json"}}, body)
	require.NoError(t, err)

	binding, err := core.DB().GetCore().GetOne(ctx, `
SELECT source_cost_price, cost_price
FROM product_goods_channel_binding
WHERE goods_id = ? AND platform_account_id = ? AND supplier_goods_no = '2' AND is_deleted = 0
LIMIT 1
`, goodsID, accountID)
	require.NoError(t, err)
	require.Equal(t, "11.0000", strings.TrimSpace(binding["source_cost_price"].String()))
	require.Equal(t, "11.0000", strings.TrimSpace(binding["cost_price"].String()))

	logCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM provider_price_notify_log WHERE provider_code = 'xingquanyi'`)
	require.NoError(t, err)
	require.Equal(t, 1, logCount.Int())
}

func TestTradeOrderLogic_HandleProviderPriceNotify_SyncGoodsNameWhenPrimaryBinding(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES ('P-PRICE-002', 1, '旧商品名', 'card_secret', 'channel', 1, ?, '29.9000', 1, 5, 1, ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('星权益账号', 'xingquanyi', '星权益', 35, ?, 1, 'xqy.example.com', '6', 'secretXYZ', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, sync_goods_name_enabled, route_mode, created_at, updated_at)
VALUES (?, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '2', '旧上游名', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	params := map[string]string{
		"customer_id":  "6",
		"timestamp":    strconv.FormatInt(now.Unix(), 10),
		"event_type":   "price_changed",
		"event_data":   `{"price":"11.000"}`,
		"product_id":   "2",
		"product_name": "新商品名",
		"product_type": "1",
	}
	payload := map[string]any{
		"customer_id":  6,
		"timestamp":    now.Unix(),
		"event_type":   params["event_type"],
		"event_data":   params["event_data"],
		"product_id":   2,
		"product_name": params["product_name"],
		"product_type": 1,
	}
	payload["sign"] = xingquanyiSign("secretXYZ", params)
	body, _ := json.Marshal(payload)

	logic := NewTradeOrderLogic(core, nil, nil)
	_, _, err = logic.HandleProviderPriceNotify(ctx, "xingquanyi", http.Header{"Content-Type": []string{"application/json"}}, body)
	require.NoError(t, err)

	updated, err := core.DB().GetCore().GetOne(ctx, `SELECT name FROM product_goods WHERE id = ?`, goodsID)
	require.NoError(t, err)
	require.Equal(t, "新商品名", strings.TrimSpace(updated["name"].String()))
}

func TestTradeOrderLogic_HandleProviderPriceNotify_BindingNotFoundUpdatesLogProcessResult(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	defer core.Close()

	ctx := context.Background()
	now := core.Now()

	subjectResult, err := core.DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('星权益账号', 'xingquanyi', '星权益', 35, ?, 1, 'xqy.example.com', '6', 'secretXYZ', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)

	params := map[string]string{
		"customer_id":  "6",
		"timestamp":    strconv.FormatInt(now.Unix(), 10),
		"event_type":   "price_changed",
		"event_data":   `{"price":"11.000"}`,
		"product_id":   "2",
		"product_name": "爱奇艺黄金会员年卡",
		"product_type": "1",
	}
	payload := map[string]any{
		"customer_id":  6,
		"timestamp":    now.Unix(),
		"event_type":   params["event_type"],
		"event_data":   params["event_data"],
		"product_id":   2,
		"product_name": params["product_name"],
		"product_type": 1,
	}
	payload["sign"] = xingquanyiSign("secretXYZ", params)
	body, _ := json.Marshal(payload)

	logic := NewTradeOrderLogic(core, nil, nil)
	_, _, err = logic.HandleProviderPriceNotify(ctx, "xingquanyi", http.Header{"Content-Type": []string{"application/json"}}, body)
	require.NoError(t, err)

	record, err := core.DB().GetCore().GetOne(ctx, `
SELECT process_result
FROM provider_price_notify_log
WHERE provider_code = 'xingquanyi' AND idempotency_key = ?
LIMIT 1
`, payload["sign"])
	require.NoError(t, err)
	require.Equal(t, "binding_not_found", strings.TrimSpace(record["process_result"].String()))
}
