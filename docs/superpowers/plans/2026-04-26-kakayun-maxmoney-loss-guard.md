# Kakayun Maxmoney Loss Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 卡卡云下单时按商品亏本配置和渠道原始进货价传递 `maxmoney`，防止上游调价导致亏本。

**Architecture:** 订单逻辑层负责读取 `source_cost_price` 和商品库存配置，并计算 `maxmoney`；supplier provider 只把 `CreateOrderInput.MaxMoney` 转成卡卡云协议字段。防亏本计算独立放在 `internal/logic/order/order_loss_guard.go`，避免把业务规则塞进 provider。

**Tech Stack:** Go、GoFrame 数据库访问、shopspring/decimal、httptest、testify、Markdown 文档。

---

## File Structure

- Modify: `internal/library/supplierplatform/provider/types.go`
  - `CreateOrderInput` 增加 `MaxMoney string`，表示上游下单防亏本最大进货总金额。
- Modify: `internal/library/supplierplatform/provider/providers.go`
  - 卡卡云创建订单 payload 在 `MaxMoney` 非空时增加 `maxmoney`，再计算签名。
- Modify: `internal/library/supplierplatform/provider/providers_test.go`
  - 验证卡卡云创建订单请求包含 `maxmoney` 且签名覆盖该字段。
- Modify: `internal/logic/order/order_channel.go`
  - 候选渠道结构和查询补充 `source_cost_price`。
- Modify: `internal/logic/order/order_reorder.go`
  - `reorderConfig` 补充 `AllowLossSaleEnabled` 和 `MaxLossAmount`。
- Create: `internal/logic/order/order_loss_guard.go`
  - 独立实现卡卡云 `maxmoney` 计算，输入为候选渠道、库存配置、订单金额快照和数量。
- Create: `internal/logic/order/order_loss_guard_test.go`
  - 覆盖不允许亏本、允许亏本、原始进货总额更低、非法金额四类行为。
- Modify: `internal/logic/order/order_submit.go`
  - 提交前计算 `maxmoney`，并填入 `supplierprovider.CreateOrderInput`。
- Modify: `test/integration/order_worker_test.go`
  - 验证正常提交、允许亏本、补单切换渠道时的 `maxmoney`。
- Modify: `docs/development.md`
  - 补充订单提交到卡卡云时的 `maxmoney` 规则。
- Modify: `docs/module-map.md`
  - 补充订单履约和第三方对接边界。
- Modify: `docs/testing.md`
  - 补充卡卡云防亏本聚焦回归命令。

---

### Task 1: Provider 支持 MaxMoney 字段

**Files:**
- Modify: `internal/library/supplierplatform/provider/types.go`
- Modify: `internal/library/supplierplatform/provider/providers.go`
- Modify: `internal/library/supplierplatform/provider/providers_test.go`

- [ ] **Step 1: 写 provider 失败测试**

把 `internal/library/supplierplatform/provider/providers_test.go` 中 `TestKakayunOrderProviderBuildCreateRequest` 替换为下面版本：

```go
func TestKakayunOrderProviderBuildCreateRequest(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)
	account := AccountConfig{
		ProviderCode: "kakayun",
		Domain:       "qqlogin.yxp8.cn",
		TokenID:      "10052",
		SecretKey:    "9aa3034b6beba7cf5bfcf6089218a674",
	}
	now := time.Unix(1735002156, 0)
	req, err := provider.BuildCreateOrderRequest(context.Background(), account, now, "http://qqlogin.yxp8.cn", CreateOrderInput{
		SupplierGoodsNo:   "720938",
		Quantity:          1,
		Account:           "13088888888",
		SupplierUSOrderNo: "O20260424153000123456-T1",
		MaxMoney:          "11.0000",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "http://qqlogin.yxp8.cn/dockapiv3/order/create", req.URL.String())

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, "10052", payload["userid"])
	require.Equal(t, float64(1735002156), payload["timestamp"])
	require.Equal(t, "720938", payload["goodsid"])
	require.Equal(t, float64(1), payload["buynum"])
	require.Equal(t, "13088888888", payload["attach"])
	require.Equal(t, "O20260424153000123456-T1", payload["usorderno"])
	require.Equal(t, "11.0000", payload["maxmoney"])

	expectedPayload := map[string]any{
		"userid":    "10052",
		"timestamp": now.Unix(),
		"goodsid":   "720938",
		"buynum":    1,
		"attach":    "13088888888",
		"usorderno": "O20260424153000123456-T1",
		"maxmoney":  "11.0000",
	}
	require.Equal(t, kakayunSign(expectedPayload, account.SecretKey), payload["sign"])
}
```

- [ ] **Step 2: 运行 provider 测试确认失败**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProviderBuildCreateRequest -count=1 -timeout 60s
```

Expected: FAIL，失败点为 `CreateOrderInput.MaxMoney undefined` 或请求体缺少 `maxmoney`。

- [ ] **Step 3: 实现 provider 最小改动**

在 `internal/library/supplierplatform/provider/types.go` 中更新 `CreateOrderInput`：

```go
// CreateOrderInput 是上游下单接口所需的最小业务参数。
type CreateOrderInput struct {
	SupplierGoodsNo   string
	Quantity          int
	Account           string
	SupplierUSOrderNo string
	MaxMoney          string
}
```

在 `internal/library/supplierplatform/provider/providers.go` 的 `kakayunProvider.BuildCreateOrderRequest` 中，创建 payload 后、计算签名前加入：

```go
	if maxMoney := strings.TrimSpace(input.MaxMoney); maxMoney != "" {
		payload["maxmoney"] = maxMoney
	}
```

最终该方法应保持这个顺序：先组装基础字段，再按需加入 `maxmoney`，最后执行 `payload["sign"] = kakayunSign(payload, account.SecretKey)`。

- [ ] **Step 4: 运行 provider 测试确认通过**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProviderBuildCreateRequest -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 5: 提交 provider 改动**

```bash
git add internal/library/supplierplatform/provider/types.go internal/library/supplierplatform/provider/providers.go internal/library/supplierplatform/provider/providers_test.go
git commit -m "feat: pass kakayun maxmoney in order request"
```

---

### Task 2: 实现订单层 maxmoney 计算 helper

**Files:**
- Create: `internal/logic/order/order_loss_guard.go`
- Create: `internal/logic/order/order_loss_guard_test.go`

- [ ] **Step 1: 写计算 helper 失败测试**

创建 `internal/logic/order/order_loss_guard_test.go`：

```go
package orderlogic

import (
	"testing"

	"myjob/internal/library/channelpricing"

	"github.com/stretchr/testify/require"
)

func TestKakayunMaxMoneyDisallowLossUsesLowerOfSourceTotalAndOrderAmount(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "9.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.NoError(t, err)
	require.Equal(t, "10.0000", maxMoney)
}

func TestKakayunMaxMoneyAllowLossAddsConfiguredTotalLoss(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "20.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "2.5000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.NoError(t, err)
	require.Equal(t, "12.5000", maxMoney)
}

func TestKakayunMaxMoneyCapsAtSourceTotalWhenSourceIsLower(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "5.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "30.0000"},
		2,
	)

	require.NoError(t, err)
	require.Equal(t, "22.0000", maxMoney)
}

func TestKakayunMaxMoneyRejectsInvalidMoney(t *testing.T) {
	_, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "bad"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "原始进货价")
}
```

- [ ] **Step 2: 运行订单 helper 测试确认失败**

Run:

```bash
go test ./internal/logic/order -run TestKakayunMaxMoney -count=1 -timeout 60s
```

Expected: FAIL，失败点为 `undefined: kakayunMaxMoney`、`unknown field SourceCostPrice` 或 `unknown field AllowLossSaleEnabled`。

- [ ] **Step 3: 新增计算 helper**

创建 `internal/logic/order/order_loss_guard.go`：

```go
package orderlogic

import (
	"fmt"
	"strings"

	"myjob/internal/library/channelpricing"

	"github.com/shopspring/decimal"
)

// kakayunMaxMoney 按卡卡云“最大进货总金额”口径计算防亏本阈值。
func kakayunMaxMoney(candidate orderChannelCandidate, config reorderConfig, snapshot channelpricing.OrderPriceSnapshot, quantity int) (string, error) {
	if quantity <= 0 {
		return "", fmt.Errorf("购买数量必须大于0")
	}

	sourceUnit, err := decimal.NewFromString(strings.TrimSpace(candidate.SourceCostPrice))
	if err != nil || sourceUnit.IsNegative() {
		return "", fmt.Errorf("原始进货价格式错误")
	}
	orderAmount, err := decimal.NewFromString(strings.TrimSpace(snapshot.OrderAmount))
	if err != nil || orderAmount.IsNegative() {
		return "", fmt.Errorf("订单金额格式错误")
	}

	allowedLoss := decimal.Zero
	if config.AllowLossSaleEnabled == 1 {
		allowedLoss, err = decimal.NewFromString(strings.TrimSpace(config.MaxLossAmount))
		if err != nil || allowedLoss.IsNegative() {
			return "", fmt.Errorf("允许亏本金额格式错误")
		}
	}

	sourceTotal := sourceUnit.Mul(decimal.NewFromInt(int64(quantity))).Round(4)
	salesCeiling := orderAmount.Add(allowedLoss).Round(4)
	maxMoney := sourceTotal
	if salesCeiling.LessThan(maxMoney) {
		maxMoney = salesCeiling
	}
	if maxMoney.IsNegative() {
		return "", fmt.Errorf("最大进货金额格式错误")
	}
	return maxMoney.Round(4).StringFixed(4), nil
}
```

- [ ] **Step 4: 扩展结构体字段让测试编译**

在 `internal/logic/order/order_channel.go` 的 `orderChannelCandidate` 增加字段：

```go
	SourceCostPrice    string `db:"source_cost_price"`
```

在 `internal/logic/order/order_reorder.go` 的 `reorderConfig` 增加字段：

```go
	AllowLossSaleEnabled int
	MaxLossAmount        string
```

- [ ] **Step 5: 运行订单 helper 测试确认通过**

Run:

```bash
go test ./internal/logic/order -run TestKakayunMaxMoney -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 6: 提交 helper 改动**

```bash
git add internal/logic/order/order_loss_guard.go internal/logic/order/order_loss_guard_test.go internal/logic/order/order_channel.go internal/logic/order/order_reorder.go
git commit -m "feat: calculate kakayun maxmoney"
```

---

### Task 3: 接入订单提交链路

**Files:**
- Modify: `internal/logic/order/order_channel.go`
- Modify: `internal/logic/order/order_submit.go`
- Modify: `internal/logic/order/order_reorder.go`

- [ ] **Step 1: 写查询和提交接入改动**

在 `internal/logic/order/order_channel.go` 的候选渠道 SELECT 中补充 `b.source_cost_price`。字段顺序保持结构体附近清晰：

```sql
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.source_cost_price,
    b.cost_price,
```

在 `internal/logic/order/order_submit.go` 的 `submitOrder` 中，选出 `candidate` 和 `supplierUSOrderNo` 后、创建 attempt 前计算金额快照和 `maxmoney`：

```go
	priceSnapshot, err := channelpricing.OrderSnapshot(candidate.pricingRule(), order.Quantity)
	if err != nil {
		return err
	}
	maxMoney, err := kakayunMaxMoney(candidate, config, priceSnapshot, order.Quantity)
	if err != nil {
		return err
	}
	attempt, claimed, err := l.claimOrderWithPendingAttempt(ctx, order, candidate, attemptNo, supplierUSOrderNo, priceSnapshot)
```

同时把 `CreateOrderInput` 增加 `MaxMoney`：

```go
	result, err := l.executeCreateOrder(ctx, provider, account, supplierprovider.CreateOrderInput{
		SupplierGoodsNo:   candidate.SupplierGoodsNo,
		Quantity:          order.Quantity,
		Account:           order.Account,
		SupplierUSOrderNo: supplierUSOrderNo,
		MaxMoney:          maxMoney,
	})
```

更新 `claimOrderWithPendingAttempt` 函数签名，避免重复计算快照：

```go
func (l *OrderLogic) claimOrderWithPendingAttempt(ctx context.Context, order entity.ExternalOrder, candidate orderChannelCandidate, attemptNo int, supplierUSOrderNo string, priceSnapshot channelpricing.OrderPriceSnapshot) (entity.ExternalOrderAttempt, bool, error) {
	now := l.core.Now()
	nextPollAt := now.Add(pollIntervalDuration(l.core.Config().OpenOrder.PollIntervalSeconds))
```

删除该函数内部原有的：

```go
	priceSnapshot, err := channelpricing.OrderSnapshot(candidate.pricingRule(), order.Quantity)
	if err != nil {
		return entity.ExternalOrderAttempt{}, false, err
	}
```

在 `internal/logic/order/order_submit.go` 的 `loadReorderConfig` 中扩展查询：

```go
	row := struct {
		SmartReorderEnabled   int    `db:"smart_reorder_enabled"`
		ReorderTimeoutEnabled int    `db:"reorder_timeout_enabled"`
		ReorderTimeoutMinutes int    `db:"reorder_timeout_minutes"`
		OrderStrategy         string `db:"order_strategy"`
		AllowLossSaleEnabled  int    `db:"allow_loss_sale_enabled"`
		MaxLossAmount         string `db:"max_loss_amount"`
	}{OrderStrategy: "fixed_order", MaxLossAmount: "0.0000"}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy,
       allow_loss_sale_enabled, max_loss_amount
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return reorderConfig{OrderStrategy: "fixed_order", MaxLossAmount: "0.0000"}, nil
	}
```

返回值补充字段：

```go
	return reorderConfig{
		SmartEnabled:          row.SmartReorderEnabled,
		TimeoutEnabled:        row.ReorderTimeoutEnabled,
		TimeoutMinutes:        row.ReorderTimeoutMinutes,
		OrderStrategy:         row.OrderStrategy,
		AllowLossSaleEnabled:  row.AllowLossSaleEnabled,
		MaxLossAmount:         row.MaxLossAmount,
	}, nil
```

- [ ] **Step 2: 运行订单包测试**

Run:

```bash
go test ./internal/logic/order -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 3: 提交订单链路改动**

```bash
git add internal/logic/order/order_channel.go internal/logic/order/order_submit.go internal/logic/order/order_reorder.go
git commit -m "feat: wire kakayun maxmoney into order submit"
```

---

### Task 4: 补订单集成测试

**Files:**
- Modify: `test/integration/order_worker_test.go`

- [ ] **Step 1: 修改基础提交测试，验证默认不允许亏本**

在 `TestOrderWorkerSubmitsPendingOrderToKakayun` 的 handler 中，解码 payload 后加入断言：

```go
		require.Equal(t, "10.0000", payload["maxmoney"])
```

该测试使用 helper 创建 `source_cost_price=10.0000`、商品售价 `20.0000`、数量 `1`，所以 `min(10.0000, 20.0000)` 应为 `10.0000`。

- [ ] **Step 2: 修改补单测试，验证每个渠道重新计算**

在 `TestOrderWorkerReordersOnlyInsideConfiguredWindow` 的 `/dockapiv3/order/create` 分支中解码 payload，并按调用次数断言：

```go
		case "/dockapiv3/order/create":
			failCount++
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			if failCount == 1 {
				require.Equal(t, "10.0000", payload["maxmoney"])
			} else {
				require.Equal(t, "11.0000", payload["maxmoney"])
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":1,"message":"下单成功","data":{"orderno":"SD20260424%04d","usorderno":"O-T%d"}}`, failCount, failCount)))
```

- [ ] **Step 3: 新增允许亏本集成测试**

在 `TestOrderWorkerSubmitsPendingOrderToKakayun` 后新增测试：

```go
func TestOrderWorkerPassesKakayunMaxMoneyWithAllowedLoss(t *testing.T) {
	captured := make([]map[string]any, 0, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/dockapiv3/order/create", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		captured = append(captured, payload)
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240099","usorderno":"` + payload["usorderno"].(string) + `"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "允许亏本品牌", "视频会员", "爱奇艺")
	subjectID := h.createSubject(t, token, "允许亏本渠道主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "允许亏本商品", "10.0000")
	platformID := h.createKakayunPlatform(t, token, "允许亏本云发卡", subjectID, 0, strings.TrimPrefix(server.URL, "http://"))
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478599",
		"supplier_goods_name": "允许亏本测试商品",
		"source_cost_price":   "20.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	saveConfig := h.request(http.MethodPut, "/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   0,
		"reorder_timeout_enabled": 0,
		"reorder_timeout_minutes": 0,
		"order_strategy":          "fixed_order",
		"sync_cost_price_enabled": 0,
		"sync_goods_name_enabled": 0,
		"allow_loss_sale_enabled": 1,
		"max_loss_amount":         "2.5000",
		"combo_goods_enabled":     0,
	}, token)
	require.Equal(t, 0, saveConfig.Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "13800138000",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	require.Len(t, captured, 1)
	require.Equal(t, "12.5000", captured[0]["maxmoney"])
}
```

- [ ] **Step 4: 运行集成聚焦测试**

Run:

```bash
go test ./test/integration -run 'TestOrderWorker(SubmitsPendingOrderToKakayun|PassesKakayunMaxMoneyWithAllowedLoss|ReordersOnlyInsideConfiguredWindow)' -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 5: 提交集成测试改动**

```bash
git add test/integration/order_worker_test.go
git commit -m "test: cover kakayun maxmoney order submission"
```

---

### Task 5: 同步文档

**Files:**
- Modify: `docs/development.md`
- Modify: `docs/module-map.md`
- Modify: `docs/testing.md`

- [ ] **Step 1: 更新开发文档**

在 `docs/development.md` 的“订单金额快照”段落后增加：

```markdown
卡卡云下单会传 `maxmoney` 作为上游防亏本最大进货总金额。该值使用渠道绑定的 `source_cost_price` 计算原始进货总额，再与订单销售金额加允许亏本金额取较小值：`min(source_cost_price * quantity, order_amount + allowed_loss)`。商品库存配置未开启亏本销售时，`allowed_loss` 固定为 `0`；开启时使用 `max_loss_amount`，且该金额按订单总额计入。
```

- [ ] **Step 2: 更新模块地图**

在 `docs/module-map.md` 的“订单履约”或“第三方对接”边界描述中补充一句：

```markdown
卡卡云下单会按订单层计算出的 `maxmoney` 传递防亏本上限，计算口径使用渠道绑定 `source_cost_price`，provider 只负责协议字段透传和签名。
```

- [ ] **Step 3: 更新测试文档**

在 `docs/testing.md` 卡卡云订单 provider 聚焦回归附近增加：

````markdown
卡卡云下单防亏本聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProviderBuildCreateRequest -count=1 -timeout 60s
go test ./internal/logic/order -run TestKakayunMaxMoney -count=1 -timeout 60s
go test ./test/integration -run 'TestOrderWorker(SubmitsPendingOrderToKakayun|PassesKakayunMaxMoneyWithAllowedLoss|ReordersOnlyInsideConfiguredWindow)' -count=1 -timeout 60s
```
````

- [ ] **Step 4: 提交文档改动**

```bash
git add docs/development.md docs/module-map.md docs/testing.md
git commit -m "docs: document kakayun maxmoney loss guard"
```

---

### Task 6: 全量验证

**Files:**
- Read: all modified files from previous tasks.

- [ ] **Step 1: 运行 provider 聚焦测试**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProvider -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 2: 运行订单逻辑测试**

Run:

```bash
go test ./internal/logic/order -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 3: 运行订单集成聚焦测试**

Run:

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 4: 运行全量测试**

Run:

```bash
go test ./... -count=1 -timeout 60s
```

Expected: PASS。

- [ ] **Step 5: 运行全量构建**

Run:

```bash
go build ./...
```

Expected: command exits 0。

- [ ] **Step 6: 运行 lint**

Run:

```bash
golangci-lint run --timeout=5m
```

Expected: command exits 0。

- [ ] **Step 7: 检查最终 diff**

Run:

```bash
git status --short
git diff --stat HEAD~4..HEAD
```

Expected: 工作区干净；diff 只包含 provider、order、集成测试和文档相关改动。
