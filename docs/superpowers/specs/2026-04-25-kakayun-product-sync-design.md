# 卡卡云商品信息定时同步设计

日期：2026-04-25

## 背景

商品渠道绑定已经保存了对接商品编号、对接商品名、原始进货价、比较成本价和单条自动改价规则。商品维度库存配置也已经有 `sync_cost_price_enabled` 和 `sync_goods_name_enabled` 两个开关，但这两个开关目前只保存和回显，没有真实同步卡卡云商品信息。

卡卡云文档提供商品详情接口 `/dockapiv3/goods/details`，返回 `goodsname` 和 `goodsprice` 等字段。因此本次要把现有开关接到定时同步链路：用户打开同步进价或同步商品名称后，系统每分钟自动同步卡卡云商品信息；不开关时保持人工维护值。

## 目标

1. 建立通用供应商商品信息查询能力，先实现卡卡云 provider。
2. 每分钟执行一次后台同步任务，只处理打开了同步开关的商品渠道绑定。
3. `sync_cost_price_enabled=1` 时同步卡卡云 `goodsprice` 到绑定原始进货价，并重新计算比较成本价和税额快照。
4. `sync_goods_name_enabled=1` 时同步卡卡云 `goodsname` 到绑定对接商品名称。
5. 关闭对应开关时，系统不覆盖该字段的人工填写值。
6. 同步进价后，现有商品列表最低成本和绑定自动改价计算自然使用新的成本价。

## 非目标

- 不接卡卡云商品订阅推送和公网回调。
- 不新增库存数字段，不保存 `goodsstock`。
- 不按卡卡云 `goodsstatus` 自动关闭绑定或商品。
- 不新增前端接口或新的开关字段；继续使用现有库存配置开关。
- 不实现其他渠道的商品信息同步；只预留 provider 接口，其他渠道本次跳过。
- 不把同步失败直接展示到前端；本次先写应用日志，不新增同步日志表。

## 现状依据

- 商品渠道绑定协议在 `api/product_goods_channel.go`，新增和编辑绑定会保存 `supplier_goods_name`、`source_cost_price`、`cost_price` 等字段。
- 库存配置协议在 `api/product_goods_channel_config.go`，已经包含 `sync_cost_price_enabled` 和 `sync_goods_name_enabled`。
- 后台商品逻辑已经按职责拆在 `internal/logic/admin/product_goods_channel*.go` 和 `internal/logic/admin/product_goods_channel_config*.go`。
- 卡卡云 provider 当前只实现余额、下单和查单能力，位于 `internal/library/supplierplatform/provider/providers.go`。
- 卡卡云文档 `platform_docs/kakayun.md` 中的商品详情接口返回 `goodsname`、`goodsprice`、`stock`、`goodsstatus` 等字段。

## 架构设计

### Provider 能力

在 `internal/library/supplierplatform/provider` 增加供应商商品信息查询接口，例如：

- `ProductInfoProvider`：定义商品详情请求构建与响应解析能力。
- `ProductInfoInput`：包含上游商品编号。
- `ProductInfoResult`：包含标准化商品编号、商品名、进货价、原始响应。

卡卡云 provider 实现该接口：

- 请求地址：`/dockapiv3/goods/details`。
- 请求字段：`userid`、`timestamp`、`goodsid`、`sign`。
- 签名沿用卡卡云现有签名规则。
- 响应成功码为 `code=1`。
- `data.goodsname` 映射为标准商品名。
- `data.goodsprice` 映射为标准进货价，并统一格式化为 4 位小数。

新增 `LookupProductInfo(code string)` 注册表。其他 provider 不注册商品信息能力，因此同步器会识别为不支持并跳过。

### 后台同步逻辑

在 `internal/logic/admin` 新增更具体的同步文件，例如 `product_goods_channel_sync.go`，不扩大 `common.go` 或 `helper.go`。

同步逻辑负责：

- 查询需要同步的绑定。
- 按 provider 能力调用商品详情接口。
- 根据库存配置开关决定是否覆盖进价和名称。
- 价格变化时复用现有税价计算逻辑重新生成 `cost_price`、`tax_adjust_direction`、`tax_adjust_rate`、`tax_adjust_amount`。
- 单条失败不影响整轮任务。

查询范围：

- 商品未删除，供货方式为渠道供货。
- 商品库存配置中 `sync_cost_price_enabled=1` 或 `sync_goods_name_enabled=1`。
- 绑定未删除，绑定有 `supplier_goods_no`。
- 渠道账号未删除且启用。
- provider 支持商品信息查询；本次实际只有卡卡云。

### 定时任务

新增轻量后台同步器，启动时由应用装配层挂载，每 60 秒触发一次。

同步器规则：

- 上一轮未结束时，下一轮直接跳过，避免并发堆积。
- 每轮限制最大处理数量，防止一次同步过多绑定拖慢后台。
- 同一轮内按 `platform_account_id + supplier_goods_no` 去重请求，同一个卡卡云商品信息复用结果。
- 卡卡云商品详情接口有限速，同步器不做无限并发；实现时按串行或低并发处理。

本次不新增配置项，60 秒周期作为固定行为实现。后续如果需要可再把周期和批量大小纳入系统配置。

## 数据流

1. 用户在商品库存配置中打开 `sync_cost_price_enabled` 或 `sync_goods_name_enabled`。
2. 定时同步器每分钟扫描符合条件的绑定。
3. 同步器根据绑定的供应商账号找到商品信息 provider。
4. 卡卡云 provider 请求 `/dockapiv3/goods/details`。
5. 同步器解析标准商品信息并按开关更新：
   - 只开同步名称：仅更新 `supplier_goods_name`。
   - 只开同步进价：更新 `source_cost_price`，并重新计算成本快照。
   - 两个都开：同时更新名称和价格相关字段。
6. 商品列表和渠道绑定列表继续使用已有查询逻辑，自动体现新的名称、成本和利润后价格。

## 错误处理

- 上游请求失败、超时、非 JSON、业务失败：跳过本条并记录日志。
- 卡卡云返回空商品名：不覆盖原有名称。
- 卡卡云返回空价格或价格格式非法：不覆盖原有价格。
- 税率配置缺失或非法：价格同步跳过，名称同步仍可继续。
- provider 不支持商品信息查询：跳过，不作为业务错误。
- 数据库更新失败：记录错误，继续处理后续绑定。

失败场景必须坚持“不用坏数据覆盖人工值”的原则。

## 测试设计

### Provider 单测

覆盖卡卡云商品详情：

- 请求路径、请求体、签名字段正确。
- 成功响应解析出商品名和 4 位小数进价。
- 业务失败、非 JSON、缺失 data、非法价格返回错误。

### 逻辑测试

覆盖同步开关：

- 只开同步名称时，只覆盖 `supplier_goods_name`。
- 只开同步进价时，只覆盖 `source_cost_price` 和成本快照。
- 两个都开时同时覆盖。
- 两个都关时不进入同步。
- 上游异常时原字段保持不变。
- 税率缺失时价格不同步，但名称仍可同步。

### 集成测试

使用 `httptest` 模拟卡卡云 `/dockapiv3/goods/details`：

- 定时同步入口调用后，卡卡云绑定能被更新。
- 不支持商品信息 provider 的渠道会被跳过。
- 重复执行不会产生并发堆积或破坏数据。

### 回归验证

交付前执行：

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

如果本地环境缺少 lint 工具或外部依赖导致验证无法执行，最终说明必须写明未执行项、原因和风险。

## 文档同步

实现时需要同步：

- `docs/module-map.md`：商品管理和第三方对接增加供应商商品信息同步说明。
- `docs/development.md`：补充新增 provider 商品信息能力的开发流程。
- `docs/testing.md`：补充卡卡云商品信息同步的聚焦测试命令。

本设计不新增配置项、不改变启动方式、不改变 API 文件名，因此不需要更新根 `README.md`。

## 验收标准

1. 打开同步进价后，卡卡云商品价格变化会在下一轮同步中更新到绑定原始进货价和比较成本价。
2. 打开同步名称后，卡卡云商品名变化会在下一轮同步中更新到绑定对接商品名。
3. 关闭任一同步开关后，对应字段不会被定时任务覆盖。
4. 卡卡云请求失败或返回非法数据时，原绑定数据保持不变。
5. 其他 provider 未实现商品信息查询时，不影响卡卡云同步和订单履约。
6. 现有自动改价规则无需改表；同步成本价后，已有利润后价格计算自动反映新成本。

## 风险与约束

- 卡卡云商品详情接口有限速，不能无限并发同步；实现必须限制每轮处理量。
- 同步价格会影响后续订单成本和列表利润后价格，但不会回写历史订单。
- 税率配置缺失会阻止价格安全重算，因此价格同步需要跳过而不是使用不完整成本。
- 本次不做前端提示和同步日志表，排查同步失败主要依赖应用日志。
