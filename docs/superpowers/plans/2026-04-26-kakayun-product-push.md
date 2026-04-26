# Kakayun Product Push Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Kakayun product price push callbacks, automatic product subscription, subscription management, and unified price change logs.

**Architecture:** Keep the callback URL generic with `{providerCode}/{platformAccountId}` so multiple supplier accounts can coexist safely. Kakayun provider owns signing, subscription requests, and push parsing; admin logic owns local subscription records, price updates, and price-change logs. Existing channel cost and effective sell price calculations remain the single source of truth.

**Tech Stack:** Go, GoFrame HTTP/router/middleware, MySQL/SQLite schema bootstrap, `shopspring/decimal`, existing `supplierplatform/provider` adapters, contract tests under `test/contract`.

---

## File Structure

Create these files:

- `api/supplier_product_subscription.go`: admin subscription list, cancel, and resubscribe protocol.
- `api/product_goods_channel_price_change.go`: admin automatic price change log protocol.
- `api/supplier_product_callback.go`: open supplier product-change callback protocol.
- `internal/controller/admin/supplier_product_subscription.go`: thin admin handlers for subscription list/cancel/resubscribe.
- `internal/controller/admin/product_goods_channel_price_change.go`: thin admin handler for price change list.
- `internal/controller/open/supplier_product_callback.go`: open callback handler that writes plain `ok`.
- `internal/service/supplier_product_subscription.go`: service interfaces for subscription, price change, and callback capabilities.
- `internal/model/entity/supplier_product_subscription.go`: subscription and price change log entity structs.
- `internal/app/supplier_product_push_schema.go`: idempotent schema creation for the two new tables.
- `manifest/sql/009_supplier_product_push.sql`: MySQL schema for subscription and price change log tables.
- `internal/library/supplierplatform/provider/product_push_types.go`: provider interfaces and DTOs for subscription and push parsing.
- `internal/library/supplierplatform/provider/kakayun_product_push.go`: Kakayun subscription and push implementation.
- `internal/library/supplierplatform/provider/kakayun_product_push_test.go`: provider unit tests.
- `internal/logic/admin/product_goods_channel_subscription.go`: subscription orchestration and admin subscription list/cancel/resubscribe logic.
- `internal/logic/admin/product_goods_channel_price_change.go`: shared price update and price-change log logic plus admin list query.
- `internal/logic/admin/product_goods_channel_subscription_test.go`: subscription logic tests.
- `internal/logic/admin/product_goods_channel_price_change_test.go`: push/monitor price-change tests.
- `test/contract/supplier_product_subscription_contract_test.go`: admin subscription and open callback contract tests.
- `test/contract/product_goods_channel_price_change_contract_test.go`: price-change list contract tests.

Modify these files:

- `internal/app/bootstrap.go`: call the new schema ensure method.
- `internal/app/schema.go`: include both tables in SQLite and MySQL bootstrap schema.
- `internal/app/mysql_schema_comment_test.go`: add `009_supplier_product_push.sql` to manifest comment checks.
- `internal/dao/tables.go`: add table constants and model helpers for the new tables.
- `internal/library/supplierplatform/provider/registry.go`: register Kakayun subscription and push providers.
- `internal/logic/admin/common.go`: expose new service interfaces from `Services`.
- `internal/logic/admin/product_goods_channel_write.go`: trigger auto-subscription after create/update succeeds.
- `internal/logic/admin/product_goods_channel_sync.go`: refactor monitor apply path to write price-change logs.
- `internal/bootstrap/application.go`: register new open and admin controllers under the correct middleware.
- `test/contract/api_layout_test.go`: add new flat `api/*.go` files to expected layout.
- `docs/module-map.md`: document subscription, callback, and price-change capabilities.
- `docs/development.md`: document provider subscription/push development flow.
- `docs/testing.md`: add focused test commands.
- `docs/superpowers/README.md`: add this implementation plan entry.

---

### Task 1: Schema, DAO, And Entities

**Files:**
- Create: `manifest/sql/009_supplier_product_push.sql`
- Create: `internal/app/supplier_product_push_schema.go`
- Create: `internal/model/entity/supplier_product_subscription.go`
- Modify: `internal/app/schema.go`
- Modify: `internal/app/bootstrap.go`
- Modify: `internal/app/mysql_schema_comment_test.go`
- Modify: `internal/dao/tables.go`
- Test: `internal/app/supplier_product_push_schema_test.go`

- [ ] **Step 1: Write the failing schema test**

Create `internal/app/supplier_product_push_schema_test.go`:

```go
package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureSupplierProductPushSchemaCreatesTables(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS supplier_product_subscription`)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS product_goods_channel_price_change_log`)
	require.NoError(t, err)

	err = core.ensureSupplierProductPushSchema(ctx)
	require.NoError(t, err)

	subscriptionColumns := loadColumnNames(t, core, "supplier_product_subscription")
	require.Contains(t, subscriptionColumns, "provider_code")
	require.Contains(t, subscriptionColumns, "platform_account_id")
	require.Contains(t, subscriptionColumns, "supplier_goods_no")
	require.Contains(t, subscriptionColumns, "callback_url")
	require.Contains(t, subscriptionColumns, "status")
	require.Contains(t, subscriptionColumns, "last_action")
	require.Contains(t, subscriptionColumns, "last_error")
	require.Contains(t, subscriptionColumns, "request_snapshot")
	require.Contains(t, subscriptionColumns, "response_snapshot")
	require.Contains(t, subscriptionColumns, "subscribed_at")
	require.Contains(t, subscriptionColumns, "canceled_at")

	priceLogColumns := loadColumnNames(t, core, "product_goods_channel_price_change_log")
	require.Contains(t, priceLogColumns, "source")
	require.Contains(t, priceLogColumns, "provider_code")
	require.Contains(t, priceLogColumns, "platform_account_id")
	require.Contains(t, priceLogColumns, "binding_id")
	require.Contains(t, priceLogColumns, "goods_id")
	require.Contains(t, priceLogColumns, "old_source_cost_price")
	require.Contains(t, priceLogColumns, "new_source_cost_price")
	require.Contains(t, priceLogColumns, "old_effective_sell_price")
	require.Contains(t, priceLogColumns, "new_effective_sell_price")
	require.Contains(t, priceLogColumns, "change_amount")
	require.Contains(t, priceLogColumns, "description")
	require.Contains(t, priceLogColumns, "raw_payload")
}

func loadColumnNames(t *testing.T, core *Core, table string) []string {
	t.Helper()
	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	err := core.DB().GetCore().GetScan(context.Background(), &rows, `SHOW COLUMNS FROM `+table)
	require.NoError(t, err)

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Field)
	}
	return names
}
```

- [ ] **Step 2: Run the schema test and verify it fails**

Run:

```bash
go test ./internal/app -run TestEnsureSupplierProductPushSchemaCreatesTables -count=1 -timeout 60s
```

Expected: fail because `ensureSupplierProductPushSchema` is not defined.

- [ ] **Step 3: Add MySQL manifest schema**

Create `manifest/sql/009_supplier_product_push.sql`:

```sql
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS supplier_product_subscription (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订阅记录ID',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  goods_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '本地商品ID',
  binding_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '渠道绑定ID',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  callback_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '订阅回调地址',
  status VARCHAR(32) NOT NULL COMMENT '订阅状态',
  last_action VARCHAR(32) NOT NULL DEFAULT '' COMMENT '最近动作',
  last_error VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  request_snapshot TEXT NOT NULL COMMENT '最近请求快照',
  response_snapshot TEXT NOT NULL COMMENT '最近响应快照',
  subscribed_at DATETIME NULL COMMENT '最近订阅成功时间',
  canceled_at DATETIME NULL COMMENT '最近取消成功时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_supplier_product_subscription_active (provider_code, platform_account_id, supplier_goods_no),
  KEY idx_supplier_product_subscription_status (status, updated_at),
  KEY idx_supplier_product_subscription_goods (goods_id, binding_id),
  KEY idx_supplier_product_subscription_platform (platform_account_id, updated_at)
) COMMENT='供应商商品订阅记录表';

CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '改价记录ID',
  source VARCHAR(32) NOT NULL COMMENT '来源',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  binding_id BIGINT UNSIGNED NOT NULL COMMENT '渠道绑定ID',
  goods_id BIGINT UNSIGNED NOT NULL COMMENT '本地商品ID',
  goods_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '本地商品编码快照',
  goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '本地商品名称快照',
  goods_icon VARCHAR(500) NOT NULL DEFAULT '' COMMENT '商品图标快照',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  old_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前原始进货价',
  new_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后原始进货价',
  old_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前比较成本价',
  new_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后比较成本价',
  old_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前利润后价格',
  new_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后利润后价格',
  change_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润后价格变化值',
  description TEXT NOT NULL COMMENT '变动描述',
  raw_payload TEXT NOT NULL COMMENT '原始载荷',
  changed_at DATETIME NOT NULL COMMENT '变动时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  KEY idx_price_change_log_changed (changed_at, id),
  KEY idx_price_change_log_goods (goods_id, changed_at),
  KEY idx_price_change_log_supplier (provider_code, platform_account_id, supplier_goods_no, changed_at),
  KEY idx_price_change_log_source (source, changed_at)
) COMMENT='商品渠道自动改价记录表';
```

- [ ] **Step 4: Add entity structs**

Create `internal/model/entity/supplier_product_subscription.go`:

```go
package entity

import (
	"database/sql"
	"time"
)

// SupplierProductSubscription 表示供应商商品推送订阅的本地状态。
type SupplierProductSubscription struct {
	ID                  int64        `db:"id" json:"id"`
	ProviderCode        string       `db:"provider_code" json:"provider_code"`
	PlatformAccountID   int64        `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName string       `db:"platform_account_name" json:"platform_account_name"`
	GoodsID             int64        `db:"goods_id" json:"goods_id"`
	BindingID           int64        `db:"binding_id" json:"binding_id"`
	SupplierGoodsNo     string       `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName   string       `db:"supplier_goods_name" json:"supplier_goods_name"`
	CallbackURL         string       `db:"callback_url" json:"callback_url"`
	Status              string       `db:"status" json:"status"`
	LastAction          string       `db:"last_action" json:"last_action"`
	LastError           string       `db:"last_error" json:"last_error"`
	RequestSnapshot     string       `db:"request_snapshot" json:"request_snapshot"`
	ResponseSnapshot    string       `db:"response_snapshot" json:"response_snapshot"`
	SubscribedAt        sql.NullTime `db:"subscribed_at" json:"subscribed_at"`
	CanceledAt          sql.NullTime `db:"canceled_at" json:"canceled_at"`
	CreatedAt           time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time    `db:"updated_at" json:"updated_at"`
}

// ProductGoodsChannelPriceChangeLog 表示一次由监控或推送触发的渠道价格变化。
type ProductGoodsChannelPriceChangeLog struct {
	ID                    int64     `db:"id" json:"id"`
	Source                string    `db:"source" json:"source"`
	ProviderCode          string    `db:"provider_code" json:"provider_code"`
	PlatformAccountID     int64     `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName   string    `db:"platform_account_name" json:"platform_account_name"`
	BindingID             int64     `db:"binding_id" json:"binding_id"`
	GoodsID               int64     `db:"goods_id" json:"goods_id"`
	GoodsCode             string    `db:"goods_code" json:"goods_code"`
	GoodsName             string    `db:"goods_name" json:"goods_name"`
	GoodsIcon             string    `db:"goods_icon" json:"goods_icon"`
	SupplierGoodsNo       string    `db:"supplier_goods_no" json:"supplier_goods_no"`
	SupplierGoodsName     string    `db:"supplier_goods_name" json:"supplier_goods_name"`
	OldSourceCostPrice    string    `db:"old_source_cost_price" json:"old_source_cost_price"`
	NewSourceCostPrice    string    `db:"new_source_cost_price" json:"new_source_cost_price"`
	OldCostPrice          string    `db:"old_cost_price" json:"old_cost_price"`
	NewCostPrice          string    `db:"new_cost_price" json:"new_cost_price"`
	OldEffectiveSellPrice string    `db:"old_effective_sell_price" json:"old_effective_sell_price"`
	NewEffectiveSellPrice string    `db:"new_effective_sell_price" json:"new_effective_sell_price"`
	ChangeAmount          string    `db:"change_amount" json:"change_amount"`
	Description           string    `db:"description" json:"description"`
	RawPayload            string    `db:"raw_payload" json:"raw_payload"`
	ChangedAt             time.Time `db:"changed_at" json:"changed_at"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}
```

- [ ] **Step 5: Add schema ensure method**

Create `internal/app/supplier_product_push_schema.go` with driver-specific table creation. Use this exact shape:

```go
package app

import "context"

// ensureSupplierProductPushSchema 确保供应商商品订阅和自动改价记录表存在。
func (c *Core) ensureSupplierProductPushSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    goods_id INTEGER NOT NULL DEFAULT 0,
    binding_id INTEGER NOT NULL DEFAULT 0,
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    callback_url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    last_action TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    subscribed_at DATETIME NULL,
    canceled_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(provider_code, platform_account_id, supplier_goods_no)
)`); err != nil {
			return err
		}
		if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    binding_id INTEGER NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL DEFAULT '',
    goods_name TEXT NOT NULL DEFAULT '',
    goods_icon TEXT NOT NULL DEFAULT '',
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    old_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    new_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    change_amount TEXT NOT NULL DEFAULT '0.0000',
    description TEXT NOT NULL,
    raw_payload TEXT NOT NULL,
    changed_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
)`); err != nil {
			return err
		}
		return nil
	}

	if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订阅记录ID',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  goods_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '本地商品ID',
  binding_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '渠道绑定ID',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  callback_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '订阅回调地址',
  status VARCHAR(32) NOT NULL COMMENT '订阅状态',
  last_action VARCHAR(32) NOT NULL DEFAULT '' COMMENT '最近动作',
  last_error VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  request_snapshot TEXT NOT NULL COMMENT '最近请求快照',
  response_snapshot TEXT NOT NULL COMMENT '最近响应快照',
  subscribed_at DATETIME NULL COMMENT '最近订阅成功时间',
  canceled_at DATETIME NULL COMMENT '最近取消成功时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_supplier_product_subscription_active (provider_code, platform_account_id, supplier_goods_no),
  KEY idx_supplier_product_subscription_status (status, updated_at),
  KEY idx_supplier_product_subscription_goods (goods_id, binding_id),
  KEY idx_supplier_product_subscription_platform (platform_account_id, updated_at)
) COMMENT='供应商商品订阅记录表'`); err != nil {
		return err
	}
	if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '改价记录ID',
  source VARCHAR(32) NOT NULL COMMENT '来源',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  binding_id BIGINT UNSIGNED NOT NULL COMMENT '渠道绑定ID',
  goods_id BIGINT UNSIGNED NOT NULL COMMENT '本地商品ID',
  goods_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '本地商品编码快照',
  goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '本地商品名称快照',
  goods_icon VARCHAR(500) NOT NULL DEFAULT '' COMMENT '商品图标快照',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  old_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前原始进货价',
  new_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后原始进货价',
  old_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前比较成本价',
  new_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后比较成本价',
  old_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前利润后价格',
  new_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后利润后价格',
  change_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润后价格变化值',
  description TEXT NOT NULL COMMENT '变动描述',
  raw_payload TEXT NOT NULL COMMENT '原始载荷',
  changed_at DATETIME NOT NULL COMMENT '变动时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  KEY idx_price_change_log_changed (changed_at, id),
  KEY idx_price_change_log_goods (goods_id, changed_at),
  KEY idx_price_change_log_supplier (provider_code, platform_account_id, supplier_goods_no, changed_at),
  KEY idx_price_change_log_source (source, changed_at)
) COMMENT='商品渠道自动改价记录表'`); err != nil {
		return err
	}
	return nil
}
```

- [ ] **Step 6: Wire schema bootstrap and base schema strings**

Modify `internal/app/bootstrap.go` after `ensureExternalOrderAttemptSchema(ctx)`:

```go
	if err := c.ensureSupplierProductPushSchema(ctx); err != nil {
		return err
	}
```

Modify `internal/app/schema.go` by adding SQLite and MySQL create-table statements for both tables. Keep them text-equivalent to Step 5 so test DBs and runtime bootstrap match.

- [ ] **Step 7: Add DAO constants and manifest comment coverage**

Modify `internal/dao/tables.go`:

```go
	TableSupplierProductSubscription       = "supplier_product_subscription"
	TableProductGoodsChannelPriceChangeLog = "product_goods_channel_price_change_log"
```

Add helpers:

```go
func SupplierProductSubscriptionModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableSupplierProductSubscription).Ctx(ctx).Safe()
}

func ProductGoodsChannelPriceChangeLogModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableProductGoodsChannelPriceChangeLog).Ctx(ctx).Safe()
}
```

Modify `internal/app/mysql_schema_comment_test.go`:

```go
for _, name := range []string{
	"001_schema.sql", "005_supplier_platform.sql", "006_product_goods_channel_binding.sql",
	"007_product_goods_channel_config.sql", "009_supplier_product_push.sql",
} {
```

- [ ] **Step 8: Run schema tests**

Run:

```bash
go test ./internal/app -run 'TestEnsureSupplierProductPushSchemaCreatesTables|TestMySQLSchemaIncludesTableAndColumnComments|TestManifestMySQLSchemaFilesIncludeTableAndColumnComments' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 9: Commit schema work**

```bash
git add manifest/sql/009_supplier_product_push.sql internal/app/schema.go internal/app/bootstrap.go internal/app/supplier_product_push_schema.go internal/app/supplier_product_push_schema_test.go internal/app/mysql_schema_comment_test.go internal/dao/tables.go internal/model/entity/supplier_product_subscription.go
git commit -m "feat: add supplier product push schema"
```

---

### Task 2: Kakayun Provider Subscription And Push Parsing

**Files:**
- Create: `internal/library/supplierplatform/provider/product_push_types.go`
- Create: `internal/library/supplierplatform/provider/kakayun_product_push.go`
- Create: `internal/library/supplierplatform/provider/kakayun_product_push_test.go`
- Modify: `internal/library/supplierplatform/provider/registry.go`

- [ ] **Step 1: Write provider tests first**

Create `internal/library/supplierplatform/provider/kakayun_product_push_test.go`:

```go
package supplierprovider

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKakayunProductSubscriptionProviderBuildRequests(t *testing.T) {
	provider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)

	account := AccountConfig{ProviderCode: "kakayun", TokenID: "10052", SecretKey: "secretXYZ"}
	now := time.Unix(1735002156, 0)

	getURLReq, err := provider.BuildGetReceiveURLsRequest(context.Background(), account, now)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, getURLReq.Method)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/user/geturl", getURLReq.URL.String())
	getURLBody := decodeJSONBodyAny(t, readRequestBody(t, getURLReq))
	require.Equal(t, "10052", getURLBody["userid"])
	require.Equal(t, float64(1735002156), getURLBody["timestamp"])
	require.NotEmpty(t, getURLBody["sign"])

	setURLReq, err := provider.BuildSetReceiveURLRequest(context.Background(), account, now, ProductReceiveURLInput{ReceiveURL: "https://example.com/callback"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/user/seturl", setURLReq.URL.String())
	setURLBody := decodeJSONBodyAny(t, readRequestBody(t, setURLReq))
	require.Equal(t, "https://example.com/callback", setURLBody["receiveurl"])

	subscribeReq, err := provider.BuildSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/subscribe", subscribeReq.URL.String())
	subscribeBody := decodeJSONBodyAny(t, readRequestBody(t, subscribeReq))
	require.Equal(t, "2582531", subscribeBody["goodsid"])

	cancelReq, err := provider.BuildCancelSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/cancelsubscribe", cancelReq.URL.String())
	cancelBody := decodeJSONBodyAny(t, readRequestBody(t, cancelReq))
	require.Equal(t, "2582531", cancelBody["goodsid"])
}

func TestKakayunProductSubscriptionProviderParsesResponses(t *testing.T) {
	provider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)

	urls, message, err := provider.ParseGetReceiveURLsResponse(http.StatusOK, []byte(`{"code":1,"msg":"success","data":[{"url":"https://a.example.com/cb","createtime":1740139648}]}`))
	require.NoError(t, err)
	require.Equal(t, "success", message)
	require.Len(t, urls, 1)
	require.Equal(t, "https://a.example.com/cb", urls[0].URL)
	require.Equal(t, int64(1740139648), urls[0].CreatedAtUnix)

	message, err = provider.ParseMutationResponse(http.StatusOK, []byte(`{"code":1,"msg":"成功"}`))
	require.NoError(t, err)
	require.Equal(t, "成功", message)

	_, err = provider.ParseMutationResponse(http.StatusOK, []byte(`{"code":0,"msg":"签名错误"}`))
	require.Error(t, err)
}

func TestKakayunProductChangePushProviderVerifyAndParse(t *testing.T) {
	provider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)

	account := AccountConfig{ProviderCode: "kakayun", SecretKey: "secretXYZ"}
	payload := map[string]any{
		"goodsid":     2582531,
		"goodsprice":  "52.9901",
		"goodsstock":  985,
		"goodsstatus": 1,
		"goodstype":   1,
		"goodsname":   "API直充接口测试",
		"update_time": 1735002156,
		"timestamp":   1735002156,
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	raw := marshalJSONForTest(t, payload)

	result, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), raw)
	require.NoError(t, err)
	require.Equal(t, "2582531", result.SupplierGoodsNo)
	require.Equal(t, "API直充接口测试", result.GoodsName)
	require.Equal(t, "52.9901", result.GoodsPrice.StringFixed(4))
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "1", result.GoodsStatus)
	require.Contains(t, result.Raw, "API直充接口测试")
}

func TestKakayunProductChangePushProviderRejectsInvalidSignAndExpiredTimestamp(t *testing.T) {
	provider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "kakayun", SecretKey: "secretXYZ"}

	payload := map[string]any{
		"goodsid":    "2582531",
		"goodsprice": "52.9901",
		"goodsname":  "API直充接口测试",
		"timestamp":  strconv.FormatInt(time.Unix(1735002156, 0).Unix(), 10),
		"sign":       "bad-sign",
	}
	_, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)

	payload["sign"] = kakayunSign(payload, account.SecretKey)
	_, err = provider.ParseProductChangePush(account, time.Unix(1735003000, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)
}

func TestLookupProductPushOnlyRegistersKakayun(t *testing.T) {
	subscriptionProvider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", subscriptionProvider.Code())

	pushProvider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", pushProvider.Code())

	_, ok = LookupProductSubscription("kayixin")
	require.False(t, ok)
	_, ok = LookupProductChangePush("kayixin")
	require.False(t, ok)
}
```

Add this helper to the same test file:

```go
func marshalJSONForTest(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}
```

Remember to import `encoding/json`.

- [ ] **Step 2: Run provider tests and verify they fail**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider|TestLookupProductPush' -count=1 -timeout 60s
```

Expected: fail because provider interfaces and lookup functions do not exist.

- [ ] **Step 3: Add provider DTOs and interfaces**

Create `internal/library/supplierplatform/provider/product_push_types.go`:

```go
package supplierprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// ProductReceiveURLInput 描述设置供应商商品变动接收 URL 的请求参数。
type ProductReceiveURLInput struct {
	ReceiveURL    string
	OldReceiveURL string
}

// ProductReceiveURLItem 表示供应商已配置的商品变动接收 URL。
type ProductReceiveURLItem struct {
	URL           string
	CreatedAtUnix int64
}

// ProductSubscribeInput 描述订阅或取消订阅单个上游商品所需参数。
type ProductSubscribeInput struct {
	SupplierGoodsNo string
}

// ProductChangePushResult 表示供应商商品变动推送解析后的稳定数据。
type ProductChangePushResult struct {
	SupplierGoodsNo string
	GoodsName       string
	GoodsPrice      decimal.Decimal
	GoodsPriceValid bool
	GoodsStatus     string
	Raw             string
}

// ProductSubscriptionProvider 定义供应商商品推送订阅能力。
type ProductSubscriptionProvider interface {
	Code() string
	Name() string
	BuildGetReceiveURLsRequest(ctx context.Context, account AccountConfig, now time.Time) (*http.Request, error)
	ParseGetReceiveURLsResponse(statusCode int, body []byte) ([]ProductReceiveURLItem, string, error)
	BuildSetReceiveURLRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductReceiveURLInput) (*http.Request, error)
	BuildSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error)
	BuildCancelSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error)
	ParseMutationResponse(statusCode int, body []byte) (string, error)
}

// ProductChangePushProvider 定义供应商商品变动推送验签和解析能力。
type ProductChangePushProvider interface {
	Code() string
	Name() string
	ParseProductChangePush(account AccountConfig, now time.Time, body []byte) (ProductChangePushResult, error)
}
```

- [ ] **Step 4: Implement Kakayun provider methods**

Create `internal/library/supplierplatform/provider/kakayun_product_push.go`:

```go
package supplierprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const kakayunProductPushBaseURL = "http://public.kky.v3.api.kakayun.vip"
const kakayunPushTimestampSkew = 5 * time.Minute

func (kakayunProvider) BuildGetReceiveURLsRequest(ctx context.Context, account AccountConfig, now time.Time) (*http.Request, error) {
	payload := map[string]any{
		"userid":    strings.TrimSpace(account.TokenID),
		"timestamp": now.Unix(),
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	return newJSONRequest(ctx, kakayunProductPushBaseURL+"/dockapiv3/user/geturl", payload, map[string]string{"User-Agent": "curl/7.81.0"})
}

func (kakayunProvider) ParseGetReceiveURLsResponse(statusCode int, body []byte) ([]ProductReceiveURLItem, string, error) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, "", errors.New("卡卡云接收URL列表 HTTP 状态异常: " + strconv.Itoa(statusCode))
	}
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, "", err
	}
	if codeString(payload["code"]) != "1" {
		message := responseMessage(payload)
		if message == "" {
			message = "卡卡云接收URL列表查询失败"
		}
		return nil, message, errors.New(message)
	}
	items := make([]ProductReceiveURLItem, 0)
	if data, ok := payload["data"].([]any); ok {
		for _, raw := range data {
			itemMap, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			items = append(items, ProductReceiveURLItem{
				URL:           strings.TrimSpace(codeString(itemMap["url"])),
				CreatedAtUnix: int64FromValue(itemMap["createtime"]),
			})
		}
	}
	return items, responseMessage(payload), nil
}

func (kakayunProvider) BuildSetReceiveURLRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductReceiveURLInput) (*http.Request, error) {
	payload := map[string]any{
		"userid":    strings.TrimSpace(account.TokenID),
		"timestamp": now.Unix(),
	}
	if strings.TrimSpace(input.ReceiveURL) != "" {
		payload["receiveurl"] = strings.TrimSpace(input.ReceiveURL)
	}
	if strings.TrimSpace(input.OldReceiveURL) != "" {
		payload["oldreceiveurl"] = strings.TrimSpace(input.OldReceiveURL)
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	return newJSONRequest(ctx, kakayunProductPushBaseURL+"/dockapiv3/user/seturl", payload, map[string]string{"User-Agent": "curl/7.81.0"})
}

func (kakayunProvider) BuildSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error) {
	return kakayunProductSubscribeRequest(ctx, account, now, input, "/dockapiv3/goods/subscribe")
}

func (kakayunProvider) BuildCancelSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error) {
	return kakayunProductSubscribeRequest(ctx, account, now, input, "/dockapiv3/goods/cancelsubscribe")
}

func kakayunProductSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput, path string) (*http.Request, error) {
	payload := map[string]any{
		"userid":    strings.TrimSpace(account.TokenID),
		"timestamp": now.Unix(),
		"goodsid":   strings.TrimSpace(input.SupplierGoodsNo),
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	return newJSONRequest(ctx, kakayunProductPushBaseURL+path, payload, map[string]string{"User-Agent": "curl/7.81.0"})
}

func (kakayunProvider) ParseMutationResponse(statusCode int, body []byte) (string, error) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return "", errors.New("卡卡云订阅接口 HTTP 状态异常: " + strconv.Itoa(statusCode))
	}
	payload, err := decodeJSONMap(body)
	if err != nil {
		return "", err
	}
	message := responseMessage(payload)
	if codeString(payload["code"]) != "1" {
		if message == "" {
			message = "卡卡云订阅接口失败"
		}
		return message, errors.New(message)
	}
	if message == "" {
		message = "成功"
	}
	return message, nil
}

func (kakayunProvider) ParseProductChangePush(account AccountConfig, now time.Time, body []byte) (ProductChangePushResult, error) {
	raw := string(body)
	payload, err := decodeJSONMap(body)
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	timestamp, err := int64FromRequired(payload["timestamp"])
	if err != nil {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送缺少有效时间戳")
	}
	if delta := now.Sub(time.Unix(timestamp, 0)); delta > kakayunPushTimestampSkew || delta < -kakayunPushTimestampSkew {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送时间戳已过期")
	}
	expected := kakayunSign(payload, account.SecretKey)
	if !strings.EqualFold(expected, strings.TrimSpace(codeString(payload["sign"]))) {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送签名错误")
	}
	price, err := decimalFromValue(payload["goodsprice"])
	priceValid := err == nil && !price.IsNegative()
	if priceValid {
		price = price.Round(4)
	}
	return ProductChangePushResult{
		SupplierGoodsNo: codeString(payload["goodsid"]),
		GoodsName:       strings.TrimSpace(codeString(payload["goodsname"])),
		GoodsPrice:      price,
		GoodsPriceValid: priceValid,
		GoodsStatus:     codeString(payload["goodsstatus"]),
		Raw:             raw,
	}, nil
}

func int64FromValue(value any) int64 {
	result, _ := int64FromRequired(value)
	return result
}

func int64FromRequired(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	case string:
		return strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
	default:
		return 0, fmt.Errorf("字段不是有效整数")
	}
}
```

Add `encoding/json` to imports in this file for `json.Number`.

- [ ] **Step 5: Register provider lookups**

Modify `internal/library/supplierplatform/provider/registry.go`:

```go
var defaultProductSubscriptionRegistry = map[string]ProductSubscriptionProvider{
	"kakayun": kakayunProvider{},
}

var defaultProductChangePushRegistry = map[string]ProductChangePushProvider{
	"kakayun": kakayunProvider{},
}

// LookupProductSubscription 根据 provider_code 查找商品推送订阅适配器实现。
func LookupProductSubscription(code string) (ProductSubscriptionProvider, bool) {
	provider, ok := defaultProductSubscriptionRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}

// LookupProductChangePush 根据 provider_code 查找商品变动推送适配器实现。
func LookupProductChangePush(code string) (ProductChangePushProvider, bool) {
	provider, ok := defaultProductChangePushRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}
```

- [ ] **Step 6: Run provider tests**

Run:

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider|TestLookupProductPush' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 7: Commit provider work**

```bash
git add internal/library/supplierplatform/provider/product_push_types.go internal/library/supplierplatform/provider/kakayun_product_push.go internal/library/supplierplatform/provider/kakayun_product_push_test.go internal/library/supplierplatform/provider/registry.go
git commit -m "feat: add kakayun product push provider"
```

---

### Task 3: Shared Price Change Apply Logic

**Files:**
- Create: `internal/logic/admin/product_goods_channel_price_change.go`
- Create: `internal/logic/admin/product_goods_channel_price_change_test.go`
- Modify: `internal/logic/admin/product_goods_channel_sync.go`

- [ ] **Step 1: Write failing logic tests**

Create `internal/logic/admin/product_goods_channel_price_change_test.go` with tests that seed a channel binding, apply a push price, and verify both binding update and log insertion:

```go
package adminlogic

import (
	"context"
	"testing"

	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestApplyProductGoodsChannelPriceChangePushUpdatesPriceAndWritesLog(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 0, "qqlogin.yxp8.cn", 1, 0)
	logic := NewProductGoodsLogic(core)

	candidate := loadSinglePriceChangeCandidate(t, core, goodsID)
	result, err := logic.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
		SupplierGoodsNo: candidate.SupplierGoodsNo,
		GoodsName:       "推送后名称",
		GoodsPrice:      decimal.RequireFromString("12.0000"),
		GoodsPriceValid: true,
		Raw:             `{"goodsid":"SKU-100","goodsprice":"12.0000"}`,
	}, productGoodsChannelPriceChangeSourcePush)
	require.NoError(t, err)
	require.True(t, result.Updated)
	require.True(t, result.PriceChanged)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "12.0000", row.SourceCostPrice)
	require.Equal(t, "12.5400", row.CostPrice)

	var count int
	err = core.DB().GetCore().GetScan(ctx, &count, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE source = 'push' AND binding_id = ?`, candidate.BindingID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestApplyProductGoodsChannelPriceChangeDoesNotLogWhenPriceUnchanged(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)
	goodsID := seedProductGoodsSyncGoods(t, core, 1, 0, "qqlogin.yxp8.cn", 1, 0)
	logic := NewProductGoodsLogic(core)
	candidate := loadSinglePriceChangeCandidate(t, core, goodsID)

	result, err := logic.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
		SupplierGoodsNo: candidate.SupplierGoodsNo,
		GoodsPrice:      decimal.RequireFromString("10.0000"),
		GoodsPriceValid: true,
		Raw:             `{"goodsid":"SKU-100","goodsprice":"10.0000"}`,
	}, productGoodsChannelPriceChangeSourcePush)
	require.NoError(t, err)
	require.False(t, result.PriceChanged)

	var count int
	err = core.DB().GetCore().GetScan(ctx, &count, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE binding_id = ?`, candidate.BindingID)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
```

Import `myjob/internal/app`. Add `loadSinglePriceChangeCandidate` helper that selects the same fields required by the new candidate struct. Use seed helpers already present in `product_goods_channel_sync_test.go`.

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/logic/admin -run 'TestApplyProductGoodsChannelPriceChange' -count=1 -timeout 60s
```

Expected: fail because `applyProductGoodsChannelPriceChange` does not exist.

- [ ] **Step 3: Implement shared price-change apply method**

Create `internal/logic/admin/product_goods_channel_price_change.go` with:

```go
const (
	productGoodsChannelPriceChangeSourceMonitor = "monitor"
	productGoodsChannelPriceChangeSourcePush    = "push"
)

type productGoodsChannelPriceChangeApplyResult struct {
	Updated      bool
	PriceChanged bool
}
```

Define an enriched candidate struct, or extend `productGoodsChannelSyncCandidate`, so it includes:

```go
GoodsCode
GoodsName
GoodsIcon
DefaultSellPrice
PlatformAccountName
```

Implement:

```go
func (l *ProductGoodsLogic) applyProductGoodsChannelPriceChange(ctx context.Context, candidate productGoodsChannelSyncCandidate, info supplierprovider.ProductInfoResult, source string) (productGoodsChannelPriceChangeApplyResult, error)
```

The method must:

1. Reject mismatched `info.SupplierGoodsNo`.
2. Skip when `candidate.SyncCostPriceEnabled != 1`.
3. Compute new cost snapshot with `computeChannelCostSnapshot`.
4. Compute old/new effective sell price with `computeChannelEffectiveSellPrice`.
5. Compare normalized `source_cost_price` and `cost_price`.
6. Update `product_goods_channel_binding`.
7. Insert a log only if price changed.

Use this insert shape:

```go
_, err := tx.Exec(`
INSERT INTO product_goods_channel_price_change_log (
    source, provider_code, platform_account_id, platform_account_name, binding_id,
    goods_id, goods_code, goods_name, goods_icon, supplier_goods_no, supplier_goods_name,
    old_source_cost_price, new_source_cost_price, old_cost_price, new_cost_price,
    old_effective_sell_price, new_effective_sell_price, change_amount,
    description, raw_payload, changed_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, source, candidate.ProviderCode, candidate.PlatformAccountID, candidate.PlatformAccountName, candidate.BindingID,
candidate.GoodsID, candidate.GoodsCode, candidate.GoodsName, candidate.GoodsIcon, candidate.SupplierGoodsNo, newSupplierName,
candidate.CurrentSourceCostPrice, snapshot.SourceCostPrice, candidate.CurrentCostPrice, snapshot.CostPrice,
oldEffectiveSellPrice, newEffectiveSellPrice, changeAmount, description, info.Raw, now, now)
```

Build `description` with a deterministic Chinese string:

```go
fmt.Sprintf("来源:%s；货源:%s；上游商品:%s；进价:%s -> %s；比较成本:%s -> %s；利润后价格:%s -> %s",
	source, candidate.PlatformAccountName, candidate.SupplierGoodsNo,
	candidate.CurrentSourceCostPrice, snapshot.SourceCostPrice,
	candidate.CurrentCostPrice, snapshot.CostPrice,
	oldEffectiveSellPrice, newEffectiveSellPrice)
```

- [ ] **Step 4: Refactor monitor sync to use shared method**

Modify `internal/logic/admin/product_goods_channel_sync.go`:

- Extend `loadProductGoodsChannelSyncCandidates` select list with goods code/name/icon, default sell price, platform account name, and current cost price.
- Replace price update branch inside `applyProductGoodsChannelProductInfo` with a call to `applyProductGoodsChannelPriceChange(..., productGoodsChannelPriceChangeSourceMonitor)`.
- Keep name-only sync support for `sync_goods_name_enabled = 1` without price changes.
- Preserve existing tests in `product_goods_channel_sync_test.go`.

- [ ] **Step 5: Run logic tests**

Run:

```bash
go test ./internal/logic/admin -run 'TestApplyProductGoodsChannelPriceChange|TestSyncChannelBindingsOnce|TestSaveInventoryConfigTriggersImmediateSyncWhenSwitchEnabled' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 6: Commit price-change apply logic**

```bash
git add internal/logic/admin/product_goods_channel_price_change.go internal/logic/admin/product_goods_channel_price_change_test.go internal/logic/admin/product_goods_channel_sync.go
git commit -m "feat: record product channel price changes"
```

---

### Task 4: Subscription Logic And Auto-Subscribe On Binding Save

**Files:**
- Create: `internal/logic/admin/product_goods_channel_subscription.go`
- Create: `internal/logic/admin/product_goods_channel_subscription_test.go`
- Modify: `internal/logic/admin/product_goods_channel_write.go`

- [ ] **Step 1: Write subscription logic tests**

Create `internal/logic/admin/product_goods_channel_subscription_test.go`:

```go
package adminlogic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
	defer server.Close()

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL

	ctx := context.Background()
	binding := seedKakayunSubscriptionBinding(t, core)
	err = logic.autoSubscribeProductGoodsChannelBinding(ctx, binding, "https://public.example.com/api/open/supplier-platforms/kakayun/1/product-change-callback")
	require.NoError(t, err)

	require.Contains(t, requests, "/dockapiv3/user/geturl")
	require.Contains(t, requests, "/dockapiv3/user/seturl")
	require.Contains(t, requests, "/dockapiv3/goods/subscribe")

	var status string
	err = core.DB().GetCore().GetScan(ctx, &status, `SELECT status FROM supplier_product_subscription WHERE binding_id = ?`, binding.BindingID)
	require.NoError(t, err)
	require.Equal(t, supplierProductSubscriptionStatusSubscribed, status)
}

func TestAutoSubscribeKakayunBindingRecordsFailureWithoutReturningError(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"msg":"签名错误"}`))
	}))
	defer server.Close()

	logic := NewProductGoodsLogic(core)
	logic.httpClient = server.Client()
	logic.productPushBaseURL = server.URL

	binding := seedKakayunSubscriptionBinding(t, core)
	err = logic.autoSubscribeProductGoodsChannelBinding(context.Background(), binding, "https://public.example.com/api/open/supplier-platforms/kakayun/1/product-change-callback")
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
```

Import `myjob/internal/app`. Add seed helper in the test file that inserts subject, platform account with `provider_code='kakayun'`, goods, config, and binding.

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/logic/admin -run 'TestAutoSubscribeKakayunBinding' -count=1 -timeout 60s
```

Expected: fail because subscription logic does not exist.

- [ ] **Step 3: Implement subscription orchestration**

Create `internal/logic/admin/product_goods_channel_subscription.go` with constants:

```go
const (
	supplierProductSubscriptionStatusSubscribed = "subscribed"
	supplierProductSubscriptionStatusFailed     = "failed"
	supplierProductSubscriptionStatusCanceled   = "canceled"

	supplierProductSubscriptionActionSubscribe   = "subscribe"
	supplierProductSubscriptionActionResubscribe = "resubscribe"
	supplierProductSubscriptionActionCancel      = "cancel"
)
```

Add a field to `ProductGoodsLogic` in `product_goods.go`:

```go
productPushBaseURL string
```

Default empty means provider uses Kakayun public URL. Tests can override it.

Implement:

```go
type productGoodsChannelSubscriptionTarget struct {
	BindingID           int64
	GoodsID             int64
	SupplierGoodsNo     string
	SupplierGoodsName   string
	ProviderCode        string
	PlatformAccountID   int64
	PlatformAccountName string
	TokenID             string
	SecretKey           string
	ExtraConfig         string
}

func (l *ProductGoodsLogic) autoSubscribeProductGoodsChannelBinding(ctx context.Context, target productGoodsChannelSubscriptionTarget, callbackURL string) error
```

Rules:

- If `target.ProviderCode != "kakayun"`, return nil.
- If `callbackURL == ""`, write failed subscription record with reason “无法构造回调 URL” and return nil.
- Use `LookupProductSubscription`.
- Read URL list, set URL if missing, then subscribe product.
- Save request/response snapshots for the last attempted operation.
- Upsert `supplier_product_subscription`.

Add helper:

```go
func (l *ProductGoodsLogic) upsertSupplierProductSubscription(ctx context.Context, target productGoodsChannelSubscriptionTarget, status, action, callbackURL, lastError, requestSnapshot, responseSnapshot string, subscribedAt, canceledAt any) error
```

Use SQLite/MySQL upsert based on driver, following the project pattern in `product_goods_channel_config_write.go`.

- [ ] **Step 4: Build callback URL from request context**

In the same file add:

```go
func supplierProductCallbackURLFromContext(ctx context.Context, providerCode string, platformAccountID int64) (string, error)
```

Use `g.RequestFromCtx(ctx)`. Logic:

```go
proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
if proto == "" && r.TLS != nil {
	proto = "https"
}
if proto == "" {
	proto = "http"
}
host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
if host == "" {
	host = strings.TrimSpace(r.Host)
}
if host == "" {
	return "", errors.New("无法构造回调 URL")
}
return fmt.Sprintf("%s://%s/api/open/supplier-platforms/%s/%d/product-change-callback", proto, host, providerCode, platformAccountID), nil
```

The comment above this function must explain that deployment proxy headers decide the public URL because no system domain parameter is stored.

- [ ] **Step 5: Trigger auto-subscribe after create/update**

Modify `CreateChannelBinding` and `UpdateChannelBinding` in `internal/logic/admin/product_goods_channel_write.go`.

After successful DB mutation and operation log:

```go
if target, targetErr := l.loadProductGoodsChannelSubscriptionTarget(ctx, createdID); targetErr == nil {
	callbackURL, callbackErr := supplierProductCallbackURLFromContext(ctx, target.ProviderCode, target.PlatformAccountID)
	if callbackErr != nil {
		callbackURL = ""
	}
	_ = l.autoSubscribeProductGoodsChannelBinding(context.Background(), target, callbackURL)
}
```

For update use `req.BindingId`. Use `context.Background()` for outbound subscription so it is not canceled by GoFrame request cleanup. Do not return subscription errors to the caller.

- [ ] **Step 6: Run subscription tests and channel binding regression**

Run:

```bash
go test ./internal/logic/admin -run 'TestAutoSubscribeKakayunBinding|TestSyncChannelBindingsOnce' -count=1 -timeout 60s
go test ./test/contract -run TestProductGoodsChannelBindingFlows -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 7: Commit subscription logic**

```bash
git add internal/logic/admin/product_goods.go internal/logic/admin/product_goods_channel_subscription.go internal/logic/admin/product_goods_channel_subscription_test.go internal/logic/admin/product_goods_channel_write.go
git commit -m "feat: auto subscribe kakayun channel bindings"
```

---

### Task 5: Open Product Change Callback

**Files:**
- Create: `api/supplier_product_callback.go`
- Create: `internal/controller/open/supplier_product_callback.go`
- Modify: `internal/service/supplier_product_subscription.go`
- Modify: `internal/logic/admin/product_goods_channel_price_change.go`
- Modify: `internal/bootstrap/application.go`
- Test: `test/contract/supplier_product_subscription_contract_test.go`

- [ ] **Step 1: Write open callback contract test**

Create `test/contract/supplier_product_subscription_contract_test.go` with this first test:

```go
package contract_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenSupplierProductChangeCallbackReturnsPlainOK(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	h.saveFinanceTaxConfig(t, token, "4.5", "3.8")

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

	saveConfig := h.putJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":     0,
		"reorder_timeout_enabled":   0,
		"reorder_timeout_minutes":   0,
		"order_strategy":            "fixed_order",
		"sync_cost_price_enabled":   1,
		"sync_goods_name_enabled":   0,
		"allow_loss_sale_enabled":   0,
		"max_loss_amount":           "0.0000",
		"combo_goods_enabled":       0,
	}, token)
	require.Equal(t, 0, saveConfig.Code)

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
}
```

Add helpers to the same file:

```go
func (h *testHarness) createKakayunSupplierPlatformAccount(t *testing.T, token, name string, subjectID int64, hasTax int, tokenID, secretKey string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           "qqlogin.yxp8.cn",
		"backup_domain":    "",
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
	var data struct{ ID int64 `json:"id"` }
	require.NoError(t, json.Unmarshal(res.Data, &data))
	return data.ID
}
```

Add local `kakayunContractSign` using the same sorted non-empty parameter rule. Import `crypto/md5`, `encoding/hex`, `encoding/json`, `fmt`, `sort`, and `strings`.

- [ ] **Step 2: Run callback contract test and verify failure**

Run:

```bash
go test ./test/contract -run TestOpenSupplierProductChangeCallbackReturnsPlainOK -count=1 -timeout 60s
```

Expected: fail because route does not exist.

- [ ] **Step 3: Add API protocol**

Create `api/supplier_product_callback.go`:

```go
package api

import "github.com/gogf/gf/v2/frame/g"

// SupplierProductChangeCallbackReq 用于接收第三方平台商品信息变动推送。
type SupplierProductChangeCallbackReq struct {
	g.Meta            `path:"/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback" method:"post" tags:"开放回调" summary:"供应商商品变动回调" dc:"第三方平台商品价格等信息变动后的通用回调入口"`
	ProviderCode      string `json:"providerCode" in:"path" v:"required#平台编码不能为空" dc:"供应商适配器编码"`
	PlatformAccountID int64  `json:"platformAccountId" in:"path" v:"required#平台账号ID不能为空" dc:"平台账号ID"`
}

// SupplierProductChangeCallbackRes 表示供应商商品变动回调已处理。
type SupplierProductChangeCallbackRes struct{}
```

- [ ] **Step 4: Add service interface**

Create or modify `internal/service/supplier_product_subscription.go`:

```go
package service

import (
	"context"

	adminapi "myjob/api"
)

// SupplierProductCallbackService 定义开放供应商商品变动回调处理能力。
type SupplierProductCallbackService interface {
	HandleSupplierProductChangeCallback(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq, body []byte) error
}
```

- [ ] **Step 5: Add open controller that writes plain ok**

Create `internal/controller/open/supplier_product_callback.go`:

```go
package opencontroller

import (
	"context"
	"io"

	adminapi "myjob/api"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// SupplierProductCallbackController 提供供应商商品信息变动开放回调。
type SupplierProductCallbackController struct {
	svc service.SupplierProductCallbackService
}

// NewSupplierProductCallback 创建供应商商品变动回调控制器。
func NewSupplierProductCallback(svc service.SupplierProductCallbackService) *SupplierProductCallbackController {
	return &SupplierProductCallbackController{svc: svc}
}

// ProductChange 接收供应商商品变动推送，成功时按上游要求返回纯文本 ok。
func (c *SupplierProductCallbackController) ProductChange(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq) (*adminapi.SupplierProductChangeCallbackRes, error) {
	request := g.RequestFromCtx(ctx)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	if err := c.svc.HandleSupplierProductChangeCallback(ctx, req, body); err != nil {
		return nil, err
	}
	request.Response.Write("ok")
	return &adminapi.SupplierProductChangeCallbackRes{}, nil
}
```

- [ ] **Step 6: Implement callback logic**

In `internal/logic/admin/product_goods_channel_price_change.go` add:

```go
func (l *ProductGoodsLogic) HandleSupplierProductChangeCallback(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq, body []byte) error
```

Flow:

1. Load platform account by `req.PlatformAccountID`.
2. Reject if deleted, disabled, or provider mismatch.
3. Lookup product change push provider.
4. Parse and verify push with `AccountConfig`.
5. Load all active bindings for `platform_account_id + supplier_goods_no`.
6. For each binding with `sync_cost_price_enabled = 1`, call `applyProductGoodsChannelPriceChange(..., sourcePush)`.
7. If no bindings or switches closed, return nil so Kakayun receives `ok`.
8. If a targeted DB update fails, return an error so Kakayun retries.

- [ ] **Step 7: Wire open controller in bootstrap**

Modify `internal/bootstrap/application.go`:

```go
supplierProductCallbackCtrl := opencontroller.NewSupplierProductCallback(services.ProductGoodsLogic)
```

Bind under `/api/open`:

```go
group.Bind(supplierProductCallbackCtrl)
```

- [ ] **Step 8: Run callback tests**

Run:

```bash
go test ./test/contract -run TestOpenSupplierProductChangeCallbackReturnsPlainOK -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 9: Commit callback work**

```bash
git add api/supplier_product_callback.go internal/controller/open/supplier_product_callback.go internal/service/supplier_product_subscription.go internal/logic/admin/product_goods_channel_price_change.go internal/bootstrap/application.go test/contract/supplier_product_subscription_contract_test.go
git commit -m "feat: handle supplier product change callbacks"
```

---

### Task 6: Admin Subscription And Price Change APIs

**Files:**
- Create: `api/supplier_product_subscription.go`
- Create: `api/product_goods_channel_price_change.go`
- Create: `internal/controller/admin/supplier_product_subscription.go`
- Create: `internal/controller/admin/product_goods_channel_price_change.go`
- Modify: `internal/service/supplier_product_subscription.go`
- Modify: `internal/logic/admin/common.go`
- Modify: `internal/logic/admin/product_goods_channel_subscription.go`
- Modify: `internal/logic/admin/product_goods_channel_price_change.go`
- Modify: `internal/bootstrap/application.go`
- Test: `test/contract/supplier_product_subscription_contract_test.go`
- Test: `test/contract/product_goods_channel_price_change_contract_test.go`

- [ ] **Step 1: Write admin contract tests**

Extend `test/contract/supplier_product_subscription_contract_test.go`:

```go
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

	resubscribe := h.postJSON("/api/admin/supplier-product-subscriptions/"+int64ToString(subscriptionID)+"/resubscribe", map[string]any{}, token)
	require.Equal(t, 0, resubscribe.Code)
}
```

Create `test/contract/product_goods_channel_price_change_contract_test.go`:

```go
package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductGoodsChannelPriceChangeList(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	h.seedProductGoodsChannelPriceChange(t, "push", "PRICE-CHANGE-001", "2582531")

	list := h.getJSON("/api/admin/product-goods-channel-price-changes?page=1&page_size=20&source=push&keyword=PRICE-CHANGE-001", token)
	require.Equal(t, 0, list.Code)
	require.Contains(t, string(list.Data), "PRICE-CHANGE-001")
	require.Contains(t, string(list.Data), "2582531")
	require.Contains(t, string(list.Data), "push")
	require.Contains(t, string(list.Data), "变动前")
}
```

Add seed helpers in those test files using `h.app.Core().DB()` if available. If `Core()` is not exposed, add this method to `bootstrap.Application`:

```go
// Core 返回应用运行时核心，仅用于测试和装配验证。
func (a *Application) Core() *app.Core { return a.core }
```

- [ ] **Step 2: Run contract tests and verify failure**

Run:

```bash
go test ./test/contract -run 'TestSupplierProductSubscriptionListCancelAndResubscribe|TestProductGoodsChannelPriceChangeList' -count=1 -timeout 60s
```

Expected: fail because API routes do not exist.

- [ ] **Step 3: Add API protocol files**

Create `api/supplier_product_subscription.go` with exported comments for every type:

```go
package api

import "github.com/gogf/gf/v2/frame/g"

// SupplierProductSubscriptionListReq 用于分页查询供应商商品推送订阅记录。
type SupplierProductSubscriptionListReq struct {
	g.Meta          `path:"/supplier-product-subscriptions" method:"get" tags:"第三方对接" summary:"商品订阅记录" security:"BearerAuth" dc:"分页查询供应商商品推送订阅记录"`
	Page            int    `json:"page" dc:"页码"`
	PageSize        int    `json:"page_size" dc:"每页条数"`
	Keyword         string `json:"keyword" dc:"商品名称关键词"`
	SupplierGoodsNo string `json:"supplier_goods_no" dc:"上游商品编号"`
	PlatformID      int64  `json:"platform_id" dc:"平台账号ID"`
	Status          string `json:"status" dc:"订阅状态"`
	StartAt         string `json:"start_at" dc:"开始时间"`
	EndAt           string `json:"end_at" dc:"结束时间"`
}

// SupplierProductSubscriptionListRes 返回供应商商品推送订阅记录列表。
type SupplierProductSubscriptionListRes struct {
	List       []SupplierProductSubscriptionItem `json:"list" dc:"订阅记录列表"`
	Pagination PaginationRes                     `json:"pagination" dc:"分页信息"`
}

// SupplierProductSubscriptionItem 是订阅记录列表单行数据。
type SupplierProductSubscriptionItem struct {
	ID                  int64  `json:"id" dc:"订阅记录ID"`
	GoodsID             int64  `json:"goods_id" dc:"本地商品ID"`
	GoodsName           string `json:"goods_name" dc:"商品名称"`
	GoodsIcon           string `json:"goods_icon" dc:"商品图标"`
	ProviderCode        string `json:"provider_code" dc:"供应商编码"`
	PlatformAccountID   int64  `json:"platform_account_id" dc:"平台账号ID"`
	PlatformAccountName string `json:"platform_account_name" dc:"平台账号名称"`
	SupplierGoodsNo     string `json:"supplier_goods_no" dc:"上游商品编号"`
	SupplierGoodsName   string `json:"supplier_goods_name" dc:"上游商品名称"`
	CallbackURL         string `json:"callback_url" dc:"回调地址"`
	Status              string `json:"status" dc:"订阅状态"`
	LastAction          string `json:"last_action" dc:"最近动作"`
	LastError           string `json:"last_error" dc:"最近失败原因"`
	SubscribedAt        string `json:"subscribed_at" dc:"订阅时间"`
	CanceledAt          string `json:"canceled_at" dc:"取消时间"`
	UpdatedAt           string `json:"updated_at" dc:"更新时间"`
}

// SupplierProductSubscriptionCancelReq 用于取消单条供应商商品订阅。
type SupplierProductSubscriptionCancelReq struct {
	g.Meta `path:"/supplier-product-subscriptions/{id}/cancel" method:"post" tags:"第三方对接" summary:"取消商品订阅" security:"BearerAuth" dc:"取消指定供应商商品推送订阅"`
	ID     int64 `json:"id" in:"path" v:"required#订阅记录ID不能为空" dc:"订阅记录ID"`
}

// SupplierProductSubscriptionCancelRes 表示取消订阅成功。
type SupplierProductSubscriptionCancelRes struct{}

// SupplierProductSubscriptionResubscribeReq 用于重新订阅单条供应商商品。
type SupplierProductSubscriptionResubscribeReq struct {
	g.Meta `path:"/supplier-product-subscriptions/{id}/resubscribe" method:"post" tags:"第三方对接" summary:"重新订阅商品" security:"BearerAuth" dc:"重新订阅指定供应商商品推送"`
	ID     int64 `json:"id" in:"path" v:"required#订阅记录ID不能为空" dc:"订阅记录ID"`
}

// SupplierProductSubscriptionResubscribeRes 表示重新订阅成功。
type SupplierProductSubscriptionResubscribeRes struct{}
```

Create `api/product_goods_channel_price_change.go` with analogous comments:

```go
// ProductGoodsChannelPriceChangeListReq 用于分页查询商品渠道自动改价记录。
type ProductGoodsChannelPriceChangeListReq struct {
	g.Meta          `path:"/product-goods-channel-price-changes" method:"get" tags:"商品管理" summary:"自动改价记录" security:"BearerAuth" dc:"分页查询监控或推送触发的商品渠道改价记录"`
	Page            int    `json:"page" dc:"页码"`
	PageSize        int    `json:"page_size" dc:"每页条数"`
	Source          string `json:"source" dc:"来源类型"`
	Keyword         string `json:"keyword" dc:"本地商品编号或名称"`
	SupplierGoodsNo string `json:"supplier_goods_no" dc:"上游商品编号"`
	PlatformID      int64  `json:"platform_id" dc:"平台账号ID"`
	StartAt         string `json:"start_at" dc:"开始时间"`
	EndAt           string `json:"end_at" dc:"结束时间"`
}
```

Define `ProductGoodsChannelPriceChangeListRes` and `ProductGoodsChannelPriceChangeItem` with all fields from the spec.

- [ ] **Step 4: Add controllers and services**

Create thin controllers mirroring existing style:

```go
// SupplierProductSubscriptionController 提供供应商商品订阅记录相关 HTTP handler。
type SupplierProductSubscriptionController struct {
	svc service.SupplierProductSubscriptionService
}
```

Methods:

```go
List
Cancel
Resubscribe
```

Create price-change controller:

```go
// ProductGoodsChannelPriceChangeController 提供商品渠道改价记录查询 HTTP handler。
type ProductGoodsChannelPriceChangeController struct {
	svc service.ProductGoodsChannelPriceChangeService
}
```

Add service interfaces:

```go
type SupplierProductSubscriptionService interface {
	ListSupplierProductSubscriptions(ctx context.Context, req *adminapi.SupplierProductSubscriptionListReq) (*adminapi.SupplierProductSubscriptionListRes, error)
	CancelSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionCancelReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionCancelRes, error)
	ResubscribeSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionResubscribeReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionResubscribeRes, error)
}

type ProductGoodsChannelPriceChangeService interface {
	ListProductGoodsChannelPriceChanges(ctx context.Context, req *adminapi.ProductGoodsChannelPriceChangeListReq) (*adminapi.ProductGoodsChannelPriceChangeListRes, error)
}
```

- [ ] **Step 5: Implement list/cancel/resubscribe logic**

In `product_goods_channel_subscription.go` implement:

```go
func (l *ProductGoodsLogic) ListSupplierProductSubscriptions(ctx context.Context, req *adminapi.SupplierProductSubscriptionListReq) (*adminapi.SupplierProductSubscriptionListRes, error)
func (l *ProductGoodsLogic) CancelSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionCancelReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionCancelRes, error)
func (l *ProductGoodsLogic) ResubscribeSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionResubscribeReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionResubscribeRes, error)
```

List query must left join `product_goods` and filter by keyword, supplier goods no, platform id, status, and time range.

Cancel and resubscribe must:

- Load subscription record.
- Load platform account and provider.
- Call provider mutation.
- Update local status and snapshots.
- Write operation log.

In `product_goods_channel_price_change.go` implement:

```go
func (l *ProductGoodsLogic) ListProductGoodsChannelPriceChanges(ctx context.Context, req *adminapi.ProductGoodsChannelPriceChangeListReq) (*adminapi.ProductGoodsChannelPriceChangeListRes, error)
```

- [ ] **Step 6: Wire services and routes**

Modify `internal/logic/admin/common.go`:

```go
SupplierProductSubscription service.SupplierProductSubscriptionService
ProductGoodsChannelPriceChange service.ProductGoodsChannelPriceChangeService
SupplierProductCallback service.SupplierProductCallbackService
```

Assign all three to `productGoods` in `NewServices`.

Modify `internal/bootstrap/application.go`:

```go
subscriptionCtrl := admincontroller.NewSupplierProductSubscription(services.SupplierProductSubscription)
priceChangeCtrl := admincontroller.NewProductGoodsChannelPriceChange(services.ProductGoodsChannelPriceChange)
```

Bind subscription controller under `supplier.index`; bind price-change controller under `product.goods`.

- [ ] **Step 7: Run admin API tests**

Run:

```bash
go test ./test/contract -run 'TestSupplierProductSubscriptionListCancelAndResubscribe|TestProductGoodsChannelPriceChangeList' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 8: Commit admin APIs**

```bash
git add api/supplier_product_subscription.go api/product_goods_channel_price_change.go internal/controller/admin/supplier_product_subscription.go internal/controller/admin/product_goods_channel_price_change.go internal/service/supplier_product_subscription.go internal/logic/admin/common.go internal/logic/admin/product_goods_channel_subscription.go internal/logic/admin/product_goods_channel_price_change.go internal/bootstrap/application.go test/contract/supplier_product_subscription_contract_test.go test/contract/product_goods_channel_price_change_contract_test.go
git commit -m "feat: add supplier subscription and price change APIs"
```

---

### Task 7: API Layout, Docs, And Focused Verification

**Files:**
- Modify: `test/contract/api_layout_test.go`
- Modify: `docs/module-map.md`
- Modify: `docs/development.md`
- Modify: `docs/testing.md`
- Modify: `docs/superpowers/README.md`

- [ ] **Step 1: Update API layout contract**

Modify expected API file list in `test/contract/api_layout_test.go`:

```go
"product_goods_channel_price_change.go",
"supplier_product_callback.go",
"supplier_product_subscription.go",
```

Keep `api/` flat. Do not add subdirectories.

- [ ] **Step 2: Update module map**

In `docs/module-map.md`:

- 商品管理主要能力 add: `自动改价记录、卡卡云推送改价记录`。
- 第三方对接主要能力 add: `卡卡云商品订阅、取消订阅、重新订阅、商品变动推送回调`。
- 路由摘要 add:

```markdown
| 商品渠道改价记录 | `/api/admin/product-goods-channel-price-changes` | `product.goods` |
| 供应商商品订阅 | `/api/admin/supplier-product-subscriptions*` | `supplier.index` |
| 供应商商品变动回调 | `/api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback` | 上游签名 |
```

- [ ] **Step 3: Update development docs**

In `docs/development.md`, add a subsection under supplier provider development:

```markdown
### 供应商商品推送订阅

商品变动推送 provider 需要实现 `ProductSubscriptionProvider` 和 `ProductChangePushProvider`。回调 URL 统一使用 `/api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback`，通过平台账号 ID 找密钥验签。新增渠道绑定后的订阅失败只能写入 `supplier_product_subscription`，不得阻断本地绑定保存。
```

- [ ] **Step 4: Update testing docs**

In `docs/testing.md`, add focused commands:

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider|TestLookupProductPush' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestAutoSubscribeKakayunBinding|TestApplyProductGoodsChannelPriceChange' -count=1 -timeout 60s
go test ./test/contract -run 'TestOpenSupplierProductChangeCallbackReturnsPlainOK|TestSupplierProductSubscriptionListCancelAndResubscribe|TestProductGoodsChannelPriceChangeList' -count=1 -timeout 60s
```

- [ ] **Step 5: Update superpowers index**

In `docs/superpowers/README.md`, add:

```markdown
- `plans/2026-04-26-kakayun-product-push.md`：卡卡云商品价格推送、订阅记录和改价记录实施计划。
```

- [ ] **Step 6: Run focused tests**

Run:

```bash
go test ./internal/app -run 'TestEnsureSupplierProductPushSchemaCreatesTables|TestMySQLSchemaIncludesTableAndColumnComments|TestManifestMySQLSchemaFilesIncludeTableAndColumnComments' -count=1 -timeout 60s
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider|TestLookupProductPush' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestAutoSubscribeKakayunBinding|TestApplyProductGoodsChannelPriceChange|TestSyncChannelBindingsOnce' -count=1 -timeout 60s
go test ./test/contract -run 'TestAPIProtocolLayout|TestOpenSupplierProductChangeCallbackReturnsPlainOK|TestSupplierProductSubscriptionListCancelAndResubscribe|TestProductGoodsChannelPriceChangeList' -count=1 -timeout 60s
```

Expected: all pass.

- [ ] **Step 7: Commit docs and layout**

```bash
git add test/contract/api_layout_test.go docs/module-map.md docs/development.md docs/testing.md docs/superpowers/README.md
git commit -m "docs: document kakayun product push"
```

---

### Task 8: Full Verification

**Files:**
- No new files.

- [ ] **Step 1: Run all tests**

```bash
go test ./... -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 2: Run build**

```bash
go build ./...
```

Expected: pass.

- [ ] **Step 3: Run lint**

```bash
golangci-lint run --timeout=5m
```

Expected: pass.

- [ ] **Step 4: Inspect git diff**

```bash
git status --short
git diff --stat
```

Expected: only intended files changed, no generated or unrelated churn.

- [ ] **Step 5: Final commit if verification changed docs or fixes**

If verification required small fixes:

```bash
git add <changed-files>
git commit -m "test: verify kakayun product push"
```

If no fixes were needed, do not create an empty commit.

---

## Self-Review Checklist

- [ ] Spec requirement “通用 URL 使用 providerCode + platformAccountId” is implemented in Task 5.
- [ ] Spec requirement “不新增公网域名参数” is preserved in Task 4 callback URL construction.
- [ ] Spec requirement “只在同步进价开启时同步进价” is covered in Task 3 and Task 5.
- [ ] Spec requirement “不改 default_sell_price” is preserved by only updating channel binding cost fields.
- [ ] Spec requirement “只记录真实价格变化” is covered in Task 3 tests.
- [ ] Spec requirement “推送和监控都记录” is covered by Task 3 monitor refactor and Task 5 push path.
- [ ] Spec requirement “新增/编辑卡卡云绑定自动订阅，历史不补” is covered in Task 4.
- [ ] Spec requirement “订阅失败不阻断保存” is covered in Task 4 tests.
- [ ] Spec requirement “取消订阅保留本地记录” is covered in Task 6.
- [ ] Spec requirement “列表接口查询订阅和改价记录” is covered in Task 6.
- [ ] API files stay flat under `api/`, covered in Task 7.
- [ ] Final verification commands match project AGENTS requirements.
