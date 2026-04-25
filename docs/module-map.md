# 模块职责地图

本文是当前仓库业务域、文件归属、路由前缀和权限边界的主参考。开发流程看 `development.md`，测试策略看 `testing.md`，迁移背景看 `migration.md`。

## 核心业务流速览

| 模块 | 核心业务流 |
| --- | --- |
| 认证与会话 | 用户登录后按账号状态和短信二验条件签发会话，后续接口通过 JWT 和 Redis 会话校验身份。 |
| 员工管理 | 超管或授权用户维护后台员工资料、状态和商务归属，删除后进入回收站再决定恢复或保留。 |
| 用户组与授权 | 后台通过用户组绑定菜单权限，登录态用户的可见菜单和接口权限都来自该授权结果。 |
| 主体管理 | 维护结算或业务主体基础资料，供商品、平台账号等后续业务选择和归属。 |
| 品牌管理 | 维护三级品牌树和品牌图片，商品创建时从品牌树选择归属。 |
| 行业管理 | 维护行业分类，并把行业和一级品牌关联起来，支持后续筛选和展示。 |
| 商品模板管理 | 维护充值模板和账号校验类型，商品配置时复用模板字段和展示规则。 |
| 商品购买数量限制策略 | 维护后台限购策略基础数据，当前只负责策略管理，不进入真实购买校验链。 |
| 商品管理 | 后台维护商品主档、渠道绑定和库存配置，开放订单会按商品编码定位可用商品。 |
| 第三方对接 | 后台维护供应商平台账号，余额刷新和订单履约通过 provider 适配器访问上游。 |
| 订单履约 | 开放接口创建订单后进入待提交队列，worker 提交上游、轮询状态，并在窗口内按规则补单。 |
| 短信配置 | 超管维护短信 provider 配置，运行态读取配置后用于登录二验验证码发送。 |
| 系统参数配置 | 超管按分组维护系统参数，保存时兼容旧单组写法和新多分组写法。 |
| 审计日志 | 登录、关键后台操作和异常行为落库，后台按管理员、时间和关键字查询。 |

## 业务域映射

### 认证与会话

- 协议：`api/auth.go`
- controller：`internal/controller/admin/auth.go`、`internal/controller/admin/session.go`
- service：`AuthService`（`internal/service/auth.go`）
- logic：`internal/logic/admin/auth.go`
- 路由前缀：`/api/admin/auth/*`
- 进入条件：登录、发送验证码、校验验证码无需登录；`me` 和退出登录需要登录。
- 主要能力：账号密码登录、条件短信二验、发送验证码、校验验证码、当前登录信息、退出登录。

### 员工管理

- 协议：`api/user.go`
- controller：`internal/controller/admin/user.go`
- service：`UserService`（`internal/service/user.go`）
- logic：`internal/logic/admin/user*.go`
- 路由前缀：`/api/admin/users*`
- 权限：`admin.list`
- 主要能力：员工列表、回收站、新增、编辑、删除、恢复、启停、余额通知、批量设置/取消商务。

### 用户组与授权

- 协议：`api/group.go`
- controller：`internal/controller/admin/group.go`
- service：`GroupService`（`internal/service/group.go`）
- logic：`internal/logic/admin/group.go`
- 路由前缀：`/api/admin/groups*`、`/api/admin/menus/tree`
- 权限：`admin.department`
- 主要能力：用户组列表、增删改、状态切换、权限读取、权限保存、菜单树。
- 边界：用户组授权和菜单树只暴露 `super_only = 0` 的菜单。

### 主体管理

- 协议：`api/subject.go`
- controller：`internal/controller/admin/subject.go`
- service：`SubjectService`（`internal/service/subject.go`）
- logic：`internal/logic/admin/subject.go`
- 路由前缀：`/api/admin/subjects*`
- 权限：`subject.manage`
- 主要能力：主体列表、新增主体、编辑主体。

### 品牌管理

- 协议：`api/brand.go`
- controller：`internal/controller/admin/brand.go`
- service：`BrandService`（`internal/service/brand.go`）
- logic：`internal/logic/admin/brand*.go`
- 路由前缀：`/api/admin/brands*`
- 权限：`product.brand`
- 主要能力：一级品牌分页、子级懒加载、新增、编辑、删除、排序、显隐切换、本地图片上传。
- 边界：品牌层级支持到三级，本地上传目录由 upload 配置决定。

### 行业管理

- 协议：`api/industry.go`
- controller：`internal/controller/admin/industry.go`
- service：`IndustryService`（`internal/service/industry.go`）
- logic：`internal/logic/admin/industry*.go`
- 路由前缀：`/api/admin/industries*`
- 权限：`product.industry`
- 主要能力：行业列表、增删改、排序、一级品牌选择器、行业品牌关联增删排序。
- 边界：行业只接受一级品牌关联。

### 商品模板管理

- 协议：`api/product_template.go`
- controller：`internal/controller/admin/product_template.go`
- service：`ProductTemplateService`（`internal/service/product_template.go`）
- logic：`internal/logic/admin/product_template*.go`
- 路由前缀：`/api/admin/product-templates*`
- 权限：`product.template`
- 主要能力：列表、关键词搜索、模板类型筛选、共享状态筛选、新增、编辑、单删、批删、验证方式枚举。
- 边界：当前只保存 `validate_type` 规则，不校验真实充值账号输入值。

### 商品购买数量限制策略

- 协议：`api/purchase_limit.go`
- controller：`internal/controller/admin/purchase_limit.go`
- service：`PurchaseLimitService`（`internal/service/purchase_limit.go`）
- logic：`internal/logic/admin/purchase_limit*.go`
- 路由前缀：`/api/admin/purchase-limit-strategies*`
- 权限：`product.purchase_limit`
- 主要能力：分页列表、关键词搜索、新增、编辑、删除、启停、限制类型和周期类型枚举。
- 边界：当前只做后台策略管理，不接真实购买前校验链。

### 商品管理

- 协议：`api/product_goods.go`、`api/product_goods_channel.go`、`api/product_goods_channel_config.go`
- controller：`internal/controller/admin/product_goods.go`、`internal/controller/admin/product_goods_channel.go`、`internal/controller/admin/product_goods_channel_config.go`
- service：`ProductGoodsService`（`internal/service/product_goods.go`）
- logic：`internal/logic/admin/product_goods*.go`、`internal/logic/admin/product_goods_channel*.go`、`internal/logic/admin/product_goods_channel_config*.go`
- 路由前缀：`/api/admin/products*`、`/api/admin/products/{goodsId}/channel-bindings*`、`/api/admin/products/{goodsId}/inventory-config`
- 权限：`product.goods`
- 主要能力：商品列表、详情、表单选项、新增、编辑、删除、启停、渠道摘要、库存配置、渠道绑定弹窗、单条自动改价、卡卡云商品名称和进货价同步。
- 边界：商品主档、渠道绑定和库存配置保持同 package 多文件拆分；商品关闭后不触发上游商品信息同步。

### 第三方对接

- 协议：`api/supplier_platform.go`
- controller：`internal/controller/admin/supplier_platform.go`
- service：`SupplierPlatformService`（`internal/service/supplier_platform.go`）
- logic：`internal/logic/admin/supplier_platform*.go`
- provider：`internal/library/supplierplatform/provider/*`
- 路由前缀：`/api/admin/supplier-platform-types`、`/api/admin/supplier-platforms*`
- 权限：`supplier.index`
- 主要能力：平台类型字典、平台账号分页、详情、增删改、启停、余额刷新、余额日志落库、平台关闭后级联关停商品绑定、卡卡云商品详情适配器。
- 边界：`platform_docs/` 保存渠道原始协议，`docs/` 保存本仓库实现说明；卡卡云商品详情因平台接口限制固定走公共域名，下单、查单和其他 provider 仍使用账号配置域名。

### 订单履约

- 协议：`api/open_order.go`、`api/order.go`
- 开放 controller：`internal/controller/open/order.go`
- 后台 controller：`internal/controller/admin/order.go`
- service：`OrderService`（`internal/service/order.go`）
- logic：`internal/logic/order/*.go`
- provider：`internal/library/supplierplatform/provider/*`
- 路由前缀：`/api/open/orders*`、`/api/admin/orders`
- 进入条件：开放订单使用固定 `open_order.token`；后台订单记录使用 `order.manage`。
- 主要能力：开放下单、开放查单、待提交扫描、云发卡提交、订单轮询、窗口内补单、后台订单列表筛选和统计。
- 边界：一期只支持直充、渠道供货商品；对外查单不暴露渠道、成本和利润。

### 短信配置

- 协议：`api/settings_sms.go`（薄入口：`api/settings.go`）
- controller：`internal/controller/admin/settings_sms.go`（声明：`internal/controller/admin/settings.go`）
- service：`SMSConfigService`（`internal/service/sms_config.go`）
- logic：`internal/logic/admin/config.go`
- 路由前缀：`/api/admin/settings/sms`
- 进入条件：super-only
- 主要能力：读取脱敏短信配置状态、保存阿里云 AccessKey、签名、模板和验证码时效配置。

### 系统参数配置

- 协议：`api/settings_system.go`（薄入口：`api/settings.go`）
- controller：`internal/controller/admin/settings_system.go`（声明：`internal/controller/admin/settings.go`）
- service：`SystemConfigService`（`internal/service/system_config.go`）
- logic：`internal/logic/admin/system_config.go`
- 路由前缀：`/api/admin/settings/system`
- 进入条件：super-only
- 主要能力：按分组读取系统参数、一次返回全部分组、兼容旧单组写法、多分组批量保存。
- 边界：当前内置 `finance`、`integration` 两组参数。

### 审计日志

- 协议：`api/log.go`
- controller：`internal/controller/admin/operation_log.go`、`internal/controller/admin/login_log.go`
- service：`AuditLogService`（`internal/service/audit_log.go`）
- logic：`internal/logic/admin/log.go`
- 路由前缀：`/api/admin/logs/*`
- 权限：操作日志 `admin.action`，登录日志 `admin.loginlog`
- 主要能力：操作日志分页查询、登录日志分页查询、管理员 ID 和时间范围过滤、操作日志关键字过滤。

## 目录地图

### `api`

扁平协议目录，只放请求/响应结构、路由协议元信息、协议相关枚举和别名。`api/settings.go` 是设置协议薄入口，具体协议在 `api/settings_sms.go` 和 `api/settings_system.go`。

### `internal/controller`

HTTP 协议适配层。`admin` 承载后台接口，`open` 承载开放订单接口。

### `internal/service`

服务接口边界层，只定义接口，不写实现。

### `internal/logic`

业务编排层。`admin` 承载后台业务域，`order` 承载订单履约链路。

### `internal/app`

运行时核心层，负责配置、MySQL、Redis、短信、区域解析、审计、schema/seed、会话和公共查询能力。

### `internal/library`

跨模块基础能力库，包括 `auth`、`sms`、`audit`、`region`、`supplierplatform/provider`。

### `manifest`

运行配置和 SQL 初始化资源。MySQL 表结构或注释变化必须同步 `manifest/sql/*.sql` 和 `internal/app/schema.go`。

### `test`

契约测试、集成测试和 fixture 说明。测试策略主文档是 `docs/testing.md`。

## 路由与权限摘要

| 模块 | 路由前缀 | 进入条件 |
| --- | --- | --- |
| 认证登录 / 短信发送 / 短信验证 | `/api/admin/auth/login`、`/api/admin/auth/sms/*` | 无需登录 |
| 会话信息 / 退出登录 | `/api/admin/auth/me`、`/api/admin/auth/session` | 需要登录 |
| 员工管理 | `/api/admin/users*` | `admin.list` |
| 用户组与授权 | `/api/admin/groups*`、`/api/admin/menus/tree` | `admin.department` |
| 主体管理 | `/api/admin/subjects*` | `subject.manage` |
| 品牌管理 | `/api/admin/brands*` | `product.brand` |
| 行业管理 | `/api/admin/industries*` | `product.industry` |
| 商品模板管理 | `/api/admin/product-templates*` | `product.template` |
| 商品购买数量限制策略 | `/api/admin/purchase-limit-strategies*` | `product.purchase_limit` |
| 商品管理 | `/api/admin/products*` | `product.goods` |
| 第三方对接 | `/api/admin/supplier-platform-types`、`/api/admin/supplier-platforms*` | `supplier.index` |
| 开放订单 | `/api/open/orders*` | 固定 `open_order.token` |
| 后台订单记录 | `/api/admin/orders` | `order.manage` |
| 短信配置 | `/api/admin/settings/sms` | super-only |
| 系统参数配置 | `/api/admin/settings/system` | super-only |
| 操作日志 | `/api/admin/logs/operations` | `admin.action` |
| 登录日志 | `/api/admin/logs/logins` | `admin.loginlog` |
