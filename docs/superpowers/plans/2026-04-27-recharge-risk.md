# Recharge Risk Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build backend recharge-account risk controls that let admins configure blocking rules and make open-order creation fail locally before any upstream submission.

**Architecture:** Add a focused admin domain for recharge risk rule CRUD and risk records, with its own API, controller, service, logic, entities, schema, menu permission, tests, and docs. Keep order fulfillment logic responsible only for checking active rules during open-order creation and writing an atomic failed order plus risk record when a rule matches.

**Tech Stack:** Go, GoFrame HTTP binding, GoFrame DB (`gdb`), MySQL production schema, SQLite/MySQL test bootstrap schema, existing `code/message/data` response middleware, existing contract and integration test harnesses.

---

## File Structure

- Create `api/recharge_risk.go`: flat protocol file for rule management and record list requests/responses.
- Create `internal/model/entity/recharge_risk.go`: database row structs and list item structs for the recharge risk domain.
- Create `internal/service/recharge_risk.go`: service interface for admin recharge risk operations.
- Create `internal/controller/admin/recharge_risk.go`: thin HTTP handlers with actor/IP forwarding.
- Create `internal/logic/admin/recharge_risk.go`: `RechargeRiskLogic` declaration.
- Create `internal/logic/admin/recharge_risk_query.go`: rule list and record list queries.
- Create `internal/logic/admin/recharge_risk_write.go`: create, update, status, and soft-delete writes.
- Create `internal/logic/admin/recharge_risk_validate.go`: input normalization, status parsing, uniqueness helpers.
- Create `internal/logic/admin/recharge_risk_mapper.go`: record-to-response mapping.
- Create `internal/app/recharge_risk_schema.go`: idempotent startup schema creation for both tables.
- Create `internal/logic/order/order_risk.go`: open-order risk matching and atomic failed-order record creation.
- Modify `internal/app/schema.go`: include both new tables in SQLite and MySQL embedded schemas.
- Modify `manifest/sql/008_external_order.sql`: add MySQL DDL for recharge risk tables near order tables.
- Modify `manifest/sql/002_seed_menu.sql`: add `order.recharge_risk` permission and default group grant.
- Modify `internal/app/bootstrap.go`: call `ensureRechargeRiskSchema`, seed menu ID `15`, grant default group, and keep comments accurate.
- Modify `internal/logic/admin/common.go`: add `RechargeRisk` service to `Services` and construct `&RechargeRiskLogic{core: core}`.
- Modify `internal/bootstrap/application.go`: instantiate and bind `RechargeRiskController` under `order.recharge_risk`.
- Modify `internal/logic/order/order_create.go`: check risk after basic validation and before candidate-channel loading.
- Modify `test/contract/api_layout_test.go`: require `api/recharge_risk.go`.
- Create `test/contract/recharge_risk_contract_test.go`: OpenAPI, permission, rule CRUD/list, record list coverage.
- Modify `test/integration/order_worker_test.go`: add open-order risk interception integration tests.
- Modify docs: `README.md` if module index changes, `docs/module-map.md`, `docs/development.md`, `docs/testing.md`, `test/contract/README.md`, `docs/superpowers/README.md`.

## Task 1: Schema And Entities

**Files:**
- Create: `internal/model/entity/recharge_risk.go`
- Create: `internal/app/recharge_risk_schema.go`
- Modify: `internal/app/schema.go`
- Modify: `manifest/sql/008_external_order.sql`
- Modify: `internal/app/bootstrap.go`
- Test: `internal/app/order_schema_test.go`

- [ ] **Step 1: Write the failing schema test**

Add these assertions to `TestExternalOrderSchemaContainsRequiredTablesAndIndexes` in `internal/app/order_schema_test.go`:

```go
require.Contains(t, schema, "recharge_risk_rule")
require.Contains(t, schema, "recharge_risk_record")
require.Contains(t, schema, "uk_recharge_risk_rule_active")
require.Contains(t, schema, "idx_recharge_risk_rule_match")
require.Contains(t, schema, "idx_recharge_risk_record_account")
require.Contains(t, schema, "idx_recharge_risk_record_keyword")
```

Add this new test in the same file:

```go
func TestRechargeRiskSchemaContainsComments(t *testing.T) {
	require.Contains(t, mysqlSchema, "COMMENT='充值账号风控规则表'")
	require.Contains(t, mysqlSchema, "COMMENT='充值账号风控拦截记录表'")
	for _, column := range []string{"充值账号", "商品名关键词", "风控原因", "累计拦截次数", "拦截时间"} {
		require.Contains(t, mysqlSchema, column)
	}
}
```

- [ ] **Step 2: Run schema tests to verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(ExternalOrderSchemaContainsRequiredTablesAndIndexes|RechargeRiskSchemaContainsComments)' -count=1 -timeout 60s
```

Expected: fail because the new tables and indexes are not in the embedded schemas yet.

- [ ] **Step 3: Add entity structs**

Create `internal/model/entity/recharge_risk.go`:

```go
package entity

import (
	"database/sql"
	"time"
)

// RechargeRiskRule 表示一条充值账号风控规则。
type RechargeRiskRule struct {
	ID             int64      `db:"id" json:"id"`
	Account        string     `db:"account" json:"account"`
	GoodsKeyword   string     `db:"goods_keyword" json:"goods_keyword"`
	Reason         string     `db:"reason" json:"reason"`
	Status         int        `db:"status" json:"status"`
	HitCount       int        `db:"hit_count" json:"hit_count"`
	CreatedByID    int64      `db:"created_by_id" json:"created_by_id"`
	CreatedByName  string     `db:"created_by_name" json:"created_by_name"`
	UpdatedByID    int64      `db:"updated_by_id" json:"updated_by_id"`
	UpdatedByName  string     `db:"updated_by_name" json:"updated_by_name"`
	IsDeleted      int        `db:"is_deleted" json:"is_deleted"`
	DeletedAt      sql.NullTime `db:"deleted_at" json:"deleted_at"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// RechargeRiskRecord 表示一次开放下单被本地风控拦截的流水。
type RechargeRiskRecord struct {
	ID                 int64     `db:"id" json:"id"`
	RuleID             int64     `db:"rule_id" json:"rule_id"`
	OrderID            int64     `db:"order_id" json:"order_id"`
	OrderNo            string    `db:"order_no" json:"order_no"`
	Account            string    `db:"account" json:"account"`
	GoodsID            int64     `db:"goods_id" json:"goods_id"`
	GoodsCode          string    `db:"goods_code" json:"goods_code"`
	GoodsName          string    `db:"goods_name" json:"goods_name"`
	MatchedKeyword     string    `db:"matched_keyword" json:"matched_keyword"`
	Reason             string    `db:"reason" json:"reason"`
	RequestTokenMasked string    `db:"request_token_masked" json:"request_token_masked"`
	InterceptedAt      time.Time `db:"intercepted_at" json:"intercepted_at"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
```

- [ ] **Step 4: Add startup schema helper**

Create `internal/app/recharge_risk_schema.go`:

```go
package app

import (
	"context"
	"database/sql"
)

// ensureRechargeRiskSchema 确保充值风控规则和拦截流水表存在。
func (c *Core) ensureRechargeRiskSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		return execStatements(ctx, func(sqlText string, args ...any) (sql.Result, error) {
			return c.DB().Exec(ctx, sqlText, args...)
		}, sqliteRechargeRiskSchema)
	}
	return execStatements(ctx, func(sqlText string, args ...any) (sql.Result, error) {
		return c.DB().Exec(ctx, sqlText, args...)
	}, mysqlRechargeRiskSchema)
}
```

- [ ] **Step 5: Add schema constants**

Add these constants near `schema.go` or inside `internal/app/recharge_risk_schema.go`:

```go
const sqliteRechargeRiskSchema = `
CREATE TABLE IF NOT EXISTS recharge_risk_rule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account TEXT NOT NULL,
    goods_keyword TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    hit_count INTEGER NOT NULL DEFAULT 0,
    created_by_id INTEGER NOT NULL DEFAULT 0,
    created_by_name TEXT NOT NULL DEFAULT '',
    updated_by_id INTEGER NOT NULL DEFAULT 0,
    updated_by_name TEXT NOT NULL DEFAULT '',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_recharge_risk_rule_active
    ON recharge_risk_rule(account, goods_keyword, is_deleted);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_match
    ON recharge_risk_rule(account, status, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_keyword
    ON recharge_risk_rule(goods_keyword, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_status
    ON recharge_risk_rule(status, is_deleted, updated_at);
CREATE TABLE IF NOT EXISTS recharge_risk_record (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL,
    account TEXT NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL,
    goods_name TEXT NOT NULL,
    matched_keyword TEXT NOT NULL,
    reason TEXT NOT NULL,
    request_token_masked TEXT NOT NULL DEFAULT '',
    intercepted_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_account
    ON recharge_risk_record(account, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_keyword
    ON recharge_risk_record(matched_keyword, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_rule
    ON recharge_risk_record(rule_id, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_order
    ON recharge_risk_record(order_no);
`

const mysqlRechargeRiskSchema = `
CREATE TABLE IF NOT EXISTS recharge_risk_rule (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '风控规则ID',
    account VARCHAR(255) NOT NULL COMMENT '充值账号',
    goods_keyword VARCHAR(255) NOT NULL COMMENT '商品名关键词',
    reason VARCHAR(512) NOT NULL DEFAULT '' COMMENT '风控原因',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用，0停用',
    hit_count INT NOT NULL DEFAULT 0 COMMENT '累计拦截次数',
    created_by_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID快照',
    created_by_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人名称快照',
    updated_by_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新人ID快照',
    updated_by_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人名称快照',
    is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    deleted_at DATETIME NULL COMMENT '删除时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_recharge_risk_rule_active (account, goods_keyword, is_deleted),
    KEY idx_recharge_risk_rule_match (account, status, is_deleted, id),
    KEY idx_recharge_risk_rule_keyword (goods_keyword, is_deleted, id),
    KEY idx_recharge_risk_rule_status (status, is_deleted, updated_at)
) COMMENT='充值账号风控规则表';
CREATE TABLE IF NOT EXISTS recharge_risk_record (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '拦截记录ID',
    rule_id BIGINT UNSIGNED NOT NULL COMMENT '命中规则ID',
    order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    order_no VARCHAR(40) NOT NULL COMMENT '订单号',
    account VARCHAR(255) NOT NULL COMMENT '充值账号',
    goods_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID快照',
    goods_code VARCHAR(32) NOT NULL COMMENT '商品编码快照',
    goods_name VARCHAR(255) NOT NULL COMMENT '商品名称快照',
    matched_keyword VARCHAR(255) NOT NULL COMMENT '命中关键词快照',
    reason VARCHAR(512) NOT NULL COMMENT '风控原因快照',
    request_token_masked VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开放下单token脱敏快照',
    intercepted_at DATETIME NOT NULL COMMENT '拦截时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    KEY idx_recharge_risk_record_account (account, intercepted_at, id),
    KEY idx_recharge_risk_record_keyword (matched_keyword, intercepted_at, id),
    KEY idx_recharge_risk_record_rule (rule_id, intercepted_at, id),
    KEY idx_recharge_risk_record_order (order_no)
) COMMENT='充值账号风控拦截记录表';
`
```

- [ ] **Step 6: Wire startup schema**

In `internal/app/bootstrap.go`, call the helper after `ensureExternalOrderAttemptSchema(ctx)`:

```go
if err := c.ensureRechargeRiskSchema(ctx); err != nil {
	return err
}
```

- [ ] **Step 7: Add DDL to embedded and manifest schemas**

Append the SQLite recharge risk DDL before `system_config` in the `sqliteSchema` string in `internal/app/schema.go`.

Append the MySQL recharge risk DDL before `system_config` in the `mysqlSchema` string in `internal/app/schema.go`.

Append the MySQL recharge risk DDL to `manifest/sql/008_external_order.sql` after `external_order_attempt`.

- [ ] **Step 8: Run schema tests**

Run:

```bash
go test ./internal/app -run 'Test(ExternalOrderSchemaContainsRequiredTablesAndIndexes|RechargeRiskSchemaContainsComments|ManifestMySQLSchemaFilesIncludeTableAndColumnComments)' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 9: Commit schema work**

```bash
git add internal/model/entity/recharge_risk.go internal/app/recharge_risk_schema.go internal/app/schema.go internal/app/bootstrap.go internal/app/order_schema_test.go manifest/sql/008_external_order.sql
git commit -m "feat: add recharge risk schema"
```

## Task 2: API, Permission, Service, Controller, Wiring

**Files:**
- Create: `api/recharge_risk.go`
- Create: `internal/service/recharge_risk.go`
- Create: `internal/controller/admin/recharge_risk.go`
- Create: `internal/logic/admin/recharge_risk.go`
- Modify: `internal/logic/admin/common.go`
- Modify: `internal/bootstrap/application.go`
- Modify: `internal/app/bootstrap.go`
- Modify: `manifest/sql/002_seed_menu.sql`
- Modify: `test/contract/api_layout_test.go`
- Test: `test/contract/recharge_risk_contract_test.go`

- [ ] **Step 1: Write failing contract tests for route exposure and permission**

Create `test/contract/recharge_risk_contract_test.go`:

```go
package contract_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_RechargeRiskPathsExposed(t *testing.T) {
	h := newTestHarness(t)
	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/recharge-risks/rules")
	require.Contains(t, res.body, "/api/admin/recharge-risks/records")
}

func TestRechargeRiskPermissionSeeded(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	me := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, me.Code)
	var data struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(me.Data, &data))
	require.Contains(t, data.Permissions, "order.recharge_risk")
}

func TestRechargeRiskRequiresPermission(t *testing.T) {
	h := newTestHarness(t)
	limitedToken := h.createLimitedUserToken(t, h.loginAdmin(t), 0)
	res := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20", limitedToken)
	require.Equal(t, 403, res.Code)
}
```

Add `"recharge_risk.go"` to the expected filenames in `test/contract/api_layout_test.go`.

- [ ] **Step 2: Run contract tests to verify they fail**

Run:

```bash
go test ./test/contract -run 'Test(OpenAPI_RechargeRiskPathsExposed|RechargeRiskPermissionSeeded|RechargeRiskRequiresPermission|APIProtocolLayout)' -count=1 -timeout 60s
```

Expected: fail because the API file, routes, and permission do not exist yet.

- [ ] **Step 3: Add API protocol**

Create `api/recharge_risk.go`:

```go
package api

import "github.com/gogf/gf/v2/frame/g"

// RechargeRiskRuleListReq 用于分页查询充值账号风控规则。
type RechargeRiskRuleListReq struct {
	g.Meta       `path:"/recharge-risks/rules" method:"get" tags:"充值风控" summary:"风控规则列表" security:"BearerAuth" dc:"分页查询充值账号风控规则"`
	Page         int    `json:"page" dc:"页码"`
	PageSize     int    `json:"page_size" dc:"每页条数"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Status       string `json:"status" dc:"状态：1启用，0停用，空或-1表示全部"`
}

// RechargeRiskRuleListRes 返回风控规则列表与分页信息。
type RechargeRiskRuleListRes struct {
	List       []RechargeRiskRuleItem `json:"list" dc:"规则列表"`
	Pagination PaginationRes          `json:"pagination" dc:"分页信息"`
}

// RechargeRiskRuleItem 是风控规则列表展示项。
type RechargeRiskRuleItem struct {
	ID             int64  `json:"id" dc:"规则ID"`
	Account        string `json:"account" dc:"充值账号"`
	GoodsKeyword   string `json:"goods_keyword" dc:"商品名关键词"`
	Reason         string `json:"reason" dc:"风控原因"`
	Status         int    `json:"status" dc:"状态"`
	StatusText     string `json:"status_text" dc:"状态文案"`
	HitCount       int    `json:"hit_count" dc:"已拦截次数"`
	CreatedByName  string `json:"created_by_name" dc:"创建人"`
	UpdatedByName  string `json:"updated_by_name" dc:"更新人"`
	CreatedAt      string `json:"created_at" dc:"创建时间"`
	UpdatedAt      string `json:"updated_at" dc:"更新时间"`
}

// RechargeRiskRuleCreateReq 用于新增充值账号风控规则。
type RechargeRiskRuleCreateReq struct {
	g.Meta       `path:"/recharge-risks/rules" method:"post" tags:"充值风控" summary:"新增风控规则" security:"BearerAuth" dc:"新增充值账号风控规则"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Reason       string `json:"reason" dc:"风控原因"`
	Status       int    `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleCreateRes 返回新增后的规则 ID。
type RechargeRiskRuleCreateRes struct {
	ID int64 `json:"id" dc:"规则ID"`
}

// RechargeRiskRuleUpdateReq 用于编辑充值账号风控规则。
type RechargeRiskRuleUpdateReq struct {
	g.Meta       `path:"/recharge-risks/rules/{id}" method:"put" tags:"充值风控" summary:"编辑风控规则" security:"BearerAuth" dc:"编辑充值账号风控规则"`
	ID           int64  `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Reason       string `json:"reason" dc:"风控原因"`
	Status       int    `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleUpdateRes 表示风控规则编辑成功。
type RechargeRiskRuleUpdateRes struct{}

// RechargeRiskRuleStatusReq 用于启用或停用充值账号风控规则。
type RechargeRiskRuleStatusReq struct {
	g.Meta `path:"/recharge-risks/rules/{id}/status" method:"patch" tags:"充值风控" summary:"修改风控规则状态" security:"BearerAuth" dc:"启用或停用充值账号风控规则"`
	ID     int64 `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
	Status int   `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleStatusRes 表示风控规则状态修改成功。
type RechargeRiskRuleStatusRes struct{}

// RechargeRiskRuleDeleteReq 用于软删除充值账号风控规则。
type RechargeRiskRuleDeleteReq struct {
	g.Meta `path:"/recharge-risks/rules/{id}" method:"delete" tags:"充值风控" summary:"删除风控规则" security:"BearerAuth" dc:"软删除充值账号风控规则"`
	ID     int64 `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
}

// RechargeRiskRuleDeleteRes 表示风控规则删除成功。
type RechargeRiskRuleDeleteRes struct{}

// RechargeRiskRecordListReq 用于分页查询充值账号风控拦截记录。
type RechargeRiskRecordListReq struct {
	g.Meta       `path:"/recharge-risks/records" method:"get" tags:"充值风控" summary:"风控记录列表" security:"BearerAuth" dc:"分页查询充值账号风控拦截记录"`
	Page         int    `json:"page" dc:"页码"`
	PageSize     int    `json:"page_size" dc:"每页条数"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	StartTime    string `json:"start_time" dc:"拦截开始时间"`
	EndTime      string `json:"end_time" dc:"拦截结束时间"`
}

// RechargeRiskRecordListRes 返回风控拦截记录列表与分页信息。
type RechargeRiskRecordListRes struct {
	List       []RechargeRiskRecordItem `json:"list" dc:"拦截记录列表"`
	Pagination PaginationRes            `json:"pagination" dc:"分页信息"`
}

// RechargeRiskRecordItem 是风控拦截记录列表展示项。
type RechargeRiskRecordItem struct {
	ID             int64  `json:"id" dc:"记录ID"`
	RuleID         int64  `json:"rule_id" dc:"规则ID"`
	OrderNo        string `json:"order_no" dc:"订单号"`
	Account        string `json:"account" dc:"充值账号"`
	MatchedKeyword string `json:"matched_keyword" dc:"命中关键词"`
	GoodsCode      string `json:"goods_code" dc:"商品编码"`
	GoodsName      string `json:"goods_name" dc:"商品名称"`
	Reason         string `json:"reason" dc:"风控原因"`
	InterceptedAt  string `json:"intercepted_at" dc:"拦截时间"`
}
```

- [ ] **Step 4: Add service interface and controller**

Create `internal/service/recharge_risk.go`:

```go
package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// RechargeRiskService 定义充值账号风控规则和拦截记录后台能力。
type RechargeRiskService interface {
	ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error)
	CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleCreateRes, error)
	UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleUpdateRes, error)
	UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleStatusRes, error)
	DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleDeleteRes, error)
	ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error)
}
```

Create `internal/controller/admin/recharge_risk.go`:

```go
package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// RechargeRiskController 提供充值账号风控规则和拦截记录后台 HTTP handler。
type RechargeRiskController struct {
	svc service.RechargeRiskService
}

// NewRechargeRisk 创建 RechargeRiskController。
func NewRechargeRisk(svc service.RechargeRiskService) *RechargeRiskController {
	return &RechargeRiskController{svc: svc}
}

// ListRules 返回充值账号风控规则分页列表。
func (c *RechargeRiskController) ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error) {
	return c.svc.ListRules(ctx, req)
}

// CreateRule 新增充值账号风控规则。
func (c *RechargeRiskController) CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq) (*adminapi.RechargeRiskRuleCreateRes, error) {
	return c.svc.CreateRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateRule 编辑充值账号风控规则。
func (c *RechargeRiskController) UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq) (*adminapi.RechargeRiskRuleUpdateRes, error) {
	return c.svc.UpdateRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateRuleStatus 启用或停用充值账号风控规则。
func (c *RechargeRiskController) UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq) (*adminapi.RechargeRiskRuleStatusRes, error) {
	return c.svc.UpdateRuleStatus(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// DeleteRule 软删除充值账号风控规则。
func (c *RechargeRiskController) DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq) (*adminapi.RechargeRiskRuleDeleteRes, error) {
	return c.svc.DeleteRule(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// ListRecords 返回充值账号风控拦截记录分页列表。
func (c *RechargeRiskController) ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error) {
	return c.svc.ListRecords(ctx, req)
}
```

Create `internal/logic/admin/recharge_risk.go`:

```go
package adminlogic

import "myjob/internal/app"

// RechargeRiskLogic 提供充值账号风控规则和拦截记录后台业务能力。
type RechargeRiskLogic struct{ core *app.Core }
```

- [ ] **Step 5: Wire services and routes**

Modify `internal/logic/admin/common.go`:

```go
RechargeRisk service.RechargeRiskService
```

Add this in `NewServices`:

```go
RechargeRisk: &RechargeRiskLogic{core: core},
```

Add this assertion in `internal/logic/admin/recharge_risk.go`:

```go
var _ service.RechargeRiskService = (*RechargeRiskLogic)(nil)
```

Import `myjob/internal/service` in that file.

Modify `internal/bootstrap/application.go`:

```go
rechargeRiskCtrl := admincontroller.NewRechargeRisk(services.RechargeRisk)
```

Bind it in a new admin group:

```go
group.Group("", func(group *ghttp.RouterGroup) {
	group.Middleware(guard.Require("order.recharge_risk", false))
	group.Bind(rechargeRiskCtrl)
})
```

- [ ] **Step 6: Seed menu permission**

Modify `defaultMenus()` in `internal/app/bootstrap.go`:

```go
{ID: 15, ParentID: 0, Name: "充值风控", Code: "order.recharge_risk", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 15},
```

Modify `ensureDefaultGroupAuth` list:

```go
for _, menuID := range []int64{1, 2, 3, 4, 5, 7, 8, 10, 11, 12, 13, 14, 15} {
```

Modify `manifest/sql/002_seed_menu.sql` by adding menu ID `15` and default group grant `(1, 15, NOW())`.

- [ ] **Step 7: Add temporary service method stubs so routes compile**

Create `internal/logic/admin/recharge_risk_query.go`:

```go
package adminlogic

import (
	"context"

	adminapi "myjob/api"
)

// ListRules 分页查询充值账号风控规则。
func (l *RechargeRiskLogic) ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error) {
	return &adminapi.RechargeRiskRuleListRes{List: []adminapi.RechargeRiskRuleItem{}, Pagination: adminapi.PaginationRes{}}, nil
}

// ListRecords 分页查询充值账号风控拦截记录。
func (l *RechargeRiskLogic) ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error) {
	return &adminapi.RechargeRiskRecordListRes{List: []adminapi.RechargeRiskRecordItem{}, Pagination: adminapi.PaginationRes{}}, nil
}
```

Create `internal/logic/admin/recharge_risk_write.go`:

```go
package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// CreateRule 新增充值账号风控规则。
func (l *RechargeRiskLogic) CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleCreateRes, error) {
	return nil, apiErr(consts.CodeInternalError, "风控规则新增未实现")
}

// UpdateRule 编辑充值账号风控规则。
func (l *RechargeRiskLogic) UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleUpdateRes, error) {
	return nil, apiErr(consts.CodeInternalError, "风控规则编辑未实现")
}

// UpdateRuleStatus 启用或停用充值账号风控规则。
func (l *RechargeRiskLogic) UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleStatusRes, error) {
	return nil, apiErr(consts.CodeInternalError, "风控规则状态修改未实现")
}

// DeleteRule 软删除充值账号风控规则。
func (l *RechargeRiskLogic) DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleDeleteRes, error) {
	return nil, apiErr(consts.CodeInternalError, "风控规则删除未实现")
}
```

- [ ] **Step 8: Run contract route tests**

Run:

```bash
go test ./test/contract -run 'Test(OpenAPI_RechargeRiskPathsExposed|RechargeRiskPermissionSeeded|RechargeRiskRequiresPermission|APIProtocolLayout)' -count=1 -timeout 60s
```

Expected: pass for route exposure, permission seeding, and forbidden access.

- [ ] **Step 9: Commit API and wiring**

```bash
git add api/recharge_risk.go internal/service/recharge_risk.go internal/controller/admin/recharge_risk.go internal/logic/admin/recharge_risk.go internal/logic/admin/recharge_risk_query.go internal/logic/admin/recharge_risk_write.go internal/logic/admin/common.go internal/bootstrap/application.go internal/app/bootstrap.go manifest/sql/002_seed_menu.sql test/contract/api_layout_test.go test/contract/recharge_risk_contract_test.go
git commit -m "feat: expose recharge risk admin routes"
```

## Task 3: Rule Management Logic

**Files:**
- Modify: `internal/logic/admin/recharge_risk_query.go`
- Modify: `internal/logic/admin/recharge_risk_write.go`
- Create: `internal/logic/admin/recharge_risk_validate.go`
- Create: `internal/logic/admin/recharge_risk_mapper.go`
- Test: `test/contract/recharge_risk_contract_test.go`

- [ ] **Step 1: Add failing rule CRUD/list contract test**

Append to `test/contract/recharge_risk_contract_test.go`:

```go
func TestRechargeRiskRuleCRUDAndFilters(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	invalid := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":        "",
		"goods_keyword":  "剪映",
		"reason":         "测试原因",
		"status":         1,
	}, token)
	require.Equal(t, 400, invalid.Code)

	create := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":        "risk-account-001",
		"goods_keyword":  "剪映",
		"reason":         "客户多次提交错误账号",
		"status":         1,
	}, token)
	require.Equal(t, 0, create.Code)
	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotZero(t, createData.ID)

	duplicate := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":        "risk-account-001",
		"goods_keyword":  "剪映",
		"reason":         "重复",
		"status":         1,
	}, token)
	require.Equal(t, 409, duplicate.Code)

	list := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&account=risk-account-001&goods_keyword=剪映&status=1", token)
	require.Equal(t, 0, list.Code)
	var listData struct {
		List []struct {
			ID           int64  `json:"id"`
			Account      string `json:"account"`
			GoodsKeyword string `json:"goods_keyword"`
			Reason       string `json:"reason"`
			Status       int    `json:"status"`
			StatusText   string `json:"status_text"`
			HitCount     int    `json:"hit_count"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(list.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, createData.ID, listData.List[0].ID)
	require.Equal(t, "risk-account-001", listData.List[0].Account)
	require.Equal(t, "剪映", listData.List[0].GoodsKeyword)
	require.Equal(t, "客户多次提交错误账号", listData.List[0].Reason)
	require.Equal(t, 1, listData.List[0].Status)
	require.Equal(t, "启用", listData.List[0].StatusText)
	require.Equal(t, 0, listData.List[0].HitCount)

	update := h.putJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID), map[string]any{
		"account":        "risk-account-001",
		"goods_keyword":  "醒图",
		"reason":         "更新后的风控原因",
		"status":         0,
	}, token)
	require.Equal(t, 0, update.Code)

	disabled := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&goods_keyword=醒图&status=0", token)
	require.Equal(t, 0, disabled.Code)
	require.NoError(t, json.Unmarshal(disabled.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, "停用", listData.List[0].StatusText)

	enable := h.patchJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID)+"/status", map[string]any{"status": 1}, token)
	require.Equal(t, 0, enable.Code)

	deleted := h.deleteJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, deleted.Code)

	empty := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&account=risk-account-001", token)
	require.Equal(t, 0, empty.Code)
	require.NoError(t, json.Unmarshal(empty.Data, &listData))
	require.Empty(t, listData.List)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./test/contract -run TestRechargeRiskRuleCRUDAndFilters -count=1 -timeout 60s
```

Expected: fail because write methods return internal errors.

- [ ] **Step 3: Add validation helpers**

Create `internal/logic/admin/recharge_risk_validate.go`:

```go
package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"myjob/internal/consts"
)

const (
	rechargeRiskStatusDisabled = 0
	rechargeRiskStatusEnabled  = 1
)

type normalizedRechargeRiskRuleInput struct {
	Account      string
	GoodsKeyword string
	Reason       string
	Status       int
}

func normalizeRechargeRiskRuleInput(account, goodsKeyword, reason string, status int) (normalizedRechargeRiskRuleInput, error) {
	account = strings.TrimSpace(account)
	goodsKeyword = strings.TrimSpace(goodsKeyword)
	reason = strings.TrimSpace(reason)
	if account == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("充值账号不能为空")
	}
	if goodsKeyword == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("匹配关键词不能为空")
	}
	if reason == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("风控原因不能为空")
	}
	if status != rechargeRiskStatusDisabled && status != rechargeRiskStatusEnabled {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("状态值错误")
	}
	return normalizedRechargeRiskRuleInput{Account: account, GoodsKeyword: goodsKeyword, Reason: reason, Status: status}, nil
}

func normalizeRechargeRiskStatusFilter(value string) (int, bool, error) {
	switch strings.TrimSpace(value) {
	case "", "-1":
		return 0, false, nil
	case "0":
		return rechargeRiskStatusDisabled, true, nil
	case "1":
		return rechargeRiskStatusEnabled, true, nil
	default:
		return 0, false, fmt.Errorf("状态筛选值错误")
	}
}

func rechargeRiskStatusText(status int) string {
	if status == rechargeRiskStatusEnabled {
		return "启用"
	}
	return "停用"
}

func archivedRechargeRiskKeyword(keyword string, id int64) string {
	return fmt.Sprintf("%s#deleted#%d", keyword, id)
}

func (l *RechargeRiskLogic) ensureRechargeRiskRuleUnique(ctx context.Context, account, keyword string, excludeID int64) error {
	conditions := []string{"account = ?", "goods_keyword = ?", "is_deleted = 0"}
	args := []any{account, keyword}
	if excludeID > 0 {
		conditions = append(conditions, "id <> ?")
		args = append(args, excludeID)
	}
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_rule WHERE `+strings.Join(conditions, " AND "), args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "风控规则重复校验失败")
	}
	if count.Int() > 0 {
		return apiErr(consts.CodeConflict, "相同充值账号和关键词的风控规则已存在")
	}
	return nil
}
```

- [ ] **Step 4: Add mappers**

Create `internal/logic/admin/recharge_risk_mapper.go`:

```go
package adminlogic

import (
	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

func rechargeRiskRuleItemFromEntity(row entity.RechargeRiskRule) adminapi.RechargeRiskRuleItem {
	return adminapi.RechargeRiskRuleItem{
		ID:            row.ID,
		Account:       row.Account,
		GoodsKeyword:  row.GoodsKeyword,
		Reason:        row.Reason,
		Status:        row.Status,
		StatusText:    rechargeRiskStatusText(row.Status),
		HitCount:      row.HitCount,
		CreatedByName: row.CreatedByName,
		UpdatedByName: row.UpdatedByName,
		CreatedAt:     formatAppTime(row.CreatedAt),
		UpdatedAt:     formatAppTime(row.UpdatedAt),
	}
}

func rechargeRiskRecordItemFromEntity(row entity.RechargeRiskRecord) adminapi.RechargeRiskRecordItem {
	return adminapi.RechargeRiskRecordItem{
		ID:             row.ID,
		RuleID:         row.RuleID,
		OrderNo:        row.OrderNo,
		Account:        row.Account,
		MatchedKeyword: row.MatchedKeyword,
		GoodsCode:      row.GoodsCode,
		GoodsName:      row.GoodsName,
		Reason:         row.Reason,
		InterceptedAt:  formatAppTime(row.InterceptedAt),
	}
}
```

- [ ] **Step 5: Implement rule list**

Replace `ListRules` in `internal/logic/admin/recharge_risk_query.go` with:

```go
func (l *RechargeRiskLogic) ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"is_deleted = 0"}
	args := make([]any, 0, 8)
	if account := strings.TrimSpace(req.Account); account != "" {
		conditions = append(conditions, "account LIKE ?")
		args = append(args, "%"+account+"%")
	}
	if keyword := strings.TrimSpace(req.GoodsKeyword); keyword != "" {
		conditions = append(conditions, "goods_keyword LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if status, ok, err := normalizeRechargeRiskStatusFilter(req.Status); err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	} else if ok {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_rule WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则列表查询失败")
	}
	rows := make([]entity.RechargeRiskRule, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, account, goods_keyword, reason, status, hit_count, created_by_id, created_by_name,
       updated_by_id, updated_by_name, is_deleted, deleted_at, created_at, updated_at
FROM recharge_risk_rule
WHERE `+whereClause+`
ORDER BY updated_at DESC, id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则列表查询失败")
	}
	items := make([]adminapi.RechargeRiskRuleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rechargeRiskRuleItemFromEntity(row))
	}
	return &adminapi.RechargeRiskRuleListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}
```

Add imports in `recharge_risk_query.go`:

```go
import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)
```

- [ ] **Step 6: Implement rule writes**

Replace the stubs in `internal/logic/admin/recharge_risk_write.go`:

```go
// CreateRule 新增充值账号风控规则，并记录操作日志。
func (l *RechargeRiskLogic) CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleCreateRes, error) {
	normalized, err := normalizeRechargeRiskRuleInput(req.Account, req.GoodsKeyword, req.Reason, req.Status)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if err = l.ensureRechargeRiskRuleUnique(ctx, normalized.Account, normalized.GoodsKeyword, 0); err != nil {
		return nil, err
	}
	now := l.core.Now()
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO recharge_risk_rule (
    account, goods_keyword, reason, status, hit_count,
    created_by_id, created_by_name, updated_by_id, updated_by_name,
    is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, 0, ?, ?, ?, ?, 0, ?, ?)
`, normalized.Account, normalized.GoodsKeyword, normalized.Reason, normalized.Status, actor.ID, actor.RealName, actor.ID, actor.RealName, now, now)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增充值风控规则：%s / %s", normalized.Account, normalized.GoodsKeyword), ip)
	return &adminapi.RechargeRiskRuleCreateRes{ID: id}, nil
}
```

Add these methods in the same file:

```go
// UpdateRule 编辑充值账号风控规则，并保留累计拦截次数。
func (l *RechargeRiskLogic) UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleUpdateRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	if _, err := l.getRechargeRiskRule(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	normalized, err := normalizeRechargeRiskRuleInput(req.Account, req.GoodsKeyword, req.Reason, req.Status)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if err = l.ensureRechargeRiskRuleUnique(ctx, normalized.Account, normalized.GoodsKeyword, req.ID); err != nil {
		return nil, err
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET account = ?, goods_keyword = ?, reason = ?, status = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, normalized.Account, normalized.GoodsKeyword, normalized.Reason, normalized.Status, actor.ID, actor.RealName, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑充值风控规则：%d", req.ID), ip)
	return &adminapi.RechargeRiskRuleUpdateRes{}, nil
}

// UpdateRuleStatus 启用或停用充值账号风控规则。
func (l *RechargeRiskLogic) UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleStatusRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	if req.Status != rechargeRiskStatusDisabled && req.Status != rechargeRiskStatusEnabled {
		return nil, apiErr(consts.CodeBadRequest, "状态值错误")
	}
	if _, err := l.getRechargeRiskRule(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET status = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, req.Status, actor.ID, actor.RealName, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则状态修改失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("修改充值风控规则状态：%d -> %s", req.ID, rechargeRiskStatusText(req.Status)), ip)
	return &adminapi.RechargeRiskRuleStatusRes{}, nil
}

// DeleteRule 软删除充值账号风控规则，并归档关键词以避免唯一约束阻塞后续重建。
func (l *RechargeRiskLogic) DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleDeleteRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	rule, err := l.getRechargeRiskRule(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	now := l.core.Now()
	if _, err = l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET goods_keyword = ?, is_deleted = 1, deleted_at = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, archivedRechargeRiskKeyword(rule.GoodsKeyword, req.ID), now, actor.ID, actor.RealName, now, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除充值风控规则：%d -> %s / %s", req.ID, rule.Account, rule.GoodsKeyword), ip)
	return &adminapi.RechargeRiskRuleDeleteRes{}, nil
}
```

Add helper:

```go
func (l *RechargeRiskLogic) getRechargeRiskRule(ctx context.Context, id int64) (entity.RechargeRiskRule, error) {
	row := entity.RechargeRiskRule{}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT id, account, goods_keyword, reason, status, hit_count, created_by_id, created_by_name,
       updated_by_id, updated_by_name, is_deleted, deleted_at, created_at, updated_at
FROM recharge_risk_rule
WHERE id = ? AND is_deleted = 0
`, id)
	return row, err
}
```

Add `fmt` to imports in `recharge_risk_write.go`.

- [ ] **Step 7: Run rule tests**

```bash
go test ./test/contract -run TestRechargeRiskRuleCRUDAndFilters -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 8: Commit rule management**

```bash
git add internal/logic/admin/recharge_risk_query.go internal/logic/admin/recharge_risk_write.go internal/logic/admin/recharge_risk_validate.go internal/logic/admin/recharge_risk_mapper.go test/contract/recharge_risk_contract_test.go
git commit -m "feat: manage recharge risk rules"
```

## Task 4: Risk Record List

**Files:**
- Modify: `internal/logic/admin/recharge_risk_query.go`
- Test: `test/contract/recharge_risk_contract_test.go`

- [ ] **Step 1: Add failing record list test**

Append to `test/contract/recharge_risk_contract_test.go`:

```go
func TestRechargeRiskRecordListFilters(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	now := h.app.Core().Now()
	_, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO recharge_risk_rule (
    account, goods_keyword, reason, status, hit_count,
    created_by_id, created_by_name, updated_by_id, updated_by_name,
    is_deleted, created_at, updated_at
) VALUES ('record-account-001', '微博', '记录筛选原因', 1, 1, 1, 'admin', 1, 'admin', 0, ?, ?)
`, now, now)
	require.NoError(t, err)
	ruleIDValue, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `SELECT id FROM recharge_risk_rule WHERE account = ?`, "record-account-001")
	require.NoError(t, err)
	ruleID := ruleIDValue.Int64()
	_, err = h.app.Core().DB().Exec(context.Background(), `
INSERT INTO recharge_risk_record (
    rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
    matched_keyword, reason, request_token_masked, intercepted_at, created_at
) VALUES (?, 1001, 'ORISKRECORD001', 'record-account-001', 2001, 'G-RISK-001', '新浪微博会员', '微博', '记录筛选原因', 'test***oken', ?, ?)
`, ruleID, now, now)
	require.NoError(t, err)

	res := h.getJSON("/api/admin/recharge-risks/records?page=1&page_size=20&account=record-account-001&goods_keyword=微博&start_time="+now.Add(-time.Minute).Format("2006-01-02 15:04:05")+"&end_time="+now.Add(time.Minute).Format("2006-01-02 15:04:05"), token)
	require.Equal(t, 0, res.Code)
	var data struct {
		List []struct {
			RuleID         int64  `json:"rule_id"`
			OrderNo        string `json:"order_no"`
			Account        string `json:"account"`
			MatchedKeyword string `json:"matched_keyword"`
			GoodsCode      string `json:"goods_code"`
			GoodsName      string `json:"goods_name"`
			Reason         string `json:"reason"`
			InterceptedAt  string `json:"intercepted_at"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.Len(t, data.List, 1)
	require.Equal(t, "ORISKRECORD001", data.List[0].OrderNo)
	require.Equal(t, "record-account-001", data.List[0].Account)
	require.Equal(t, "微博", data.List[0].MatchedKeyword)
	require.Equal(t, "G-RISK-001", data.List[0].GoodsCode)
	require.Equal(t, "新浪微博会员", data.List[0].GoodsName)
	require.Equal(t, "记录筛选原因", data.List[0].Reason)

	invalidTime := h.getJSON("/api/admin/recharge-risks/records?page=1&page_size=20&start_time=bad-time", token)
	require.Equal(t, 400, invalidTime.Code)
}
```

Add imports to the test file:

```go
import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./test/contract -run TestRechargeRiskRecordListFilters -count=1 -timeout 60s
```

Expected: fail because `ListRecords` still returns an empty list.

- [ ] **Step 3: Implement record list**

Replace `ListRecords` in `internal/logic/admin/recharge_risk_query.go`:

```go
func (l *RechargeRiskLogic) ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 8)
	if account := strings.TrimSpace(req.Account); account != "" {
		conditions = append(conditions, "account LIKE ?")
		args = append(args, "%"+account+"%")
	}
	if keyword := strings.TrimSpace(req.GoodsKeyword); keyword != "" {
		conditions = append(conditions, "matched_keyword LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if start := strings.TrimSpace(req.StartTime); start != "" {
		parsed, err := app.ParseQueryTime(start)
		if err != nil {
			return nil, apiErr(consts.CodeBadRequest, "拦截开始时间格式错误")
		}
		conditions = append(conditions, "intercepted_at >= ?")
		args = append(args, parsed)
	}
	if end := strings.TrimSpace(req.EndTime); end != "" {
		parsed, err := app.ParseQueryTime(end)
		if err != nil {
			return nil, apiErr(consts.CodeBadRequest, "拦截结束时间格式错误")
		}
		conditions = append(conditions, "intercepted_at <= ?")
		args = append(args, parsed)
	}
	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_record WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控记录列表查询失败")
	}
	rows := make([]entity.RechargeRiskRecord, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
       matched_keyword, reason, request_token_masked, intercepted_at, created_at
FROM recharge_risk_record
WHERE `+whereClause+`
ORDER BY intercepted_at DESC, id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控记录列表查询失败")
	}
	items := make([]adminapi.RechargeRiskRecordItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rechargeRiskRecordItemFromEntity(row))
	}
	return &adminapi.RechargeRiskRecordListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}
```

- [ ] **Step 4: Run record tests**

```bash
go test ./test/contract -run TestRechargeRiskRecordListFilters -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 5: Commit record list**

```bash
git add internal/logic/admin/recharge_risk_query.go test/contract/recharge_risk_contract_test.go
git commit -m "feat: list recharge risk records"
```

## Task 5: Open Order Risk Interception

**Files:**
- Create: `internal/logic/order/order_risk.go`
- Modify: `internal/logic/order/order_create.go`
- Test: `test/integration/order_worker_test.go`

- [ ] **Step 1: Add failing integration test for enabled rule interception**

Add this test near other order worker tests in `test/integration/order_worker_test.go`:

```go
func TestOpenOrderCreatesFailedOrderWhenRechargeRiskMatches(t *testing.T) {
	var createCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createCount.Add(1)
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功"}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "风控品牌", "剪辑工具", "剪映")
	subjectID := h.createSubject(t, token, "风控主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "剪映专业版会员", "20.0000")
	platformID := h.createKakayunPlatform(t, token, "风控云发卡", subjectID, 0, strings.TrimPrefix(server.URL, "http://"))
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478601",
		"supplier_goods_name": "剪映专业版会员",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	rule := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":        "bad-jianying-account",
		"goods_keyword":  "剪映",
		"reason":         "客户多次提交错误剪映账号",
		"status":         1,
	}, token)
	require.Equal(t, 0, rule.Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "bad-jianying-account",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	var createData struct {
		OrderNo    string `json:"order_no"`
		StatusCode string `json:"status_code"`
		StatusText string `json:"status_text"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotEmpty(t, createData.OrderNo)
	require.Equal(t, "failed", createData.StatusCode)
	require.Equal(t, "失败", createData.StatusText)

	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	require.EqualValues(t, 0, createCount.Load())
	order := h.loadOrder(t, createData.OrderNo)
	require.Equal(t, "failed", order.Status)
	require.Contains(t, order.LastReceipt, "客户多次提交错误剪映账号")
	require.EqualValues(t, 0, h.scalarInt(t, `SELECT COUNT(*) FROM external_order_attempt WHERE order_id = ?`, order.ID))
	require.EqualValues(t, 1, h.scalarInt(t, `SELECT hit_count FROM recharge_risk_rule WHERE account = ? AND is_deleted = 0`, "bad-jianying-account"))
	require.EqualValues(t, 1, h.scalarInt(t, `SELECT COUNT(*) FROM recharge_risk_record WHERE order_no = ? AND matched_keyword = ?`, createData.OrderNo, "剪映"))
}
```

- [ ] **Step 2: Add failing integration test for disabled rule**

Add this test:

```go
func TestOpenOrderDoesNotRiskBlockWhenRuleDisabled(t *testing.T) {
	var createCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createCount.Add(1)
		_, _ = w.Write([]byte(`{"code":1,"message":"下单成功","data":{"orderno":"SDRISKDISABLED","usorderno":"O-T1"}}`))
	}))
	defer server.Close()

	h := newOrderIntegrationHarness(t)
	token := h.loginAdmin(t)
	leafBrandID := h.createBrandPath(t, token, "停用风控品牌", "修图工具", "醒图")
	subjectID := h.createSubject(t, token, "停用风控主体", 0)
	goodsID := h.createDirectRechargeGoods(t, token, leafBrandID, "醒图会员", "20.0000")
	platformID := h.createKakayunPlatform(t, token, "停用风控云发卡", subjectID, 0, strings.TrimPrefix(server.URL, "http://"))
	require.Equal(t, 0, h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "2478602",
		"supplier_goods_name": "醒图会员",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token).Code)
	require.Equal(t, 0, h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":        "disabled-risk-account",
		"goods_keyword":  "醒图",
		"reason":         "停用规则不应拦截",
		"status":         0,
	}, token).Code)

	detail := h.getJSON("/api/admin/products/"+int64ToString(goodsID), token)
	require.Equal(t, 0, detail.Code)
	var goodsDetail struct {
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(detail.Data, &goodsDetail))

	create := h.postJSON("/api/open/orders", map[string]any{
		"token":    "test-open-order-token",
		"goods_id": goodsDetail.GoodsCode,
		"account":  "disabled-risk-account",
		"quantity": 1,
	}, "")
	require.Equal(t, 0, create.Code)
	var createData struct {
		OrderNo    string `json:"order_no"`
		StatusCode string `json:"status_code"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.Equal(t, "pending_submit", createData.StatusCode)

	require.NoError(t, h.orderService.SubmitPendingOnce(context.Background()))
	require.EqualValues(t, 1, createCount.Load())
	require.EqualValues(t, 0, h.scalarInt(t, `SELECT COUNT(*) FROM recharge_risk_record WHERE order_no = ?`, createData.OrderNo))
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./test/integration -run 'TestOpenOrder(CreatesFailedOrderWhenRechargeRiskMatches|DoesNotRiskBlockWhenRuleDisabled)' -count=1 -timeout 60s
```

Expected: first test fails because open-order creation does not check risk rules yet.

- [ ] **Step 4: Add order risk helper**

Create `internal/logic/order/order_risk.go`:

```go
package orderlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	adminapi "myjob/api"
	runtimeapp "myjob/internal/app"

	"github.com/gogf/gf/v2/database/gdb"
)

var errOrderNoRetriesExhausted = errors.New("订单号生成重试失败")

type rechargeRiskMatch struct {
	RuleID       int64  `db:"id"`
	GoodsKeyword string `db:"goods_keyword"`
	Reason       string `db:"reason"`
}

func (l *OrderLogic) matchRechargeRisk(ctx context.Context, account, goodsName string) (rechargeRiskMatch, bool, error) {
	rows := make([]rechargeRiskMatch, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, goods_keyword, reason
FROM recharge_risk_rule
WHERE account = ? AND status = 1 AND is_deleted = 0
ORDER BY id ASC
`, account); err != nil {
		return rechargeRiskMatch{}, false, err
	}
	for _, row := range rows {
		if strings.TrimSpace(row.GoodsKeyword) != "" && strings.Contains(goodsName, row.GoodsKeyword) {
			return row, true, nil
		}
	}
	return rechargeRiskMatch{}, false, nil
}

func (l *OrderLogic) createRiskFailedOpenOrder(ctx context.Context, req *adminapi.OpenOrderCreateReq, goods openOrderGoods, account, unitPrice, orderAmount string, match rechargeRiskMatch, now time.Time) (string, error) {
	orderNo := ""
	for attempt := 0; attempt < maxOpenOrderCreateAttempts; attempt++ {
		orderNo = l.nextOrderNo()
		err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			result, err := tx.Exec(`
INSERT INTO external_order (
    order_no, goods_id, goods_code, goods_name, goods_type, supply_type, subject_id, subject_name,
    has_tax, account, quantity, unit_price, order_amount, cost_amount, profit_amount,
    status, attempt_count, last_receipt, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '0.0000', ?, ?, 0, ?, ?, ?)
`, orderNo, goods.ID, goods.GoodsCode, goods.Name, goods.GoodsType, goods.SupplyType, nullableInt64Arg(goods.SubjectID), goods.SubjectName, goods.HasTax,
				account, req.Quantity, unitPrice, orderAmount, orderAmount, OrderStatusFailed, riskOrderReceipt(match.Reason), now, now)
			if err != nil {
				return err
			}
			orderID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(`
INSERT INTO recharge_risk_record (
    rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
    matched_keyword, reason, request_token_masked, intercepted_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, match.RuleID, orderID, orderNo, account, goods.ID, goods.GoodsCode, goods.Name, match.GoodsKeyword, match.Reason, runtimeapp.MaskSecret(req.Token), now, now); err != nil {
				return err
			}
			_, err = tx.Exec(`UPDATE recharge_risk_rule SET hit_count = hit_count + 1, updated_at = ? WHERE id = ? AND is_deleted = 0`, now, match.RuleID)
			return err
		})
		if err != nil {
			if isOrderNoUniqueConflict(err) {
				continue
			}
			return "", err
		}
		return orderNo, nil
	}
	return "", errOrderNoRetriesExhausted
}

func riskOrderReceipt(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "充值账号命中风控"
	}
	return "充值账号命中风控：" + reason
}
```

- [ ] **Step 5: Hook risk check into open-order creation**

Modify `CreateOpenOrder` in `internal/logic/order/order_create.go` after quantity validation and before `loadCandidateChannels`:

```go
now := l.core.Now()
unitPrice, err := normalizeOrderMoney(goods.DefaultSellPrice)
if err != nil {
	return nil, apiErr(consts.CodeBadRequest, "商品售价格式错误")
}
orderAmount, err := multiplyOrderMoney(unitPrice, req.Quantity)
if err != nil {
	return nil, apiErr(consts.CodeBadRequest, "订单金额计算失败")
}
riskMatch, matched, err := l.matchRechargeRisk(ctx, account, goods.Name)
if err != nil {
	return nil, apiErr(consts.CodeInternalError, "充值风控规则查询失败")
}
if matched {
	orderNo, err := l.createRiskFailedOpenOrder(ctx, req, goods, account, unitPrice, orderAmount, riskMatch, now)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控失败订单创建失败")
	}
	return &adminapi.OpenOrderCreateRes{
		OrderNo:    orderNo,
		StatusCode: OrderStatusFailed,
		StatusText: orderStatusText(OrderStatusFailed),
		CreatedAt:  formatAppTime(now),
	}, nil
}
```

Remove the later duplicate `now`, `unitPrice`, and `orderAmount` declarations from the original method. Keep candidate loading after the risk branch.

- [ ] **Step 6: Run risk interception tests**

```bash
go test ./test/integration -run 'TestOpenOrder(CreatesFailedOrderWhenRechargeRiskMatches|DoesNotRiskBlockWhenRuleDisabled)' -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 7: Run order regression tests**

```bash
go test ./test/integration -run 'TestOrderWorker(SubmitsPendingOrderToKakayun|PassesKakayunMaxMoneyWithAllowedLoss|FailsOrderWhenKakayunMaxMoneyCannotBeCalculated)' -count=1 -timeout 60s
go test ./test/contract -run TestOpenOrder -count=1 -timeout 60s
```

Expected: pass.

- [ ] **Step 8: Commit order interception**

```bash
git add internal/logic/order/order_risk.go internal/logic/order/order_create.go test/integration/order_worker_test.go
git commit -m "feat: block risky recharge orders"
```

## Task 6: Documentation, Full Verification, Cleanup

**Files:**
- Modify: `docs/module-map.md`
- Modify: `docs/development.md`
- Modify: `docs/testing.md`
- Modify: `test/contract/README.md`
- Modify: `docs/superpowers/README.md`
- Modify: `README.md` only if its module index or validation commands become stale

- [ ] **Step 1: Update module map**

Add a “充值风控” subsection under `docs/module-map.md` near “订单履约”:

```markdown
### 充值风控

- 协议：`api/recharge_risk.go`
- controller：`internal/controller/admin/recharge_risk.go`
- service：`RechargeRiskService`（`internal/service/recharge_risk.go`）
- logic：`internal/logic/admin/recharge_risk*.go`、订单命中点 `internal/logic/order/order_risk.go`
- 路由前缀：`/api/admin/recharge-risks*`
- 权限：`order.recharge_risk`
- 主要能力：配置充值账号风控规则、启停规则、查询风控规则列表、查询拦截记录；开放下单命中启用规则时创建失败订单并写入拦截流水。
- 边界：规则匹配只使用充值账号精确匹配和商品名关键词包含；不实现正则、优先级、批量导入导出或前端页面代码。
```

Add route summary row:

```markdown
| 充值风控 | `/api/admin/recharge-risks*` | `order.recharge_risk` |
```

- [ ] **Step 2: Update development docs**

Add to `docs/development.md` near order fulfillment notes:

```markdown
开放下单会在创建订单前检查启用的充值风控规则。匹配口径为充值账号精确匹配，并要求当前商品名称包含规则中的商品关键词。命中后系统仍创建 `external_order`，但状态直接为 `failed`，写入 `last_receipt` 和 `recharge_risk_record`，同时递增规则 `hit_count`；该订单不会触发 worker，也不会生成 `external_order_attempt`。
```

- [ ] **Step 3: Update testing docs**

Add to `docs/testing.md` under focused regressions:

```markdown
充值风控聚焦回归：

```bash
go test ./test/contract -run 'TestRechargeRisk' -count=1 -timeout 60s
go test ./test/integration -run 'TestOpenOrder(CreatesFailedOrderWhenRechargeRiskMatches|DoesNotRiskBlockWhenRuleDisabled)' -count=1 -timeout 60s
go test ./internal/app -run 'Test(ExternalOrderSchemaContainsRequiredTablesAndIndexes|RechargeRiskSchemaContainsComments)' -count=1 -timeout 60s
```
```

- [ ] **Step 4: Update contract README**

Add this bullet to `test/contract/README.md`:

```markdown
- 充值风控规则、权限、OpenAPI 暴露和风控记录列表。
```

- [ ] **Step 5: Update superpowers index**

Add the implementation plan line to `docs/superpowers/README.md`:

```markdown
- `plans/2026-04-27-recharge-risk.md`：充值账号风控管理实施计划。
```

- [ ] **Step 6: Run gofmt**

```bash
gofmt -w api/recharge_risk.go internal/model/entity/recharge_risk.go internal/service/recharge_risk.go internal/controller/admin/recharge_risk.go internal/logic/admin/recharge_risk*.go internal/app/recharge_risk_schema.go internal/app/bootstrap.go internal/bootstrap/application.go internal/logic/order/order_risk.go internal/logic/order/order_create.go internal/app/order_schema_test.go test/contract/api_layout_test.go test/contract/recharge_risk_contract_test.go test/integration/order_worker_test.go
```

- [ ] **Step 7: Run full verification**

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

Expected: all pass. If `golangci-lint` is not installed, record that exact failure in the final implementation summary and rely on CI lint.

- [ ] **Step 8: Inspect git diff**

```bash
git status --short
git diff --stat
```

Expected: only recharge risk implementation, tests, schema, and documentation files are modified.

- [ ] **Step 9: Commit docs and final cleanup**

```bash
git add README.md docs/module-map.md docs/development.md docs/testing.md test/contract/README.md docs/superpowers/README.md
git commit -m "docs: document recharge risk controls"
```

If `README.md` did not need changes, omit it from `git add`.

## Self-Review Checklist

- Spec coverage: the plan includes backend rule CRUD, rule status, record listing, account exact match, goods-name keyword match, failed order creation, no upstream submission, permission seeding, tests, and docs.
- File boundaries: no rule CRUD goes into `internal/logic/order`; order logic only checks active rules and writes atomic failed-order state.
- API layout: `api/recharge_risk.go` stays flat under `api/`.
- Risk record consistency: order failure, risk record insert, and hit-count increment happen in one transaction.
- YAGNI: no regex, priority, batch import/export, frontend code, or provider changes.
- Validation: final verification commands are `go test ./... -count=1 -timeout 60s`, `go build ./...`, and `golangci-lint run --timeout=5m`.
