# Product Goods Tax Subject Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为商品管理补充 `subject_id`，实现“含税商品必须选择含税主体，不含税商品不保存主体”。

**Architecture:** 在 `product_goods` 主表新增可空 `subject_id`，商品写入统一通过 `normalizeProductGoodsInput` 校验主体规则；详情和表单聚合接口通过关联 `admin_subject` 暴露回显和下拉，不改当前商品列表返回结构。

**Tech Stack:** Go, GoFrame HTTP/OpenAPI, SQL(MySQL/SQLite 双 schema), testify contract tests

---

### Task 1: 先补商品主体约束的失败测试

**Files:**
- Modify: `test/contract/product_goods_contract_test.go`
- Test: `go test ./test/contract -run ProductGoods -count=1`

- [ ] **Step 1: 写主体相关契约测试**

```go
func TestProductGoodsCRUDAndFilters(t *testing.T) {
    // 断言 form-options 返回 subjects
    // 断言 has_tax=1 且未传 subject_id 返回 400
    // 断言 has_tax=1 且传非含税主体返回 400
    // 断言 has_tax=0 时 subject_id 最终回显为 nil
}
```

- [ ] **Step 2: 运行测试确认当前行为失败**

Run: `go test ./test/contract -run ProductGoods -count=1`
Expected: FAIL，报商品接口缺少 `subject_id/subject_name/subjects` 或含税主体校验不符。

### Task 2: 扩商品 API、schema 和实体

**Files:**
- Modify: `api/product_goods.go`
- Modify: `internal/model/entity/product_goods.go`
- Modify: `internal/app/schema.go`
- Modify: `manifest/sql/001_schema.sql`

- [ ] **Step 1: 给商品请求/详情/表单下拉增加主体字段**

- [ ] **Step 2: 给 `product_goods` 表增加可空 `subject_id`**

- [ ] **Step 3: 复跑商品契约测试，确认失败点收敛到逻辑实现**

Run: `go test ./test/contract -run ProductGoods -count=1`

### Task 3: 实现主体校验、详情回显和表单下拉

**Files:**
- Modify: `internal/logic/admin/product_goods.go`

- [ ] **Step 1: 在商品统一校验里加入含税主体规则**

- [ ] **Step 2: 让创建/编辑在 `has_tax=0` 时自动清空 `subject_id`**

- [ ] **Step 3: 详情关联主体名称，表单聚合返回含税主体下拉**

- [ ] **Step 4: 小步复跑商品契约测试直到通过**

Run: `go test ./test/contract -run ProductGoods -count=1`
Expected: PASS

### Task 4: 全量回归

**Files:**
- Test: `go test ./... -count=1`

- [ ] **Step 1: 跑全量测试**

Run: `go test ./... -count=1`
Expected: PASS

- [ ] **Step 2: 检查 diff，只包含商品主体补点相关改动**

Run: `git diff --stat`
