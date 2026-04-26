# 卡卡云商品价格推送与订阅设计

日期：2026-04-26

## 背景

商品渠道绑定已经支持卡卡云商品详情定时同步：当商品库存配置打开 `sync_cost_price_enabled` 后，系统会把卡卡云 `goodsprice` 同步到渠道绑定的 `source_cost_price`，再按税态计算 `cost_price`。如果绑定本身打开 `is_auto_change`，现有价格模型会基于 `cost_price` 和利润规则计算 `effective_sell_price`。

当前缺口是：卡卡云商品价格变化只能依赖定时监控，不能接收上游实时推送；也没有订阅记录列表和改价记录列表。卡卡云文档支持商品信息推送模式：设置接收 URL、订阅商品、取消订阅商品，商品价格、库存和状态变化后会向接收 URL 推送。

本次一期只实现卡卡云。回调 URL 要保持通用，后续其他平台复用同一 URL 形态。

## 已确认口径

1. 定价口径沿用现有 A 方案：上游价格变化只同步进价，不直接修改商品主档 `default_sell_price`。
2. `effective_sell_price` 是渠道绑定按自动改价规则计算出的可售价格，不是单独落库的商品主档售价。
3. 只有商品库存配置 `sync_cost_price_enabled = 1` 时，才同步上游进价。
4. 同步进价开启后，无论绑定是否打开自动改价，都更新进价；绑定 `is_auto_change = 1` 时，`effective_sell_price` 会跟随成本自动变化；未打开自动改价时只是同步进货价。
5. 改价记录只记录价格真的发生变化的情况；推送和定时监控两种来源都要记录。
6. 不新增公网域名系统参数。部署到公网后，由实际请求域名承载回调地址。
7. 通用回调 URL 使用 `providerCode + platformAccountId` 区分第三方渠道账号：

```text
/api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback
```

8. 后续新增或编辑卡卡云渠道绑定时自动订阅；历史已有绑定不主动批量补订阅。
9. 自动订阅失败不阻断渠道绑定保存，只记录订阅失败状态和失败原因。
10. 取消订阅成功后，本地订阅记录保留，状态改为已取消，不物理删除。

## 目标

1. 增加通用上游商品变动回调入口，一期支持卡卡云价格推送。
2. 卡卡云推送改价后，在满足同步进价开关的前提下立即同步本地渠道进价和比较成本价。
3. 定时监控和推送同步都写入统一的自动改价记录表。
4. 新增卡卡云商品订阅记录列表，能查询已订阅、失败和已取消的记录。
5. 新增取消订阅接口，允许后台取消已订阅的卡卡云商品。
6. 新增重新订阅能力，便于订阅失败或已取消后在列表中恢复订阅。
7. 新增卡卡云 provider 订阅、取消订阅、接收 URL 检查和商品推送验签解析能力。

## 非目标

- 不实现其他 provider 的商品推送和订阅。
- 不新增公网域名配置项。
- 不直接改商品主档 `default_sell_price`。
- 不保存或同步卡卡云库存、上下架状态到现有商品绑定状态。
- 不为历史已有卡卡云绑定做启动扫描补订阅。
- 不重构现有订单履约价格模型。
- 不把订阅失败变成渠道绑定保存失败。

## 现状依据

- 商品渠道绑定协议在 `api/product_goods_channel.go`，包含 `supplier_goods_no`、`source_cost_price`、`cost_price`、`is_auto_change`、`add_type`、`default_price`。
- 商品库存配置协议在 `api/product_goods_channel_config.go`，包含 `sync_cost_price_enabled`。
- 定时同步入口在 `internal/logic/admin/product_goods_channel_sync.go`，已经能查询卡卡云商品详情并更新进价。
- 成本价计算在 `internal/logic/admin/product_goods_channel_price.go`，应继续复用 `computeChannelCostSnapshot`。
- `effective_sell_price` 由 `internal/library/channelpricing/pricing.go` 根据默认售价或自动改价规则计算。
- 卡卡云 provider 位于 `internal/library/supplierplatform/provider`，已有余额、下单、查单和商品详情能力。
- 卡卡云文档的推送接口要求签名校验成功后返回纯文本 `ok`。

## 架构设计

### 通用回调入口

新增开放回调接口：

```text
POST /api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback
```

该接口不依赖后台登录态。处理流程：

1. 读取路径中的 `providerCode` 和 `platformAccountId`。
2. 查询 `supplier_platform_account`，要求账号存在、未删除、启用，并且账号 `provider_code` 与路径一致。
3. 根据 `providerCode` 查找商品变动推送 provider。
4. 使用平台账号密钥验签并解析推送内容。
5. 调用商品渠道价格同步逻辑。
6. 卡卡云成功处理后写入纯文本 `ok`，由响应中间件识别已写入响应并跳过 JSON 包装。

卡卡云推送中没有稳定的商户 ID 字段用于反查账号，因此 `platformAccountId` 必须出现在 URL 中。这样同一个系统配置多个卡卡云账号时不会混淆密钥和订阅关系。

### Provider 能力

在 `internal/library/supplierplatform/provider` 增加商品订阅与推送接口：

```text
ProductSubscriptionProvider
ProductChangePushProvider
```

卡卡云一期实现：

- 获取接收 URL 列表：`/dockapiv3/user/geturl`
- 设置接收 URL：`/dockapiv3/user/seturl`
- 订阅商品：`/dockapiv3/goods/subscribe`
- 取消订阅商品：`/dockapiv3/goods/cancelsubscribe`
- 推送验签和解析：商品信息变动推送 payload

订阅前先读取卡卡云接收 URL 列表：

- 目标 URL 已存在：不重复设置 URL，直接订阅商品。
- 目标 URL 不存在：调用 `seturl` 新增。
- 卡卡云接收 URL 已满或设置失败：订阅状态记为失败，不阻断本地绑定保存。

### 回调 URL 组装

不新增系统域名参数。新增/编辑渠道绑定发生在后台 HTTP 请求内，因此订阅时从当前请求组装公网回调地址：

1. 优先使用 `X-Forwarded-Proto` 和 `X-Forwarded-Host`。
2. 其次使用请求 TLS 状态和 `Host`。
3. 拼接通用路径：

```text
{scheme}://{host}/api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback
```

如果当前上下文无法获得请求域名，则订阅动作标记失败，失败原因写明“无法构造回调 URL”。这不会影响渠道绑定保存。

### 商品渠道价格同步

新增或调整商品渠道同步内部方法，使推送和定时监控复用同一套更新逻辑：

```text
上游 goodsprice
  -> source_cost_price
  -> cost_price / tax_adjust_*
  -> effective_sell_price
```

更新约束：

- 商品必须未删除、启用、渠道供货。
- 渠道绑定必须未删除，且 `supplier_goods_no` 与推送 `goodsid` 匹配。
- 平台账号必须未删除、启用，且账号 ID 与 URL 中一致。
- 商品库存配置 `sync_cost_price_enabled = 1` 才允许同步进价。
- 卡卡云价格非法、为空或负数时，不覆盖本地价格。
- 税率配置缺失时，不覆盖价格字段。
- 只在 `source_cost_price` 或 `cost_price` 经 4 位小数归一化后发生变化时写改价记录。

## 数据设计

### 订阅记录表

建议表名：

```text
supplier_product_subscription
```

字段：

```text
id
provider_code
platform_account_id
platform_account_name
goods_id
binding_id
supplier_goods_no
supplier_goods_name
callback_url
status
last_action
last_error
request_snapshot
response_snapshot
subscribed_at
canceled_at
created_at
updated_at
```

状态枚举：

```text
subscribed
failed
canceled
```

动作枚举：

```text
subscribe
resubscribe
cancel
```

唯一约束建议：

```text
provider_code + platform_account_id + supplier_goods_no
```

口径：

- 一条平台账号商品订阅对象保留一条当前记录。
- 新增/编辑绑定触发订阅时，若记录存在则更新状态、快照、时间和关联绑定信息。
- 取消订阅成功后状态改为 `canceled`，记录保留。
- 重新订阅成功后状态改回 `subscribed`。

### 自动改价记录表

建议表名：

```text
product_goods_channel_price_change_log
```

字段：

```text
id
source
provider_code
platform_account_id
platform_account_name
binding_id
goods_id
goods_code
goods_name
goods_icon
supplier_goods_no
supplier_goods_name
old_source_cost_price
new_source_cost_price
old_cost_price
new_cost_price
old_effective_sell_price
new_effective_sell_price
change_amount
description
raw_payload
changed_at
created_at
```

来源枚举：

```text
monitor
push
```

记录口径：

- 推送和定时监控都写入该表。
- 只记录价格真实变化，不记录无变化推送或无变化扫描。
- `change_amount = new_effective_sell_price - old_effective_sell_price`。
- 未开启绑定自动改价时，`old_effective_sell_price` 和 `new_effective_sell_price` 仍按现有规则等于商品默认售价；描述中写明“未开启自动改价，仅同步进价”。
- `description` 组装给前端展示的说明文本，包括来源、平台、上游商品、变价前后、税价调整、自动改价规则和是否触发利润后价格变化。

## 核心流程

### 新增或编辑渠道绑定自动订阅

```text
保存渠道绑定成功
  -> 判断平台账号 provider_code 是否为 kakayun
  -> 从当前请求组装通用回调 URL
  -> 调用卡卡云 geturl 检查 URL 是否已设置
  -> 必要时调用 seturl 新增接收 URL
  -> 调用 goods/subscribe 订阅 supplier_goods_no
  -> 写入或更新 supplier_product_subscription
```

订阅失败不回滚渠道绑定事务。因为渠道绑定是本地业务主流程，订阅只是上游推送增强能力；网络失败、卡卡云限流或接收 URL 数量限制不能阻止本地商品维护。

### 取消订阅

```text
后台调用取消订阅接口
  -> 查询本地订阅记录
  -> 校验 provider_code = kakayun
  -> 调用卡卡云 cancelsubscribe
  -> 成功后本地状态改为 canceled
  -> 失败时保留原状态并写 last_error
```

取消订阅不删除本地记录。列表可以继续展示取消时间和最近一次响应，方便追查。

### 重新订阅

```text
后台调用重新订阅接口
  -> 查询本地订阅记录和平台账号
  -> 重新组装当前公网回调 URL
  -> 确保卡卡云接收 URL 存在
  -> 调用 goods/subscribe
  -> 成功后状态改为 subscribed
```

重新订阅用于失败恢复和已取消订阅恢复。它是后台显式操作，操作失败时需要更新 `status = failed`、`last_error` 和请求/响应快照，并向调用方返回业务错误，避免运营误判为已恢复订阅。

### 上游推送改价

```text
卡卡云 POST 通用回调
  -> 用 platformAccountId 找平台账号
  -> providerCode 必须匹配 kakayun
  -> 用该账号 secret_key 验签
  -> 校验 timestamp 有效期
  -> 解析 goodsid / goodsname / goodsprice
  -> 查找该平台账号 + goodsid 下的有效绑定
  -> 只处理 sync_cost_price_enabled = 1 的绑定
  -> 更新 source_cost_price / cost_price / tax_adjust_*
  -> 价格真实变化时写 source = push 的改价记录
  -> 返回 ok
```

返回策略：

- 验签失败、账号不存在、provider 不匹配：不返回 `ok`。
- 有效签名但没有绑定、同步开关关闭或价格无变化：返回 `ok`，避免卡卡云无意义重试。
- 目标绑定应更新但数据库保存失败：不返回 `ok`，让卡卡云按其重试策略再次推送。

### 定时监控记录改造

现有定时同步器保留。需要把“应用上游商品信息”的内部方法增加来源参数：

```text
source = monitor
source = push
```

定时监控发现价格变化时写 `source = monitor` 的改价记录。推送与监控共用同一张日志表和同一套价格差异判断。

## API 设计

### 开放接口

```text
POST /api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback
```

返回：

- 成功：纯文本 `ok`
- 失败：普通 HTTP 错误响应，不返回 `ok`

### 管理端订阅接口

```text
GET  /api/admin/supplier-product-subscriptions
POST /api/admin/supplier-product-subscriptions/{id}/cancel
POST /api/admin/supplier-product-subscriptions/{id}/resubscribe
```

列表筛选：

```text
商品名称
渠道商品编号
渠道账号
订阅状态
时间段
```

列表展示：

```text
商品名称
商品图标
货源平台
货源商品编号
订阅时间
订阅状态
失败原因
操作
```

权限建议使用 `supplier.index`，因为订阅动作依赖第三方平台账号。

### 管理端改价记录接口

```text
GET /api/admin/product-goods-channel-price-changes
```

列表筛选：

```text
来源类型 monitor / push
本地商品编号或名称
对接商品编号
渠道账号
时间段
```

列表展示：

```text
商品名称
商品图标
货源平台
类型
描述
变动前
变动后
变化
变动时间
```

权限建议使用 `product.goods`，因为记录反映商品渠道价格变化。

## 错误处理与幂等

- 推送验签失败不处理业务，不返回 `ok`。
- 卡卡云推送时间戳超过有效期不处理业务，不返回 `ok`。
- 推送商品不存在绑定时只记录日志并返回 `ok`。
- 同一推送被重复投递时，若本地价格已更新到相同值，不再写重复改价记录。
- 订阅 URL 已存在时不重复调用 `seturl`。
- 订阅或取消订阅失败时，保留上游请求和响应快照，便于排查。
- 改价记录写入失败时，不应阻断价格更新；但需要写应用日志。价格更新是主业务结果，日志失败属于可追踪但不回滚的辅助失败。

## 安全设计

- 开放回调入口必须根据 `platformAccountId` 读取账号密钥，并使用 provider 规则验签。
- 路径 `providerCode` 必须与账号 `provider_code` 一致。
- 不允许通过请求体中的任意字段决定使用哪个平台账号。
- 推送 payload 的 `sign` 不落入签名原文。
- 空值字段不参与卡卡云签名，沿用卡卡云文档规则。
- 接收 URL 组装只使用请求 host 和代理头，不读取用户提交的任意 URL 参数。

## 测试设计

### Provider 单测

- 卡卡云 `geturl` 请求、签名和响应解析。
- 卡卡云 `seturl` 请求、签名和响应解析。
- 卡卡云 `goods/subscribe` 请求支持单个商品。
- 卡卡云 `goods/cancelsubscribe` 请求支持单个商品。
- 卡卡云商品推送验签成功。
- 卡卡云商品推送验签失败。
- 商品推送价格解析为 4 位小数。

### Logic 单测

- 新增卡卡云渠道绑定后自动订阅成功并落库。
- 自动订阅失败不阻断绑定保存，并记录失败状态。
- 推送改价时 `sync_cost_price_enabled = 1` 才更新价格。
- 推送改价时绑定 `is_auto_change = 1` 会产生新的 `effective_sell_price`。
- 未开启自动改价时只同步进价，`effective_sell_price` 仍按默认售价计算。
- 价格无变化时不写改价记录。
- 定时监控同步价格变化时写 `source = monitor`。
- 上游推送同步价格变化时写 `source = push`。
- 取消订阅成功后状态变为 `canceled`。
- 重新订阅成功后状态变为 `subscribed`。

### Contract 测试

- OpenAPI 暴露新增管理端订阅列表、取消订阅、重新订阅和改价记录列表。
- 开放回调路由存在。
- 卡卡云回调成功时返回纯文本 `ok`，不会被包装为 `{code,message,data}`。
- 新增 `api/*.go` 文件后，`test/contract/api_layout_test.go` 同步更新。

### Schema 测试

- `manifest/sql/*.sql` 和 `internal/app/schema.go` 都包含两张新表。
- MySQL 表和字段都有注释。
- SQLite 测试 schema 与 MySQL schema 字段保持一致。
- 启动补表逻辑能为旧库补齐新表。

### 验证命令

交付前至少执行：

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

聚焦回归建议：

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestProductGoodsChannelSubscription|TestProductGoodsChannelPriceChange' -count=1 -timeout 60s
go test ./test/contract -run 'TestProductGoodsChannelPriceChange|TestSupplierProductSubscription|TestOpenSupplierProductChangeCallback' -count=1 -timeout 60s
```

## 文档同步

实现时需要同步：

- `docs/module-map.md`：商品管理增加改价记录；第三方对接增加卡卡云订阅记录和商品推送。
- `docs/development.md`：补充新增 provider 订阅和推送能力的开发约定。
- `docs/testing.md`：补充卡卡云订阅、推送和改价记录聚焦测试命令。
- `docs/superpowers/README.md`：增加本设计文档索引。
- `test/contract/README.md`：如新增 API 协议文件被布局测试锚定，需要同步说明。

## 验收标准

1. 新增或编辑卡卡云渠道绑定后，系统会尝试订阅该上游商品。
2. 订阅失败不会导致渠道绑定保存失败，订阅列表能看到失败原因。
3. 订阅列表能按商品、渠道商品编号、渠道账号、状态和时间查询。
4. 后台可以取消已订阅的卡卡云商品，取消后本地记录保留为已取消。
5. 后台可以重新订阅失败或已取消的卡卡云商品。
6. 卡卡云推送签名正确且商品开启同步进价时，本地渠道进价会立即更新。
7. 商品未开启同步进价时，卡卡云推送不会覆盖本地进价。
8. 绑定开启自动改价时，价格变化后 `effective_sell_price` 按已有利润规则变化。
9. 系统不会直接修改商品主档 `default_sell_price`。
10. 推送和定时监控都能写入自动改价记录。
11. 价格无变化时不写自动改价记录。
12. 卡卡云回调成功返回纯文本 `ok`。
13. 多个卡卡云平台账号共存时，回调 URL 中的 `platformAccountId` 能准确区分账号和密钥。

## 风险与约束

- 回调 URL 从当前请求 host 组装，部署时反向代理必须正确传递 `Host` 或 `X-Forwarded-Host`、`X-Forwarded-Proto`。
- 卡卡云最多可设置 3 条接收 URL；如果同一账号已配置满，自动订阅会失败但不影响本地绑定。
- 历史已有绑定不自动补订阅，只有后续新增或编辑的卡卡云绑定会触发订阅。
- 推送和定时监控可能先后处理同一价格变化，必须用价格差异判断避免重复改价记录。
- 改价会影响后续订单的渠道成本和利润后价格，不回写历史订单金额快照。
- 卡卡云推送还包含库存和上下架状态，本次不处理这些字段，后续若要接入需另行设计。
