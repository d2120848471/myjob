# 第三方平台启停状态与商品绑定联动实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给第三方平台账号增加业务启停状态，并把商品渠道绑定入口、摘要统计和关闭级联行为统一到同一口径。

**Architecture:** 在 `supplier_platform_account` 增加 `status` 字段，平台管理接口承载该字段并在关闭平台时同步关停历史绑定。商品渠道绑定查询与校验统一要求平台账号“未软删且已启用”，重新开启平台只恢复平台本身，不自动恢复绑定。

**Tech Stack:** GoFrame、MySQL/SQLite dual schema、contract tests

---

### Task 1: 平台账号状态字段

**Files:**
- Modify: `api/supplier_platform.go`
- Modify: `internal/model/entity/supplier_platform.go`
- Modify: `internal/logic/admin/supplier_platform_query.go`
- Modify: `internal/logic/admin/supplier_platform_mapper.go`

- [ ] 在平台新增/编辑协议中承载 `status`，并让详情/列表回显该字段
- [ ] 保持新增默认开启，编辑未显式传值时沿用当前状态

### Task 2: 平台关闭级联关停绑定

**Files:**
- Modify: `internal/logic/admin/supplier_platform_validate.go`
- Modify: `internal/logic/admin/supplier_platform_write.go`

- [ ] 校验平台状态只能为 `0/1`
- [ ] 平台从开启改关闭时，在同一事务内把该平台下 `product_goods_channel_binding.dock_status` 批量更新为关闭
- [ ] 平台重新开启时不自动恢复绑定

### Task 3: 商品渠道绑定联动过滤

**Files:**
- Modify: `internal/logic/admin/product_goods_channel_validate.go`
- Modify: `internal/logic/admin/product_goods_channel_query.go`

- [ ] 商品渠道绑定新增/编辑时拒绝已关闭平台
- [ ] 表单选项、绑定列表、商品列表渠道摘要统一过滤关闭平台

### Task 4: Schema、文档与验证

**Files:**
- Modify: `internal/app/schema.go`
- Modify: `internal/app/bootstrap.go`
- Modify: `manifest/sql/005_supplier_platform.sql`
- Modify: `test/contract/supplier_platform_contract_test.go`
- Modify: `test/contract/product_goods_channel_contract_test.go`
- Modify: `README.md`
- Modify: `docs/module-map.md`
- Modify: `docs/testing.md`
- Modify: `docs/product_goods_channel_order_requirements.md`

- [ ] 补齐 SQLite/MySQL schema 和启动期补列逻辑
- [ ] 增加契约测试覆盖默认启用、关闭级联关停绑定、重新开启不恢复绑定、关闭平台过滤和绑定拦截
- [ ] 运行 `go test ./... -count=1 -timeout 60s`、`go build ./...`、`golangci-lint run --timeout=5m`
