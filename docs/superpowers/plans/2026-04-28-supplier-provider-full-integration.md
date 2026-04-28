# Supplier Provider Full Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接入卡易信、卡速售、星权益、优卡云、聚浪云、星海、飞速源的下单、查单、商品同步和防亏损能力，并保留卡卡云订阅专属边界。

**Architecture:** 继续使用 `internal/library/supplierplatform/provider` 的 provider 适配器边界，业务层只消费标准化结果。订单层新增 segment 表和聚合逻辑，用一个本地 attempt 承载一个或多个真实上游请求。商品同步保持现有主动 worker，非卡卡云不进入订阅列表。

**Tech Stack:** Go 1.24、GoFrame v2、MySQL/SQLite schema、`shopspring/decimal`、`stretchr/testify`、现有 `channelpricing` 和 provider helper。

---

## File Structure

- Modify: `internal/library/supplierplatform/provider/types.go`
  - Add `SafetyPriceMode`, `OrderProviderCapabilities`, `SafePrice` and `ProductInfoListProvider`.
- Modify: `internal/library/supplierplatform/provider/common.go`
  - Add deterministic JSON and form helpers needed by multiple providers.
- Modify: `internal/library/supplierplatform/provider/providers.go`
  - Keep only existing provider structs, names and balance provider implementations; new order/product logic belongs in focused files.
- Create: `internal/library/supplierplatform/provider/order_providers.go`
  - Implement order create/query providers for kayixin, kasushou, xingquanyi, youkayun, julangyun, xinghai, feisuyuan.
- Create: `internal/library/supplierplatform/provider/product_info_providers.go`
  - Implement product detail/list parsers for all non-kakayun platforms.
- Create: `internal/library/supplierplatform/provider/product_push_providers.go`
  - Implement only push providers whose signatures can be verified with the current body-only interface: kasushou, xingquanyi, youkayun.
- Modify: `internal/library/supplierplatform/provider/registry.go`
  - Register new order, product info and selected product push providers; keep product subscription registry as kakayun only.
- Modify: `internal/library/supplierplatform/provider/providers_test.go`
  - Add provider request/parse tests.
- Modify: `internal/library/supplierplatform/provider/kakayun_product_push_test.go`
  - Update lookup tests to reflect new push providers and subscription boundary.
- Modify: `manifest/sql/008_external_order.sql`
  - Add `external_order_attempt_segment`.
- Modify: `internal/app/schema.go`
  - Add SQLite/MySQL schema for segment table.
- Modify: `internal/app/order_bootstrap.go`
  - Add `ensureExternalOrderAttemptSegmentSchema`.
- Modify: `internal/app/bootstrap.go`
  - Call the new schema ensure function.
- Modify: `internal/app/order_schema_test.go`
  - Assert segment table exists and bootstrap is idempotent.
- Modify: `internal/model/entity/order.go`
  - Add `ExternalOrderAttemptSegment`.
- Modify: `internal/logic/order/order_loss_guard.go`
  - Replace `kakayunMaxMoney` with generic segment safety-price calculation.
- Modify: `internal/logic/order/order_loss_guard_test.go`
  - Add generic safety-price tests.
- Modify: `internal/logic/order/order_channel.go`
  - Remove `a.provider_code = 'kakayun'` SQL filter; filter by registered provider in Go.
- Modify: `internal/logic/order/order_submit.go`
  - Submit segments and aggregate attempt result.
- Modify: `internal/logic/order/order_poll.go`
  - Poll segments and aggregate status.
- Create: `internal/logic/order/order_segment.go`
  - Segment creation, status aggregation, persistence and query helpers.
- Modify: `test/integration/order_worker_test.go`
  - Add non-kakayun order and feisuyuan split-order integration tests.
- Modify: `internal/logic/admin/product_goods_channel_sync.go`
  - Use list-based product info cache when provider supports it.
- Modify: `internal/logic/admin/product_goods_channel_subscription.go`
  - Keep auto subscribe and subscription mutations kakayun-only.
- Modify: `internal/logic/admin/product_goods_channel_subscription_test.go`
  - Add non-kakayun no-subscription assertion.
- Modify: `internal/logic/admin/product_goods_channel_sync_test.go`
  - Add non-kakayun product info sync test.
- Modify: `docs/development.md`, `docs/module-map.md`, `docs/testing.md`
  - Document multi-provider order, segment split, safety price and monitoring behavior.

---

### Task 1: Provider Capability Types And Subscription Boundary

**Files:**
- Modify: `internal/library/supplierplatform/provider/types.go`
- Create: `internal/library/supplierplatform/provider/order_providers.go`
- Modify: `internal/library/supplierplatform/provider/providers_test.go`
- Modify: `internal/library/supplierplatform/provider/kakayun_product_push_test.go`

- [ ] **Step 1: Write failing capability and subscription-boundary tests**

Add these tests to `internal/library/supplierplatform/provider/providers_test.go`:

```go
func TestOrderProviderCapabilities(t *testing.T) {
	tests := []struct {
		code        string
		provider    interface{ Capabilities() OrderProviderCapabilities }
		maxQty      int
		safetyMode  SafetyPriceMode
		safetyField string
	}{
		{code: "kakayun", provider: kakayunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "maxmoney"},
		{code: "kayixin", provider: kayixinProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "safePrice"},
		{code: "kasushou", provider: kasushouProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "safe_price"},
		{code: "xingquanyi", provider: xingquanyiProvider{}, maxQty: 0, safetyMode: SafetyPriceModeUnit, safetyField: "safe_cost"},
		{code: "youkayun", provider: youkayunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "maxmoney"},
		{code: "julangyun", provider: julangyunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "accessPrice"},
		{code: "xinghai", provider: xinghaiProvider{}, maxQty: 0, safetyMode: SafetyPriceModeUnit, safetyField: "itemPrice"},
		{code: "feisuyuan", provider: feisuyuanProvider{}, maxQty: 1, safetyMode: SafetyPriceModeUnsupported, safetyField: ""},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			capabilities := tc.provider.Capabilities()
			require.Equal(t, tc.maxQty, capabilities.MaxQuantityPerCreate)
			require.Equal(t, tc.safetyMode, capabilities.SafetyPrice.Mode)
			require.Equal(t, tc.safetyField, capabilities.SafetyPrice.FieldName)
		})
	}
}
```

Update `internal/library/supplierplatform/provider/kakayun_product_push_test.go`:

```go
func TestLookupProductSubscriptionOnlyRegistersKakayun(t *testing.T) {
	subscriptionProvider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", subscriptionProvider.Code())

	for _, code := range []string{"kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"} {
		t.Run(code, func(t *testing.T) {
			_, ok := LookupProductSubscription(code)
			require.False(t, ok)
		})
	}
}
```

- [ ] **Step 2: Run capability tests and verify failure**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run 'TestOrderProviderCapabilities|TestLookupProductSubscriptionOnlyRegistersKakayun' -count=1 -timeout 60s
```

Expected: fail because capability types and provider capability methods do not exist yet.

- [ ] **Step 3: Add provider capability types**

In `internal/library/supplierplatform/provider/types.go`, add:

```go
// SafetyPriceMode 表示上游防亏损字段的金额口径。
type SafetyPriceMode string

const (
	SafetyPriceModeUnsupported SafetyPriceMode = "unsupported"
	SafetyPriceModeTotal       SafetyPriceMode = "total"
	SafetyPriceModeUnit        SafetyPriceMode = "unit"
)

// SafetyPriceCapability 描述 provider 是否支持把本地防亏损阈值透传给上游。
type SafetyPriceCapability struct {
	Mode      SafetyPriceMode
	FieldName string
}

// OrderProviderCapabilities 描述订单 provider 对数量拆分和安全价字段的支持情况。
type OrderProviderCapabilities struct {
	MaxQuantityPerCreate int
	SafetyPrice          SafetyPriceCapability
}
```

Update `OrderProvider`:

```go
type OrderProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	Capabilities() OrderProviderCapabilities
	BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error)
	ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error)
	BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error)
	ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error)
}
```

Add `SafePrice` to `CreateOrderInput` while keeping `MaxMoney` for compatibility during the task:

```go
// CreateOrderInput 是上游下单接口所需的最小业务参数。
type CreateOrderInput struct {
	SupplierGoodsNo   string
	Quantity          int
	Account           string
	SupplierUSOrderNo string
	// MaxMoney 兼容历史卡卡云调用；新 provider 应优先读取 SafePrice。
	MaxMoney  string
	SafePrice string
}
```

- [ ] **Step 4: Add capability methods**

In `internal/library/supplierplatform/provider/order_providers.go`, add capability methods first:

```go
package supplierprovider

func (kakayunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "maxmoney"}}
}
func (kayixinProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "safePrice"}}
}
func (kasushouProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "safe_price"}}
}
func (xingquanyiProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnit, FieldName: "safe_cost"}}
}
func (youkayunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "maxmoney"}}
}
func (julangyunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "accessPrice"}}
}
func (xinghaiProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnit, FieldName: "itemPrice"}}
}
func (feisuyuanProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{MaxQuantityPerCreate: 1, SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnsupported}}
}
```

Keep `defaultProductSubscriptionRegistry` unchanged with only `kakayun`; non-kakayun order and product-info registry entries are added only after the concrete adapters compile in Tasks 2 and 3.

- [ ] **Step 5: Run capability tests**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run 'TestOrderProviderCapabilities|TestLookupProductSubscriptionOnlyRegistersKakayun' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add internal/library/supplierplatform/provider/types.go internal/library/supplierplatform/provider/order_providers.go internal/library/supplierplatform/provider/providers_test.go internal/library/supplierplatform/provider/kakayun_product_push_test.go
git commit -m "feat: declare supplier provider capabilities"
```

---

### Task 2: Order Provider Request And Response Adapters

**Files:**
- Modify: `internal/library/supplierplatform/provider/common.go`
- Modify: `internal/library/supplierplatform/provider/order_providers.go`
- Modify: `internal/library/supplierplatform/provider/registry.go`
- Modify: `internal/library/supplierplatform/provider/providers_test.go`

- [ ] **Step 1: Write failing order provider tests**

Add tests that assert each non-kakayun provider builds requests and parses statuses. Use these concrete fixtures in `providers_test.go`:

```go
func TestOrderProviderRegistriesIncludeAllConfiguredPlatforms(t *testing.T) {
	codes := []string{"kakayun", "kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupOrder(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
}

func TestMultiPlatformOrderProvidersBuildCreateRequests(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 123000000, time.UTC)
	account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: map[string]any{}}
	input := CreateOrderInput{SupplierGoodsNo: "10001", Quantity: 2, Account: "13800138000", SupplierUSOrderNo: "O20260428123045123456-T1-S1", SafePrice: "20.0000"}
	tests := []struct {
		code      string
		path      string
		assert    func(t *testing.T, req *http.Request, body []byte)
	}{
		{"kayixin", "/api/v3/order/create", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, float64(10001), payload["goodsId"])
			require.Equal(t, float64(2), payload["count"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["outerNumber"])
			require.Equal(t, float64(20), payload["safePrice"])
			require.Equal(t, []any{map[string]any{"name": "充值账号", "value": "13800138000"}}, payload["attach"])
		}},
		{"kasushou", "/api/v1/order/buy", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, float64(10001), payload["id"])
			require.Equal(t, float64(2), payload["quantity"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["external_orderno"])
			require.Equal(t, "20.0000", payload["safe_price"])
			require.Equal(t, map[string]any{"recharge_account": "13800138000"}, payload["attach"])
		}},
		{"xingquanyi", "/api/buy", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBody(t, body)
			require.Equal(t, "10001", payload["product_id"])
			require.Equal(t, "2", payload["quantity"])
			require.Equal(t, "13800138000", payload["recharge_account"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["outer_order_id"])
			require.Equal(t, "20.0000", payload["safe_cost"])
		}},
		{"youkayun", "/api/buygoods", func(t *testing.T, req *http.Request, body []byte) {
			fields := decodeMultipartFields(t, req, body)
			require.Equal(t, "10001", fields["goodsid"])
			require.Equal(t, "2", fields["quantity"])
			require.Equal(t, "13800138000", fields["accountname"])
			require.Equal(t, "O20260428123045123456-T1-S1", fields["outorderno"])
			require.Equal(t, "20.0000", fields["maxmoney"])
		}},
		{"julangyun", "/api/recharge/order/submit", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, "10001", payload["goodsCode"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["accessOrderNo"])
			require.Equal(t, "13800138000", payload["rechargeAccount"])
			require.Equal(t, float64(2), payload["orderNum"])
			require.Equal(t, float64(20), payload["accessPrice"])
		}},
		{"xinghai", "/api/order/submit", func(t *testing.T, req *http.Request, body []byte) {
			values, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			require.Equal(t, "10001", values.Get("itemId"))
			require.Equal(t, "2", values.Get("amount"))
			require.Equal(t, "13800138000", values.Get("uuid"))
			require.Equal(t, "O20260428123045123456-T1-S1", values.Get("outOrderId"))
			require.Equal(t, "20.0000", values.Get("itemPrice"))
		}},
		{"feisuyuan", "/recharge/order", func(t *testing.T, req *http.Request, body []byte) {
			values, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			require.Equal(t, "10001", values.Get("productId"))
			require.Equal(t, "1", values.Get("number"))
			require.Equal(t, "13800138000", values.Get("rechargeAccount"))
			require.Equal(t, "O20260428123045123456-T1-S1", values.Get("outTradeNo"))
			require.Empty(t, values.Get("maxmoney"))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupOrder(tc.code)
			require.True(t, ok)
			req, err := provider.BuildCreateOrderRequest(context.Background(), account, now, "http://platform.example.com", input)
			require.NoError(t, err)
			require.Equal(t, "http://platform.example.com"+tc.path, req.URL.String())
			tc.assert(t, req, readRequestBody(t, req))
		})
	}
}
```

Also add parse tests:

```go
func TestMultiPlatformOrderProvidersParseCreateAndQuery(t *testing.T) {
	tests := []struct {
		code          string
		createBody    string
		querySuccess  string
		queryFailed   string
		queryProgress string
	}{
		{"kayixin", `{"code":1000,"msg":"购买成功","data":{"orderNumber":"KYX001"}}`, `{"code":1000,"data":{"orderNumber":"KYX001","outerNumber":"OUT001","status":3,"result":"完成"}}`, `{"code":1000,"data":{"status":4,"result":"退款"}}`, `{"code":1000,"data":{"status":2,"result":"处理中"}}`},
		{"kasushou", `{"code":200,"msg":"下单成功","data":{"ordersn":"KSS001","external_orderno":"OUT001"}}`, `{"code":200,"data":[{"ordersn":"KSS001","external_orderno":"OUT001","status":3,"recharge_hints":"完成"}]}`, `{"code":200,"data":[{"status":4,"recharge_hints":"取消"}]}`, `{"code":200,"data":[{"status":2,"recharge_hints":"处理中"}]}`},
		{"xingquanyi", `{"code":"ok","message":"","data":{"order_id":"XQY001","state":101}}`, `{"code":"ok","data":{"id":"XQY001","outer_order_id":"OUT001","state":200,"recharge_info":"完成"}}`, `{"code":"ok","data":{"state":500,"recharge_info":"失败"}}`, `{"code":"ok","data":{"state":101,"recharge_info":"处理中"}}`},
		{"youkayun", `{"code":1000,"msg":"获取成功","data":{"ordersn":"YKY001","outorderno":"OUT001"}}`, `{"code":1000,"data":{"ordersn":"YKY001","outer_order_no":"OUT001","status":3}}`, `{"code":1000,"data":{"status":5}}`, `{"code":1000,"data":{"status":2}}`},
		{"julangyun", `{"code":200,"message":"处理成功","data":{"returnOrderNo":"JLY001","accessOrderNo":"OUT001","orderStatus":20}}`, `{"code":200,"data":{"returnOrderNo":"JLY001","accessOrderNo":"OUT001","orderStatus":30}}`, `{"code":200,"data":{"orderStatus":40}}`, `{"code":200,"data":{"orderStatus":20}}`},
		{"xinghai", `{"code":"00","msg":"下单成功","orderId":"XH001","outOrderId":"OUT001"}`, `{"code":"00","orderId":"XH001","outOrderId":"OUT001","orderStatus":"2","orderDesc":"成功"}`, `{"code":"00","orderStatus":"3","orderDesc":"失败"}`, `{"code":"00","orderStatus":"1","orderDesc":"处理中"}`},
		{"feisuyuan", `{"code":"2000","message":"ok"}`, `{"code":"0000","status":"01","message":"success","outTradeNo":"OUT001"}`, `{"code":"0000","status":"03","message":"fail","outTradeNo":"OUT001"}`, `{"code":"0000","status":"02","message":"pending","outTradeNo":"OUT001"}`},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupOrder(tc.code)
			require.True(t, ok)
			create, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(tc.createBody))
			require.NoError(t, err)
			require.True(t, create.Accepted)
			require.Equal(t, SupplierOrderStatusProcessing, create.Status)

			success, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.querySuccess))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusSuccess, success.Status)

			failed, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.queryFailed))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusFailed, failed.Status)

			processing, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.queryProgress))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusProcessing, processing.Status)
		})
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./internal/library/supplierplatform/provider -run 'TestOrderProviderRegistriesIncludeAllConfiguredPlatforms|TestMultiPlatformOrderProviders' -count=1 -timeout 60s
```

Expected: fail because non-kakayun order adapters are not registered and implemented.

- [ ] **Step 3: Add helper functions**

In `internal/library/supplierplatform/provider/common.go`, add:

```go
func orderedJSONBody(payload map[string]any) (string, error) {
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	builder.WriteString("{")
	for index, key := range keys {
		if index > 0 {
			builder.WriteString(",")
		}
		keyRaw, err := json.Marshal(key)
		if err != nil {
			return "", err
		}
		valueRaw, err := json.Marshal(payload[key])
		if err != nil {
			return "", err
		}
		builder.Write(keyRaw)
		builder.WriteString(":")
		builder.Write(valueRaw)
	}
	builder.WriteString("}")
	return builder.String(), nil
}

func newSortedJSONRequest(ctx context.Context, requestURL string, payload map[string]any, headers map[string]string) (*http.Request, error) {
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, requestURL, body, headers)
}

func safePriceValue(input CreateOrderInput) string {
	if strings.TrimSpace(input.SafePrice) != "" {
		return strings.TrimSpace(input.SafePrice)
	}
	return strings.TrimSpace(input.MaxMoney)
}
```

- [ ] **Step 4: Implement order provider methods**

In `internal/library/supplierplatform/provider/order_providers.go`, implement each provider using the documented paths and status maps:

```go
func (kayixinProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	payload := map[string]any{
		"goodsId":     mustIntString(input.SupplierGoodsNo),
		"count":       input.Quantity,
		"notifyUrl":   "",
		"outerNumber": strings.TrimSpace(input.SupplierUSOrderNo),
		"safePrice":   decimalStringOrZero(safePriceValue(input)),
		"sku":         "",
		"attach":      []map[string]string{{"name": "充值账号", "value": strings.TrimSpace(input.Account)}},
	}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	headers := kayixinHeaders(account, now, body)
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/order/create", body, headers)
}
```

Use equivalent concrete mappings:

- kayixin query: `/api/v3/order/getDetail`, JSON body `orderNumber` and `outerNumber`; success statuses `3` and `9`, processing `1`, `2`, `7`, `8`, failed `4`, refunded `5` as failed for this system.
- kasushou create: `/api/v1/order/buy`, JSON body `id`, `url`, `external_orderno`, `safe_price`, `mark`, `quantity`, `attach.recharge_account`; status success `3`, processing `1`, `2`, failed `4`, `5`, `-1`.
- xingquanyi create: `/api/buy`, JSON body with common sign params, `product_id`, `recharge_account`, `quantity`, `outer_order_id`, `notify_url`, `safe_cost`; status success `200`, processing `100`, `101`, unknown `501`, failed `500`.
- youkayun create: `/api/buygoods`, multipart `userid`, `goodsid`, `quantity`, `accountname`, `outorderno`, `callbackurl`, `maxmoney`, `sign`; status success `3`, processing `1`, `2`, failed `5`.
- julangyun create: `/api/recharge/order/submit`, JSON with headers `userCode`, `timestamp`, `sign`; status success `30`, processing `20`, failed `40`, `50`.
- xinghai create: `/api/order/submit`, form `appId`, `outOrderId`, `uuid`, `itemId`, `amount`, `itemPrice`, `callbackUrl`, `timestamp`, `sign`; create code `00` accepted, `-16` and `-99` unknown, documented hard failures failed.
- feisuyuan create: `/recharge/order`, form `merchantId`, `outTradeNo`, `productId`, `rechargeAccount`, `accountType=0`, `number=1`, `timeStamp`, `notifyUrl`, `version=1.0`, `sign`; code `2000` accepted, `1999` unknown, other non-success failed.

Add these helper functions in the same file:

```go
func mustIntString(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func decimalStringOrZero(value string) string {
	amount, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return "0"
	}
	return amount.Round(4).StringFixed(4)
}

func firstDataItem(payload map[string]any) map[string]any {
	data, ok := payload["data"].([]any)
	if !ok || len(data) == 0 {
		return map[string]any{}
	}
	row, _ := data[0].(map[string]any)
	if row == nil {
		return map[string]any{}
	}
	return row
}
```

In `internal/library/supplierplatform/provider/registry.go`, update the order registry after every listed provider has concrete create/query methods:

```go
var defaultOrderRegistry = map[string]OrderProvider{
	"kakayun":    kakayunProvider{},
	"kayixin":    kayixinProvider{},
	"kasushou":   kasushouProvider{},
	"xingquanyi": xingquanyiProvider{},
	"youkayun":   youkayunProvider{},
	"julangyun":  julangyunProvider{},
	"xinghai":    xinghaiProvider{},
	"feisuyuan":  feisuyuanProvider{},
}
```

- [ ] **Step 5: Run order provider tests**

```bash
go test ./internal/library/supplierplatform/provider -run 'TestOrderProviderRegistriesIncludeAllConfiguredPlatforms|TestMultiPlatformOrderProviders|TestKakayunOrderProvider|TestOrderProviderCapabilities' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add internal/library/supplierplatform/provider/common.go internal/library/supplierplatform/provider/order_providers.go internal/library/supplierplatform/provider/registry.go internal/library/supplierplatform/provider/providers_test.go
git commit -m "feat: add supplier order providers"
```

---

### Task 3: Product Info Providers And Active Monitoring

**Files:**
- Create or modify: `internal/library/supplierplatform/provider/product_info_providers.go`
- Modify: `internal/library/supplierplatform/provider/types.go`
- Modify: `internal/library/supplierplatform/provider/registry.go`
- Modify: `internal/library/supplierplatform/provider/providers_test.go`
- Modify: `internal/logic/admin/product_goods_channel_sync.go`
- Modify: `internal/logic/admin/product_goods_channel_sync_test.go`

- [ ] **Step 1: Write failing product info provider tests**

Add to `providers_test.go`:

```go
func TestProductInfoProviderRegistriesIncludeAllConfiguredPlatforms(t *testing.T) {
	codes := []string{"kakayun", "kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupProductInfo(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
}

func TestMultiPlatformProductInfoProvidersBuildAndParse(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 0, time.UTC)
	account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: map[string]any{}}
	tests := []struct {
		code string
		path string
		body string
	}{
		{"kayixin", "/api/v3/goods/getDetail", `{"code":1000,"msg":"success","data":{"goodsId":10001,"name":"卡易信会员","salesPrice":12.34,"status":1}}`},
		{"kasushou", "/api/v1/goods/info", `{"code":200,"msg":"成功","data":{"id":10001,"goods_name":"卡速售会员","goods_price":"23.4500","status":1}}`},
		{"xingquanyi", "/api/product", `{"code":"ok","message":"","data":{"id":10001,"product_name":"星权益会员","name":"月卡","price":"34.5600"}}`},
		{"youkayun", "/api/goodsdetails", `{"code":1000,"msg":"查询成功","data":{"id":10001,"goods_name":"优卡云会员","goods_price":"45.6700","status":1}}`},
		{"julangyun", "/api/recharge/goods/detail", `{"code":200,"message":"处理成功","data":{"goodsCode":"10001","goodsName":"聚浪云会员","goodsPrice":56.78,"goodsStatus":1}}`},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupProductInfo(tc.code)
			require.True(t, ok)
			req, err := provider.BuildProductInfoRequest(context.Background(), account, now, "http://platform.example.com", ProductInfoInput{SupplierGoodsNo: "10001"})
			require.NoError(t, err)
			require.Equal(t, "http://platform.example.com"+tc.path, req.URL.String())

			info, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(tc.body))
			require.NoError(t, err)
			require.Equal(t, "10001", info.SupplierGoodsNo)
			require.NotEmpty(t, info.GoodsName)
			require.True(t, info.GoodsPriceValid)
		})
	}
}

func TestListBasedProductInfoProvidersParseMatchedItem(t *testing.T) {
	tests := []struct {
		code string
		body string
		name string
		price string
	}{
		{"xinghai", `{"code":"00","msg":"查询成功","data":[{"itemId":"10001","itemName":"星海会员","price":"67.8900"},{"itemId":"10002","itemName":"其他","price":"1"}]}`, "星海会员", "67.8900"},
		{"feisuyuan", `{"code":"0000","products":[{"product_id":"10001","channel_price":"78.9000","item_name":"飞速源会员"},{"product_id":"10002","channel_price":"1","item_name":"其他"}]}`, "飞速源会员", "78.9000"},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupProductInfo(tc.code)
			require.True(t, ok)
			listProvider, ok := provider.(ProductInfoListProvider)
			require.True(t, ok)
			results, err := listProvider.ParseProductInfoListResponse(http.StatusOK, []byte(tc.body))
			require.NoError(t, err)
			require.Equal(t, tc.name, results["10001"].GoodsName)
			require.Equal(t, tc.price, results["10001"].GoodsPrice.StringFixed(4))
		})
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./internal/library/supplierplatform/provider -run 'TestProductInfoProviderRegistriesIncludeAllConfiguredPlatforms|TestMultiPlatformProductInfoProviders|TestListBasedProductInfoProviders' -count=1 -timeout 60s
```

Expected: fail because non-kakayun product info providers are not registered and implemented.

- [ ] **Step 3: Add list provider interface**

In `types.go`, add:

```go
// ProductInfoListProvider 表示只能或更适合通过商品列表同步商品信息的平台能力。
type ProductInfoListProvider interface {
	ProductInfoProvider
	BuildProductInfoListRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error)
	ParseProductInfoListResponse(statusCode int, body []byte) (map[string]ProductInfoResult, error)
}
```

- [ ] **Step 4: Implement product info providers**

In `product_info_providers.go`, implement detail providers and list providers:

```go
func (kayixinProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	body := map[string]any{"goodsId": mustIntString(input.SupplierGoodsNo)}
	raw, err := orderedJSONBody(body)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/goods/getDetail", raw, kayixinHeaders(account, now, raw))
}

func (kayixinProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return ProductInfoResult{Raw: string(body)}, err
	}
	if codeString(payload["code"]) != "1000" {
		return ProductInfoResult{Raw: string(body)}, errors.New(defaultOrderMessage(responseMessage(payload), "卡易信商品详情查询失败"))
	}
	data := nestedMap(payload, "data")
	return productInfoFromFields(string(body), data["goodsId"], data["name"], data["salesPrice"])
}
```

Use `productInfoFromFields`:

```go
func productInfoFromFields(raw string, idValue, nameValue, priceValue any) (ProductInfoResult, error) {
	name := strings.TrimSpace(codeString(nameValue))
	price, err := decimalFromValue(priceValue)
	priceValid := err == nil && !price.IsNegative()
	if !priceValid && name == "" {
		return ProductInfoResult{Raw: raw}, errors.New("供应商商品详情缺少可同步字段")
	}
	if priceValid {
		price = price.Round(4)
	}
	return ProductInfoResult{SupplierGoodsNo: codeString(idValue), GoodsName: name, GoodsPrice: price, GoodsPriceValid: priceValid, Raw: raw}, nil
}
```

Field mappings:

- kasushou: path `/api/v1/goods/info`, fields `id`, `goods_name`, `goods_price`, success code `200`.
- xingquanyi: path `/api/product`, fields `id`, combined name `product_name` plus optional `name`, price `price`, success `ok`.
- youkayun: path `/api/goodsdetails`, multipart fields `userid`, `goodsid`, `sign`, response `id`, `goods_name`, `goods_price`, success `1000`.
- julangyun: path `/api/recharge/goods/detail`, JSON/header signed request, response `goodsCode`, `goodsName`, `goodsPrice`, success `200`.
- xinghai list: path `/api/item/query`, form `appId`, `timestamp`, `sign`, parse `data[].itemId`, `itemName`, `price`, success `00`.
- feisuyuan list: path `/recharge/product`, form `merchantId`, `timeStamp`, `version`, `sign`, parse `products[].product_id`, `item_name`, `channel_price`, success `0000`.

In `internal/library/supplierplatform/provider/registry.go`, update the product info registry after every listed provider has concrete detail or list methods:

```go
var defaultProductInfoRegistry = map[string]ProductInfoProvider{
	"kakayun":    kakayunProvider{},
	"kayixin":    kayixinProvider{},
	"kasushou":   kasushouProvider{},
	"xingquanyi": xingquanyiProvider{},
	"youkayun":   youkayunProvider{},
	"julangyun":  julangyunProvider{},
	"xinghai":    xinghaiProvider{},
	"feisuyuan":  feisuyuanProvider{},
}
```

- [ ] **Step 5: Update sync logic for list providers**

In `product_goods_channel_sync.go`, modify cache logic so list providers cache by platform account:

```go
func (l *ProductGoodsLogic) fetchProductGoodsChannelProductInfo(ctx context.Context, candidate productGoodsChannelSyncCandidate, cache map[string]supplierprovider.ProductInfoResult) (supplierprovider.ProductInfoResult, error) {
	provider, ok := supplierprovider.LookupProductInfo(candidate.ProviderCode)
	if !ok {
		return supplierprovider.ProductInfoResult{}, errProductInfoProviderUnsupported
	}
	cacheKey := fmt.Sprintf("%d:%s", candidate.PlatformAccountID, candidate.SupplierGoodsNo)
	if cached, exists := cache[cacheKey]; exists {
		return cached, nil
	}
	if listProvider, ok := provider.(supplierprovider.ProductInfoListProvider); ok {
		return l.fetchProductGoodsChannelProductInfoFromList(ctx, candidate, cache, listProvider)
	}
	return l.fetchProductGoodsChannelProductInfoByDetail(ctx, candidate, cache, provider)
}
```

Move the existing detail-request body into `fetchProductGoodsChannelProductInfoByDetail`. Add `fetchProductGoodsChannelProductInfoFromList` that caches each parsed item using `fmt.Sprintf("%d:%s", candidate.PlatformAccountID, supplierGoodsNo)`.

- [ ] **Step 6: Run product info and admin sync tests**

```bash
go test ./internal/library/supplierplatform/provider -run 'TestProductInfoProviderRegistriesIncludeAllConfiguredPlatforms|TestMultiPlatformProductInfoProviders|TestListBasedProductInfoProviders|TestKakayunProductInfoProvider' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestProductGoodsChannelSync' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 7: Commit**

```bash
git add internal/library/supplierplatform/provider/types.go internal/library/supplierplatform/provider/product_info_providers.go internal/library/supplierplatform/provider/registry.go internal/library/supplierplatform/provider/providers_test.go internal/logic/admin/product_goods_channel_sync.go internal/logic/admin/product_goods_channel_sync_test.go
git commit -m "feat: add supplier product info providers"
```

---

### Task 4: Segment Schema And Entity

**Files:**
- Modify: `manifest/sql/008_external_order.sql`
- Modify: `internal/app/schema.go`
- Modify: `internal/app/order_bootstrap.go`
- Modify: `internal/app/bootstrap.go`
- Modify: `internal/app/order_schema_test.go`
- Modify: `internal/model/entity/order.go`

- [ ] **Step 1: Write failing schema tests**

In `internal/app/order_schema_test.go`, add:

```go
func TestExternalOrderAttemptSegmentSchemaExists(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	columns := loadColumnNames(t, core, "external_order_attempt_segment")
	for _, column := range []string{
		"id", "order_id", "attempt_id", "segment_no", "quantity", "provider_code",
		"supplier_goods_no", "supplier_us_order_no", "supplier_order_no",
		"supplier_status", "refund_status", "request_snapshot", "response_snapshot",
		"receipt", "status", "submitted_at", "last_checked_at", "created_at", "updated_at",
	} {
		require.Contains(t, columns, column)
	}
}

func TestEnsureExternalOrderAttemptSegmentSchemaIsIdempotent(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS external_order_attempt_segment`)
	require.NoError(t, err)
	require.NoError(t, core.ensureExternalOrderAttemptSegmentSchema(ctx))
	require.NoError(t, core.ensureExternalOrderAttemptSegmentSchema(ctx))
	require.Contains(t, loadColumnNames(t, core, "external_order_attempt_segment"), "supplier_us_order_no")
}
```

- [ ] **Step 2: Run schema tests and verify failure**

```bash
go test ./internal/app -run 'TestExternalOrderAttemptSegmentSchema' -count=1 -timeout 60s
```

Expected: fail because the table and ensure method do not exist.

- [ ] **Step 3: Add MySQL migration SQL**

Append to `manifest/sql/008_external_order.sql` after `external_order_attempt`:

```sql
CREATE TABLE IF NOT EXISTS external_order_attempt_segment (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '尝试子单ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  attempt_id BIGINT UNSIGNED NOT NULL COMMENT '尝试ID',
  segment_no INT NOT NULL COMMENT '子单序号',
  quantity INT NOT NULL COMMENT '子单数量',
  provider_code VARCHAR(32) NOT NULL COMMENT '适配器编码',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_us_order_no VARCHAR(80) NOT NULL COMMENT '上游商家单号',
  supplier_order_no VARCHAR(128) NOT NULL DEFAULT '' COMMENT '上游订单号',
  supplier_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游原始状态',
  refund_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游退款状态',
  request_snapshot TEXT NOT NULL COMMENT '请求快照',
  response_snapshot TEXT NOT NULL COMMENT '响应快照',
  receipt VARCHAR(512) NOT NULL DEFAULT '' COMMENT '上游回执',
  status VARCHAR(32) NOT NULL COMMENT '子单状态',
  submitted_at DATETIME NULL COMMENT '提交时间',
  last_checked_at DATETIME NULL COMMENT '最近查单时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_external_order_attempt_segment_no (attempt_id, segment_no),
  UNIQUE KEY uk_external_order_segment_supplier_us (provider_code, supplier_us_order_no),
  KEY idx_external_order_segment_attempt (attempt_id, id),
  KEY idx_external_order_segment_order (order_id, id)
) COMMENT='外部订单渠道尝试子单表';
```

- [ ] **Step 4: Add SQLite and MySQL schema strings**

In `internal/app/schema.go`, add matching `CREATE TABLE IF NOT EXISTS external_order_attempt_segment` blocks to both SQLite and MySQL schema constants, near `external_order_attempt`.

- [ ] **Step 5: Add bootstrap ensure method**

In `internal/app/order_bootstrap.go`, add `ensureExternalOrderAttemptSegmentSchema(ctx context.Context) error` that creates the full table for SQLite and MySQL. Use the same short `lock_wait_timeout` pattern as `ensureExternalOrderAttemptSchema`.

In `internal/app/bootstrap.go`, call it immediately after `ensureExternalOrderAttemptSchema`:

```go
if err := c.ensureExternalOrderAttemptSegmentSchema(ctx); err != nil {
	return err
}
```

- [ ] **Step 6: Add entity**

In `internal/model/entity/order.go`, add:

```go
// ExternalOrderAttemptSegment 表示一次渠道尝试下拆分出的真实上游子单。
type ExternalOrderAttemptSegment struct {
	ID                int64        `db:"id" json:"id"`
	OrderID           int64        `db:"order_id" json:"order_id"`
	AttemptID         int64        `db:"attempt_id" json:"attempt_id"`
	SegmentNo         int          `db:"segment_no" json:"segment_no"`
	Quantity          int          `db:"quantity" json:"quantity"`
	ProviderCode      string       `db:"provider_code" json:"provider_code"`
	SupplierGoodsNo   string       `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierUSOrderNo string       `db:"supplier_us_order_no" json:"supplier_us_order_no"`
	SupplierOrderNo   string       `db:"supplier_order_no" json:"supplier_order_no"`
	SupplierStatus    string       `db:"supplier_status" json:"supplier_status"`
	RefundStatus      string       `db:"refund_status" json:"refund_status"`
	RequestSnapshot   string       `db:"request_snapshot" json:"request_snapshot"`
	ResponseSnapshot  string       `db:"response_snapshot" json:"response_snapshot"`
	Receipt           string       `db:"receipt" json:"receipt"`
	Status            string       `db:"status" json:"status"`
	SubmittedAt       sql.NullTime `db:"submitted_at" json:"submitted_at"`
	LastCheckedAt     sql.NullTime `db:"last_checked_at" json:"last_checked_at"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
}
```

- [ ] **Step 7: Run schema tests**

```bash
go test ./internal/app -run 'TestExternalOrderAttemptSegmentSchema|TestOrderSchema' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 8: Commit**

```bash
git add manifest/sql/008_external_order.sql internal/app/schema.go internal/app/order_bootstrap.go internal/app/bootstrap.go internal/app/order_schema_test.go internal/model/entity/order.go
git commit -m "feat: add order attempt segment schema"
```

---

### Task 5: Generic Safety Price Calculation

**Files:**
- Modify: `internal/logic/order/order_loss_guard.go`
- Modify: `internal/logic/order/order_loss_guard_test.go`
- Modify: `internal/logic/order/order_submit.go`

- [ ] **Step 1: Write failing safety-price tests**

Replace `kakayunMaxMoney` tests with generic tests:

```go
func TestSegmentSafetyPriceTotalModeDisallowLoss(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "9.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "20.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeTotal, FieldName: "maxmoney"}},
		2,
		2,
	)
	require.NoError(t, err)
	require.Equal(t, "20.0000", result.Value)
	require.True(t, result.SendToSupplier)
}

func TestSegmentSafetyPriceUnitModeUsesUnitCeiling(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "20.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "4.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "40.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnit, FieldName: "safe_cost"}},
		2,
		2,
	)
	require.NoError(t, err)
	require.Equal(t, "20.0000", result.Value)
	require.True(t, result.SendToSupplier)
}

func TestSegmentSafetyPriceUnsupportedRunsLocalGuard(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "9.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnsupported}},
		1,
		1,
	)
	require.NoError(t, err)
	require.False(t, result.SendToSupplier)
	require.Empty(t, result.Value)
}

func TestSegmentSafetyPriceUnsupportedRejectsLocalLoss(t *testing.T) {
	_, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnsupported}},
		1,
		1,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "防亏损")
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./internal/logic/order -run TestSegmentSafetyPrice -count=1 -timeout 60s
```

Expected: fail because `computeSegmentSafetyPrice` does not exist.

- [ ] **Step 3: Implement generic safety price**

In `order_loss_guard.go`, replace `kakayunMaxMoney` with:

```go
type segmentSafetyPrice struct {
	Value          string
	SendToSupplier bool
}

func computeSegmentSafetyPrice(candidate orderChannelCandidate, config reorderConfig, snapshot channelpricing.OrderPriceSnapshot, capabilities supplierprovider.OrderProviderCapabilities, orderQuantity, segmentQuantity int) (segmentSafetyPrice, error) {
	if orderQuantity <= 0 || segmentQuantity <= 0 {
		return segmentSafetyPrice{}, fmt.Errorf("购买数量必须大于0")
	}
	sourceUnit, err := decimal.NewFromString(strings.TrimSpace(candidate.SourceCostPrice))
	if err != nil || sourceUnit.IsNegative() {
		return segmentSafetyPrice{}, fmt.Errorf("原始进货价格式错误")
	}
	orderAmount, err := decimal.NewFromString(strings.TrimSpace(snapshot.OrderAmount))
	if err != nil || orderAmount.IsNegative() {
		return segmentSafetyPrice{}, fmt.Errorf("订单金额格式错误")
	}
	allowedLoss := decimal.Zero
	if config.AllowLossSaleEnabled == 1 {
		allowedLoss, err = decimal.NewFromString(strings.TrimSpace(config.MaxLossAmount))
		if err != nil || allowedLoss.IsNegative() {
			return segmentSafetyPrice{}, fmt.Errorf("允许亏本金额格式错误")
		}
	}
	orderQty := decimal.NewFromInt(int64(orderQuantity))
	segmentQty := decimal.NewFromInt(int64(segmentQuantity))
	segmentOrderAmount := orderAmount.Div(orderQty).Mul(segmentQty).Round(4)
	segmentAllowedLoss := allowedLoss.Div(orderQty).Mul(segmentQty).Round(4)
	segmentSourceTotal := sourceUnit.Mul(segmentQty).Round(4)
	ceiling := segmentOrderAmount.Add(segmentAllowedLoss).Round(4)
	if capabilities.SafetyPrice.Mode == supplierprovider.SafetyPriceModeUnsupported && segmentSourceTotal.GreaterThan(ceiling) {
		return segmentSafetyPrice{}, fmt.Errorf("本地防亏损校验失败：进货总额%s超过允许上限%s", segmentSourceTotal.StringFixed(4), ceiling.StringFixed(4))
	}
	if capabilities.SafetyPrice.Mode == supplierprovider.SafetyPriceModeUnsupported {
		return segmentSafetyPrice{}, nil
	}
	value := segmentSourceTotal
	if ceiling.LessThan(value) {
		value = ceiling
	}
	if capabilities.SafetyPrice.Mode == supplierprovider.SafetyPriceModeUnit {
		value = value.Div(segmentQty).Round(4)
	}
	return segmentSafetyPrice{Value: value.StringFixed(4), SendToSupplier: true}, nil
}
```

- [ ] **Step 4: Update submit path to use SafePrice**

In `order_submit.go`, remove the call to `kakayunMaxMoney`. Segment submission in Task 6 will pass `SafePrice`. Until Task 6, pass a single-segment value:

```go
safetyPrice, err := computeSegmentSafetyPrice(candidate, config, priceSnapshot, provider.Capabilities(), order.Quantity, order.Quantity)
if err != nil {
	return l.markOrderFailed(ctx, order.ID, 0, "防亏损金额计算失败："+err.Error())
}
```

Then set `SafePrice`:

```go
SafePrice: safetyPrice.Value,
MaxMoney:  safetyPrice.Value,
```

- [ ] **Step 5: Run safety tests**

```bash
go test ./internal/logic/order -run 'TestSegmentSafetyPrice|TestOrderWorkerPassesKakayunMaxMoneyWithAllowedLoss' -count=1 -timeout 60s
```

Expected: pass. Keep the existing integration test name unchanged in this task to avoid unrelated test churn.

- [ ] **Step 6: Commit**

```bash
git add internal/logic/order/order_loss_guard.go internal/logic/order/order_loss_guard_test.go internal/logic/order/order_submit.go
git commit -m "feat: generalize supplier safety price"
```

---

### Task 6: Segment Submission And Polling

**Files:**
- Create: `internal/logic/order/order_segment.go`
- Modify: `internal/logic/order/order_submit.go`
- Modify: `internal/logic/order/order_poll.go`
- Modify: `internal/logic/order/order_channel.go`
- Modify: `internal/logic/order/order_test.go`
- Modify: `test/integration/order_worker_test.go`

- [ ] **Step 1: Write failing unit tests for segmentation**

Add to `internal/logic/order/order_test.go`:

```go
func TestBuildOrderSegmentsRespectsProviderMaxQuantity(t *testing.T) {
	segments := buildOrderSegments("O20260428123045123456-T1", 3, supplierprovider.OrderProviderCapabilities{MaxQuantityPerCreate: 1})
	require.Equal(t, []orderSegmentPlan{
		{SegmentNo: 1, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S1"},
		{SegmentNo: 2, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S2"},
		{SegmentNo: 3, Quantity: 1, SupplierUSOrderNo: "O20260428123045123456-T1-S3"},
	}, segments)
}

func TestBuildOrderSegmentsUsesSingleSegmentWhenUnlimited(t *testing.T) {
	segments := buildOrderSegments("O20260428123045123456-T1", 3, supplierprovider.OrderProviderCapabilities{})
	require.Equal(t, []orderSegmentPlan{{SegmentNo: 1, Quantity: 3, SupplierUSOrderNo: "O20260428123045123456-T1-S1"}}, segments)
}

func TestAggregateSegmentStatuses(t *testing.T) {
	success := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess},
		{Status: OrderAttemptStatusSuccess},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusSuccess, success.OrderStatus)
	require.Equal(t, OrderAttemptStatusSuccess, success.AttemptStatus)

	failed := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess},
		{Status: OrderAttemptStatusFailed, Receipt: "失败"},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusFailed, failed.OrderStatus)
	require.Equal(t, OrderAttemptStatusFailed, failed.AttemptStatus)

	processing := aggregateSegmentStatuses([]entity.ExternalOrderAttemptSegment{
		{Status: OrderAttemptStatusSuccess},
		{Status: OrderAttemptStatusProcessing},
	})
	require.Equal(t, supplierprovider.SupplierOrderStatusProcessing, processing.OrderStatus)
	require.Equal(t, OrderAttemptStatusProcessing, processing.AttemptStatus)
}
```

- [ ] **Step 2: Run unit tests and verify failure**

```bash
go test ./internal/logic/order -run 'TestBuildOrderSegments|TestAggregateSegmentStatuses' -count=1 -timeout 60s
```

Expected: fail because segment planner and aggregator do not exist.

- [ ] **Step 3: Implement `order_segment.go`**

Create `internal/logic/order/order_segment.go`:

```go
package orderlogic

import (
	"context"
	"database/sql"
	"strings"
	"time"

	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

type orderSegmentPlan struct {
	SegmentNo         int
	Quantity          int
	SupplierUSOrderNo string
}

type segmentAggregateResult struct {
	OrderStatus   string
	AttemptStatus string
	Receipt       string
	Message       string
}

func buildOrderSegments(baseSupplierUSOrderNo string, quantity int, capabilities supplierprovider.OrderProviderCapabilities) []orderSegmentPlan {
	maxQty := capabilities.MaxQuantityPerCreate
	if maxQty <= 0 || maxQty >= quantity {
		return []orderSegmentPlan{{SegmentNo: 1, Quantity: quantity, SupplierUSOrderNo: baseSupplierUSOrderNo + "-S1"}}
	}
	segments := make([]orderSegmentPlan, 0, (quantity+maxQty-1)/maxQty)
	remaining := quantity
	for remaining > 0 {
		segmentQty := maxQty
		if remaining < segmentQty {
			segmentQty = remaining
		}
		segmentNo := len(segments) + 1
		segments = append(segments, orderSegmentPlan{SegmentNo: segmentNo, Quantity: segmentQty, SupplierUSOrderNo: baseSupplierUSOrderNo + "-S" + intToString(segmentNo)})
		remaining -= segmentQty
	}
	return segments
}

func aggregateSegmentStatuses(segments []entity.ExternalOrderAttemptSegment) segmentAggregateResult {
	if len(segments) == 0 {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusUnknown, AttemptStatus: OrderAttemptStatusUnknown, Message: "上游子单为空"}
	}
	hasUnknown := false
	hasProcessing := false
	receipts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if strings.TrimSpace(segment.Receipt) != "" {
			receipts = append(receipts, segment.Receipt)
		}
		switch segment.Status {
		case OrderAttemptStatusFailed:
			return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusFailed, AttemptStatus: OrderAttemptStatusFailed, Receipt: strings.Join(receipts, "；"), Message: defaultOrderMessage(segment.Receipt, "上游子单失败")}
		case OrderAttemptStatusUnknown:
			hasUnknown = true
		case OrderAttemptStatusProcessing, OrderAttemptStatusSubmitted, OrderAttemptStatusPending:
			hasProcessing = true
		}
	}
	if hasUnknown {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusUnknown, AttemptStatus: OrderAttemptStatusUnknown, Receipt: strings.Join(receipts, "；"), Message: "上游子单状态无法确认"}
	}
	if hasProcessing {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusProcessing, AttemptStatus: OrderAttemptStatusProcessing, Receipt: strings.Join(receipts, "；"), Message: "上游子单处理中"}
	}
	return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusSuccess, AttemptStatus: OrderAttemptStatusSuccess, Receipt: strings.Join(receipts, "；"), Message: "全部上游子单成功"}
}
```

Add persistence helpers in the same file:

```go
func insertAttemptSegments(ctx context.Context, tx gdb.TX, attempt entity.ExternalOrderAttempt, candidate orderChannelCandidate, plans []orderSegmentPlan, now time.Time) error {
	for _, plan := range plans {
		if _, err := tx.Exec(`
INSERT INTO external_order_attempt_segment (
    order_id, attempt_id, segment_no, quantity, provider_code, supplier_goods_no,
    supplier_us_order_no, supplier_order_no, supplier_status, refund_status,
    request_snapshot, response_snapshot, receipt, status, submitted_at, last_checked_at,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, '', '', '', '', '', '等待上游提交结果', ?, NULL, NULL, ?, ?)
`, attempt.OrderID, attempt.ID, plan.SegmentNo, plan.Quantity, candidate.ProviderCode, candidate.SupplierGoodsNo,
			plan.SupplierUSOrderNo, OrderAttemptStatusPending, now, now); err != nil {
			return err
		}
	}
	return nil
}

func (l *OrderLogic) loadAttemptSegments(ctx context.Context, attemptID int64) ([]entity.ExternalOrderAttemptSegment, error) {
	rows := make([]entity.ExternalOrderAttemptSegment, 0)
	err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT * FROM external_order_attempt_segment WHERE attempt_id = ? ORDER BY segment_no ASC`, attemptID)
	return rows, err
}
```

- [ ] **Step 4: Update claim to create segments**

In `order_submit.go`, change `claimOrderWithPendingAttempt` signature to accept `plans []orderSegmentPlan`. After inserting the attempt and assigning `attempt.ID`, set `attempt.OrderID = order.ID`, then call `insertAttemptSegments(ctx, tx, attempt, candidate, plans, now)`.

- [ ] **Step 5: Update submit execution to loop segments**

Replace single `executeCreateOrder` call with a segment loop:

```go
plans := buildOrderSegments(supplierUSOrderNo, order.Quantity, provider.Capabilities())
attempt, claimed, err := l.claimOrderWithPendingAttempt(ctx, order, candidate, attemptNo, supplierUSOrderNo, priceSnapshot, plans)
```

Then execute each segment using `CreateOrderInput{Quantity: plan.Quantity, SupplierUSOrderNo: plan.SupplierUSOrderNo, SafePrice: safetyPrice.Value}`. Update each segment row after response. After loop, load segments, aggregate, update attempt using aggregate, then update order as existing logic does.

- [ ] **Step 6: Update poll execution to loop segments**

In `order_poll.go`, load segments for the current attempt. If no segment rows exist, keep compatibility by synthesizing one segment from attempt fields. For each non-terminal segment, query provider with that segment's `SupplierOrderNo` and `SupplierUSOrderNo`, update segment row, then aggregate all segments and apply existing `applyPollSuccess`, `applyPollProcessing`, `applyPollAttemptFailed`, or `applyPollUnknown`.

- [ ] **Step 7: Remove kakayun-only candidate filter**

In `order_channel.go`, remove:

```sql
  AND a.provider_code = 'kakayun'
```

After scanning rows, filter unsupported providers:

```go
if _, ok := supplierprovider.LookupOrder(row.ProviderCode); !ok {
	continue
}
```

Add import for `supplierprovider`.

- [ ] **Step 8: Run unit tests**

```bash
go test ./internal/logic/order -run 'TestBuildOrderSegments|TestAggregateSegmentStatuses|TestSelectCandidate' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 9: Commit**

```bash
git add internal/logic/order/order_segment.go internal/logic/order/order_submit.go internal/logic/order/order_poll.go internal/logic/order/order_channel.go internal/logic/order/order_test.go
git commit -m "feat: add order segment aggregation"
```

---

### Task 7: Integration Tests For Non-Kakayun And Split Orders

**Files:**
- Modify: `test/integration/order_worker_test.go`
- Modify: `internal/logic/order/order_submit.go`
- Modify: `internal/logic/order/order_poll.go`

- [ ] **Step 1: Add generic platform helper**

In `test/integration/order_worker_test.go`, add:

```go
func (h *orderIntegrationHarness) createSupplierPlatform(t *testing.T, token, name string, typeID int, subjectID int64, hasTax int, host, tokenID string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           host,
		"backup_domain":    host,
		"type_id":          typeID,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             1,
		"crowd_name":       "订单群",
	}, token)
	require.Equal(t, 0, res.Code)
	var data struct{ ID int64 `json:"id"` }
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}
```

- [ ] **Step 2: Write non-kakayun integration test**

Add:

```go
func TestOrderWorkerSubmitsPendingOrderToYoukayun(t *testing.T) {
	var captured map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/buygoods":
			fields := parseRequestFieldsForIntegration(t, r)
			captured = map[string]string{
				"goodsid":    fields.Get("goodsid"),
				"account":    fields.Get("accountname"),
				"outorderno": fields.Get("outorderno"),
				"maxmoney":   fields.Get("maxmoney"),
			}
			_, _ = w.Write([]byte(`{"code":1000,"msg":"获取成功","data":{"ordersn":"YKY202604280001","outorderno":"` + fields.Get("outorderno") + `","money":"10.0000"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "优卡云品牌", "视频会员", "优酷")
	subjectID := h.createSubject(t, token, "优卡云主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "优卡云订单商品", "20.0000")
	platformID := h.createSupplierPlatform(t, token, "优卡云订单平台", 7, subjectID, 0, strings.TrimPrefix(server.URL, "http://"), "10052")
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "10001",
		"supplier_goods_name": "优卡云测试商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct{ GoodsCode string `json:"goods_code"` }
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))
	create := h.postJSON("/api/open/orders", map[string]any{"token": "test-open-order-token", "goods_id": goodsDetail.GoodsCode, "account": "13800138000", "quantity": 1}, "")
	require.Equal(t, 0, create.Code)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))

	require.Equal(t, "10001", captured["goodsid"])
	require.Equal(t, "13800138000", captured["account"])
	require.Contains(t, captured["outorderno"], "-T1-S1")
	require.Equal(t, "10.0000", captured["maxmoney"])
}
```

- [ ] **Step 3: Write feisuyuan split integration test**

Add:

```go
func TestOrderWorkerSplitsFeisuyuanQuantityIntoSegments(t *testing.T) {
	var received []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/recharge/order", r.URL.Path)
		body := readAllForIntegration(t, r)
		values, err := url.ParseQuery(body)
		require.NoError(t, err)
		require.Equal(t, "1", values.Get("number"))
		received = append(received, values.Get("outTradeNo"))
		_, _ = w.Write([]byte(`{"code":"2000","message":"ok"}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "飞速源品牌", "视频会员", "芒果")
	subjectID := h.createSubject(t, token, "飞速源主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "飞速源订单商品", "20.0000")
	_, err := h.app.Core().DB().Exec(context.Background(), `UPDATE product_goods SET max_purchase_qty = 3 WHERE id = ?`, goodsID)
	require.NoError(t, err)
	platformID := h.createSupplierPlatform(t, token, "飞速源订单平台", 56, subjectID, 0, strings.TrimPrefix(server.URL, "http://"), "23329")
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "106",
		"supplier_goods_name": "飞速源测试商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct{ GoodsCode string `json:"goods_code"` }
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))
	create := h.postJSON("/api/open/orders", map[string]any{"token": "test-open-order-token", "goods_id": goodsDetail.GoodsCode, "account": "13800138000", "quantity": 3}, "")
	require.Equal(t, 0, create.Code)
	var createData struct{ OrderNo string `json:"order_no"` }
	require.NoError(t, json.Unmarshal(create.Data, &createData))

	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	require.Len(t, received, 3)
	require.Contains(t, received[0], "-S1")
	require.Contains(t, received[1], "-S2")
	require.Contains(t, received[2], "-S3")
	order := h.loadOrder(t, createData.OrderNo)
	attempt := h.loadCurrentAttempt(t, order.ID)
	require.EqualValues(t, 3, h.scalarInt(t, `SELECT COUNT(*) FROM external_order_attempt_segment WHERE attempt_id = ?`, attempt.ID))
}
```

Add helper:

```go
func readAllForIntegration(t *testing.T, r *http.Request) string {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return string(raw)
}

func parseRequestFieldsForIntegration(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		require.NoError(t, r.ParseMultipartForm(1<<20))
		return url.Values(r.MultipartForm.Value)
	}
	body := readAllForIntegration(t, r)
	values, err := url.ParseQuery(body)
	require.NoError(t, err)
	return values
}
```

- [ ] **Step 4: Run integration tests**

```bash
go test ./test/integration -run 'TestOrderWorker(SubmitsPendingOrderToYoukayun|SplitsFeisuyuanQuantityIntoSegments)' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 5: Run existing order worker regression**

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add test/integration/order_worker_test.go internal/logic/order/order_submit.go internal/logic/order/order_poll.go
git commit -m "test: cover multi-provider order segments"
```

---

### Task 8: Product Push Providers And Subscription Boundary

**Files:**
- Create: `internal/library/supplierplatform/provider/product_push_providers.go`
- Modify: `internal/library/supplierplatform/provider/registry.go`
- Modify: `internal/library/supplierplatform/provider/kakayun_product_push_test.go`
- Modify: `internal/logic/admin/product_goods_channel_subscription_test.go`

- [ ] **Step 1: Write failing push registry and parse tests**

In `kakayun_product_push_test.go`, add:

```go
func TestLookupProductChangePushRegistersSupportedNonKakayunProviders(t *testing.T) {
	for _, code := range []string{"kakayun", "kasushou", "xingquanyi", "youkayun"} {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupProductChangePush(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
	for _, code := range []string{"kayixin", "julangyun", "xinghai", "feisuyuan"} {
		t.Run(code, func(t *testing.T) {
			_, ok := LookupProductChangePush(code)
			require.False(t, ok)
		})
	}
}

func TestKasushouProductChangePushProviderParse(t *testing.T) {
	provider, ok := LookupProductChangePush("kasushou")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "kasushou", SecretKey: "secretXYZ"}
	payload := map[string]any{"id": "10001", "goods_price": "12.3400", "status": "1", "time": "1735002156123"}
	payload["sign"] = sha1Lower("1735002156123" + `{"id":"10001","time":"1735002156123"}` + "secretXYZ")
	result, err := provider.ParseProductChangePush(account, time.UnixMilli(1735002156123), marshalJSONForTest(t, payload))
	require.NoError(t, err)
	require.Equal(t, "10001", result.SupplierGoodsNo)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "12.3400", result.GoodsPrice.StringFixed(4))
	require.Equal(t, "1", result.GoodsStatus)
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./internal/library/supplierplatform/provider -run 'TestLookupProductChangePushRegistersSupportedNonKakayunProviders|TestKasushouProductChangePushProviderParse' -count=1 -timeout 60s
```

Expected: fail because push providers are not registered.

- [ ] **Step 3: Implement push providers**

Create `product_push_providers.go`. Implement:

- kayixin: keep unregistered because the current `ProductChangePushProvider` interface only receives body bytes and cannot verify `X-App-Id`, `X-Timestamp`, `X-Signature`; rely on active monitoring for price/name changes.
- kasushou: verify body `sign` with documented `id` and `time`.
- xingquanyi: verify body sign using common xingquanyi signing over non-empty params excluding `sign`; parse `product_id`, `product_name`, `event_data.price`.
- youkayun: verify `sign` using sorted query excluding sign plus secret; parse `goods_id`, `goods_price`, `status`.

- [ ] **Step 4: Update registry**

In `registry.go`, register only verified push providers:

```go
var defaultProductChangePushRegistry = map[string]ProductChangePushProvider{
	"kakayun":    kakayunProvider{},
	"kasushou":   kasushouProvider{},
	"xingquanyi": xingquanyiProvider{},
	"youkayun":   youkayunProvider{},
}
```

- [ ] **Step 5: Add non-kakayun no-subscription test**

In `internal/logic/admin/product_goods_channel_subscription_test.go`, add or update a test that creates a non-kakayun binding and asserts:

```go
require.EqualValues(t, 0, scalarAdminTestInt(t, core, `SELECT COUNT(*) FROM supplier_product_subscription WHERE provider_code <> 'kakayun'`))
```

- [ ] **Step 6: Run push and subscription tests**

```bash
go test ./internal/library/supplierplatform/provider -run 'Test.*ProductChangePush|TestLookupProductSubscriptionOnlyRegistersKakayun' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'Test.*SupplierProductSubscription|TestAutoSubscribeKakayunBinding' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 7: Commit**

```bash
git add internal/library/supplierplatform/provider/product_push_providers.go internal/library/supplierplatform/provider/registry.go internal/library/supplierplatform/provider/kakayun_product_push_test.go internal/logic/admin/product_goods_channel_subscription_test.go
git commit -m "feat: add verifiable product push providers"
```

---

### Task 9: Documentation And Contract Updates

**Files:**
- Modify: `docs/development.md`
- Modify: `docs/module-map.md`
- Modify: `docs/testing.md`

- [ ] **Step 1: Update development docs**

In `docs/development.md`, update the supplier provider section with:

```markdown
## 多供应商订单与商品同步 provider

订单 provider 通过 `internal/library/supplierplatform/provider.OrderProvider` 暴露统一下单和查单能力。新增平台时必须声明 `Capabilities()`，明确单次最大提交数量和安全价字段口径；订单业务层根据该能力拆分 `external_order_attempt_segment`，provider 只负责协议字段和状态映射。

商品详情同步优先实现 `ProductInfoProvider`。如果平台只有全量商品列表接口，则同时实现 `ProductInfoListProvider`，业务侧按平台账号和商品编号缓存本轮列表结果，避免跨账号污染。

商品订阅记录只面向卡卡云。其他平台即使支持后台配置商品通知，也不写入 `supplier_product_subscription`，不提供取消或重订阅接口，价格名称靠主动同步兜底。
```

- [ ] **Step 2: Update module map**

In `docs/module-map.md`, update third-party and order sections to mention:

```markdown
- 主要能力：平台类型字典、平台账号分页、详情、增删改、启停、余额刷新、余额日志落库、多供应商下单/查单 provider、商品详情主动同步、卡卡云商品订阅、商品变动推送回调。
- 边界：订单履约通过 provider capability 判断是否拆单和是否传上游安全价；卡卡云以外的平台不进入订阅列表，不调用上游订阅或取消订阅。
```

Update order section:

```markdown
- 主要能力：开放下单、开放查单、待提交扫描、多供应商提交、segment 子单聚合轮询、窗口内补单、后台订单列表筛选和统计。
- 边界：本地订单响应结构不暴露 segment；segment 仅用于保存拆单平台的真实上游请求和查单状态。
```

- [ ] **Step 3: Update testing docs**

In `docs/testing.md`, add focused commands:

```markdown
多供应商 provider 聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run 'TestMultiPlatform(Order|ProductInfo)|Test.*ProductChangePush|TestOrderProviderCapabilities' -count=1 -timeout 60s
```

订单拆单和防亏损聚焦回归：

```bash
go test ./internal/logic/order -run 'TestSegmentSafetyPrice|TestBuildOrderSegments|TestAggregateSegmentStatuses' -count=1 -timeout 60s
go test ./test/integration -run 'TestOrderWorker(SubmitsPendingOrderToYoukayun|SplitsFeisuyuanQuantityIntoSegments)' -count=1 -timeout 60s
```
```

- [ ] **Step 4: Run doc-related checks**

```bash
go test ./test/contract -run 'TestOrderDocsMentionOpenOrderAndWorker|TestAPILayout' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add docs/development.md docs/module-map.md docs/testing.md
git commit -m "docs: document multi-provider order integration"
```

---

### Task 10: Full Verification

**Files:**
- Verification-only task; defects found here are fixed in the task that introduced them before final handoff.

- [ ] **Step 1: Run provider package tests**

```bash
go test ./internal/library/supplierplatform/provider -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 2: Run order package tests**

```bash
go test ./internal/logic/order -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 3: Run admin product tests**

```bash
go test ./internal/logic/admin -run 'Test.*ProductGoodsChannel.*|Test.*SupplierProduct.*' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 4: Run integration order worker tests**

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 5: Run full test suite**

```bash
go test ./... -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 6: Build**

```bash
go build ./...
```

Expected: pass with no output.

- [ ] **Step 7: Lint**

```bash
golangci-lint run --timeout=5m
```

Expected: pass with no output.

- [ ] **Step 8: Inspect git status**

```bash
git status --short
```

Expected: no unstaged files except intentional final documentation updates.
