package tradelogic

import (
	"bytes"
	"context"
	"encoding/json"
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

type testOrderProvider struct {
	supportsNativeQuantity bool
}

func (p testOrderProvider) Code() string { return "test" }
func (p testOrderProvider) Name() string { return "测试Provider" }
func (p testOrderProvider) CandidateBaseURLs(account supplierprovider.AccountConfig) []string {
	return []string{strings.TrimRight(account.Domain, "/")}
}
func (p testOrderProvider) SupportsNativeQuantity() bool { return p.supportsNativeQuantity }
func (p testOrderProvider) BuildCreateOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.CreateOrderInput, baseURL string) (*http.Request, error) {
	payload := map[string]any{
		"provider_request_order_no": input.ProviderRequestOrderNo,
		"supplier_goods_no":         input.SupplierGoodsNo,
		"quantity":                  input.Quantity,
		"payload":                   input.Payload,
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/create", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(raw)), nil }
	return req, nil
}
func (p testOrderProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*supplierprovider.CreateOrderResult, error) {
	if statusCode != http.StatusOK {
		return &supplierprovider.CreateOrderResult{
			FinalFailed:   true,
			ErrorCategory: "server_error",
			ErrorMessage:  "http status not ok",
			RawPayload:    string(body),
		}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return &supplierprovider.CreateOrderResult{
		Accepted:       payload["accepted"].(bool),
		ChannelOrderNo: strings.TrimSpace(payload["channel_order_no"].(string)),
		UpstreamStatus: strings.TrimSpace(payload["status"].(string)),
		RawPayload:     string(body),
	}, nil
}
func (p testOrderProvider) BuildQueryOrderRequest(ctx context.Context, account supplierprovider.AccountConfig, input supplierprovider.QueryOrderInput, baseURL string) (*http.Request, error) {
	return nil, nil
}
func (p testOrderProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*supplierprovider.QueryOrderResult, error) {
	return nil, nil
}

func TestTradeOrderLogic_CreateOrder_IdempotentAndAttemptWritten(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/create" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		_ = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accepted":true,"channel_order_no":"CH001","status":"accepted"}`))
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

	goodsCode := "P-ORDER-001"
	goodsResult, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, product_template_id, default_sell_price, min_purchase_qty, max_purchase_qty, status, created_at, updated_at)
VALUES (?, 1, '交易商品A', 'card_secret', 'channel', 1, ?, ?, '29.9000', 1, 5, 1, ?, ?)
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

	provider := testOrderProvider{supportsNativeQuantity: true}
	lookup := func(code string) (supplierprovider.OrderProvider, bool) {
		if strings.TrimSpace(code) == "test" {
			return provider, true
		}
		return nil, false
	}
	logic := NewTradeOrderLogic(core, lookup, upstream.Client())

	input := CreateTradeOrderInput{
		CallerID:      100,
		ClientOrderNo: "C001",
		GoodsCode:     goodsCode,
		Quantity:      1,
		PayloadJSON:   `{"mobile":"13800138000"}`,
		RequestIP:     "127.0.0.1",
		RequestedAt:   time.Time{},
	}

	first, err := logic.CreateOrder(ctx, input)
	require.NoError(t, err)
	require.NotEmpty(t, first.OrderNo)
	require.Equal(t, "processing", first.Status)

	second, err := logic.CreateOrder(ctx, input)
	require.NoError(t, err)
	require.Equal(t, first.OrderNo, second.OrderNo)

	orderCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order WHERE caller_id = ? AND client_order_no = ?`, input.CallerID, input.ClientOrderNo)
	require.NoError(t, err)
	require.Equal(t, 1, orderCount.Int())

	attemptCount, err := core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM trade_order_attempt WHERE order_id = ?`, first.ID)
	require.NoError(t, err)
	require.Equal(t, 1, attemptCount.Int())
}
