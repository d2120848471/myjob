# 卡卡云下单 maxmoney 防亏本设计

## 背景

当前卡卡云下单请求没有传 `maxmoney`。卡卡云文档说明该字段表示“最大进货总金额”，用于阻止上游实际进货金额超过商户允许范围，避免因上游调价导致亏本。

商品库存配置已经存在 `allow_loss_sale_enabled` 和 `max_loss_amount`，渠道绑定也已经保存 `source_cost_price`。本次设计把这些现有数据用于卡卡云下单防亏本，不新增数据库字段。

## 目标

- 卡卡云下单时必须在请求中携带 `maxmoney`。
- `maxmoney` 按订单总金额口径传递，不按单价传递。
- 防亏本计算以渠道绑定的 `source_cost_price` 为进货价口径。
- 商品不允许亏本时，允许亏损额视为 `0`。
- 商品允许亏本时，把 `max_loss_amount` 计入允许上限。
- 每次智能补单切换渠道时，都按当前渠道重新计算 `maxmoney`。

## 非目标

- 不改变商品库存配置接口和数据库结构。
- 不改变现有订单金额快照的展示口径。
- 不回算历史订单。
- 不把亏本保护扩展到非卡卡云 provider。
- 不在本次改动中重构订单 worker 或渠道选择策略。

## 已确认规则

`maxmoney` 的计算公式为：

```text
sourceTotal = source_cost_price * quantity
allowedLoss = allow_loss_sale_enabled == 1 ? max_loss_amount : 0
salesCeiling = order_amount + allowedLoss
maxmoney = min(sourceTotal, salesCeiling)
```

说明：

- `source_cost_price` 来自 `product_goods_channel_binding.source_cost_price`，表示卡卡云原始进货价。
- `quantity` 来自外部订单购买数量。
- `order_amount` 使用当前已选渠道提交时计算得到的订单销售总金额。
- `max_loss_amount` 是商品库存配置中的允许亏本金额，按订单总额计入，不按数量重复放大。
- `maxmoney` 保留 4 位小数。

## 架构与边界

本次改动采用“订单逻辑计算，provider 只透传协议字段”的方案。

- `internal/logic/order` 负责读取候选渠道和库存配置，并计算 `maxmoney`。
- `internal/library/supplierplatform/provider` 只负责把 `CreateOrderInput.MaxMoney` 写入卡卡云下单 payload。
- 卡卡云签名必须基于包含 `maxmoney` 的 payload 计算。
- provider 不感知 `allow_loss_sale_enabled`、`max_loss_amount`、订单金额快照等业务规则。

这样可以保持第三方适配器只承担协议转换，避免把商品库存配置规则耦合到 provider 层。

## 数据流

1. `loadCandidateChannels` 查询候选渠道时补充 `source_cost_price`。
2. `loadReorderConfig` 查询商品库存配置时补充 `allow_loss_sale_enabled` 和 `max_loss_amount`；如果没有配置记录，默认按不允许亏本处理。
3. 提交订单前，订单逻辑按当前候选渠道计算订单金额快照。
4. 订单逻辑使用金额快照、渠道 `source_cost_price`、库存配置计算 `maxmoney`。
5. `executeCreateOrder` 把 `MaxMoney` 放入 `CreateOrderInput`。
6. 卡卡云 provider 在 `MaxMoney` 非空时写入请求体字段 `maxmoney`。
7. 卡卡云 provider 使用包含 `maxmoney` 的 payload 生成签名。

补单时会重新选择候选渠道，因此第 3 到第 7 步会按新渠道重新执行。

## 错误处理

- `source_cost_price` 解析失败时，不允许继续向卡卡云下单，避免缺少防亏本保护。
- `order_amount` 解析失败时，沿用订单提交错误处理，不强行请求上游。
- `max_loss_amount` 解析失败时，沿用订单提交错误处理，不强行请求上游。
- `allow_loss_sale_enabled != 1` 时，`max_loss_amount` 一律视为 `0`。
- 计算出的 `maxmoney` 小于 0 时，按金额异常处理。
- 卡卡云返回“存在亏本”等业务失败时，沿用现有失败解析和智能补单逻辑。

## 文件计划

预计修改：

- `internal/logic/order/order_channel.go`
  - 候选渠道结构和查询补充 `source_cost_price`。
- `internal/logic/order/order_reorder.go`
  - 扩展库存配置快照字段。
- `internal/logic/order/order_loss_guard.go`
  - 放置 `maxmoney` 计算 helper，保持防亏本计算职责独立。
- `internal/logic/order/order_submit.go`
  - 提交前计算并填充 `CreateOrderInput.MaxMoney`。
- `internal/library/supplierplatform/provider/types.go`
  - `CreateOrderInput` 增加 `MaxMoney`。
- `internal/library/supplierplatform/provider/providers.go`
  - 卡卡云下单 payload 增加 `maxmoney` 并参与签名。

预计测试：

- `internal/library/supplierplatform/provider/providers_test.go`
  - 验证卡卡云下单请求包含 `maxmoney`。
  - 验证签名覆盖 `maxmoney`。
- `internal/logic/order/order_loss_guard_test.go`
  - 覆盖不允许亏本、允许亏本和取最小值规则。
- `test/integration/order_worker_test.go`
  - 验证订单提交到卡卡云时传递正确 `maxmoney`。
  - 验证允许亏本时使用 `min(sourceTotal, orderAmount + maxLossAmount)`。

预计文档：

- `docs/development.md`
  - 订单金额快照或供应商对接说明中补充卡卡云 `maxmoney` 规则。
- `docs/module-map.md`
  - 订单履约或第三方对接边界中补充防亏本保护。
- `docs/testing.md`
  - 补充卡卡云下单防亏本聚焦回归命令。

## 验证计划

聚焦验证：

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProvider -count=1 -timeout 60s
go test ./internal/logic/order -count=1 -timeout 60s
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

交付前全量验证：

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

## 验收标准

- 卡卡云创建订单请求体包含 `maxmoney`。
- `maxmoney` 使用 `source_cost_price` 计算，不使用 `cost_price`。
- 不允许亏本时，`maxmoney = min(source_cost_price * quantity, order_amount)`。
- 允许亏本时，`maxmoney = min(source_cost_price * quantity, order_amount + max_loss_amount)`。
- 卡卡云签名包含 `maxmoney`。
- 智能补单时，第二个渠道使用自己的 `source_cost_price` 重新计算 `maxmoney`。
- 现有订单金额快照字段行为不被改变。

## 风险与约束

- `max_loss_amount` 按订单总额计入；如果未来产品希望按单件亏损额度处理，需要单独调整配置语义和文档。
- 卡卡云文档写的是“最大进货总金额”，实现必须保持总额口径。
- 如果历史渠道绑定缺少有效 `source_cost_price`，相关订单会在提交前失败或进入现有异常处理，需要运营先补齐渠道进货价。
