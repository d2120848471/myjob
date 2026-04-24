# MyJob Admin Backend

MyJob Admin Backend 是当前仓库根目录下运行的 GoFrame 单体后台项目。
当前代码已经不再依赖历史 `admin/` 子工程；仓库根就是唯一需要维护、启动和发布的后端入口。

## 当前状态

- 主入口在仓库根：`main.go` -> `internal/cmd` -> `internal/bootstrap`
- 后台接口挂在 `/api/admin`，外部开放下单接口挂在 `/api/open`
- 统一响应格式为 `code / message / data`
- 运行时默认使用 MySQL + Redis，配置来源是 `manifest/config/config.local.yaml` 或 `ADMIN_CONFIG`
- OpenAPI 文档默认暴露在 `/api.json`，Swagger UI 默认暴露在 `/swagger/`

## 当前功能概览

### 认证与会话

- 支持账号密码登录
- 当用户是首次登录，或本次登录 IP 与上次登录 IP 不一致时，登录接口会返回：
  - `need_sms_verify=true`
  - `login_token`
  - 脱敏手机号 `phone`
  - 触发原因 `reason`，当前可能是 `first_login` 或 `ip_changed`
- 短信二验通过后才会签发正式登录 token
- 当前登录信息和退出登录都走 `/api/admin/auth/*`

### 短信验证码

- 运行态短信 provider 默认是 `aliyun`
- 测试态默认使用 `mock` sender
- 发送验证码前会先写入 Redis 临时验证码和发送锁
- 如果短信发送失败，会立即删除 Redis 中的验证码和发送锁，避免残留错误登录态
- 验证码校验错误最多允许 5 次，超限后会清理临时登录态和验证码缓存

### 后台业务能力

- 员工管理：列表、新增、编辑、删除、恢复、启停、余额通知、批量设置/取消商务
- 用户组与授权：列表、增删改、状态切换、菜单树、菜单授权、已授权菜单回显
- 主体管理：列表、新增、编辑
- 品牌管理：一级品牌分页、品牌子级懒加载（支持三级）、增删改、排序、显隐切换、本地图片上传
- 行业管理：行业 CRUD、一级品牌选择器、行业品牌关联增删排、行业引用校验
- 商品模板管理：列表、新增、编辑、删除/批删、验证方式枚举
- 商品购买数量限制策略：列表、新增、编辑、删除、启停、枚举
- 商品管理：列表、详情、表单下拉选项、新增、编辑、删除、启停
- 商品渠道绑定：商品列表渠道摘要、库存配置详情/保存、绑定弹窗列表、表单选项、新增、编辑、删除、单条自动改价
- 第三方对接：平台类型字典、平台账号 CRUD/启停、手动余额刷新、余额日志落库、平台关闭后级联关停商品绑定
- 订单履约：外部开放下单/查单、云发卡异步提交与轮询、后台订单记录列表和基础统计
- 短信配置：读取、保存、脱敏展示
- 系统参数配置：按组读取、单组/多组保存、配置校验与批量回滚
- 审计日志：操作日志、登录日志

### 权限模型

- 超级管理员使用 `group_id = 0`
- 普通用户按用户组权限码鉴权
- 短信配置接口是 super-only 能力，普通用户组不会拿到 `config.sms` 菜单授权
- 用户组授权和菜单树只会暴露 `super_only = 0` 的菜单项

## 快速开始

### 1. 启动依赖

```bash
docker compose up -d mysql redis
```

### 2. 检查本地默认超管

- 用户名：`admin`
- 手机号：`15881767197`
- 密码：`abc123`

> 当前仓库默认使用 `manifest/config/config.local.yaml` 中写死的本地超管凭证启动，不再要求每次手工 `export SUPER_ADMIN_*`。

### 3. 检查配置来源

启动时按下面的顺序加载配置：

1. `ADMIN_CONFIG` 指向的配置文件
2. 默认回退到 `manifest/config/config.local.yaml`

例如：

```bash
export ADMIN_CONFIG=manifest/config/config.local.yaml
```

### 4. 启动服务

```bash
go run .
```

默认监听地址来自配置文件，当前本地配置是 `:8080`。

### 5. 上传目录说明

- 品牌图片默认保存在 `storage/uploads/brands/<yyyymmdd>/`
- 对外静态访问前缀默认是 `/uploads`
- 测试态会自动切到临时目录，并在测试结束后清理

### 6. 运行验证

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

> CI 会执行 `go test/go build/golangci-lint`（见 `.github/workflows/ci.yml`）。当前 `.golangci.yml` 启用 `govet/staticcheck/ineffassign/unused/typecheck`。
>
> 其中 `test` job 会先启动本地 MySQL service，并复用默认 DSN `root:root123456@tcp(127.0.0.1:3306)/admin` 的口径，确保远端也能创建并重置测试库 `admin_test`。
>
> 契约测试和大部分应用级测试会通过 `NewTestCore()` 使用 MySQL 测试库 `admin_test`；测试启动时会自动建库并清空旧表，不会污染日常运行的 `admin`。
>
> MySQL 表结构和字段注释存在两条维护链路：Docker 首次建库使用 `manifest/sql/*.sql`，应用启动自建表使用 `internal/app/schema.go`。凡是修改 MySQL schema 或 `COMMENT`，必须同步这两处。

## 目录说明

```text
.
├── README.md
├── api/
│   ├── auth.go
│   ├── brand.go
│   ├── common.go
│   ├── group.go
│   ├── industry.go
│   ├── log.go
│   ├── open_order.go
│   ├── order.go
│   ├── product_goods.go
│   ├── product_goods_channel.go
│   ├── product_goods_channel_config.go
│   ├── product_template.go
│   ├── purchase_limit.go
│   ├── settings.go
│   ├── settings_sms.go
│   ├── settings_system.go
│   ├── subject.go
│   ├── supplier_platform.go
│   └── user.go
├── docs/
│   ├── architecture.md
│   ├── development.md
│   ├── migration.md
│   ├── module-map.md
│   ├── overview.md
│   ├── testing.md
│   └── superpowers/
│       └── README.md
├── platform_docs/
│   ├── README.md
│   └── *.md
├── hack/
├── internal/
│   ├── app
│   ├── bootstrap
│   ├── cmd
│   ├── consts
│   ├── controller/admin
│   ├── controller/open
│   ├── dao
│   ├── library
│   ├── logic/admin
│   ├── logic/order
│   ├── middleware
│   ├── model
│   └── service
├── manifest/
│   ├── config
│   └── sql
├── resource/
└── test/
    ├── contract
    ├── fixture
    └── integration
```

### `api/`

`api/` 只放对外请求/响应协议定义，不放业务逻辑。
当前协议目录已经拍平成仓库根下的 `api/*.go`，不再使用历史嵌套协议包路径。

说明：

- `api/settings.go` 为薄入口/说明文件；短信配置与系统参数配置协议拆分到 `api/settings_sms.go` 与 `api/settings_system.go`
- `api/product_goods.go` 保留商品主档协议；商品渠道绑定与库存配置协议独立放到 `api/product_goods_channel.go`、`api/product_goods_channel_config.go`
- `api/open_order.go` 放外部开放下单/查单协议，`api/order.go` 放后台订单记录协议
- `api/common.go` 只保留跨业务域复用的通用别名（例如分页），单域 `Item/Enum` 放回对应领域协议文件

### `docs`

项目内部手写文档目录，主要记录当前仓库自己的架构、模块边界、开发方式、测试方式和迁移背景。
其中 `docs/superpowers/README.md` 是规格与实施计划类文档的薄入口；只有当某轮需求确实沉淀了可复用的 spec 或 plan 时，才继续在该目录下补充对应文档。

### `platform_docs`

第三方渠道原始对接文档归档目录。
`platform_docs/README.md` 维护平台总览，其他 `platform_docs/*.md` 按渠道拆分，保留外部接口的签名规则、字段约束、状态码、模块说明和示例，供第三方供货平台接入、余额刷新排障和 provider 实现对照使用。

该目录和 `docs/` 的边界如下：

- `platform_docs/` 记录渠道侧原始协议与接口资料
- `docs/` 记录本仓库自己的实现、结构、测试与运行说明

### `internal/bootstrap`

负责把配置、运行时 `Core`、控制器、服务、路由和中间件组装成可运行应用。

### `internal/controller/admin`

负责 HTTP 协议适配，方法签名统一是 GoFrame 标准 `Req/Res + error`。

### `internal/controller/open`

负责开放接口协议适配，当前提供 `/api/open/orders` 创建订单与查单入口，不依赖后台登录态。

### `internal/service`

定义 controller 依赖的服务接口，是模块边界层。

### `internal/logic/admin`

负责业务编排和规则控制，调用 `app`、`library` 与数据库访问能力。

该层普遍采用“同 package 多文件按职责拆分”，例如：

- `brand*.go`
- `industry*.go`
- `product_goods*.go`
- `product_goods_channel*.go`
- `product_template*.go`
- `purchase_limit*.go`
- `user*.go`

### `internal/logic/order`

负责订单履约编排，当前按职责拆分为创建、查单、后台列表、选渠、提交、轮询、补单和 worker 生命周期等文件。订单逻辑不放进商品、第三方平台或通用 helper 文件。

### `internal/app`

负责运行时核心能力，包括：

- 配置加载
- MySQL / Redis 初始化
- 会话签发与校验
- 菜单与种子初始化
- 短信配置读取
- IP 归属地解析
- 审计写入辅助

当前公共能力已按职责拆分为多个同 package 文件（例如 `pagination.go`、`mask.go`、`menu_tree.go`、`auth_session.go` 等），避免继续堆回历史 `helpers.go`。
其中 `schema.go` 负责应用启动期的内置建表语句；如果改 MySQL 表结构或表/字段注释，需要和 `manifest/sql/*.sql` 保持同步。

### `internal/library`

跨模块基础能力库：

- `auth`：JWT、Redis key、会话存储
- `sms`：短信 sender 抽象、mock sender、阿里云 sender
- `audit`：审计日志写入器
- `region`：IP 归属地解析与脱敏工具
- `supplierplatform/provider`：第三方供货平台适配器（余额刷新等对接能力）
  - 其中云发卡适配器同时支持开放订单下单和查单

### `manifest/config`

当前只保留本地真实开发配置 `config.local.yaml`。

### `manifest/sql`

初始化 SQL 资源目录，包含 schema、菜单、超级管理员模板和系统配置种子。
这里的 MySQL 建表语句用于 Docker 首次初始化数据库；如果调整 MySQL 表结构或表/字段注释，需要和 `internal/app/schema.go` 一起修改。

- `001_schema.sql`：基础表结构
- `002_seed_menu.sql`：菜单与权限种子
- `003_seed_admin.sql.tmpl`：超级管理员初始化模板
- `004_seed_config.sql`：系统参数初始值
- `005_supplier_platform.sql`：第三方平台类型、账号启停状态和余额日志结构
- `006_product_goods_channel_binding.sql`：商品渠道绑定表结构
- `007_product_goods_channel_config.sql`：商品库存配置表结构
- `008_external_order.sql`：外部订单主表和渠道尝试表结构

### `hack`

- `hack/bootstrap-admin.sh`：根据手机号和 bcrypt hash 生成超级管理员 SQL
- `hack/gen-dao.sh`：按连接串生成 DAO / DO / Entity / table 元数据

### `test/contract`

契约测试入口，当前主要验证：

- 扁平 API 协议目录
- 统一响应包裹
- OpenAPI / Swagger 暴露
- 登录 / 短信二验 / 会话
- 系统参数配置的单组读取、多组批量保存、super-only 鉴权与回滚
- 品牌、行业、本地上传主流程
- 商品模板、购买数量限制策略、商品管理、商品渠道绑定、第三方对接主流程
- 开放订单创建/查单、后台订单列表权限、筛选和统计
- 员工、用户组、主体、短信配置、日志等核心流程

### `test/integration`

集成测试目录当前包含 runtime smoke test、第三方平台余额刷新集成回归和订单 worker 集成回归。
其中 runtime smoke test 仍由显式环境开关控制；整体也不是完整的 MySQL / Redis / 短信 / 日志闭环回归集。

## 文档索引

- `docs/overview.md`：项目定位与当前功能状态
- `docs/architecture.md`：启动链路、请求流、认证流、权限与运行时结构
- `docs/module-map.md`：模块、接口前缀、权限与目录地图
- `docs/development.md`：开发依赖、配置、脚本和日常命令
- `docs/testing.md`：测试分层、当前覆盖范围与执行命令
- `docs/migration.md`：从历史后台迁到当前根应用结构的迁移背景
- `docs/superpowers/README.md`：规格与实施计划文档入口；当前记录本轮开放订单与云发卡异步履约计划
- `platform_docs/README.md`：第三方渠道原始接口文档总览；具体平台文档按 `platform_docs/*.md` 拆分
