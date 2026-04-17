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

type failCreateProvider struct{}

func (failCreateProvider) Code() string { return "p1" }
func (failCreateProvider) Name() string { return "失败Provider" }
func (failCreateProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (failCreateProvider) SupportsNativeQuantity() bool { return true }
func (failCreateProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create1", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (failCreateProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	return &supplierprovider.CreateOrderResult{
		FinalFailed:   true,
		ErrorCategory: "stock_not_enough",
		ErrorMessage:  "no stock",
		RawPayload:    string(body),
	}, nil
}
func (failCreateProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	return nil, nil
}
func (failCreateProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return nil, nil
}

type successCreateProvider struct{}

func (successCreateProvider) Code() string { return "p2" }
func (successCreateProvider) Name() string { return "成功Provider" }
func (successCreateProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (successCreateProvider) SupportsNativeQuantity() bool { return true }
func (successCreateProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	raw := []byte(`{}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create2", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (successCreateProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	return &supplierprovider.CreateOrderResult{
		FinalSuccess:  true,
		ChannelOrderNo: "CH-R2",
		UpstreamStatus: "success",
		RawPayload:     string(body),
	}, nil
}
func (successCreateProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	return nil, nil
}
func (successCreateProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return nil, nil
}

func TestTradeOrderLogic_Replenish_OnFinalFailedSwitchesToNextBinding(t *testing.T) {
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

	goodsCode := "P-REPLENISH-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品R', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
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

	binding2Result, err := core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price, dock_status, sort, is_auto_change, add_type, default_price, created_at, updated_at)
VALUES (?, ?, 'G002', '上游商品2', '11.0000', '11.0000', 'enabled', 20, 0, 'fixed', '0.0000', ?, ?)
`, goodsID, account2ID, now, now)
	require.NoError(t, err)
	binding2ID, _ := binding2Result.LastInsertId()

	lookup := func(code string) (supplierprovider.OrderProvider, bool) {
		switch strings.TrimSpace(code) {
		case "p1":
			return failCreateProvider{}, true
		case "p2":
			return successCreateProvider{}, true
		default:
			return nil, false
		}
	}
	logic := NewTradeOrderLogic(core, lookup, upstream.Client())

	order, err := logic.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C-R1",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
		RequestedAt:   time.Time{},
	})
	require.NoError(t, err)

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, attemptCount.Int())

	second, err := core.DB().GetCore().GetOne(ctx, `
SELECT attempt_no, binding_id, provider_code, attempt_status
FROM trade_order_attempt
WHERE order_id = ? AND attempt_no = 2
LIMIT 1
`, order.ID)
	require.NoError(t, err)
	require.Equal(t, 2, second["attempt_no"].Int())
	require.Equal(t, binding2ID, second["binding_id"].Int64())
	require.Equal(t, "p2", strings.TrimSpace(second["provider_code"].String()))
	require.NotEmpty(t, strings.TrimSpace(second["attempt_status"].String()))

	orderRow, err := core.DB().GetCore().GetOne(ctx, `SELECT status FROM trade_order WHERE id = ?`, order.ID)
	require.NoError(t, err)
	require.Equal(t, "success", strings.TrimSpace(orderRow["status"].String()))
}
