# 商品渠道绑定、真实渠道下单与对外下单接口需求文档（最终版）

## 1. 文档目的

本文档用于明确 `myjob` 一期需要落地的三类能力：

1. 商品与渠道账号的绑定管理。
2. 基于绑定关系的真实渠道下单、异步查单、自动补单。
3. 面向下游调用方的对外下单与查单接口。

本文档是需求文档，不是实现文档。重点是把功能语义、数据口径、价格规则、接口契约、状态流转和验收标准写清楚，后续开发设计、表结构迁移、Provider 适配、前端页面重构都必须以本文档为上游输入。

### 1.1 当前已落地范围

截至 2026-04-18，仓库内已经先落地“商品渠道绑定最小闭环”，当前实现范围固定为：

- 商品列表渠道摘要展示
- 商品渠道绑定弹窗列表
- 绑定表单选项读取
- 单条绑定新增、编辑、删除
- 单条自动改价编辑

当前尚未落地：

- 真实渠道下单
- 上游价格通知
- 商品级渠道配置
- 批量启停、批量删除、批量自动改价
- 一键排序或拖拽排序

后续如果继续推进真实下单、回调、补单和开放接口，应继续以本文件剩余章节为上游需求。

---

## 2. 背景与现状

### 2.1 当前仓库已具备的基础能力

当前仓库已经具备以下基础能力：

- 商品主档管理：`product_goods`
- 商品模板管理：`product_template`
- 商品购买数量限制策略：`product_purchase_limit_strategy`
- 第三方平台账号管理：`supplier_platform_account`
- 第三方平台余额刷新：`supplier_platform_balance_log`
- 第三方平台余额 Provider 适配层

### 2.2 当前仓库尚未具备的能力

当前仓库尚未具备以下能力：

- 商品级渠道绑定表
- 商品级渠道配置表或等价扩展字段
- 商品列表“渠道”摘要字段
- 商品级绑定配置、排序、批量启停、批量删除
- 绑定级自动改价配置与价格通知接收
- 渠道真实下单 Provider
- 渠道异步查单与上游回调接收
- 对外订单接口
- 对外调用方主数据与签名鉴权
- 内部交易订单表和订单尝试日志表

### 2.3 参考页面与本期定位

参考页面路径为：

- `#/product/goodslist`

已确认参考页面存在以下业务能力：

- 商品列表直接展示渠道摘要。
- 点击“渠道”进入商品级渠道绑定弹窗。
- 弹窗顶部展示商品级配置摘要。
- 弹窗支持修改配置、新增库存、一键排序、批量启停、编辑自动改价、批量删除。
- 绑定列表展示名称、对接状态、对接编号、进货价、自动改价、排序等字段。

### 2.4 本期重要决策

本期已明确以下决策，这些结论必须在后续开发中保持一致：

1. **前端允许按新 REST 接口重构，不兼容旧页面接口名。**
2. **旧接口 `joinGoodsSetSwitch` 不再沿用。**
   - 商品级配置、绑定基础字段、绑定自动改价配置，统一拆成新的 REST 资源。
3. **所有订单都按异步订单处理。**
   - 建单成功只代表“已受理”或“已提交上游”。
   - 最终成功与否必须通过“上游回调”和“定时查单”确认。
4. **上游回调纳入一期。**
   - 本期做“上游回调接收与处理”。
   - 本期不做“下游回调订阅中心 / 下游回调重试中心”。
5. **补单时间按分钟配置，语义是“当前尝试等待超时后是否切下一渠道”。**
   - 不是整张订单的总生命周期窗口。
6. **税点按百分比换算，顺序固定为“先税态换算，再加利润”。**
7. **页面展示的“进货价”，显示税态换算后的比较成本。**
8. **售价计算公式固定为：**
   - 比较成本价 `cost_price`
   - 加固定利润或百分比利润
   - 百分比利润的基数也是 `cost_price`
9. **自动改价为绑定级能力。**
   - 上游价格变动通过 POST 通知本系统。
   - 系统接收通知后更新绑定成本，并在开启自动改价时按规则重算售价。
10. **自动改价中的 `default_price` 字段语义固定为“利润值”。**
    - `add_type = fixed` 时表示固定利润金额。
    - `add_type = percent` 时表示利润百分比。
    - 后台字段名可继续沿用 `default_price`，但前端自动改价弹窗文案必须展示为“利润值”或“加价值”，不得展示为“默认售价”。
11. **商品绑定以人工维护为准。**
    - `supplier_goods_no`、`supplier_goods_name`、`source_cost_price` 支持人工录入。
    - 后台新增/编辑绑定时不主动调用 `CatalogProvider` 对上游商品做保存前校验。
12. **当 `quantity > 1` 时，允许系统按 Provider 能力拆成多次上游建单。**
    - 该规则属于同一主订单的数量拆分执行。
    - 不等同于组合商品的一单多渠道子单编排。

### 2.5 本期边界

本文档一期目标是“下单闭环”，即：

- 管理端可配置商品绑定和选路规则。
- 服务端可真实请求上游渠道下单。
- 服务端可异步查单、接收上游回调、在超时后自动补单。
- 下游调用方可通过开放接口下单与查单。

一期明确不包含以下内容：

- 退款
- 售后
- 手工补单后台页面
- 下游回调订阅中心
- 下游回调重试中心
- 自动改价历史页
- 价格推送历史页面
- 组合商品一单多渠道子单编排流程
- 调用方管理后台页面

补充说明：

- 本期只覆盖 `供货方式 = 渠道` 的商品。
- 本地库存、卡密库存、预售、代充等其他供货方式不纳入本文档范围。
- 组合商品在一期仅表示“模板复杂”，不表示“组合商品一单多渠道拆分下单”；但当 `quantity > 1` 时，允许系统基于 Provider 能力做数量拆单执行。

---

## 3. 术语定义

### 3.1 商品

平台内部售卖的标准商品，主档来自 `product_goods`。

### 3.2 渠道账号

指一个可访问第三方平台的账号配置，主档复用 `supplier_platform_account`。

### 3.3 商品绑定

指“某个商品”与“某个渠道账号下某个上游商品”的绑定关系。

### 3.4 对接商品

上游渠道平台中的商品。由以下核心信息唯一标识：

- `platform_account_id`
- `supplier_goods_no`

### 3.5 对外调用方

通过开放接口向本系统发起下单请求的外部系统或客户。

### 3.6 尝试

订单命中某一条绑定后，发起一次上游建单请求的过程，称为一次 `attempt`。

### 3.7 补单

某次尝试失败，或某次尝试在等待窗口内未得到最终结果时，系统自动切换到下一条可用绑定继续下单。

### 3.8 路由模式

商品在存在多个有效绑定时，系统如何选择首个绑定，以及后续补单时如何选择剩余绑定的策略。

### 3.9 原始进货价

绑定录入或上游通知的原始进货价，字段建议为：

- `source_cost_price`

### 3.10 比较成本价

绑定按税态换算后的标准成本价，字段建议为：

- `cost_price`

比较成本价是本系统内部统一的“可比较成本”，用于页面展示、最低价选路、自动改价、亏本判断和订单成本快照。

### 3.11 税态换算

当商品税态与渠道税态不一致时，根据系统配置的税点，对 `source_cost_price` 按百分比换算得到 `cost_price` 的过程。

### 3.12 上游价格通知

上游渠道通过 POST 通知本系统某个上游商品成本变动的过程。

### 3.13 上游回调

上游渠道通过 POST 通知本系统某次订单最终结果的过程。

### 3.14 数量拆单

当主订单 `quantity > 1` 且当前渠道不适合一次性按总数量建单时，系统将同一主订单拆成多次上游建单执行的过程。

### 3.15 履约分片

数量拆单后，每个待履约的子数量单元称为一个“履约分片”。补单、超时、回调、查单都以履约分片内的 `attempt` 作为最小执行单元。

---

## 4. 总体目标

一期落地后，应满足以下目标：

1. 一个商品可以配置多个渠道绑定。
2. 每个绑定可以独立配置对接状态、对接编号、原始进货价、比较成本价、排序、权重、时段和自动改价。
3. 商品可配置 5 种下单方式，并由系统按明确规则选择渠道。
4. 渠道订单全部按异步处理，系统必须支持“建单 + 查单 + 回调”闭环。
5. 商品可以配置智能补单和超时补单。
6. 系统能记录本次订单到底命中了哪个绑定、每次尝试为何失败、是否发生补单、何时超时。
7. 系统能接收上游价格通知，并在开启自动改价时自动按利润规则更新售价。
8. 下游调用方能通过开放接口创建订单并查询订单状态。
9. 对外接口仍保持三态输出：`processing / success / failed`。
10. 当 `quantity > 1` 时，系统应能根据渠道 Provider 能力选择“单次上传总数量”或“拆成多次上游建单”，并保持对外主订单语义不变。

---

## 5. 核心业务口径

本章是全文最关键的统一口径，后续所有模块必须遵守。

### 5.1 价格口径

#### 5.1.1 双层成本口径

系统必须同时保留两层成本：

- `source_cost_price`
  - 上游原始进货价
  - 用于录入、上游价格通知、对账、审计
- `cost_price`
  - 税态换算后的比较成本价
  - 用于页面展示、最低价选路、自动改价、亏本判断、订单快照

#### 5.1.2 页面展示口径

页面中绑定列表的“进货价”、商品列表中的“最低进价”，都展示 `cost_price`，不展示 `source_cost_price`。

当前接口会同时额外返回利润后价格：

- 绑定列表返回 `effective_sell_price`
- 商品列表摘要返回 `min_channel_effective_sell_price`

这样可以把“比较成本价”和“利润后价格”拆开，避免把列表里的“进货价”误解成最终售价。

#### 5.1.3 税态换算顺序

顺序固定为：

1. 先根据商品税态与渠道税态，对 `source_cost_price` 做税态换算，得到 `cost_price`
2. 再在 `cost_price` 基础上加固定利润或百分比利润
3. 最终得到本次可售单价

禁止先加利润再换税。

### 5.2 税态换算规则

系统需要支持两个独立税点配置：

- `untaxed_to_taxed_rate`
  - 未税转有税税点百分比
- `taxed_to_untaxed_rate`
  - 有税转未税税点百分比

换算规则固定如下：

#### 情况一：商品未税，渠道未税

- `cost_price = source_cost_price`

#### 情况二：商品有税，渠道无税

- `cost_price = source_cost_price * (1 + untaxed_to_taxed_rate / 100)`

#### 情况三：商品有税，渠道有税

- `cost_price = source_cost_price`

#### 情况四：商品无税，渠道有税

- `cost_price = source_cost_price * (1 - taxed_to_untaxed_rate / 100)`

补充规则：

- 若商品税态与渠道税态不一致，但对应税点未配置，则该绑定不可用。
- 税态不再作为“必须完全一致才能下单”的硬拦截条件。
- 税态不同不是错误，税点缺失才是错误。

### 5.3 售价计算规则

#### 5.3.1 不改价

当绑定未开启自动改价，或绑定自动改价规则配置为“不改价”时：

- 本次订单 `sale_price` 使用商品主档默认售价快照。

#### 5.3.2 固定利润

当绑定开启自动改价且 `add_type = fixed` 时：

- `effective_sell_price = cost_price + default_price`

#### 5.3.3 百分比利润

当绑定开启自动改价且 `add_type = percent` 时：

- `effective_sell_price = cost_price * (1 + default_price / 100)`

说明：

- 百分比利润的基数是 `cost_price`
- 不是 `source_cost_price`

#### 5.3.4 `default_price` 字段口径

- 字段名历史保留为 `default_price`
- 业务语义固定为“利润值”，不是最终售价
- `add_type = fixed` 时表示固定利润金额
- `add_type = percent` 时表示利润百分比
- 自动改价弹窗中的录入文案必须为“利润值”或“加价值”，不得展示为“默认售价”

### 5.4 售价锁定规则

订单创建时，系统必须先选出首个尝试绑定，再据此锁定本次订单 `sale_price`。

锁定后规则如下：

- 后续补单不得重新抬高对外售价。
- 若后续尝试绑定成本更高，则仅使用新的 `cost_price` 与已锁定的 `sale_price` 做亏本判断。
- 若亏本不允许或超出最大亏损额，则该次尝试不得发起。
- 当 `quantity > 1` 且发生数量拆单时，整张主订单共享同一个已锁定 `sale_price` 单价，不因各履约分片的补单或重试而重新定价。

### 5.5 精度规则

为避免不同模块计算口径不一致，统一采用以下精度规则：

- `source_cost_price`：保留 4 位小数
- `cost_price`：保留 4 位小数
- `tax_adjust_amount`：保留 4 位小数
- `sale_price`：保留 4 位小数
- `default_price`（利润值）：保留 4 位小数
- `total_amount = sale_price * quantity`：保留 4 位小数
- 利润计算与税态换算后，都按统一规则四舍五入到 4 位小数

### 5.6 订单状态口径

对外调用方只看三态：

- `processing`
- `success`
- `failed`

对内不得只用三态，必须拆分更细的状态，以支持异步查单、回调和补单。

### 5.7 自动改价口径

自动改价是绑定级能力，不是旧页面那种混在一个“大而全”接口里的配置集合。

自动改价必须满足以下语义：

- 上游价格通知到达后，系统先更新 `source_cost_price`
- 再重算 `cost_price`
- 若绑定开启自动改价，则按利润规则计算 `effective_sell_price`
- 若绑定未开启自动改价，则只更新成本，不自动改销售价
- 若商品开启同步进价，则刷新商品最低进价快照

---

## 6. 管理端需求

## 6.1 商品列表中的渠道摘要

商品列表新增以下只读字段：

- `bound_channels`
  - 当前商品已启用绑定的渠道名称列表
- `bound_channel_count`
  - 当前商品已启用绑定数量
- `primary_channel_name`
  - 当前首选绑定渠道名称
- `min_channel_cost`
  - 当前商品所有有效绑定中的最小 `cost_price`
- `channel_auto_price_status`
  - 当前商品是否存在启用自动改价的绑定，可用于列表筛选或标记展示

展示规则：

- 若无绑定，则渠道列为空或显示“未绑定”
- 若有 1 个绑定，则直接显示该渠道名称
- 若有多个绑定，则展示首个名称，并可配合数量提示
- 点击渠道列，进入“商品渠道绑定弹窗”
- `min_channel_cost` 永远展示比较成本价的最小值，不展示 `source_cost_price`

补充说明：

- `primary_channel_name` 的计算应尽量与当前 `route_mode` 一致。
- 若当前路由模式无法静态确定固定首选渠道，则允许展示“按规则选路”或展示当前主排序第一条绑定。

## 6.2 商品渠道绑定弹窗

弹窗包含两个区域：

1. 顶部商品级配置摘要
2. 下方绑定列表

### 6.2.1 顶部商品级配置摘要

必须展示以下字段：

- 商品名称
- 类目
- 商品面值
- 智能补单
- 补单时间设置
- 下单方式
- 同步进价
- 同步商品名称
- 亏本销售
- 默认售价
- 组合商品

顶部摘要对应的是商品主档级配置，不是单条绑定级配置。

补充说明：

- 顶部摘要中的“类目”“商品面值”“默认售价”都来自商品主档。
- 这里的“默认售价”是商品主档默认销售价，不是绑定自动改价中的利润值。

### 6.2.2 绑定列表

每一行代表一个商品绑定。

列表字段定义如下：

- `名称`
  - 展示“上游商品名 + 主体名 + 渠道名”
- `对接状态`
  - 展示当前绑定是否允许参与下单
- `对接编号`
  - 上游商品编号，即 `supplier_goods_no`
- `进货价`
  - 当前绑定比较成本价，即 `cost_price`
- `原始进货价`
  - 新增/编辑时展示，用于录入或回显 `source_cost_price`
  - 列表是否展示可由前端决定，一期至少在编辑态可见
- `自动改价`
  - 展示该绑定是否启用自动改价
- `排序`
  - 用于固定顺序、同价优先级、时段模式下的优先级比较
- `权重`
  - 百分比分配模式下使用
- `时段`
  - 按时段提交模式下使用，展示 `start_time ~ end_time`
- `充值匹配`
  - 绑定所需充值模板

### 6.2.3 弹窗操作按钮

- `修改配置`
  - 修改商品主档级配置
- `新增库存`
  - 新增商品绑定
  - 页面文案保留“新增库存”，系统语义统一为“新增渠道绑定”
- `一键排序`
  - 按系统规则重排当前商品所有绑定排序值
- `批量开启对接状态`
  - 批量启用选中绑定
- `批量关闭对接状态`
  - 批量停用选中绑定
- `编辑自动改价`
  - 修改单个绑定或批量绑定的自动改价参数
- `批量删除`
  - 批量软删除绑定

## 6.3 商品级配置功能定义

以下配置属于商品主档级，不属于单个绑定。

### 6.3.1 智能补单

字段建议：

- `smart_replenish_enabled`

业务语义：

- 开启：
  - 当前尝试明确失败，或在等待窗口内未拿到最终结果时，系统允许切换到下一个候选绑定继续尝试
- 关闭：
  - 只尝试首个绑定，不自动切换下一条绑定

可进入补单的场景分两类：

#### A. 明确失败，可立即补单

- 上游明确返回库存不足
- 上游明确返回商品不可售
- 上游明确返回余额不足
- 上游明确返回业务失败且结果确定
- 查单明确返回失败且结果确定

#### B. 等待超时，可进入补单

- 建单后长时间未得到最终结果
- 上游未回调，查单也始终未给出最终结果
- 达到当前尝试的等待超时阈值

不可进入补单的场景：

- 商品未绑定
- 商品已下架
- 调用方鉴权失败
- 调用方参数校验失败
- 幂等单重复请求
- 本地亏本保护拦截
- 本地模板校验失败
- 本地签名/配置错误导致根本未发起上游请求

### 6.3.2 补单时间设置

字段建议：

- `attempt_timeout_enabled`
- `attempt_timeout_minutes`

页面文案仍可展示为：

- `补单时间设置`
- `补单时间`

但后端语义固定为：

- 当前 attempt 的等待超时时间
- 单位：分钟

业务语义：

- 开启：
  - 某次尝试在 `attempt_timeout_minutes` 内仍未得到最终结果，则视为该次尝试超时
  - 若允许补单，则切到下一条绑定
- 关闭：
  - 不因等待时间自动触发补单
  - 但系统仍会继续查单与等待回调，直到进入全局结束条件

说明：

- 该字段不是整张订单的总窗口。
- 每次新的 attempt 都独立重新计算自己的等待截止时间。

### 6.3.3 下单方式

字段建议：

- `route_mode`

一期支持以下 5 种模式：

- `fixed_order`
- `lowest_cost_first`
- `weight_percent`
- `time_period`
- `random`

#### 1. `fixed_order` 固定顺序

候选顺序：

1. `sort asc`
2. `id asc`

首单取第一条，补单按剩余顺序依次取下一条。

#### 2. `lowest_cost_first` 进价从低到高

候选顺序：

1. `cost_price asc`
2. `sort asc`
3. `id asc`

说明：

- 这里的 `cost_price` 是税态换算后的比较成本价。

#### 3. `weight_percent` 百分比分配

规则：

- 仅保留 `weight > 0` 的绑定
- 首次命中按权重随机抽取
- 补单时从剩余未尝试绑定中继续按权重抽取
- 若所有有效绑定 `weight = 0`，则视为无可用绑定
- 禁止自动回退到其他路由模式

#### 4. `time_period` 按时段提交

规则：

- 仅保留当前时间命中时段窗口的绑定
- 若多个绑定同时命中，再按以下规则排序：
  1. `sort asc`
  2. `id asc`
- 若当前时间没有任何绑定命中有效时段，则视为无可用绑定
- 禁止自动回退到其他路由模式

时段规则：

- `start_time <= end_time` 表示当天时段
- `start_time > end_time` 表示跨天时段

#### 5. `random` 随机选择

规则：

- 从当前有效绑定中随机抽取首个绑定
- 补单时从剩余未尝试绑定中继续随机抽取
- 禁止重复命中已失败绑定

### 6.3.4 同步进价

字段建议：

- `sync_cost_enabled`

业务语义：

- 开启：
  - 商品主档中的最低进价快照由绑定表自动聚合
  - 聚合口径为所有有效绑定的最小 `cost_price`
- 关闭：
  - 商品仍可下单，但商品主档不自动刷新最低进价快照

### 6.3.5 同步商品名称

字段建议：

- `sync_goods_name_enabled`

业务语义：

- 开启：
  - 当主渠道绑定的上游商品名称变化时，可同步更新商品名称快照
- 关闭：
  - 商品名称只允许人工维护

说明：

- 一期默认只同步“主渠道绑定”的名称
- 主渠道绑定的判定应尽量与 `route_mode` 保持一致
- 若无法静态确定，则以当前排序第一条绑定为名称同步来源

### 6.3.6 亏本销售

字段建议：

- `allow_loss`
- `max_loss_amount`

业务语义：

- 关闭：
  - 若本次尝试绑定的 `cost_price` 大于已锁定 `sale_price`，则直接拒绝该次尝试
- 开启：
  - 允许继续下单，但订单需标记为亏本单
- 若配置了 `max_loss_amount`：
  - 当单笔亏损金额超过该阈值时，仍然拒绝本次尝试

判断公式：

- `loss_amount = current_attempt_cost_price - locked_sale_price`

说明：

- 这里的 `current_attempt_cost_price` 是当前尝试绑定的 `cost_price`
- 不是 `source_cost_price`

### 6.3.7 组合商品

字段建议：

- `is_bundle`

一期业务定义：

- 表示该商品的充值信息模板较复杂，可能包含多个字段
- 不表示一张订单会拆成多个上游子单

## 6.4 绑定级功能定义

以下配置属于单个商品绑定。

### 6.4.1 对接状态

字段建议：

- `dock_status`

枚举：

- `enabled`
- `disabled`

业务语义：

- `enabled`
  - 允许进入候选绑定池
- `disabled`
  - 列表仍展示，但不参与下单，也不参与补单

### 6.4.2 渠道账号

字段建议：

- `platform_account_id`

业务语义：

- 指定本绑定实际使用的渠道账号

### 6.4.3 对接商品编号

字段建议：

- `supplier_goods_no`

业务语义：

- 上游平台中的商品编号
- 下单时用于告诉上游“买的是哪个商品”

### 6.4.4 对接商品名称

字段建议：

- `supplier_goods_name`

业务语义：

- 绑定级上游商品名称快照
- 可手填，可由价格通知或后续可选的商品查询能力辅助同步
- 不作为新增/编辑绑定保存前必须实时校验的字段

### 6.4.5 原始进货价

字段建议：

- `source_cost_price`

业务语义：

- 渠道原始进货价
- 新增/编辑绑定时录入
- 上游价格通知时更新
- 用于对账和审计

### 6.4.6 比较成本价

字段建议：

- `cost_price`

业务语义：

- 按税态换算后的成本价
- 用于页面“进货价”展示、最低价选路、自动改价、亏本判断和订单快照

### 6.4.7 排序

字段建议：

- `sort`

业务语义：

- 用于固定顺序模式、同成本价场景、时段模式场景下的稳定优先级
- 值越小优先级越高

### 6.4.8 权重

字段建议：

- `weight`

业务语义：

- 仅在 `weight_percent` 模式下生效
- `0` 表示不参与权重池

### 6.4.9 时段

字段建议：

- `start_time`
- `end_time`

业务语义：

- 仅在 `time_period` 模式下生效

### 6.4.10 自动改价

字段建议：

- `is_auto_change`
- `add_type`
- `default_price`
  - 字段名保留，业务语义为利润值
- `lock_price`
- `symbol_price`
- `max_price`
- `min_price`

一期必须生效的字段：

- `is_auto_change`
- `add_type`
- `default_price`

一期只落库回显、不参与真实计算的字段：

- `lock_price`
- `symbol_price`
- `max_price`
- `min_price`

字段口径：

- `default_price` 字段名保留不变，但业务语义固定为“利润值”
- `add_type = fixed` 时，`default_price` 表示固定利润金额
- `add_type = percent` 时，`default_price` 表示利润百分比
- `default_price` 统一支持 4 位小数
- 前端自动改价弹窗中，该输入框文案必须展示为“利润值”或“加价值”，不得展示为“默认售价”

计算规则：

#### 固定利润

- `is_auto_change = true`
- `add_type = fixed`
- `effective_sell_price = cost_price + default_price`

#### 百分比利润

- `is_auto_change = true`
- `add_type = percent`
- `effective_sell_price = cost_price * (1 + default_price / 100)`

#### 不改价

以下情况视为不改价：

- `is_auto_change = false`

兼容说明：

- 旧数据中若存在 `default_price = -1`，可继续按“不改价”兼容处理
- 新前端录入和新接口语义不再依赖 `default_price = -1`

不改价时：

- 销售价沿用商品主档默认售价

### 6.4.11 充值匹配

字段建议：

- `validate_template_id`

业务语义：

- 表示该绑定要求的充值参数模板
- 对外下单时，`payload` 必须符合该模板

示例：

- 手机号
- QQ号
- 手机号或 QQ
- 邮箱
- 纯数字
- 网址

### 6.4.12 路由模式与绑定字段生效关系

不同 `route_mode` 下，绑定字段生效规则如下：

- `fixed_order`
  - 生效字段：`sort`
- `lowest_cost_first`
  - 生效字段：`cost_price`、`sort`
- `weight_percent`
  - 生效字段：`weight`
- `time_period`
  - 生效字段：`start_time`、`end_time`、`sort`
- `random`
  - 无额外排序字段强依赖，基础过滤通过即可参与随机

说明：

- 所有模式都仍然受 `dock_status`、主体匹配、模板匹配、税态可换算约束
- 所有模式都受智能补单、等待超时、亏本销售等商品级规则约束
- 不再使用“税态必须完全一致”的硬限制

## 6.5 新增/编辑绑定规则

新增绑定时必须录入以下字段：

- 选择渠道账号
- 对接商品 ID
- 原始进货价
- 对接商品名
- 充值匹配模板
- 对接状态
- 排序
- 权重（按需）
- 时段配置（按需）

新增/编辑校验规则：

- 同一商品下，`platform_account_id + supplier_goods_no` 不允许重复
- 渠道账号必须是启用状态
- 渠道账号主体必须与商品主体一致
- `source_cost_price` 必须大于等于 `0`
- 对接商品 ID 不能为空
- 若设置 `weight`，则必须大于等于 `0`
- 若配置时段，则 `start_time` 与 `end_time` 必须同时存在
- 若商品税态与渠道税态不一致，则系统必须能够找到对应税点配置，否则该绑定不可保存或不可启用
- 保存时必须同步计算并持久化 `cost_price`
- 编辑表单中，输入框文案必须明确为“原始进货价”，避免与列表展示的比较成本价混淆

补充规则：

- 后台新增/编辑绑定时，不主动调用 `CatalogProvider` 或其他上游商品查询接口校验 `supplier_goods_no` 是否存在、当前是否可售
- `supplier_goods_no`、`supplier_goods_name`、`source_cost_price` 允许人工录入，由运营侧自行核验真实性
- 若后续某个平台额外实现商品查询能力，也仅作为运营辅助，不作为保存前必经校验链路

## 6.6 一键排序规则

一键排序的排序规则固定为：

1. `cost_price asc`
2. 当前 `sort asc`
3. `id asc`

重写规则建议：

- 对参与排序的有效绑定按顺序重写为 `10, 20, 30, ...`

说明：

- 相同 `cost_price` 时，保留原先较高优先级的绑定靠前
- 已删除绑定不参与
- 停用绑定是否参与一键排序由产品决定；一期建议仅对未删除绑定全部参与，避免排序与启停状态耦合

## 6.7 自动改价与价格通知

### 6.7.1 触发源

自动改价不是手动轮询上游价格，而是由上游通过 POST 主动通知本系统。

### 6.7.2 处理流程

收到上游价格通知后，系统必须执行：

1. 验签
2. 幂等校验
3. 定位渠道账号与上游商品
4. 更新 `source_cost_price`
5. 重算 `cost_price`
6. 若绑定开启自动改价，则重算绑定的 `effective_sell_price`
7. 若商品开启同步进价，则刷新 `min_channel_cost`
8. 若商品开启同步商品名称且本绑定为主绑定，则允许同步商品名称快照
9. 记录价格通知日志

### 6.7.3 自动改价生效边界

- 自动改价改变的是“该绑定的可售价格计算结果”
- 不强制要求回写所有历史订单
- 已创建订单的价格快照不可被价格通知回写

### 6.7.4 价格通知失败处理

- 验签失败：拒收并记日志
- 找不到绑定：拒收并记日志
- 重复通知：按幂等处理，不重复更新
- 部分绑定更新失败：保留失败日志，便于人工排查

## 6.8 管理端 REST 接口契约

本期管理端按新 REST 风格重构，不兼容旧接口名。

### 6.8.1 商品级渠道配置

#### 获取商品渠道配置

- `GET /api/admin/product-goods/{goodsId}/channel-config`

返回内容至少包含：

- 商品基础信息
- `smart_replenish_enabled`
- `attempt_timeout_enabled`
- `attempt_timeout_minutes`
- `route_mode`
- `sync_cost_enabled`
- `sync_goods_name_enabled`
- `allow_loss`
- `max_loss_amount`
- `is_bundle`
- 默认售价
- 列表摘要字段

#### 更新商品渠道配置

- `PATCH /api/admin/product-goods/{goodsId}/channel-config`

### 6.8.2 绑定列表与 CRUD

#### 绑定列表

- `GET /api/admin/product-goods/{goodsId}/channel-bindings`

#### 新增绑定

- `POST /api/admin/product-goods/{goodsId}/channel-bindings`

#### 更新绑定基础字段

- `PATCH /api/admin/product-goods/{goodsId}/channel-bindings/{bindingId}`

#### 删除绑定

- `DELETE /api/admin/product-goods/{goodsId}/channel-bindings/{bindingId}`

### 6.8.3 批量操作

#### 批量启停

- `POST /api/admin/product-goods/{goodsId}/channel-bindings:batch-status`

#### 批量删除

- `POST /api/admin/product-goods/{goodsId}/channel-bindings:batch-delete`

#### 一键排序

- `POST /api/admin/product-goods/{goodsId}/channel-bindings:reorder`

### 6.8.4 自动改价配置

#### 单条绑定自动改价更新

- `PATCH /api/admin/product-goods/{goodsId}/channel-bindings/{bindingId}/auto-price`

#### 批量绑定自动改价更新

- `POST /api/admin/product-goods/{goodsId}/channel-bindings:auto-price-batch`

补充说明：

- 自动改价相关接口中的 `default_price` 请求字段，语义固定为利润值，不是最终售价
- 前端表单提交与回显时，都必须按“利润值”口径处理

说明：

- 不再保留旧接口 `joinGoodsSetSwitch`
- 绑定基础字段与自动改价配置必须拆开维护

---

## 7. 下单链路需求

## 7.1 候选绑定生成

下单前，系统必须先生成基础可用绑定集合。

基础过滤条件：

- 商品状态为启用
- 商品未删除
- 商品供货方式为 `渠道`
- 商品至少存在一条有效绑定
- 渠道账号启用
- 绑定启用
- 主体匹配
- `source_cost_price` 有效
- `cost_price` 已成功计算
- 若商品税态与渠道税态不一致，则对应税点配置存在
- 若绑定配置了 `validate_template_id`，则必须与本次订单 `payload` 匹配

在生成基础集合后，再根据商品 `route_mode` 生成最终候选顺序或抽取规则。

补充说明：

- 候选绑定生成仅依赖本地绑定配置、本地商品状态和本地模板/税点校验结果
- 不依赖 `CatalogProvider` 的实时商品查询结果作为前置条件

### 7.1.1 `fixed_order`

排序规则：

1. `sort asc`
2. `id asc`

### 7.1.2 `lowest_cost_first`

排序规则：

1. `cost_price asc`
2. `sort asc`
3. `id asc`

### 7.1.3 `weight_percent`

规则：

- 仅保留 `weight > 0` 的绑定
- 根据权重随机抽取本次首个绑定
- 补单时从剩余未尝试绑定中继续按权重抽取
- 若过滤后无绑定，则视为无可用绑定

### 7.1.4 `time_period`

规则：

- 仅保留当前时间命中 `start_time ~ end_time` 的绑定
- 再按以下规则排序：
  1. `sort asc`
  2. `id asc`
- 若过滤后无绑定，则视为无可用绑定

### 7.1.5 `random`

规则：

- 从基础可用绑定集合中随机抽取首个绑定
- 补单时从剩余未尝试绑定中继续随机抽取

## 7.2 创建订单流程

系统创建订单的标准流程如下：

1. 校验调用方身份。
2. 校验签名、时间戳、随机串、防重放。
3. 校验 `client_order_no` 幂等。
4. 校验商品是否启用、是否可售。
5. 校验购买数量是否符合购买限制策略。
6. 校验调用方传入的 `payload` 是否满足商品模板。
7. 生成基础可用绑定集合。
8. 根据商品 `route_mode` 生成首个候选绑定。
9. 若无可用绑定，则直接拒单。
10. 基于首个候选绑定计算并锁定 `sale_price`。
11. 创建内部订单，状态初始化为对外 `processing`。
12. 根据当前 `quantity` 与 Provider 能力，确定执行方式，并创建首个待执行履约分片及其第 1 次 `attempt`。
13. 对当前 `attempt` 执行亏本保护判断。
14. 调用上游渠道建单。
15. 根据建单结果：
    - 若是明确受理成功，则进入等待回调/查单状态
    - 若是明确业务失败，则判断当前履约分片是否进入补单
    - 若是结果不确定，则进入等待回调/查单状态，不得立即补单
16. 后续由“上游回调”和“定时查单任务”共同驱动当前 `attempt`、当前履约分片与主订单状态收敛。
17. 当累计成功履约数量等于订单 `quantity` 时，主订单置为 `success`。
18. 当剩余未成功数量已无任何可继续尝试的绑定，或整单已无法满足订单 `quantity` 时，主订单置为 `failed`。

## 7.2.1 `quantity > 1` 的数量拆单执行规则

当主订单 `quantity > 1` 时，系统必须先判断当前 Provider 是否支持一次性按总数量建单。

执行规则如下：

1. 若当前 Provider 支持原生数量提交，则系统可以直接以本次 `quantity` 发起单次上游建单。
2. 若当前 Provider 不支持原生数量提交，或接入实现明确要求拆分执行，则系统允许将同一主订单拆成多个“履约分片”后分别向上游建单。
3. 履约分片的默认拆分粒度建议为 `1`，如个别平台支持按部分数量建单，也可由 Provider 实现定义其他粒度。
4. 每一次真实的上游建单请求，都必须落一条 `trade_order_attempt`。
5. `trade_order_attempt` 必须记录：
   - `fulfillment_no`
   - `attempt_quantity`
6. 补单、超时、回调、查单都以“履约分片”为最小执行单元。
7. 主订单的对外 `sale_price` 仍表示统一单价，`total_amount = sale_price * quantity`，数量拆单对调用方透明。
8. 当累计成功履约数量等于订单 `quantity` 时，主订单置为 `success`。
9. 当累计成功履约数量小于订单 `quantity`，且剩余未成功数量已无任何可继续尝试的绑定时，主订单置为 `failed`。
10. 若出现“部分履约分片成功、部分履约分片最终失败”的情况，主订单对外状态仍返回 `failed`，并在内部保留成功数量、失败数量与明细审计信息。

## 7.3 补单统一规则

补单必须同时满足以下条件：

- 商品开启智能补单
- 当前 attempt 已明确失败，或已达到等待超时阈值
- 当前订单尚未成功
- 尚有未尝试的候选绑定
- 当前剩余候选绑定通过亏本保护校验

统一约束：

- 补单的最小单位是“履约分片”，不是整张主订单
- 同一履约分片内，禁止重复尝试已经失败过的同一绑定
- 不同履约分片之间，允许命中同一绑定
- 每次补单都必须新增一条 `trade_order_attempt`
- 一旦某个履约分片的某次尝试成功，则该履约分片停止后续补单
- 对“结果不确定”的 attempt，不允许立即补单，必须先等待回调或查单，直到明确失败或等待超时

各路由模式下的补单规则：

- `fixed_order / lowest_cost_first / time_period`
  - 按首次生成的候选顺序，跳过已尝试绑定，继续尝试下一条
- `weight_percent`
  - 从剩余未尝试且 `weight > 0` 的绑定中重新按权重抽取
- `random`
  - 从剩余未尝试绑定中重新随机抽取

## 7.4 attempt 等待超时规则

每次 attempt 必须记录自己的等待截止时间：

- `query_deadline_at`

计算规则：

- 若 `attempt_timeout_enabled = true`
  - `query_deadline_at = attempt_created_at + attempt_timeout_minutes`
- 若 `attempt_timeout_enabled = false`
  - 不设置等待截止时间，或置为 null

到达 `query_deadline_at` 且仍未得到最终结果时：

- 将 attempt 标记为 `timeout`
- 若允许补单，则在当前履约分片内切下一条绑定
- 若不允许补单，则当前履约分片失败；若整单因此已无法完成，则主订单最终失败

## 7.5 结果不确定场景

以下场景都属于“结果不确定”，不得立刻补单：

- 渠道接口超时
- 上游返回空响应
- 上游返回非 JSON
- 上游返回 `server_error`
- 上游返回“已受理处理中”
- 上游返回重复单但无法确认原单结果
- 查单结果仍为处理中或未知

这些场景必须进入：

- `waiting_callback`
- `querying`
- 或等价内部状态

直到满足以下任一条件：

- 回调确认成功
- 回调确认失败
- 查单确认成功
- 查单确认失败
- 达到等待超时并允许补单

## 7.6 订单状态

### 7.6.1 对外状态

对外状态固定如下：

- `processing`
- `success`
- `failed`

### 7.6.2 对内主订单状态建议

建议至少支持以下内部状态：

- `created`
- `routing`
- `processing`
- `success`
- `failed`
- `manual_review`

补充说明：

- 当 `quantity > 1` 且发生数量拆单时，若出现“部分履约分片成功、部分履约分片失败”的情况，主订单可进入 `manual_review` 或等价内部审计状态，但对外状态仍按既定三态输出。

### 7.6.3 对内 attempt 状态建议

建议至少支持以下 attempt 状态：

- `created`
- `submitted`
- `accepted`
- `waiting_callback`
- `querying`
- `success`
- `failed`
- `timeout`
- `unknown`

## 7.7 幂等规则

幂等键：

- `caller_id + client_order_no`

规则：

- 同一调用方重复提交同一 `client_order_no` 时，直接返回原订单
- 不重新发起上游请求
- 幂等命中返回的 `sale_price`、`status`、`order_no` 必须与原订单一致

## 7.8 成本与售价快照

订单创建后，必须固化以下快照：

- `binding_id`
- `platform_account_id`
- `source_cost_price_snapshot`
- `cost_price_snapshot`
- `tax_adjust_direction`
- `tax_adjust_rate`
- `tax_adjust_amount`
- `sale_price`
- `quantity`
- `total_amount`
- `loss_order`
- `loss_amount`

说明：

- 历史订单快照不得因后续价格通知回写
- 补单时可记录新的 attempt 成本信息，但主订单已锁定售价不变
- 当 `quantity > 1` 且发生数量拆单时，主订单层记录的是总 `quantity` 与统一单价快照；各履约分片的实际建单数量必须在 `trade_order_attempt.attempt_quantity` 中单独记录

---

## 8. 异步查单与上游回调需求

## 8.1 全部订单按异步处理

本期默认所有真实渠道订单都是异步订单。

即使建单接口同步返回“成功”，也只表示：

- 上游已受理
- 或上游已返回初始结果

最终成功与否仍需由以下机制确认：

- 上游回调
- 定时查单

## 8.2 定时查单任务

系统必须存在异步查单任务，对以下 attempt 周期性查单：

- `accepted`
- `waiting_callback`
- `querying`
- `unknown`

任务要求：

- 支持退避重试
- 每次查单都记录请求与响应日志
- 查单成功后立即终止该 attempt 后续查询
- 查单失败但结果明确时，写回最终失败
- 查单仍为处理中时，继续等待
- 查单超过等待截止时间时，触发超时逻辑

## 8.3 上游回调接收

本期必须支持上游回调接收。

回调处理流程：

1. 验签
2. 幂等
3. 根据 `provider_request_order_no`、`channel_order_no` 或其他稳定外键定位 attempt
4. 写入回调日志
5. 更新 attempt 最终状态
6. 若回调成功，则更新对应履约分片成功结果，并在累计成功履约数量满足订单 `quantity` 时推进主订单成功
7. 若回调失败，则在对应履约分片维度判断是否触发补单；若整单已无法完成，则主订单失败
8. 返回上游要求的 ACK 内容

## 8.4 晚到回调处理

典型场景：

- attempt A 已等待超时并补单到 attempt B
- attempt B 最终成功
- attempt A 的成功回调晚到

规则固定如下：

- 晚到回调不得覆盖主订单已确定的最终状态
- 晚到成功只记录到 attempt 审计与异常日志
- 若晚到回调与主订单最终状态冲突，订单应进入异常审计队列或人工复核队列

## 8.5 回调幂等

相同回调重复推送时：

- 不得重复推进订单状态
- 不得重复创建补单
- 仅更新回调日志或命中幂等标记

---

## 9. 真实渠道对接需求

## 9.1 Provider 能力拆分

一期必须按能力拆分 Provider，禁止把上游 HTTP 逻辑直接写进 Controller 或业务 Logic。

一期核心 Provider 建议至少拆为：

- `BalanceProvider`
  - 查询余额
- `OrderProvider`
  - 创建订单、解析下单结果、查询订单状态、解析查单结果
- `CallbackProvider`
  - 验签回调、解析回调、生成 ACK
- `PriceNotifyProvider`
  - 验签价格通知、解析价格通知

`CatalogProvider` 为可选扩展能力，不作为一期必接能力。

说明：

- 未实现 `CatalogProvider` 不影响该平台参与真实下单、查单、回调和价格通知
- 若某个平台回调和价格通知验签规则相同，可以由同一 Provider 实现多个接口
- 但系统语义上仍要按能力拆开

## 9.2 CatalogProvider（可选能力）

若某个平台提供商品查询接口，可实现 `CatalogProvider`，用于：

- 运营人工辅助查询商品详情
- 查询商品价格
- 查询商品状态
- 查询商品模板或下单参数要求（若平台支持）
- 辅助同步商品名称

补充说明：

- `CatalogProvider` 不作为新增/编辑绑定保存前的前置校验
- 不作为候选绑定生成前的前置校验
- 不作为下单前必须成功的校验链路
- 商品绑定仍支持人工录入 `supplier_goods_no`、`supplier_goods_name`、`source_cost_price`

## 9.3 OrderProvider 最低能力要求

每个 Provider 至少要实现：

- `Code()`
- `Name()`
- `CandidateBaseURLs()`
- `BuildCreateOrderRequest()`
- `ParseCreateOrderResponse()`
- `BuildQueryOrderRequest()`
- `ParseQueryOrderResponse()`

与数量拆单相关的补充要求：

- Provider 或其配置元数据必须能声明“是否支持原生 quantity 提交”
- `BuildCreateOrderRequest()` 必须接收当前实际建单数量，即当前 `attempt_quantity`
- 当 Provider 不支持原生 quantity 提交时，系统可按数量拆单规则分多次调用 `BuildCreateOrderRequest()`

建单解析结果必须能统一归一为：

- 是否已成功受理
- 是否已最终成功
- 是否已最终失败
- 是否结果不确定
- 上游订单号
- 上游状态
- 失败码
- 失败文案
- 原始响应

## 9.4 CallbackProvider 最低能力要求

至少要实现：

- `VerifyCallbackSignature()`
- `ParseCallbackPayload()`
- `BuildCallbackAck()`

回调解析结果必须能统一归一为：

- attempt 定位键
- 上游订单号
- 最终状态
- 上游状态
- 错误码
- 错误文案
- 原始回调内容

## 9.5 PriceNotifyProvider 最低能力要求

至少要实现：

- `VerifyPriceNotifySignature()`
- `ParsePriceNotifyPayload()`

价格通知解析结果至少包含：

- 渠道账号标识
- `supplier_goods_no`
- 上游商品名称（若有）
- 新的 `source_cost_price`
- 通知时间
- 外部通知流水号或幂等键
- 原始通知内容

## 9.6 渠道错误归类

系统至少能归类以下错误：

- `stock_not_enough`
- `balance_not_enough`
- `goods_not_available`
- `param_invalid`
- `sign_invalid`
- `timeout`
- `server_error`
- `unknown`

该归类结果将直接影响：

- 是否立即失败
- 是否进入等待查单/回调
- 是否允许补单

---

## 10. 对外开放接口需求

## 10.1 对外调用方主数据

建议新增调用方主数据对象：

- `open_caller`

至少包含：

- `id`
- `name`
- `app_key`
- `app_secret`
- `status`
- `allowed_ip_list`
- `sign_version`
- `remark`
- `created_at`
- `updated_at`

## 10.2 鉴权方式

对外接口不复用 admin 登录态。

一期统一采用调用方签名鉴权，请求头如下：

- `X-App-Key`
- `X-Timestamp`
- `X-Nonce`
- `X-Signature`

安全要求：

- `X-Timestamp` 与服务器时间允许误差建议不超过 300 秒
- `X-Nonce` 需要做防重放缓存，建议保存 10 分钟
- `X-Signature` 必须同时覆盖请求方法、请求路径、时间戳、随机串、请求体摘要

签名建议：

- `signature = HMAC-SHA256(app_secret, canonical_string)`

`canonical_string` 建议包含：

- HTTP Method
- Request Path
- Timestamp
- Nonce
- Body SHA256

### 10.2.1 鉴权失败错误分类

鉴权失败时直接拒绝请求，不创建订单，错误至少要区分：

- `app_key_invalid`
- `signature_invalid`
- `timestamp_expired`
- `nonce_replayed`
- `caller_disabled`
- `ip_not_allowed`

## 10.3 创建订单接口

建议路径：

- `POST /api/open/orders`

请求字段：

- `client_order_no`
  - 调用方订单号
- `goods_code`
  - 平台商品编码
- `quantity`
  - 购买数量
- `payload`
  - 充值信息对象，字段由商品模板决定

响应字段：

- `order_no`
- `client_order_no`
- `status`
- `goods_code`
- `goods_name`
- `quantity`
- `sale_price`
- `total_amount`
- `created_at`

说明：

- 下游调用方不得自行传价格
- 价格由系统按当前选中首尝试绑定的规则计算并锁定
- 当 `quantity > 1` 时，系统可以对调用方透明地选择“单次上传总数量”或“拆成多个履约分片分别向上游建单”

## 10.4 查订单详情接口

建议路径：

- `GET /api/open/orders/{order_no}`

响应字段：

- `order_no`
- `client_order_no`
- `status`
- `goods_code`
- `goods_name`
- `quantity`
- `success_quantity`
- `failed_quantity`
- `sale_price`
- `total_amount`
- `failure_reason`
- `created_at`
- `finished_at`
- `upstream_orders`
  - 数组，建议至少包含：
    - `fulfillment_no`
    - `attempt_quantity`
    - `status`
    - `binding_channel_name`
    - `binding_supplier_goods_no`
    - `channel_order_no`

补充说明：

- 当订单未发生数量拆单，且最终仅存在单一上游单号时，`upstream_orders` 也可以只返回 1 条明细
- 当订单发生数量拆单、补单或存在多个上游单号时，统一通过 `upstream_orders` 返回聚合明细，避免使用单一 `channel_order_no` 造成歧义

## 10.5 按调用方订单号查单接口

建议路径：

- `GET /api/open/orders/by-client/{client_order_no}`

返回结构与按内部订单号查单一致。

## 10.6 对外价格规则

对外价格统一规则如下：

- 调用方不传价格
- 系统在订单创建时按首个 attempt 绑定计算并锁定 `sale_price`
- `total_amount = sale_price * quantity`
- 后续补单不得重新计算对外价格
- 后续 attempt 只拿新的 `cost_price` 与已锁定 `sale_price` 做亏本校验
- 当订单发生数量拆单时，对外仍按整单统一单价与总金额表达，不向调用方暴露拆分定价

---

## 11. 数据模型需求

## 11.1 商品级渠道配置表

建议新增表：

- `product_goods_channel_config`

如不单独建表，也必须在商品主档或扩展表中承载以下字段：

- `goods_id`
- `smart_replenish_enabled`
- `attempt_timeout_enabled`
- `attempt_timeout_minutes`
- `route_mode`
- `sync_cost_enabled`
- `sync_goods_name_enabled`
- `allow_loss`
- `max_loss_amount`
- `is_bundle`
- `created_at`
- `updated_at`

## 11.2 商品绑定表

建议新增表：

- `product_goods_channel_binding`

核心字段：

- `id`
- `goods_id`
- `platform_account_id`
- `supplier_goods_no`
- `supplier_goods_name`
- `source_cost_price`
- `cost_price`
- `tax_adjust_direction`
- `tax_adjust_rate`
- `tax_adjust_amount`
- `dock_status`
- `sort`
- `weight`
- `start_time`
- `end_time`
- `validate_template_id`
- `is_auto_change`
- `add_type`
- `default_price`
- `lock_price`
- `symbol_price`
- `max_price`
- `min_price`
- `created_at`
- `updated_at`
- `deleted_at`

索引建议：

- 唯一索引：`(goods_id, platform_account_id, supplier_goods_no, deleted_at)`
- 普通索引：`goods_id`
- 普通索引：`platform_account_id`
- 普通索引：`dock_status`
- 普通索引：`sort`

## 11.3 交易订单表

建议新增表：

- `trade_order`

核心字段：

- `id`
- `order_no`
- `caller_id`
- `client_order_no`
- `goods_id`
- `binding_id`
- `platform_account_id`
- `quantity`
- `success_quantity`
- `failed_quantity`
- `payload_json`
- `sale_price`
- `total_amount`
- `source_cost_price_snapshot`
- `cost_price_snapshot`
- `tax_adjust_direction`
- `tax_adjust_rate`
- `tax_adjust_amount`
- `loss_order`
- `loss_amount`
- `channel_order_no`
  - 无数量拆单或仅存在单一成功上游单时可回填
  - 若存在多个履约分片/多个上游单号，则以 `trade_order_attempt` 聚合明细为准
- `status`
- `failure_reason`
- `created_at`
- `updated_at`
- `finished_at`

说明：

- `binding_id`、`platform_account_id` 表示首个锁价绑定快照或主参考绑定
- 若订单发生数量拆单，实际履约明细与各次上游单号统一以 `trade_order_attempt` 为准

索引建议：

- 唯一索引：`(caller_id, client_order_no)`
- 唯一索引：`order_no`
- 普通索引：`goods_id`
- 普通索引：`status`

## 11.4 订单尝试日志表

建议新增表：

- `trade_order_attempt`

核心字段：

- `id`
- `order_id`
- `binding_id`
- `platform_account_id`
- `provider_code`
- `fulfillment_no`
- `attempt_quantity`
- `attempt_no`
- `provider_request_order_no`
- `channel_order_no`
- `attempt_status`
- `upstream_status`
- `request_url`
- `request_method`
- `request_payload`
- `response_payload`
- `http_status`
- `duration_ms`
- `error_category`
- `error_code`
- `error_message`
- `query_count`
- `last_query_at`
- `next_query_at`
- `query_deadline_at`
- `callback_payload`
- `callback_received_at`
- `callback_processed_at`
- `trace_id`
- `created_at`
- `updated_at`

用途：

- 记录每次命中的绑定及结果
- 支撑补单追踪、回调追踪、异步查单与审计

索引建议：

- 普通索引：`order_id`
- 普通索引：`(order_id, fulfillment_no)`
- 普通索引：`binding_id`
- 唯一索引：`provider_request_order_no`
- 普通索引：`channel_order_no`
- 普通索引：`attempt_status`
- 普通索引：`next_query_at`

## 11.5 上游回调日志表

建议新增表：

- `provider_callback_log`

核心字段：

- `id`
- `provider_code`
- `platform_account_id`
- `provider_request_order_no`
- `channel_order_no`
- `request_headers`
- `request_body`
- `verify_result`
- `process_result`
- `ack_body`
- `created_at`

## 11.6 上游价格通知日志表

建议新增表：

- `provider_price_notify_log`

核心字段：

- `id`
- `provider_code`
- `platform_account_id`
- `supplier_goods_no`
- `request_headers`
- `request_body`
- `source_cost_price_new`
- `verify_result`
- `process_result`
- `created_at`

---

## 12. 非功能要求

## 12.1 事务要求

以下动作必须在同一事务边界中保持一致：

- 幂等判断
- 内部订单创建
- 首次 attempt 创建
- 主订单状态迁移
- 绑定命中记录

禁止先写订单，再无约束地并发请求上游。

## 12.2 审计要求

以下行为必须记录操作日志：

- 新增绑定
- 修改绑定
- 批量启停绑定
- 批量删除绑定
- 修改商品级配置
- 修改自动改价配置

## 12.3 日志与排障要求

真实下单、查单、回调、价格通知必须保留足够排障字段，包括：

- 请求地址
- 请求参数
- 响应内容
- HTTP 状态码
- 耗时
- 追踪号
- 错误分类
- 上游状态
- 最终处理结论

## 12.4 OpenAPI 要求

一期新增接口必须进入：

- `/api.json`
- `/swagger/`

统一响应结构保持为：

- `code`
- `message`
- `data`

---

## 13. 一期不做的内容

以下内容明确不在一期实现范围：

- 退款接口
- 售后接口
- 手工补单页面
- 下游回调订阅中心
- 下游回调重试中心
- 自动改价历史记录页
- 对接商品价格变化明细页
- 组合商品一单多渠道子单编排
- 调用方管理后台页面

---

## 14. 验收标准

## 14.1 管理端验收

- 商品列表能正确展示渠道摘要
- 点击渠道可进入绑定弹窗
- 新增绑定后，重新打开弹窗可正确回显
- 列表“进货价”显示比较成本价 `cost_price`
- 编辑绑定时可正确录入和回显 `source_cost_price`
- 修改原始进货价后，比较成本价和最低进价聚合正确
- 新增/编辑绑定时，后台不依赖 `CatalogProvider` 主动校验也可正常保存人工录入数据
- 税态不同但税点存在时，绑定可正常保存并参与下单
- 税态不同且税点缺失时，绑定不可保存或不可启用
- 批量启停后，绑定状态正确变化
- 自动改价配置可正确保存和回显
- 自动改价弹窗中的利润值字段按“利润值”语义展示，不以“默认售价”误导运营
- 一键排序后，排序值按规则重写
- 百分比分配模式下，权重字段可正确保存与回显
- 按时段提交模式下，时段字段可正确保存与回显

## 14.2 价格与税点验收

- 商品有税、渠道未税时，`cost_price` 正确加税点
- 商品未税、渠道有税时，`cost_price` 正确扣税点
- 商品税态与渠道税态一致时，`cost_price = source_cost_price`
- 固定利润模式下，售价等于 `cost_price + 利润`
- 百分比利润模式下，售价按 `cost_price` 为基数计算
- 自动改价中的 `default_price` 按利润值口径处理，并支持 4 位小数
- 上游价格通知到达后，系统能更新 `source_cost_price` 与 `cost_price`
- 绑定开启自动改价时，价格通知能驱动售价规则变化
- 已创建订单价格快照不会被后续价格通知回写

## 14.3 交易验收

- 单绑定商品可以成功下单
- 多绑定商品在 `fixed_order` 模式下按 `sort` 优先命中绑定
- 多绑定商品在 `lowest_cost_first` 模式下优先命中最低 `cost_price` 绑定
- 多绑定商品在 `weight_percent` 模式下，多次下单后命中分布基本符合权重配置
- 多绑定商品在 `time_period` 模式下，仅当前时间窗口命中的绑定参与下单
- 多绑定商品在 `random` 模式下，多次下单可随机命中不同绑定
- 首个绑定明确失败且开启智能补单时，可按当前模式自动切换到剩余绑定
- 首个绑定长时间未完成且达到等待超时阈值时，可自动补单
- 结果不确定的 attempt 不会在未超时前立即补单
- 关闭亏本销售时，亏本 attempt 被拦截
- 开启亏本销售但超出最大亏损额度时，attempt 仍被拦截
- 当 `quantity > 1` 且 Provider 不支持原生数量提交时，系统可对同一主订单拆成多次上游建单
- 数量拆单后，各履约分片的上游建单、补单、查单、回调都能独立收敛，并最终汇总到主订单
- 同一履约分片不会重复尝试已经失败的同一绑定
- 不同履约分片之间允许命中同一绑定
- 若部分履约分片成功、部分履约分片失败，主订单对外状态仍能稳定返回 `failed`
- 重复下单时，幂等命中原订单
- 绑定停用后，不再参与选路
- 晚到回调不会覆盖主订单已确定的最终状态

## 14.4 对外接口验收

- 调用方可以成功创建订单
- 调用方可以按内部订单号查单
- 调用方可以按自己的订单号查单
- 鉴权失败时，接口直接拒绝
- 参数不符合商品模板时，接口直接拒绝
- `processing / success / failed` 三态返回稳定一致
- 同一 `client_order_no` 重复提交时，返回原订单而非重新下单

---

## 15. 后续开发文档拆分建议

基于本文档，后续开发文档建议拆为以下几份：

1. 商品绑定模块开发文档
2. 商品渠道配置模块开发文档
3. 渠道 Provider 扩展开发文档
4. 异步查单与上游回调处理文档
5. 上游价格通知与自动改价文档
6. 交易订单模块开发文档
7. 对外开放接口开发文档
8. 表结构与迁移文档
9. 合约测试与集成测试文档

---

## 16. 结论

一期不是单纯补一个“渠道弹窗”，而是要形成完整闭环：

- 管理端可配置商品与渠道的绑定关系
- 系统可按固定顺序、最低比较成本、百分比分配、按时段提交、随机选择这 5 种规则真实请求上游下单
- 所有订单按异步处理，通过上游回调和定时查单共同收敛结果
- 当前尝试超时后可按规则自动补单
- 上游价格通知可驱动绑定成本和自动改价联动
- `CatalogProvider` 若存在，仅作为可选辅助能力，不作为保存前或下单前必经校验
- 当 `quantity > 1` 时，系统可按 Provider 能力决定是否做数量拆单执行
- 下游调用方可通过开放接口稳定下单与查单

后续开发、建表、接口设计、前端重构和联调，必须严格以本文档里的字段语义、价格公式、税态换算、状态机、补单规则、数量拆单规则和接口边界为准。
