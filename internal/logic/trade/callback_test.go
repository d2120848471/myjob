package tradelogic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"myjob/internal/app"

	"github.com/stretchr/testify/require"
)

func TestTradeOrderLogic_HandleProviderOrderCallback_MarksOrderSuccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/buygoods" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH-CB-001"}}`))
	}))
	defer upstream.Close()

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

	templateResult, err := core.DB().Exec(ctx, `
INSERT INTO product_template (title, template_type, is_shared, account_name, validate_type, created_at, updated_at)
VALUES ('手机号模板', 'local', 0, '手机号', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	templateID, _ := templateResult.LastInsertId()

	goodsCode := "P-CALLBACK-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品CB', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '100', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	logic := NewTradeOrderLogic(core, nil, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C-CB-001",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
		RequestedAt:   time.Time{},
	})
	require.NoError(t, err)
	require.NotZero(t, order.ID)

	record, err := core.DB().GetCore().GetOne(ctx, `SELECT provider_request_order_no, channel_order_no FROM trade_order_attempt WHERE order_id = ? LIMIT 1`, order.ID)
	require.NoError(t, err)
	providerRequestOrderNo := record["provider_request_order_no"].String()
	channelOrderNo := record["channel_order_no"].String()
	require.NotEmpty(t, providerRequestOrderNo)
	require.NotEmpty(t, channelOrderNo)

	callbackBody, _ := json.Marshal(map[string]any{
		"orderno":     providerRequestOrderNo,
		"outorderno":  channelOrderNo,
		"userid":      "token-a",
		"status":      3,
		"money":       "10.0000",
		"extra_field": "ignored",
	})

	ackBody, contentType, err := logic.HandleProviderOrderCallback(ctx, "youkayun", http.Header{}, callbackBody)
	require.NoError(t, err)
	require.NotEmpty(t, ackBody)
	require.NotEmpty(t, contentType)

	updated, err := core.DB().GetCore().GetOne(ctx, `SELECT status, success_quantity FROM trade_order WHERE id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, "success", updated["status"].String())
	require.Equal(t, 1, updated["success_quantity"].Int())
}

func TestTradeOrderLogic_HandleProviderOrderCallback_FinalFailedTriggersReplenish(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/buygoods" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH-CB-002"}}`))
	}))
	defer upstream.Close()

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

	templateResult, err := core.DB().Exec(ctx, `
INSERT INTO product_template (title, template_type, is_shared, account_name, validate_type, created_at, updated_at)
VALUES ('手机号模板', 'local', 0, '手机号', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	templateID, _ := templateResult.LastInsertId()

	goodsCode := "P-CALLBACK-REPL-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品CBR', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	// 开启智能补单。
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, smart_replenish_enabled, route_mode, created_at, updated_at)
VALUES (?, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	account1Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account1ID, _ := account1Result.LastInsertId()

	account2Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号B', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-b', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account2ID, _ := account2Result.LastInsertId()

	binding1Result, err := core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '100', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account1ID, now, now)
	require.NoError(t, err)
	binding1ID, _ := binding1Result.LastInsertId()

	binding2Result, err := core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '200', '上游商品B', '11.0000', '11.0000', 'enabled', 20, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account2ID, now, now)
	require.NoError(t, err)
	binding2ID, _ := binding2Result.LastInsertId()

	logic := NewTradeOrderLogic(core, nil, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      101,
		ClientOrderNo: "C-CBR-001",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
	})
	require.NoError(t, err)

	first, err := core.DB().GetCore().GetOne(ctx, `
SELECT id, provider_request_order_no, channel_order_no, binding_id
FROM trade_order_attempt
WHERE order_id = ? AND attempt_no = 1
LIMIT 1
`, order.ID)
	require.NoError(t, err)
	require.Equal(t, binding1ID, first["binding_id"].Int64())
	providerRequestOrderNo := strings.TrimSpace(first["provider_request_order_no"].String())
	channelOrderNo := strings.TrimSpace(first["channel_order_no"].String())
	require.NotEmpty(t, providerRequestOrderNo)
	require.NotEmpty(t, channelOrderNo)

	callbackBody, _ := json.Marshal(map[string]any{
		"orderno":    providerRequestOrderNo,
		"outorderno": channelOrderNo,
		"userid":     "token-a",
		"status":     5,
	})
	_, _, err = logic.HandleProviderOrderCallback(ctx, "youkayun", http.Header{}, callbackBody)
	require.NoError(t, err)

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, attemptCount.Int())

	second, err := core.DB().GetCore().GetOne(ctx, `
SELECT attempt_no, binding_id, attempt_status
FROM trade_order_attempt
WHERE order_id = ? AND attempt_no = 2
LIMIT 1
`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, second["attempt_no"].Int())
	require.Equal(t, binding2ID, second["binding_id"].Int64())
	require.NotEmpty(t, strings.TrimSpace(second["attempt_status"].String()))
}

func TestTradeOrderLogic_HandleProviderOrderCallback_AttemptNotFoundUpdatesLogProcessResult(t *testing.T) {
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
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, 'http://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)

	logic := NewTradeOrderLogic(core, nil, nil)

	callbackBody, _ := json.Marshal(map[string]any{
		"orderno":    "PR-NOT-FOUND-001",
		"outorderno": "CH-NOT-FOUND-001",
		"userid":     "token-a",
		"status":     3,
	})

	_, _, err = logic.HandleProviderOrderCallback(ctx, "youkayun", http.Header{"Content-Type": []string{"application/json"}}, callbackBody)
	require.NoError(t, err)

	record, err := core.DB().GetCore().GetOne(ctx, `
SELECT process_result
FROM provider_callback_log
WHERE provider_code = 'youkayun' AND idempotency_key = 'PR-NOT-FOUND-001'
LIMIT 1
`)
	require.NoError(t, err)
	require.Equal(t, "attempt_not_found", strings.TrimSpace(record["process_result"].String()))
}

func TestTradeOrderLogic_HandleProviderOrderCallback_LateSuccessIgnoredWhenFulfillmentAlreadySucceeded(t *testing.T) {
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

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, 'http://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	orderResult, err := core.DB().Exec(ctx, `
INSERT INTO trade_order (order_no, caller_id, client_order_no, goods_id, quantity, success_quantity, failed_quantity, status, created_at, updated_at)
VALUES ('TO-LATE-001', 100, 'C-LATE-001', 1, 1, 1, 0, 'success', ?, ?)
`, now, now)
	require.NoError(t, err)
	orderID, _ := orderResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO trade_order_attempt (order_id, binding_id, platform_account_id, provider_code, fulfillment_no, attempt_quantity, attempt_no, provider_request_order_no, attempt_status, created_at, updated_at)
VALUES (?, 1, ?, 'youkayun', 'F001', 1, 1, 'PR-LATE-001', 'timeout', ?, ?)
`, orderID, accountID, now, now)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
INSERT INTO trade_order_attempt (order_id, binding_id, platform_account_id, provider_code, fulfillment_no, attempt_quantity, attempt_no, provider_request_order_no, channel_order_no, attempt_status, created_at, updated_at)
VALUES (?, 1, ?, 'youkayun', 'F001', 1, 2, 'PR-LATE-002', 'CH-NEW', 'success', ?, ?)
`, orderID, accountID, now, now)
	require.NoError(t, err)

	logic := NewTradeOrderLogic(core, nil, nil)

	callbackBody, _ := json.Marshal(map[string]any{
		"orderno":    "PR-LATE-001",
		"outorderno": "CH-OLD",
		"userid":     "token-a",
		"status":     3,
	})

	_, _, err = logic.HandleProviderOrderCallback(ctx, "youkayun", http.Header{"Content-Type": []string{"application/json"}}, callbackBody)
	require.NoError(t, err)

	logRecord, err := core.DB().GetCore().GetOne(ctx, `
SELECT process_result
FROM provider_callback_log
WHERE provider_code = 'youkayun' AND idempotency_key = 'PR-LATE-001'
LIMIT 1
`)
	require.NoError(t, err)
	require.Equal(t, "late_success_ignored", strings.TrimSpace(logRecord["process_result"].String()))

	attemptRecord, err := core.DB().GetCore().GetOne(ctx, `
SELECT attempt_status, callback_payload
FROM trade_order_attempt
WHERE provider_request_order_no = 'PR-LATE-001'
LIMIT 1
`)
	require.NoError(t, err)
	require.Equal(t, "timeout", strings.TrimSpace(attemptRecord["attempt_status"].String()))
	require.NotEmpty(t, strings.TrimSpace(attemptRecord["callback_payload"].String()))
}

func TestTradeOrderLogic_HandleProviderOrderCallback_OrderFinalDoesNotReplenish(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH-NEW"}}`))
	}))
	defer upstream.Close()

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

	goodsCode := "P-CALLBACK-FINAL-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '终态回调商品', 'card_secret', 'channel', 1, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, smart_replenish_enabled, route_mode, created_at, updated_at)
VALUES (?, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	account1Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account1ID, _ := account1Result.LastInsertId()

	account2Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号B', 'youkayun', '优卡云', 7, ?, 0, ?, 'token-b', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account2ID, _ := account2Result.LastInsertId()

	binding1Result, err := core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '100', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account1ID, now, now)
	require.NoError(t, err)
	binding1ID, _ := binding1Result.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, '200', '上游商品B', '11.0000', '11.0000', 'enabled', 20, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account2ID, now, now)
	require.NoError(t, err)

	orderResult, err := core.DB().Exec(ctx, `
INSERT INTO trade_order (
    order_no, caller_id, client_order_no,
    goods_id, goods_code_snapshot, goods_name_snapshot,
    binding_id, platform_account_id, route_mode_snapshot,
    quantity, payload_json,
    sale_price, total_amount,
    source_cost_price_snapshot, cost_price_snapshot,
    loss_order, loss_amount,
    status, created_at, updated_at
) VALUES (
    'TO-FINAL-001', 100, 'C-FINAL-001',
    ?, ?, '终态回调商品',
    ?, ?, 'fixed_order',
    1, '{"mobile":"13800138000"}',
    '29.9000', '29.9000',
    '10.0000', '10.0000',
    0, '0.0000',
    'failed', ?, ?
)
`, goodsID, goodsCode, binding1ID, account1ID, now, now)
	require.NoError(t, err)
	orderID, _ := orderResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO trade_order_attempt (
    order_id, binding_id, platform_account_id, provider_code,
    fulfillment_no, attempt_quantity, attempt_no, provider_request_order_no,
    channel_order_no, attempt_status,
    created_at, updated_at
) VALUES (?, ?, ?, 'youkayun', 'F001', 1, 1, 'PR-FINAL-001', 'CH-OLD', 'accepted', ?, ?)
`, orderID, binding1ID, account1ID, now, now)
	require.NoError(t, err)

	logic := NewTradeOrderLogic(core, nil, upstream.Client())

	callbackBody, _ := json.Marshal(map[string]any{
		"orderno":    "PR-FINAL-001",
		"outorderno": "CH-OLD",
		"userid":     "token-a",
		"status":     5,
	})

	_, _, err = logic.HandleProviderOrderCallback(ctx, "youkayun", http.Header{"Content-Type": []string{"application/json"}}, callbackBody)
	require.NoError(t, err)

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, orderID)
	require.NoError(t, err)
	require.Equal(t, 1, attemptCount.Int())

	logRecord, err := core.DB().GetCore().GetOne(ctx, `
SELECT process_result
FROM provider_callback_log
WHERE provider_code = 'youkayun' AND idempotency_key = 'PR-FINAL-001'
LIMIT 1
`)
	require.NoError(t, err)
	require.Equal(t, "late_failed_ignored", strings.TrimSpace(logRecord["process_result"].String()))
}
