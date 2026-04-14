# 商品含税主体补充设计

**目标**

在现有商品管理一期基础上补充“含税商品必须关联主体”的后端能力，满足后续按商品维度取开票主体的场景。

**规则**

- `product_goods` 新增可空字段 `subject_id`。
- 只有 `has_tax = 1` 的商品才允许并要求选择主体。
- `has_tax = 1` 时，`subject_id` 必填，且主体必须存在并且 `admin_subject.has_tax = 1`。
- `has_tax = 0` 时，不要求选择主体；后端统一把 `subject_id` 清空为 `NULL`，不保留历史值。

**接口调整**

1. 商品创建、编辑请求增加 `subject_id`。
2. 商品详情增加 `subject_id`、`subject_name`，用于编辑页回显。
3. 商品表单聚合接口增加 `subjects` 下拉，只返回含税主体。
4. 商品列表暂不追加主体列，保持当前一期列表结构稳定。

**数据实现**

1. `internal/app/schema.go` 与 `manifest/sql/001_schema.sql` 同步为 `product_goods` 增加 `subject_id BIGINT/INTEGER NULL`。
2. 商品详情查询通过 `LEFT JOIN admin_subject` 实时取 `subject_name`，不做名称快照冗余。
3. 创建、编辑写库时都走统一校验函数，保证含税规则和主体存在性一致。

**测试策略**

- 先补契约测试，锁住以下行为：
  - 含税商品不传主体返回 `400`
  - 含税商品传非含税主体返回 `400`
  - 含税商品创建成功后详情能回显 `subject_id`、`subject_name`
  - 不含税商品即使传了主体，详情里也回显为 `null`
  - 表单聚合接口只返回含税主体
- 然后实现最小代码使测试转绿，再跑全量 `go test ./... -count=1` 回归。
