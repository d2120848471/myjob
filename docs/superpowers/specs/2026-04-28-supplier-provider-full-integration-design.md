# 全量供应商渠道对接设计

## 背景

当前商品渠道绑定、开放下单和商品同步已经具备 provider 适配器雏形，但订单履约、商品详情同步、商品订阅和商品推送实际只完整接入卡卡云。现有代码还在订单候选渠道查询中硬限制 `provider_code = 'kakayun'`，导致其他已维护的平台账号只能用于余额查询，不能参与下单、查单、同步价格和同步名称。

本设计覆盖卡卡云以外的全部已整理平台：卡易信、卡速售、星权益、优卡云、聚浪云、星海、飞速源。目标是在不改变开放订单协议的前提下，让这些平台参与下单、查单、商品名称同步、进货价同步和防亏损保护；平台不支持的能力按可观测降级处理。

## 目标

- 全部已整理供应商平台都能作为商品渠道候选参与开放订单履约。
- 全部平台尽量支持上游商品名称和进货价主动同步。
- 防亏损逻辑从卡卡云专属能力改为通用安全价能力，由 provider 按平台协议映射字段。
- 对单次只能提交 1 个数量的平台支持拆单，不因为本地订单数量大于 1 而直接放弃该渠道。
- 商品订阅列表保持卡卡云专属，不把不支持 API 订阅的平台塞进订阅记录。
- 不支持商品推送的平台依赖主动监控同步，不做伪订阅或不可验证的推送解析。

## 非目标

- 不扩展开放下单协议；第一版仍只接收商品编码、充值账号和数量。
- 不新增 SKU、面值、扩展参数等平台专属绑定字段。
- 不支持复杂多规格或需要额外充值模板字段的上游商品。
- 不实现上游订单状态回调；订单状态仍以现有 worker 轮询为主。
- 不改变后台订单列表和开放查单的响应结构。

## 方案

采用“能力矩阵 + provider 横向扩展”方案。保持现有 `OrderProvider`、`ProductInfoProvider`、`ProductSubscriptionProvider`、`ProductChangePushProvider` 边界，按平台新增适配器实现签名、请求构造、响应解析和状态映射。业务层只依赖标准化结果，不在订单或商品逻辑中解析各平台原始响应。

不新增无意义子目录；provider 仍放在 `internal/library/supplierplatform/provider` 同 package 下，按平台或职责拆分文件。订单拆单能力放在 `internal/logic/order` 下新增职责明确的文件，不塞入通用 `common.go`。

## 平台能力矩阵

| 平台 | 下单/查单 | 主动同步价格名称 | 防亏损字段 | 商品推送 | 订阅列表 |
| --- | --- | --- | --- | --- | --- |
| 卡卡云 | 已支持，保留 | 已支持，保留 | `maxmoney` 总额 | 已支持，保留 | 支持 |
| 卡易信 | 接入 | 商品详情 | `safePrice` 总额 | 文档字段明确时接入 | 不展示 |
| 卡速售 | 接入 | 商品详情 | `safe_price` 总额 | 文档字段明确时接入 | 不展示 |
| 星权益 | 接入 | 商品详情 | `safe_cost` 单价 | 文档字段明确时接入 | 不展示 |
| 优卡云 | 接入 | 商品详情 | `maxmoney` 总额 | 文档字段明确时接入 | 不展示 |
| 聚浪云 | 接入 | 商品详情 | `accessPrice` 总额 | 第一版不改价，靠主动同步 | 不展示 |
| 星海 | 接入 | 商品列表匹配 | `itemPrice` 单价 | 不支持 | 不展示 |
| 飞速源 | 接入 | 产品列表匹配 | 无上游字段，本地校验 | 不支持 | 不展示 |

## 订单履约设计

订单主流程保持现状：开放下单创建 `external_order`，worker 选择渠道绑定，提交上游，按轮询结果更新状态，失败时复用现有补单策略。

订单候选渠道查询需要移除卡卡云硬限制，改为选择所有注册了 `OrderProvider` 的平台账号。若某平台未注册订单 provider，则该账号不会成为可提交候选。

provider 新增平台能力描述，至少包含单次下单最大数量和安全价字段口径。默认单次下单数量不限制；飞速源返回 `1`。

提交前按 provider 能力拆分上游请求：

- 普通平台：一个本地 attempt 对应一个上游 segment。
- 飞速源：本地数量为 N 时拆成 N 个 segment，每个 segment 数量为 1。
- segment 的上游自传单号使用稳定后缀，例如 `O20260428120000123456-T1-S1`。

新增 `external_order_attempt_segment` 保存真实上游请求粒度的数据：attempt ID、segment 序号、segment 数量、上游单号、自传单号、状态、请求快照、响应快照和回执。

状态聚合规则：

- 全部 segment 成功：attempt 成功，订单成功。
- 任一 segment 处理中或未知：attempt 保持处理中或未知，订单继续轮询。
- 任一 segment 明确失败：attempt 失败，进入现有补单逻辑。

查单逻辑按 segment 执行。普通平台只查一个 segment；拆单平台逐个查 segment 后聚合状态。

## 防亏损设计

订单侧统一计算安全价，provider 只负责把标准化安全价映射成平台字段。

安全价输入仍来自现有本地数据：渠道绑定原始进货价、本地销售金额、是否允许亏本销售、允许亏本金额和本次 segment 数量。

总额口径平台使用 segment 总额安全价；单价口径平台使用 segment 单价安全价。不支持上游安全价字段的平台必须先通过本地防亏校验，校验失败不提交上游。

本地防亏校验规则：如果商品未开启允许亏本销售，上游绑定原始进货总额不得大于订单销售金额；如果开启，则不得大于订单销售金额加允许亏本金额。拆单时按 segment 数量拆分销售金额与成本金额，避免整体订单通过但单个 segment 明显亏损。

## 商品同步设计

`ProductGoodsChannelSyncWorker` 保持现有主动监控模式，扫描启用同步开关的渠道绑定。新增平台 `ProductInfoProvider` 后，现有同步入口即可覆盖更多平台。

同步规则保持现有语义：

- 上游商品编号和绑定编号不一致时跳过。
- 名称为空不覆盖本地名称。
- 价格无效但名称有效时只同步名称。
- 价格同步复用税价计算和自动改价记录逻辑。
- 利润配置不在同步流程中重写。

有商品详情接口的平台直接查详情；星海和飞速源使用列表接口拉取后按商品编号匹配。列表匹配必须限制在当前平台账号上下文中，避免跨账号缓存污染。

## 商品推送与订阅设计

卡卡云订阅、取消订阅、重新订阅和商品变动推送保持现状。

非卡卡云不写 `supplier_product_subscription` 记录，不展示在订阅列表，也不提供取消或重新订阅动作。运营如需开启上游商品变动通知，应在上游后台手动配置系统统一回调地址。

支持推送且字段明确的平台可以注册 `ProductChangePushProvider`，回调解析成标准 `ProductChangePushResult` 后复用现有改价逻辑。聚浪云文档未给出明确 payload 字段，第一版不做改价解析；星海和飞速源不支持商品推送，依赖主动同步。

## 数据库变更

新增 `external_order_attempt_segment` 表。核心字段：

- `id`
- `order_id`
- `attempt_id`
- `segment_no`
- `quantity`
- `provider_code`
- `supplier_goods_no`
- `supplier_order_no`
- `supplier_us_order_no`
- `supplier_status`
- `refund_status`
- `request_snapshot`
- `response_snapshot`
- `receipt`
- `status`
- `submitted_at`
- `last_checked_at`
- `created_at`
- `updated_at`

启动兜底建表或补索引逻辑放在订单 bootstrap 相关文件中，并对 MySQL 设置短锁等待，避免启动期长时间等待元数据锁。正式生产仍建议优先在维护窗口执行显式 DDL。

## 测试计划

provider 单测：

- 每个平台下单请求构造、签名和安全价字段。
- 每个平台下单响应状态映射。
- 每个平台查单响应状态映射。
- 每个平台商品信息解析。
- 支持推送的平台验签和 payload 解析。

订单逻辑测试：

- 非卡卡云渠道可成为候选渠道。
- 支持上游安全价的平台收到正确字段。
- 不支持上游安全价的平台执行本地防亏校验。
- 飞速源数量 N 拆为 N 个 segment。
- 全部 segment 成功后本地订单成功。
- 任一 segment 失败后进入现有补单路径。
- 任一 segment 未确认时本地订单保持待轮询。

商品同步测试：

- 多 provider 注册后可同步名称和价格。
- 星海、飞速源通过列表匹配同步。
- 非卡卡云绑定不会写订阅记录。
- 不支持推送的平台仍可通过主动同步改价。

验证命令在实现阶段至少执行：

```bash
go test ./internal/library/supplierplatform/provider -count=1 -timeout 60s
go test ./internal/logic/order -count=1 -timeout 60s
go test ./internal/logic/admin -run 'Test.*ProductGoodsChannel.*|Test.*SupplierProduct.*' -count=1 -timeout 60s
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

## 文档同步

实现阶段需要同步：

- `docs/development.md`：多平台 provider 接入规则、拆单 segment、防亏字段口径。
- `docs/module-map.md`：第三方对接、订单履约、商品同步能力描述。
- `docs/testing.md`：多平台 provider、拆单和主动同步聚焦测试命令。
- `test/contract/README.md`：如新增或调整合约测试口径，同步说明。

## 风险与处理

- 平台文档字段与真实响应不一致：provider 解析保持保守，未知响应按未确认处理，不直接失败。
- 拆单带来部分成功风险：segment 聚合以任一失败触发补单，保留每个 segment 快照用于人工核对。
- 列表同步可能成本较高：对星海和飞速源按平台账号和商品编号缓存单次同步结果，避免同一轮重复拉取。
- 不扩展开放下单协议会限制复杂上游商品：第一版只覆盖当前业务模型，复杂商品后续单独设计渠道绑定扩展字段。
