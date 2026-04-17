package tradelogic

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"myjob/internal/app"
	"myjob/internal/library/supplierplatform/provider"

	"github.com/stretchr/testify/require"
)

type createAcceptedQuerySuccessProvider struct{}

func (createAcceptedQuerySuccessProvider) Code() string { return "test" }
func (createAcceptedQuerySuccessProvider) Name() string { return "测试Provider" }
func (createAcceptedQuerySuccessProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (createAcceptedQuerySuccessProvider) SupportsNativeQuantity() bool { return true }
func (createAcceptedQuerySuccessProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (createAcceptedQuerySuccessProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	return &supplierprovider.CreateOrderResult{
		Accepted:      true,
		ChannelOrderNo: "CH-Q1",
		UpstreamStatus: "accepted",
		RawPayload:     string(body),
	}, nil
}
func (createAcceptedQuerySuccessProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/query", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (createAcceptedQuerySuccessProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return &supplierprovider.QueryOrderResult{
		FinalSuccess:  true,
		ChannelOrderNo: "CH-Q1",
		UpstreamStatus: "success",
		RawPayload:     string(body),
	}, nil
}

func TestTradeOrderLogic_RunQueryJob_QueriesDueAttemptsAndMarksSuccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/create":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accepted":true}`))
		case "/query":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
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

	goodsCode := "P-QUERY-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品Q', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	accountResult, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'test', '测试平台', 6, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	accountID, _ := accountResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G001', '上游商品A', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, accountID, now, now)
	require.NoError(t, err)

	lookup := func(code string) (supplierprovider.OrderProvider, bool) {
		if strings.TrimSpace(code) == "test" {
			return createAcceptedQuerySuccessProvider{}, true
		}
		return nil, false
	}
	logic := NewTradeOrderLogic(core, lookup, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C-Q1",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotZero(t, order.ID)

	_, err = core.DB().Exec(ctx, `UPDATE trade_order_attempt SET next_query_at = ? WHERE order_id = ?`, now.Add(-time.Second), order.ID)
	require.NoError(t, err)

	processed, err := logic.RunQueryJob(ctx, "trace-q1")
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	record, err := core.DB().GetCore().GetOne(ctx, `SELECT status, success_quantity FROM trade_order WHERE id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, "success", record["status"].String())
	require.Equal(t, 1, record["success_quantity"].Int())
}

type createAcceptedQueryFailedProvider struct{}

func (createAcceptedQueryFailedProvider) Code() string { return "p1" }
func (createAcceptedQueryFailedProvider) Name() string { return "失败Query Provider" }
func (createAcceptedQueryFailedProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (createAcceptedQueryFailedProvider) SupportsNativeQuantity() bool { return true }
func (createAcceptedQueryFailedProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create1", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (createAcceptedQueryFailedProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	return &supplierprovider.CreateOrderResult{
		Accepted:      true,
		ChannelOrderNo: "CH-QF-1",
		UpstreamStatus: "accepted",
		RawPayload:     string(body),
	}, nil
}
func (createAcceptedQueryFailedProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/query1", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (createAcceptedQueryFailedProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return &supplierprovider.QueryOrderResult{
		FinalFailed:    true,
		ChannelOrderNo: "CH-QF-1",
		UpstreamStatus: "failed",
		ErrorCategory:  "stock_not_enough",
		ErrorMessage:   "no stock",
		RawPayload:     string(body),
	}, nil
}

type createFinalSuccessProvider struct{}

func (createFinalSuccessProvider) Code() string { return "p2" }
func (createFinalSuccessProvider) Name() string { return "成功Create Provider" }
func (createFinalSuccessProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (createFinalSuccessProvider) SupportsNativeQuantity() bool { return true }
func (createFinalSuccessProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create2", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (createFinalSuccessProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	return &supplierprovider.CreateOrderResult{
		FinalSuccess:  true,
		ChannelOrderNo: "CH-QF-2",
		UpstreamStatus: "success",
		RawPayload:     string(body),
	}, nil
}
func (createFinalSuccessProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	return nil, nil
}
func (createFinalSuccessProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return nil, nil
}

func TestTradeOrderLogic_RunQueryJob_FinalFailedTriggersReplenishAndSuccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
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

	goodsCode := "P-QUERY-FAIL-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品QF', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, smart_replenish_enabled, route_mode, created_at, updated_at)
VALUES (?, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	account1Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号1', 'p1', '失败平台', 6, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account1ID, _ := account1Result.LastInsertId()

	account2Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号2', 'p2', '成功平台', 6, ?, 0, ?, 'token-b', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account2ID, _ := account2Result.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G001', '上游商品1', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account1ID, now, now)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G002', '上游商品2', '11.0000', '11.0000', 'enabled', 20, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account2ID, now, now)
	require.NoError(t, err)

	lookup := func(code string) (supplierprovider.OrderProvider, bool) {
		switch strings.TrimSpace(code) {
		case "p1":
			return createAcceptedQueryFailedProvider{}, true
		case "p2":
			return createFinalSuccessProvider{}, true
		default:
			return nil, false
		}
	}
	logic := NewTradeOrderLogic(core, lookup, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C-QF-1",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotZero(t, order.ID)

	_, err = core.DB().Exec(ctx, `UPDATE trade_order_attempt SET next_query_at = ? WHERE order_id = ?`, now.Add(-time.Second), order.ID)
	require.NoError(t, err)

	processed, err := logic.RunQueryJob(ctx, "trace-qf")
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, attemptCount.Int())

	orderRow, err := core.DB().GetCore().GetOne(ctx, `SELECT status, success_quantity FROM trade_order WHERE id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, "success", strings.TrimSpace(orderRow["status"].String()))
	require.Equal(t, 1, orderRow["success_quantity"].Int())
}

func TestTradeOrderLogic_RunQueryJob_TimeoutTriggersReplenish(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
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

	goodsCode := "P-QUERY-TIMEOUT-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品QT', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
`, goodsCode, subjectID, templateID, now, now)
	require.NoError(t, err)
	goodsID, _ := goodsResult.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (goods_id, smart_replenish_enabled, attempt_timeout_enabled, attempt_timeout_minutes, route_mode, created_at, updated_at)
VALUES (?, 1, 1, 1, 'fixed_order', ?, ?)
`, goodsID, now, now)
	require.NoError(t, err)

	account1Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号1', 'p1', '失败平台', 6, ?, 0, ?, 'token-a', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account1ID, _ := account1Result.LastInsertId()

	account2Result, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号2', 'p2', '成功平台', 6, ?, 0, ?, 'token-b', 'secret', ?, ?)
`, subjectID, upstream.URL, now, now)
	require.NoError(t, err)
	account2ID, _ := account2Result.LastInsertId()

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G001', '上游商品1', '10.0000', '10.0000', 'enabled', 10, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account1ID, now, now)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G002', '上游商品2', '11.0000', '11.0000', 'enabled', 20, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account2ID, now, now)
	require.NoError(t, err)

	lookup := func(code string) (supplierprovider.OrderProvider, bool) {
		switch strings.TrimSpace(code) {
		case "p1":
			return createAcceptedQuerySuccessProvider{}, true
		case "p2":
			return createFinalSuccessProvider{}, true
		default:
			return nil, false
		}
	}
	logic := NewTradeOrderLogic(core, lookup, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C-QT-1",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotZero(t, order.ID)

	_, err = core.DB().Exec(ctx, `
UPDATE trade_order_attempt
SET attempt_status = 'unknown',
    next_query_at = ?,
    query_deadline_at = ?
WHERE order_id = ?
`, now.Add(-time.Second), now.Add(-time.Second), order.ID)
	require.NoError(t, err)

	processed, err := logic.RunQueryJob(ctx, "trace-qt")
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, attemptCount.Int())
}
