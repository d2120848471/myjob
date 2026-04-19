# 模块职责地图

## 业务域到实现目录

### 认证与会话

- 协议：`api/auth.go`
- controller：`internal/controller/admin/auth.go`、`internal/controller/admin/session.go`
- service：`AuthService`（`internal/service/auth.go`）
- logic：`internal/logic/admin/auth.go`
- 路由前缀：`/api/admin/auth/*`
- 主要能力：
  - 账号密码登录
  - 条件短信二验
  - 发送验证码 / 校验验证码
  - 获取当前登录信息
  - 退出登录

### 员工管理

- 协议：`api/user.go`
- controller：`internal/controller/admin/user.go`
- service：`UserService`
- logic：`internal/logic/admin/user*.go`
- 路由前缀：`/api/admin/users*`
- 主要能力：
  - 员工列表
  - 回收站
  - 新增 / 编辑 / 删除 / 恢复
  - 启停
  - 余额通知开关
  - 批量设置 / 取消商务

### 用户组与授权

- 协议：`api/group.go`
- controller：`internal/controller/admin/group.go`
- service：`GroupService`
- logic：`internal/logic/admin/group.go`
- 路由前缀：`/api/admin/groups*`、`/api/admin/menus/tree`
- 主要能力：
  - 用户组列表、增删改、状态切换
  - 读取用户组已有菜单授权
  - 菜单树
  - 保存用户组菜单授权
- 权限边界：
  - 用户组授权只允许分配 `super_only = 0` 的菜单
  - 菜单树只返回 `super_only = 0` 的菜单项

### 主体管理

- 协议：`api/subject.go`
- controller：`internal/controller/admin/subject.go`
- service：`SubjectService`
- logic：`internal/logic/admin/subject.go`
- 路由前缀：`/api/admin/subjects*`
- 主要能力：
  - 主体列表
  - 新增主体
  - 编辑主体

### 第三方对接

- 协议：`api/supplier_platform.go`
- controller：`internal/controller/admin/supplier_platform.go`
- service：`SupplierPlatformService`
- logic：`internal/logic/admin/supplier_platform*.go`、`internal/logic/admin/supplier_platform_balance.go`
- provider 适配层：`internal/library/supplierplatform/provider/*`
- 路由前缀：`/api/admin/supplier-platform-types`、`/api/admin/supplier-platforms*`
- 主要能力：
  - 平台类型字典
  - 平台账号分页、详情、增删改、启停
  - 主体关联与平台侧 `has_tax`
  - 手动余额刷新
  - 余额刷新日志落库
  - 平台关闭时同步关停该平台下的商品渠道绑定，重新开启不自动恢复绑定
- 外部协议参考：
  - `platform_docs/README.md`
  - `platform_docs/*.md`
- 权限边界：
  - 第三方对接接口统一要求 `supplier.index`
  - 一期只写余额日志，不开放日志查询 HTTP 接口
- 文档边界：
  - `platform_docs/` 记录渠道原始接口与签名规则
  - `docs/*.md` 记录本仓库怎样落地这些渠道能力

### 品牌管理

- 协议：`api/brand.go`
- controller：`internal/controller/admin/brand.go`
- service：`BrandService`
- logic：`internal/logic/admin/brand*.go`
- 路由前缀：`/api/admin/brands*`
- 主要能力：
  - 一级品牌分页与搜索
  - 品牌子级懒加载（支持三级）
  - 新增 / 编辑 / 删除
  - 同级排序与显隐切换
  - 本地图片上传
- 权限边界：
  - 品牌接口统一要求 `product.brand`

### 行业管理

- 协议：`api/industry.go`
- controller：`internal/controller/admin/industry.go`
- service：`IndustryService`
- logic：`internal/logic/admin/industry*.go`
- 路由前缀：`/api/admin/industries*`
- 主要能力：
  - 行业列表、增删改、排序
  - 一级品牌选择器
  - 行业品牌关联增删排序
  - 行业删除 / 品牌删除前的关联校验
- 权限边界：
  - 行业接口统一要求 `product.industry`
  - 行业只接受一级品牌关联，二级和三级品牌不会出现在选择器里，也不能被保存

### 商品模板管理

- 协议：`api/product_template.go`
- controller：`internal/controller/admin/product_template.go`
- service：`ProductTemplateService`
- logic：`internal/logic/admin/product_template*.go`
- 路由前缀：`/api/admin/product-templates*`
- 主要能力：
  - 本地模板分页查询、关键词搜索
  - 模板类型筛选、共享状态筛选
  - 新增 / 编辑 / 单删 / 批删
  - 读取验证方式枚举 `/api/admin/product-templates/validate-types`
- 当前字段：
  - `title`：模板名称
  - `type`：模板类型，本期仅支持 `local`
  - `is_shared`：共享状态，仅允许 `0` 或 `1`
  - `account_name`：充值账号名称
  - `validate_type`：验证方式枚举 ID
- 当前内置验证方式枚举：
  - `1` 手机号
  - `2` QQ号
  - `3` 手机号或者QQ号
  - `4` 邮箱
  - `5` 网址
  - `6` 纯数字
  - `7` 微信号
  - `8` 手机号或者微信号
  - `9` QQ号或者微信号
  - `10` 手机号或者QQ号或微信号
  - `11` 禁止填写手机号
  - `12` 禁止填写邮箱
- 权限边界：
  - 商品模板接口统一要求 `product.template`
- 当前实现限制：
  - 本期只做模板管理，不做模板内容渲染或动态表单配置
  - 当前只保存 `validate_type` 规则，不会据此校验某次真实输入值是不是微信号、QQ号、邮箱等

### 商品购买数量限制策略

- 协议：`api/purchase_limit.go`
- controller：`internal/controller/admin/purchase_limit.go`
- service：`PurchaseLimitService`
- logic：`internal/logic/admin/purchase_limit*.go`
- 路由前缀：`/api/admin/purchase-limit-strategies*`
- 主要能力：
  - 分页列表、关键词搜索
  - 新增 / 编辑 / 删除
  - 启停状态切换
  - 读取限制类型和周期类型枚举
- 当前字段：
  - `name`：策略名称
  - `limit_type`：限制类型，当前支持同一会员 / 同一充值账号
  - `period_type`：周期类型，当前支持按天 / 按区间(分钟)
  - `period`：限制周期，必须大于 0
  - `limit_nums`：限制数量，`0` 表示不限制
  - `limit_times`：限制笔数，`0` 表示不限制
  - `status`：启停状态
- 权限边界：
  - 商品购买数量限制策略接口统一要求 `product.purchase_limit`
- 当前实现限制：
  - 本期只做策略后台管理
  - 不实现老站的“清空策略数据”
  - 不接真实购买前校验链

### 商品管理

- 协议：`api/product_goods.go`、`api/product_goods_channel.go`
- controller：`internal/controller/admin/product_goods.go`、`internal/controller/admin/product_goods_channel.go`
- service：`ProductGoodsService`
- logic：`internal/logic/admin/product_goods*.go`、`internal/logic/admin/product_goods_channel*.go`
- 路由前缀：`/api/admin/products*`、`/api/admin/products/{goodsId}/channel-bindings*`
- 权限码：`product.goods`
- 主要能力：
  - 商品列表、详情、表单选项、新增、编辑、删除、启停
  - 商品列表渠道摘要：已绑定渠道、主渠道、最低进货价、最低利润后价格、自动改价状态
  - 商品渠道绑定弹窗：列表、表单选项、新增、编辑、删除、单条自动改价
  - 仅允许选择已启用的平台账号；关闭平台后摘要与绑定列表不再把该平台视为可用渠道
- 当前价格口径：
  - 绑定列表和商品列表展示的“进货价”统一返回税态换算后的 `cost_price`
  - 利润后价格单独返回 `effective_sell_price`

### 短信配置

- 协议：`api/settings_sms.go`（薄入口：`api/settings.go`）
- controller：`internal/controller/admin/settings_sms.go`（声明：`internal/controller/admin/settings.go`）
- service：`SMSConfigService`
- logic：`internal/logic/admin/config.go`
- 路由前缀：`/api/admin/settings/sms`
- 主要能力：
  - 读取脱敏后的短信配置状态
  - 保存阿里云 AccessKey、签名、模板和验证码时效配置
- 权限边界：
  - 该模块是 super-only 接口
  - 普通用户组不会获得 `config.sms` 菜单权限

### 系统参数配置

- 协议：`api/settings_system.go`（薄入口：`api/settings.go`）
- controller：`internal/controller/admin/settings_system.go`（声明：`internal/controller/admin/settings.go`）
- service：`SystemConfigService`
- logic：`internal/logic/admin/system_config.go`
- 路由前缀：`/api/admin/settings/system`
- 主要能力：
  - 按分组读取单组系统参数，或一次返回全部分组
  - 兼容旧单组写法，同时支持多分组批量保存
  - 当前内置 `finance`、`integration` 两组参数
- 权限边界：
  - 该模块是 super-only 接口
  - 普通用户组不会获得 `config.system` 菜单权限

### 审计日志

- 协议：`api/log.go`
- controller：`internal/controller/admin/operation_log.go`、`internal/controller/admin/login_log.go`
- service：`AuditLogService`
- logic：`internal/logic/admin/log.go`
- 路由前缀：`/api/admin/logs/*`
- 主要能力：
  - 操作日志分页查询
  - 登录日志分页查询
  - 支持管理员 ID、时间范围过滤
  - 操作日志额外支持关键字过滤

### 公共能力（internal/app）

`internal/app` 是运行期核心能力层，当前已按职责拆分为多个同 package 文件（避免继续堆回 `helpers.go`）：

- `helpers.go`：历史入口说明（薄入口）
- `pagination.go`：分页工具与统一分页结构
- `mask.go`：敏感信息脱敏（AccessKey/Secret 等）
- `menu_tree.go`：菜单树与授权回显组装
- `auth_session.go`：登录态签发、校验与会话存储
- `sms_config.go`：短信配置读取与缓存
- `audit.go`：审计写入辅助
- `user_lookup.go`：登录用户/员工查询辅助
- `redis_helpers.go`：Redis key/TTL 等基础辅助

其中 `schema.go` 负责应用启动期的内置建表语句；MySQL 表结构和表/字段注释调整时，需要和 `manifest/sql/*.sql` 同步维护。

## 目录地图

### `api`

对外协议目录，当前是扁平文件结构：

- `auth.go`
- `brand.go`
- `common.go`
- `group.go`
- `industry.go`
- `log.go`
- `product_goods.go`
- `product_goods_channel.go`
- `product_template.go`
- `purchase_limit.go`
- `settings.go`
- `settings_sms.go`
- `settings_system.go`
- `supplier_platform.go`
- `subject.go`
- `user.go`

### `docs`

项目内部手写文档目录，负责记录当前仓库自己的架构、模块边界、开发流程、测试方式和迁移说明。

### `platform_docs`

第三方渠道接口文档归档目录。
`platform_docs/README.md` 负责平台总览，其他 `platform_docs/*.md` 一平台一文件，保留外部渠道的签名规则、字段约束、状态码、接口模块和示例，供 `internal/library/supplierplatform/provider/*` 的实现和排障时对照。

### `internal/bootstrap`

应用装配层，负责：

- 创建 controller / service / logic 组合
- 绑定 `/api/admin` 路由
- 注册统一响应中间件和鉴权中间件
- 注册 OpenAPI / Swagger

### `internal/controller/admin`

HTTP 协议适配层，不直接写业务规则。

### `internal/service`

模块接口边界层。

### `internal/logic/admin`

业务编排层，当前按业务域拆分为：

- `auth.go`
- `user.go`
- `group.go`
- `subject.go`
- `brand.go`
- `industry.go`
- `product_common.go`
- `product_goods*.go`
- `product_template.go`
- `purchase_limit.go`
- `supplier_platform*.go`
- `supplier_platform_balance.go`
- `money.go`
- `db_record_time.go`
- `config.go`
- `system_config.go`
- `log.go`

### `internal/app`

运行时核心层，负责配置、依赖初始化、种子引导、公共查询和通用辅助能力。
其中 `schema.go` 维护应用启动自建表所需的 MySQL / SQLite schema；MySQL 注释属于这里的职责边界。

### `internal/library`

基础能力库：

- `auth`
- `sms`
- `audit`
- `region`

### `internal/model`

当前已使用的模型子目录：

- `config`：配置结构体
- `do`：写入 / 条件对象
- `dto/admin`：分页 DTO 等内部传输对象
- `entity`：数据库实体与查询结果结构
- `runtime`：登录态、短信配置等运行时模型

### `manifest/config`

运行时配置目录，当前主要是 `config.local.yaml`，其中本地默认 MySQL 连接使用 `127.0.0.1:3306/admin`，并固定超级管理员账号 `admin / 15881767197 / abc123`。

### `manifest/sql`

数据库 schema、菜单种子、超级管理员模板和系统配置种子。

- `001_schema.sql`：基础表结构
- `002_seed_menu.sql`：菜单与权限种子
- `003_seed_admin.sql.tmpl`：超级管理员初始化模板
- `004_seed_config.sql`：系统参数初始值
- `005_supplier_platform.sql`：第三方平台类型、账号和余额日志结构
- `006_product_goods_channel_binding.sql`：商品渠道绑定表结构

这里的 schema 文件用于 Docker 首次初始化 MySQL；如果修改 MySQL 表结构、表注释或字段注释，必须和 `internal/app/schema.go` 一起修改，避免首次建库和应用自建表出现漂移。

### `test/contract`

契约测试目录，覆盖接口兼容和核心业务流。

当前也覆盖商品列表渠道摘要与商品渠道绑定弹窗的主流程。

### `test/integration`

集成测试目录，当前包括 runtime smoke test、第三方平台余额刷新回归和可选 live 验证。

## 当前路由与权限摘要

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
| 短信配置 | `/api/admin/settings/sms` | super-only |
| 系统参数配置 | `/api/admin/settings/system` | super-only |
| 操作日志 | `/api/admin/logs/operations` | `admin.action` |
| 登录日志 | `/api/admin/logs/logins` | `admin.loginlog` |
