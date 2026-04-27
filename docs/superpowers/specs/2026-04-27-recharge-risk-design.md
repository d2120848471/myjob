# 充值账号风控管理设计

## 背景

部分客户会把错误充值账号提交到渠道，导致订单进入上游后需要人工处理。需要在本系统开放下单阶段识别这些账号，并在命中指定商品关键词时直接拦截失败，避免继续提交到上游。

当前仓库只有 Go 后端、OpenAPI、后台权限和菜单种子，没有前端项目代码。因此本次设计范围是后端接口、权限、风控生效逻辑、测试和文档，前端按接口实现“风控管理 / 风控记录”页面。

## 目标

- 后台可维护充值账号风控规则，支持列表、筛选、新增、编辑、启停和删除。
- 规则按“充值账号精确匹配 + 当前订单商品名关键词包含”命中。
- 开放下单命中启用规则时创建订单并直接失败，不提交上游。
- 记录每次拦截流水，支持按充值账号、关键词和拦截时间筛选。
- 保持现有分层边界，不把新逻辑堆进 `common.go`、`helper.go` 或后台订单大文件。

## 非目标

- 不实现前端页面代码。
- 不做复杂优先级、正则匹配、批量导入导出或黑白名单分组。
- 不改变上游 provider 协议。
- 不改变开放查单响应结构，只沿用订单状态表示失败。

## 已确认决策

- 风控命中维度：充值账号 + 当前订单商品名关键词。
- 充值账号匹配：精确匹配。
- 命中后行为：创建订单并直接置为失败。
- 规则状态：支持启用和停用。
- 记录筛选：支持充值账号、商品关键词和拦截时间。
- 交付范围：只做后端接口、权限菜单、风控生效逻辑、测试和文档。
- 推荐方案：开放下单创建阶段同步风控。

## 架构边界

新增独立后台业务域“充值风控”，路由前缀为 `/api/admin/recharge-risks`，权限点为 `order.recharge_risk`。

建议文件归属：

- 协议：`api/recharge_risk.go`
- controller：`internal/controller/admin/recharge_risk.go`
- service：`internal/service/recharge_risk.go`
- logic：`internal/logic/admin/recharge_risk.go` 保留类型声明，具体实现拆到：
  - `recharge_risk_query.go`
  - `recharge_risk_write.go`
  - `recharge_risk_validate.go`
  - `recharge_risk_mapper.go`
- 订单命中点：`internal/logic/order/order_create.go` 仅调用风控检查与记录能力，不承载规则 CRUD。
- schema：同步 `manifest/sql` 和 `internal/app/schema.go`，并提供启动兜底建表。

该领域归属于订单履约的前置拦截能力，但不复用后台订单列表 controller，避免 `/api/admin/orders` 继续膨胀。

## 数据模型

### `recharge_risk_rule`

规则配置表，使用软删除。

- `id`：规则 ID。
- `account`：充值账号，精确匹配开放下单账号。
- `goods_keyword`：商品名关键词，命中当前商品名时拦截。
- `reason`：风控原因，用于订单失败原因和记录展示。
- `status`：`1` 启用，`0` 停用。
- `hit_count`：累计拦截次数，命中后递增。
- `created_by` / `updated_by`：后台操作人快照。
- `is_deleted` / `deleted_at`：软删除字段。
- `created_at` / `updated_at`：审计时间。

唯一约束：`account + goods_keyword + is_deleted`，保证同一有效账号关键词组合只有一条规则。

推荐索引：

- `(account, status, is_deleted, id)`：下单命中查询。
- `(goods_keyword, is_deleted, id)`：后台关键词筛选。
- `(status, is_deleted, updated_at)`：后台状态筛选。

### `recharge_risk_record`

拦截流水表，不做软删除。

- `id`：记录 ID。
- `rule_id`：命中的规则 ID。
- `order_id` / `order_no`：被拦截订单。
- `account`：充值账号快照。
- `goods_id` / `goods_code` / `goods_name`：商品快照。
- `matched_keyword`：命中的关键词快照。
- `reason`：风控原因快照。
- `request_token_masked`：开放 token 脱敏快照，用于排查来源，不保存明文 token。
- `intercepted_at`：拦截时间。

推荐索引：

- `(account, intercepted_at, id)`：账号 + 时间筛选。
- `(matched_keyword, intercepted_at, id)`：关键词 + 时间筛选。
- `(rule_id, intercepted_at, id)`：按规则排查。
- `(order_no)`：订单反查。

## 后台接口

所有接口使用 `order.recharge_risk` 权限。

### 规则管理

`GET /api/admin/recharge-risks/rules`

- 入参：`page`、`page_size`、`account`、`goods_keyword`、`status`。
- 返回：规则列表、分页信息。
- 列表字段：`id`、`account`、`goods_keyword`、`hit_count`、`reason`、`status`、`status_text`、`created_by`、`updated_by`、`created_at`、`updated_at`。

`POST /api/admin/recharge-risks/rules`

- 入参：`account`、`goods_keyword`、`reason`、`status`。
- 行为：创建规则，校验账号、关键词、原因非空，校验组合唯一。

`PUT /api/admin/recharge-risks/rules/{id}`

- 入参：`account`、`goods_keyword`、`reason`、`status`。
- 行为：编辑规则，保留历史 `hit_count`。

`PATCH /api/admin/recharge-risks/rules/{id}/status`

- 入参：`status`。
- 行为：启用或停用规则。

`DELETE /api/admin/recharge-risks/rules/{id}`

- 行为：软删除规则。

### 风控记录

`GET /api/admin/recharge-risks/records`

- 入参：`page`、`page_size`、`account`、`goods_keyword`、`start_time`、`end_time`。
- 返回：记录列表、分页信息。
- 列表字段：`id`、`rule_id`、`order_no`、`account`、`matched_keyword`、`goods_code`、`goods_name`、`reason`、`intercepted_at`。

## 开放下单流程

`POST /api/open/orders` 调整为：

1. 校验开放 token、商品、充值账号和数量。
2. 按 `account = 提交账号` 查询启用风控规则，并用当前商品名匹配 `goods_keyword`。
3. 未命中：保持现有流程，创建 `pending_submit` 订单并触发 worker。
4. 命中：
   - 在同一事务内创建 `external_order`，状态为 `failed`。
   - `last_receipt` 写入 `充值账号命中风控：{reason}`。
   - 插入 `recharge_risk_record`。
   - 规则 `hit_count + 1`。
   - 不触发 worker。
   - 不创建 `external_order_attempt`。
5. 开放下单响应仍返回 `code=0`，`data.status_code=failed`，并返回订单号。

命中多条规则时按 `id ASC` 取第一条，保证行为稳定。当前不设计优先级字段，避免过早扩展。

## 错误处理

- 规则账号、商品关键词、原因为空：返回参数错误。
- 重复创建有效规则：返回参数错误，提示规则已存在。
- 状态值不是 `0/1`：返回参数错误。
- 记录列表时间格式错误：返回参数错误。
- 风控记录写入失败：订单创建事务整体失败，避免出现“订单失败但无风控流水”的不一致状态。
- token 脱敏失败不应阻断订单，使用空字符串兜底。

## 测试设计

合约测试：

- OpenAPI 暴露 `/api/admin/recharge-risks/rules` 和 `/api/admin/recharge-risks/records`。
- `api/` 仍保持扁平目录，新增 `api/recharge_risk.go`。
- 菜单种子包含 `order.recharge_risk`。
- 未授权用户访问风控接口返回 `403`。

业务/集成测试：

- 创建启用规则后，开放下单命中账号精确 + 商品关键词，订单直接 `failed`。
- 命中风控时不生成 `external_order_attempt`，也不会请求上游。
- 风控记录落库，规则 `hit_count` 累加。
- 停用规则不拦截。
- 规则列表支持账号、关键词、状态筛选。
- 记录列表支持账号、关键词、时间筛选。

交付验证命令：

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

## 文档同步

实现阶段需要同步：

- `docs/module-map.md`：新增充值风控业务域、路由、权限和边界。
- `docs/development.md`：说明开放下单风控命中流程。
- `docs/testing.md`：补充风控相关聚焦测试命令。
- `test/contract/README.md`：补充风控接口合约覆盖。
- `README.md`：仅在模块索引或常用验证命令需要调整时同步。
- `docs/superpowers/README.md`：登记本设计文档和后续实施计划。

## 验收标准

- 管理员能通过接口配置、筛选、编辑、启停和删除风控规则。
- 开放下单命中启用规则时，接口立即返回失败状态的订单号。
- 命中风控不会提交上游，也不会产生渠道尝试记录。
- 后台可查询风控拦截记录，并按账号、关键词、时间筛选。
- 订单记录可看到失败订单和风控原因。
- 全量测试、构建和 lint 验证通过，或在交付说明中明确未执行原因和风险。
