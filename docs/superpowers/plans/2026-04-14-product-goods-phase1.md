# Product Goods Phase1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前 `myjob` 后端中交付商品管理一期后端闭环，并保持 SQL、菜单权限、OpenAPI 与契约测试同步。

**Architecture:** 新增独立 `ProductGoods` 模块承接商品 CRUD 和表单聚合接口；品牌、模板、购买限制策略模块只补商品反向引用校验。商品写入、软删和品牌计数更新统一放事务内处理，列表和详情通过实时关联读取展示字段。

**Tech Stack:** Go, GoFrame HTTP/OpenAPI, SQL(MySQL/SQLite 双 schema), testify contract tests

---

### Task 1: 先补失败的商品契约测试

**Files:**
- Create: `test/contract/product_goods_contract_test.go`
- Modify: `test/contract/product_contract_test.go`
- Test: `go test ./test/contract -run 'ProductGoods|ProductModulePathsExposed' -count=1`

- [ ] **Step 1: 先写商品模块契约测试**

```go
func TestOpenAPI_ProductGoodsPathsExposed(t *testing.T) {}
func TestProductGoodsSeedsStayInSync(t *testing.T) {}
func TestProductGoodsCRUDAndFilters(t *testing.T) {}
func TestProductGoodsReferenceConflicts(t *testing.T) {}
```

- [ ] **Step 2: 运行测试确认当前缺接口而失败**

Run: `go test ./test/contract -run 'ProductGoods|ProductModulePathsExposed' -count=1`
Expected: FAIL，报缺少 `/api/admin/products`、`product.goods` 或商品接口 404/校验不符。

- [ ] **Step 3: 保留现有品牌/模板/策略回归测试不删，只把商品新场景单独加进去**

- [ ] **Step 4: 小步复跑新测试，确保失败点收敛到“商品模块尚未实现”**

### Task 2: 搭商品模块 API 骨架和类型

**Files:**
- Create: `api/product_goods.go`
- Create: `internal/controller/admin/product_goods.go`
- Modify: `internal/service/interfaces.go`
- Modify: `internal/logic/admin/common.go`
- Modify: `internal/bootstrap/application.go`
- Modify: `internal/model/entity/admin.go`

- [ ] **Step 1: 定义商品 CRUD 与表单聚合请求响应结构**

```go
type ProductGoodsListReq struct {}
type ProductGoodsCreateReq struct {}
type ProductGoodsFormOptionsReq struct {}
```

- [ ] **Step 2: 增加控制器和 service 接口，把 `/products`、`/products/{id}`、`/products/form-options` 挂进 `/api/admin`**

- [ ] **Step 3: 增加 `product.goods` 权限守卫，保持 OpenAPI 暴露**

- [ ] **Step 4: 复跑契约测试，确认从 404 前进到业务实现失败**

Run: `go test ./test/contract -run 'ProductGoods|ProductModulePathsExposed' -count=1`

### Task 3: 实现商品逻辑、库表和旧模块回改

**Files:**
- Modify: `internal/app/schema.go`
- Modify: `manifest/sql/001_schema.sql`
- Modify: `internal/app/bootstrap.go`
- Modify: `manifest/sql/002_seed_menu.sql`
- Modify: `internal/logic/admin/brand.go`
- Modify: `internal/logic/admin/product_template.go`
- Modify: `internal/logic/admin/purchase_limit.go`
- Create: `internal/logic/admin/product_goods.go`
- Modify: `internal/dao/tables.go`

- [ ] **Step 1: 新增 `product_goods` 表和索引，运行时 schema 与部署 SQL 双份同步**

- [ ] **Step 2: 实现商品新增、详情、列表、编辑、软删、表单聚合和品牌递归筛选**

- [ ] **Step 3: 用事务保证品牌 `goods_count` 与商品写入一致**

- [ ] **Step 4: 给品牌、模板、策略删除补商品引用冲突校验**

- [ ] **Step 5: 新增 `product.goods` 菜单、默认组授权和路由挂载**

### Task 4: 回归验证

**Files:**
- Test: `go test ./test/contract -count=1`
- Test: `go test ./... -count=1`

- [ ] **Step 1: 先跑商品契约测试**

Run: `go test ./test/contract -run ProductGoods -count=1`
Expected: PASS

- [ ] **Step 2: 再跑全量契约和全量测试**

Run: `go test ./test/contract -count=1`
Run: `go test ./... -count=1`
Expected: PASS

- [ ] **Step 3: 检查 `git diff --stat`，确认只包含商品一期相关改动**
