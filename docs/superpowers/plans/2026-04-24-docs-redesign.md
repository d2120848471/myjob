# 文档体系重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将项目文档重构为“项目首页 + 文档门户 + 专题文档”的稳定体系，修复过时目录和功能描述漂移。

**Architecture:** 根 `README.md` 只做项目入口，`docs/README.md` 做文档门户，专题文档按解释型、参考型、How-to 和历史背景分工。业务域和路由权限事实集中到 `docs/module-map.md`，测试和 CI/lint 口径集中到 `docs/testing.md`，避免多处重复维护。

**Tech Stack:** Markdown、GoFrame 项目现有目录、Go test、golangci-lint、Git。

---

## File Structure

- Create: `docs/README.md`
  - 文档门户，按阅读目的指向 README、overview、architecture、module-map、development、testing、migration、superpowers、platform_docs。
- Create: `docs/superpowers/README.md`
  - 规格和实施计划入口，解释 `specs/` 与 `plans/` 的用途。
- Modify: `README.md`
  - 收缩为项目首页，保留定位、快速开始、验证命令、关键边界和文档索引。
- Modify: `docs/overview.md`
  - 作为解释型文档，只描述项目定位、当前能力、非目标和阅读建议。
- Modify: `docs/architecture.md`
  - 作为解释型文档，保留启动链路、分层职责、请求链路、认证/权限/短信/订单 worker 流程和运行时依赖。
- Modify: `docs/module-map.md`
  - 作为参考型文档，集中维护业务域到协议/controller/service/logic/路由/权限/边界的映射。
- Modify: `docs/development.md`
  - 作为 How-to 文档，保留本地开发、配置、SQL/schema、DAO 和常用命令。
- Modify: `docs/testing.md`
  - 作为测试主参考，集中维护测试分层、命令、CI/lint 和外部依赖边界。
- Modify: `docs/migration.md`
  - 压缩为历史背景文档，不再维护完整当前目录清单。
- Modify: `test/contract/README.md`
  - 收敛为契约测试目录内最小说明，并回链 `docs/testing.md`。
- Modify: `test/integration/README.md`
  - 收敛为集成测试目录内最小说明，并回链 `docs/testing.md`。
- Leave unchanged unless a fact check shows a direct mismatch: `platform_docs/README.md`
  - 第三方原始接口资料总览，不混入本仓库实现说明。

## Task 1: Add Documentation Portal And Superpowers Entry

**Files:**
- Create: `docs/README.md`
- Create: `docs/superpowers/README.md`
- Reference: `docs/superpowers/specs/2026-04-24-docs-redesign-design.md`

- [ ] **Step 1: Confirm current tracked docs tree**

Run:

```bash
git ls-files README.md docs test/contract/README.md test/integration/README.md platform_docs/README.md
```

Expected: output includes `README.md`, six existing `docs/*.md` files, `platform_docs/README.md`, and the existing design spec under `docs/superpowers/specs/`.

- [ ] **Step 2: Create `docs/README.md`**

Write this content:

```markdown
# 文档入口

本目录记录 MyJob 后端当前真实的架构、模块边界、开发方式、测试方式和迁移背景。

## 按目的阅读

- 快速启动项目：先读 `../README.md`，再读 `development.md`。
- 理解系统定位和能力边界：读 `overview.md`。
- 理解启动链路、分层职责和请求流：读 `architecture.md`。
- 查询业务域、路由、权限和文件归属：读 `module-map.md`。
- 跑测试、对齐 CI 或确认 lint 口径：读 `testing.md`。
- 理解历史迁移背景：读 `migration.md`。
- 查看需求设计和实施计划：读 `superpowers/README.md`。
- 查询第三方渠道原始接口资料：读 `../platform_docs/README.md`。

## 维护原则

- 当前事实只写一次：模块归属放在 `module-map.md`，测试策略放在 `testing.md`，开发命令放在 `development.md`。
- Markdown 不重复维护完整字段级 API 文档；字段细节以 `../api/*.go` 和运行时 OpenAPI `/api.json` 为准。
- 代码、路由、测试、CI 或目录结构变化时，同步检查本目录和根 `README.md`。
```

- [ ] **Step 3: Create `docs/superpowers/README.md`**

Write this content:

```markdown
# Superpowers Specs And Plans

本目录保存经过确认的需求设计、实施计划和阶段性规格文档。

## 目录约定

- `specs/`：设计规格，记录目标、范围、非目标、信息架构、验收标准和风险。
- `plans/`：实施计划，记录可执行任务、涉及文件、验证命令和提交节奏。

## 当前文档

- `specs/2026-04-24-docs-redesign-design.md`：文档体系重构设计。
- `plans/2026-04-24-docs-redesign.md`：文档体系重构实施计划。

## 维护要求

- spec 必须先被确认，再写 plan。
- plan 必须能被没有上下文的执行者按步骤推进。
- 已完成的实现结果应回写到根 `README.md` 或 `docs/*.md`，不要只停留在计划文档中。
```

- [ ] **Step 4: Verify portal links are syntactically present**

Run:

```bash
rg -n "docs/README.md|superpowers/README.md|platform_docs/README.md|module-map.md|testing.md" README.md docs/README.md docs/superpowers/README.md docs/superpowers/specs/2026-04-24-docs-redesign-design.md
```

Expected: no command error; output shows the new portal and superpowers paths.

- [ ] **Step 5: Commit Task 1**

Run:

```bash
git add docs/README.md docs/superpowers/README.md
git commit -m "docs: 添加文档门户入口"
```

Expected: commit succeeds with only the two new entry files.

## Task 2: Rewrite Root README As Project Entry

**Files:**
- Modify: `README.md`
- Reference: `docs/README.md`
- Reference: `docs/development.md`
- Reference: `docs/testing.md`

- [ ] **Step 1: Inspect current README sections**

Run:

```bash
rg -n "^#|^##|^###|docs/superpowers|目录说明|文档索引|当前功能概览|快速开始|运行验证" README.md
```

Expected: output shows the existing large README section layout and the stale `docs/superpowers/README.md` references.

- [ ] **Step 2: Replace README with concise project entry**

Rewrite `README.md` with this structure and content. Keep the wording concise; do not reintroduce the long module-by-module directory explanation.

```markdown
# MyJob Admin Backend

MyJob Admin Backend 是运行在仓库根目录的 GoFrame 单体后台项目。仓库根就是唯一需要维护、启动和发布的后端入口，历史 `admin/` 子工程不再作为运行入口。

## 当前状态

- 主入口：`main.go` -> `internal/cmd` -> `internal/bootstrap`
- 后台接口前缀：`/api/admin`
- 开放订单接口前缀：`/api/open`
- 统一响应结构：`code / message / data`
- 协议目录：扁平 `api/*.go`
- 运行依赖：MySQL + Redis
- 默认配置：`ADMIN_CONFIG` 指定文件，未指定时回退到 `manifest/config/config.local.yaml`
- OpenAPI：`/api.json`
- Swagger UI：`/swagger/`

## 核心能力

- 认证与会话：账号密码登录、条件短信二验、当前登录信息、退出登录。
- 后台管理：员工、用户组与授权、主体、品牌、行业、商品模板、购买数量限制策略、商品、商品渠道绑定和库存配置。
- 设置与审计：短信配置、系统参数配置、操作日志、登录日志。
- 第三方对接：平台类型字典、平台账号管理、余额刷新、余额日志落库。
- 订单履约：开放下单/查单、云发卡异步提交与轮询、后台订单记录列表和基础统计。

## 快速开始

### 1. 启动依赖

```bash
docker compose up -d mysql redis
```

默认端口：

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6380`

### 2. 使用本地默认超管

- 用户名：`admin`
- 手机号：`15881767197`
- 密码：`abc123`

默认凭证来自 `manifest/config/config.local.yaml`，本地启动不要求手工导出 `SUPER_ADMIN_*`。

### 3. 启动服务

```bash
go run .
```

默认监听地址来自配置文件，当前本地配置是 `:8080`。

### 4. 运行验证

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

CI 会执行 test、build 和 lint。详细测试分层、MySQL 测试库和 live 验证开关见 `docs/testing.md`。

## 关键维护边界

- `api/` 只放请求/响应协议，不放业务逻辑。
- controller 只做 HTTP 协议适配，不直接访问 DAO。
- service 只定义接口。
- logic 负责业务编排、校验、事务边界和数据聚合。
- `internal/app` 和 `internal/bootstrap` 只做运行时能力和应用装配。
- MySQL schema 修改必须同步 `manifest/sql/*.sql` 和 `internal/app/schema.go`。
- 文档描述必须以当前真实代码行为为准，不提前描述尚未实现的能力。

## 文档索引

- `docs/README.md`：文档门户，按阅读目的选择入口。
- `docs/overview.md`：项目定位、当前能力和非目标。
- `docs/architecture.md`：启动链路、分层职责、请求流和运行时依赖。
- `docs/module-map.md`：业务域、协议、controller、service、logic、路由和权限映射。
- `docs/development.md`：本地开发、配置、SQL/schema 和 DAO 说明。
- `docs/testing.md`：测试分层、执行命令、CI/lint 和外部依赖边界。
- `docs/migration.md`：历史迁移背景。
- `docs/superpowers/README.md`：规格和实施计划文档入口。
- `platform_docs/README.md`：第三方渠道原始接口资料总览。
```

- [ ] **Step 3: Verify README no longer contains stale directory tree**

Run:

```bash
rg -n "docs/superpowers/README.md|docs/README.md|目录说明|├──|└──|当前功能概览" README.md
```

Expected: output includes `docs/README.md` and `docs/superpowers/README.md`; output does not include tree characters `├──` or `└──`; output does not include old section title `当前功能概览`.

- [ ] **Step 4: Commit Task 2**

Run:

```bash
git add README.md
git commit -m "docs: 收敛项目首页说明"
```

Expected: commit succeeds with only `README.md` modified.

## Task 3: Rewrite Overview And Architecture As Explanation Docs

**Files:**
- Modify: `docs/overview.md`
- Modify: `docs/architecture.md`
- Reference: `internal/bootstrap/application.go`
- Reference: `internal/app/core.go`
- Reference: `internal/app/bootstrap.go`
- Reference: `internal/logic/order/order_worker.go`
- Reference: `internal/middleware/response.go`
- Reference: `internal/middleware/auth.go`

- [ ] **Step 1: Inspect current architecture facts**

Run:

```bash
rg -n "NewApplicationFromEnv|NewCoreFromEnv|SetOpenApiPath|SetSwaggerPath|WorkerEnabled|middleware.Response|AuthGuard|Core.Close|open_order" internal/bootstrap internal/app internal/middleware internal/logic/order docs/architecture.md docs/overview.md
```

Expected: output confirms startup chain, OpenAPI paths, worker enablement, response middleware and auth guard references.

- [ ] **Step 2: Rewrite `docs/overview.md`**

Use these sections and keep details at capability level, not file-list level:

```markdown
# 项目概览

## 项目定位

当前仓库是运行在仓库根目录的 GoFrame 单体后台。项目目标是保持既有后台业务语义，同时把后端整理成边界清晰、可维护、可测试的结构。

历史 `admin/` 子工程不再是运行入口，仓库根目录代码是唯一主实现。

## 当前运行形态

- 主进程入口：`main.go`
- 启动命令入口：`internal/cmd`
- 应用装配：`internal/bootstrap`
- 运行时核心：`internal/app`
- 配置优先级：`ADMIN_CONFIG` > `manifest/config/config.local.yaml`
- 默认依赖：MySQL 8.4、Redis 7
- 后台接口：`/api/admin/*`
- 开放订单接口：`/api/open/*`

## 当前能力

### 认证、权限与设置

- 账号密码登录。
- 首次登录或登录 IP 变化时触发短信二验。
- 超级管理员固定使用 `group_id = 0`。
- 普通用户通过用户组菜单权限码鉴权。
- 短信配置和系统参数配置属于 super-only 能力。

### 后台业务域

- 员工管理、用户组与授权、主体管理。
- 品牌管理、行业管理、商品模板管理。
- 商品购买数量限制策略、商品管理、商品渠道绑定、商品库存配置。
- 第三方平台账号管理、余额刷新和余额日志落库。
- 操作日志和登录日志查询。

### 订单履约

- 外部调用方通过固定 `open_order.token` 创建订单和查单。
- 一期只支持直充、渠道供货商品。
- 云发卡订单通过进程内 worker 异步提交和轮询。
- 后台订单记录支持筛选和基础统计。
- 对外查单不暴露渠道、成本和利润。

## 当前边界

- 不拆微服务。
- 不预埋没有真实使用场景的平台化基础设施。
- 不在文档里描述尚未实现的接口或测试覆盖。
- 第三方原始接口资料放在 `../platform_docs/`，本仓库实现说明放在 `docs/`。

## 阅读建议

- 快速启动：`../README.md`
- 架构和请求流：`architecture.md`
- 模块、路由和权限：`module-map.md`
- 开发命令：`development.md`
- 测试和 CI：`testing.md`
```

- [ ] **Step 3: Rewrite `docs/architecture.md`**

Keep the startup chain and flows, but remove duplicated business-module inventory. Use these sections:

```markdown
# 架构说明

## 启动链路

```text
main.go
  -> cmd.Main.Run()
    -> bootstrap.NewApplicationFromEnv()
      -> app.NewCoreFromEnv()
        -> app.NewCoreFromConfigFile()
          -> loadConfig()
            -> model/config.LoadFromGoFrame()
      -> newCore()
            -> initStores()
            -> bootstrap()
      -> assemble()
        -> 按配置启动订单 worker（可选）
      -> ghttp.Server.Run()
```

`main.go` 只进入命令入口；`internal/app` 初始化配置、MySQL、Redis、种子和运行时资源；`internal/bootstrap` 负责路由、中间件、controller、service 和 logic 的组合。

## 分层职责

- `api/`：请求/响应协议和 OpenAPI 元信息。
- `internal/controller/admin`：后台 HTTP 协议适配。
- `internal/controller/open`：开放订单 HTTP 协议适配。
- `internal/service`：controller 依赖的接口边界。
- `internal/logic/admin`：后台业务编排、校验、事务和审计。
- `internal/logic/order`：开放订单、后台订单列表、提交、轮询、补单和 worker 生命周期。
- `internal/app`：配置、数据库、Redis、短信、区域解析、审计、schema/seed 和会话等运行时能力。
- `internal/library`：跨模块基础能力和第三方 provider 适配。

## 请求处理链路

```text
HTTP Request
  -> middleware.Response
  -> middleware.AuthGuard（按路由分组决定是否启用）
  -> controller/admin 或 controller/open
  -> service interface
  -> logic/admin 或 logic/order
  -> app / dao / library
  -> MySQL / Redis / provider
```

`middleware.Response` 统一包装 `code / message / data`。后台受保护路由通过 `AuthGuard` 解析 Bearer token、加载用户并校验权限；`/api/open/orders*` 不走后台 Bearer 鉴权，由订单 logic 校验开放 token。

## 认证与短信流程

### 登录

```text
POST /api/admin/auth/login
  -> 校验用户名和密码
  -> 校验用户状态
  -> 判断是否需要短信二验
       -> 不需要：签发正式 token
       -> 需要：写入临时登录态并返回 login_token
```

短信二验触发条件是首次登录或登录 IP 变化。

### 发送验证码

```text
POST /api/admin/auth/sms/send
  -> 校验 login_token
  -> 读取短信配置
  -> 写入验证码缓存和发送锁
  -> 调用 sender
       -> 失败：删除验证码缓存和发送锁
       -> 成功：等待用户提交验证码
```

### 校验验证码

```text
POST /api/admin/auth/sms/verify
  -> 校验 login_token 和 sms_code
  -> 错误时累计 attempts
  -> 超过 5 次后清理临时登录态和验证码
  -> 成功后签发正式 token
```

## 权限模型

- 超级管理员：`group_id = 0`。
- 普通用户：通过用户组读取菜单权限码。
- 用户组授权和菜单树只暴露 `super_only = 0` 的菜单。
- 短信配置和系统参数配置走 super-only 保护。

## 运行时依赖

- MySQL：真实运行数据库，本地默认 `127.0.0.1:3306/admin`。
- Redis：会话、临时登录态、短信验证码、发送锁和权限缓存。
- 短信 provider：运行态默认 `aliyun`，测试态默认 `mock`。
- 审计写入器：支持同步和异步模式，由 `audit.async` 控制。
- 第三方 provider：`internal/library/supplierplatform/provider` 负责平台识别、余额刷新、云发卡下单和查单适配。
- 订单 worker：配置 `open_order.worker_enabled=true` 时启动，`Core.Close()` 会先停止 worker，再释放运行时资源。

## OpenAPI

- OpenAPI JSON：`/api.json`
- Swagger UI：`/swagger/`
- Bearer 鉴权方案和公共响应壳由启动期统一注册。

## 测试形态

契约测试通过 `NewTestApplication()` 启动应用；测试 Core 会自动创建并重置 MySQL `admin_test`，Redis 使用 `miniredis`，短信 sender 使用 mock。
```

- [ ] **Step 4: Verify overview and architecture do not duplicate module map responsibilities**

Run:

```bash
rg -n "协议：|controller：|service：|logic：|路由前缀：|权限码：|权限边界：" docs/overview.md docs/architecture.md
```

Expected: no matches. These detailed mapping labels belong in `docs/module-map.md`.

- [ ] **Step 5: Commit Task 3**

Run:

```bash
git add docs/overview.md docs/architecture.md
git commit -m "docs: 重写概览与架构说明"
```

Expected: commit succeeds with only the two docs modified.

## Task 4: Rewrite Module Map As The Main Reference

**Files:**
- Modify: `docs/module-map.md`
- Reference: `api/*.go`
- Reference: `internal/controller/admin/*.go`
- Reference: `internal/controller/open/*.go`
- Reference: `internal/service/*.go`
- Reference: `internal/logic/admin/*.go`
- Reference: `internal/logic/order/*.go`
- Reference: `internal/bootstrap/application.go`
- Reference: `manifest/sql/*.sql`

- [ ] **Step 1: Generate factual file lists**

Run:

```bash
git ls-files api internal/controller/admin internal/controller/open internal/service internal/logic/admin internal/logic/order manifest/sql | sort
```

Expected: output includes `api/open_order.go`, `api/order.go`, settings split files, product channel split files, `internal/logic/order/*.go`, and `manifest/sql/008_external_order.sql`.

- [ ] **Step 2: Confirm route and permission bindings**

Run:

```bash
rg -n "guard.Require|Group\\(\"/api/open\"|Group\\(\"/api/admin\"|SetOpenApiPath|SetSwaggerPath" internal/bootstrap/application.go
```

Expected: output shows `/api/open`, `/api/admin`, permission codes `admin.list`, `admin.department`, `subject.manage`, `product.brand`, `product.industry`, `product.template`, `product.purchase_limit`, `product.goods`, `supplier.index`, `order.manage`, `admin.action`, `admin.loginlog`, and super-only settings group.

- [ ] **Step 3: Rewrite `docs/module-map.md`**

Use this structure:

```markdown
# 模块职责地图

本文是当前仓库业务域、文件归属、路由前缀和权限边界的主参考。开发流程看 `development.md`，测试策略看 `testing.md`，迁移背景看 `migration.md`。

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
- 主要能力：商品列表、详情、表单选项、新增、编辑、删除、启停、渠道摘要、库存配置、渠道绑定弹窗、单条自动改价。
- 边界：商品主档、渠道绑定和库存配置保持同 package 多文件拆分。

### 第三方对接

- 协议：`api/supplier_platform.go`
- controller：`internal/controller/admin/supplier_platform.go`
- service：`SupplierPlatformService`（`internal/service/supplier_platform.go`）
- logic：`internal/logic/admin/supplier_platform*.go`
- provider：`internal/library/supplierplatform/provider/*`
- 路由前缀：`/api/admin/supplier-platform-types`、`/api/admin/supplier-platforms*`
- 权限：`supplier.index`
- 主要能力：平台类型字典、平台账号分页、详情、增删改、启停、余额刷新、余额日志落库、平台关闭后级联关停商品绑定。
- 边界：`platform_docs/` 保存渠道原始协议，`docs/` 保存本仓库实现说明。

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
```

- [ ] **Step 4: Verify module-map references real files**

Run:

```bash
for path in \
  api/auth.go api/user.go api/group.go api/subject.go api/brand.go api/industry.go \
  api/product_template.go api/purchase_limit.go api/product_goods.go api/product_goods_channel.go \
  api/product_goods_channel_config.go api/supplier_platform.go api/open_order.go api/order.go \
  api/settings.go api/settings_sms.go api/settings_system.go api/log.go \
  internal/controller/open/order.go internal/controller/admin/order.go internal/logic/order/order.go; do \
  test -e "$path" || { echo "missing $path"; exit 1; }
done
```

Expected: no output and exit code 0.

- [ ] **Step 5: Commit Task 4**

Run:

```bash
git add docs/module-map.md
git commit -m "docs: 重写模块职责地图"
```

Expected: commit succeeds with only `docs/module-map.md` modified.

## Task 5: Rewrite Development And Testing How-To Docs

**Files:**
- Modify: `docs/development.md`
- Modify: `docs/testing.md`
- Reference: `go.mod`
- Reference: `.github/workflows/ci.yml`
- Reference: `.golangci.yml`
- Reference: `docker-compose.yml`
- Reference: `manifest/config/config.local.yaml`
- Reference: `manifest/sql/*.sql`
- Reference: `internal/app/schema.go`
- Reference: `test/contract/*.go`
- Reference: `test/integration/*.go`

- [ ] **Step 1: Confirm toolchain, CI and lint facts**

Run:

```bash
sed -n '1,20p' go.mod
sed -n '1,180p' .github/workflows/ci.yml
sed -n '1,80p' .golangci.yml
```

Expected: Go version comes from `go.mod`; CI has test and lint jobs; lint enables `govet`, `staticcheck`, `ineffassign`, `unused`, `typecheck`.

- [ ] **Step 2: Rewrite `docs/development.md`**

Use these sections:

```markdown
# 开发说明

## 本地依赖

- Go 版本以根目录 `go.mod` 为准。
- Docker / Docker Compose。
- MySQL 8.4。
- Redis 7。
- `golangci-lint`，用于本地对齐 CI lint 口径。

## 本地启动

### 1. 启动数据库和缓存

```bash
docker compose up -d mysql redis
```

默认端口：

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6380`

### 2. 使用默认配置

默认配置文件是 `manifest/config/config.local.yaml`。如需显式指定：

```bash
export ADMIN_CONFIG=manifest/config/config.local.yaml
```

配置加载顺序：

1. `ADMIN_CONFIG`
2. `manifest/config/config.local.yaml`

### 3. 启动应用

```bash
go run .
```

本地默认超管：

- 用户名：`admin`
- 手机号：`15881767197`
- 密码：`abc123`

## 常用验证命令

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

更细的测试分层和 CI 口径见 `testing.md`。

## 配置项

当前本地配置主要包含：

- `server.address`
- `database.driver` / `database.dsn`
- `redis.addr` / `redis.password` / `redis.db`
- `auth.jwt_secret`
- `auth.access_token_ttl_minutes`
- `auth.temp_login_ttl_minutes`
- `bootstrap.super_admin_*`
- `sms.provider`
- `audit.async` / `audit.buffer_size`
- `open_order.token`
- `open_order.worker_enabled`
- `open_order.poll_interval_seconds`
- `open_order.submit_scan_interval_seconds`

`open_order.token` 支持 `${OPEN_ORDER_TOKEN:-default}` 形式从环境变量读取。订单 worker 在本地默认开启，测试态默认关闭。

## SQL 与 schema 同步

Docker 首次建库执行 `manifest/sql/*.sql`。应用启动期还会通过 `internal/app/schema.go` 补齐缺失表。

修改 MySQL 表结构、表注释或字段注释时，必须同步：

- `manifest/sql/*.sql`
- `internal/app/schema.go`

推荐静态校验：

```bash
go test ./internal/app -run 'Test(MySQLSchemaIncludesTableAndColumnComments|ManifestMySQLSchemaFilesIncludeTableAndColumnComments)' -count=1 -timeout 60s
```

需要重建本地库验收时：

```bash
docker compose down -v
docker compose up -d mysql redis
go run .
```

再查询空注释：

```bash
docker exec $(docker compose ps -q mysql) mysql -uroot -proot123456 -D admin -e "SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA='admin' AND TABLE_COMMENT = '';"
docker exec $(docker compose ps -q mysql) mysql -uroot -proot123456 -D admin -e "SELECT TABLE_NAME, COLUMN_NAME, COLUMN_COMMENT FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='admin' AND COLUMN_COMMENT = '';"
```

## 初始化 SQL

- `manifest/sql/001_schema.sql`：基础表结构。
- `manifest/sql/002_seed_menu.sql`：菜单与权限种子。
- `manifest/sql/003_seed_admin.sql.tmpl`：超级管理员 SQL 模板。
- `manifest/sql/004_seed_config.sql`：系统参数初始值。
- `manifest/sql/005_supplier_platform.sql`：第三方平台类型、账号和余额日志结构。
- `manifest/sql/006_product_goods_channel_binding.sql`：商品渠道绑定表结构。
- `manifest/sql/007_product_goods_channel_config.sql`：商品库存配置表结构。
- `manifest/sql/008_external_order.sql`：外部订单主表和渠道尝试表结构。

生成超级管理员初始化 SQL：

```bash
export SUPER_ADMIN_PHONE=15881767197
export SUPER_ADMIN_BCRYPT_HASH='$2a$10$exampleexampleexampleexampleexampleexampleexampleexample'
./hack/bootstrap-admin.sh
```

## DAO 生成

```bash
export GF_DAO_LINK='mysql:root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local'
./hack/gen-dao.sh
```

脚本会生成 DAO / DO / Entity / table 元数据。当前仓库主要使用 `internal/dao/tables.go` 作为轻量表入口，重新生成后必须检查是否带入无用生成目录。

## 日常开发约束

- controller 不直接访问 DAO。
- logic 不把 HTTP 请求结构继续下传到更深层。
- 跨模块基础能力优先收口到 `internal/library`。
- 关键流程保留简体中文注释。
- 文档描述以当前真实代码行为为准。
```

- [ ] **Step 3: Rewrite `docs/testing.md`**

Use these sections:

```markdown
# 测试说明

## 推荐默认命令

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

`go test ./...` 会覆盖包内测试、契约测试和默认可运行的集成测试。部分 live 或真实配置测试需要显式环境变量。

## 契约测试

目录：`test/contract/`

用途：

- 约束扁平 `api/*.go` 协议目录。
- 防止回退到历史嵌套协议包。
- 验证 OpenAPI / Swagger 暴露。
- 验证统一响应 `code / message / data`。
- 覆盖登录、短信二验、权限、主要后台业务、开放订单和后台订单主流程。

运行：

```bash
go test ./test/contract -count=1 -timeout 60s
```

契约测试通过 `NewTestApplication()` 启动测试态应用，底层使用 MySQL 测试库 `admin_test`、`miniredis` 和 mock 短信 sender。

## 集成测试

目录：`test/integration/`

默认可运行的集成测试使用 `httptest.Server` 模拟第三方平台或云发卡，不依赖真实外部账号。

运行：

```bash
go test ./test/integration -count=1 -timeout 60s
```

订单 worker 聚焦回归：

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

runtime smoke test 需要显式开启：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

第三方平台 live 验证需要真实账号并显式开启：

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_NAME='示例平台'
export SUPPLIER_LIVE_DOMAIN=example.com
export SUPPLIER_LIVE_BACKUP_DOMAIN=example.com
export SUPPLIER_LIVE_TOKEN_ID=1008612345
export SUPPLIER_LIVE_SECRET_KEY=secret
go test ./test/integration -run TestSupplierPlatformRefresh_LiveProviderBalance -count=1 -v
```

## 包内测试

包内测试集中在基础能力和业务逻辑边界，例如：

- `internal/app`
- `internal/library/region`
- `internal/library/sms`
- `internal/library/supplierplatform/provider`
- `internal/logic/admin`
- `internal/logic/order`

schema 和注释约束：

```bash
go test ./internal/app -run 'Test(MySQLSchemaIncludesTableAndColumnComments|ManifestMySQLSchemaFilesIncludeTableAndColumnComments)' -count=1 -timeout 60s
```

云发卡 provider 聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProvider -count=1 -timeout 60s
```

## CI 与 lint

CI workflow 位于 `.github/workflows/ci.yml`：

- `test` job 启动 MySQL 8.4 service，执行 `go test ./... -count=1 -timeout 60s` 和 `go build ./...`。
- `lint` job 执行 `golangci-lint`，参数包含 `--timeout=5m --new-from-rev=origin/main`。

`.golangci.yml` 当前启用：

- `govet`
- `staticcheck`
- `ineffassign`
- `unused`
- `typecheck`

本地建议在提交前执行：

```bash
golangci-lint run --timeout=5m
```

## 当前认知边界

文档只描述已经存在的测试覆盖，不把建议补充项写成已经覆盖。

当前尚未形成完整外部依赖闭环测试集的内容包括：

- 真实短信发送链路回归。
- 跨重启行为验证。
- 多平台批量 live 回归。
- 更细的 MySQL / Redis 行为断言。
```

- [ ] **Step 4: Verify development/testing facts align with repo**

Run:

```bash
rg -n "go test ./\\.\\.\\.|go build ./\\.\\.\\.|golangci-lint run --timeout=5m|admin_test|MYJOB_RUN_INTEGRATION|MYJOB_RUN_SUPPLIER_LIVE|008_external_order" docs/development.md docs/testing.md .github/workflows/ci.yml manifest/sql test/integration
```

Expected: output confirms the documented commands, env vars and SQL file names exist.

- [ ] **Step 5: Commit Task 5**

Run:

```bash
git add docs/development.md docs/testing.md
git commit -m "docs: 重写开发与测试说明"
```

Expected: commit succeeds with only the two docs modified.

## Task 6: Compress Migration And Test Directory READMEs

**Files:**
- Modify: `docs/migration.md`
- Modify: `test/contract/README.md`
- Modify: `test/integration/README.md`
- Reference: `docs/module-map.md`
- Reference: `docs/testing.md`

- [ ] **Step 1: Inspect duplicate facts to remove**

Run:

```bash
rg -n "api/open_order.go|api/order.go|完整|当前对外协议|当前目录|覆盖范围|运行方式|runtime smoke|订单 worker" docs/migration.md test/contract/README.md test/integration/README.md docs/module-map.md docs/testing.md
```

Expected: output shows migration and test READMEs currently duplicate facts that should primarily live in `module-map.md` or `testing.md`.

- [ ] **Step 2: Rewrite `docs/migration.md`**

Use this content:

```markdown
# 迁移说明

## 背景

这个仓库原本承载的是一套更分散、更手写的后台实现。当前迁移工作的目标，是把后端能力收口到仓库根目录，形成统一的 GoFrame 主应用结构，并固定清晰的目录职责、运行时依赖和测试入口。

## 已完成方向

### 主入口收口

当前唯一主入口位于仓库根：

- `main.go`
- `internal/cmd`
- `internal/bootstrap`

历史 `admin/` 子工程不再作为运行入口。

### 协议目录拍平

对外协议统一收口到根目录 `api/*.go`。协议目录保持扁平结构，不再回到历史嵌套协议包路径。

当前协议文件清单和职责以 `module-map.md` 为准。

### 运行时能力收口

运行时核心能力收口在：

- `internal/app`
- `internal/library/auth`
- `internal/library/sms`
- `internal/library/audit`
- `internal/library/region`
- `internal/library/supplierplatform/provider`

### 业务域按职责拆分

后台业务实现按同 package 多文件拆分，典型模式包括：

- `brand*.go`
- `industry*.go`
- `user*.go`
- `product_template*.go`
- `purchase_limit*.go`
- `product_goods*.go`
- `product_goods_channel*.go`
- `supplier_platform*.go`

订单履约单独位于 `internal/logic/order`，不放进商品、第三方平台或通用 helper 文件。

## 当前保留的兼容面

- 后台接口前缀仍为 `/api/admin/*`。
- 开放订单接口前缀为 `/api/open/*`。
- 响应壳仍为 `code / message / data`。
- MySQL、Redis、菜单权限、短信配置、系统参数配置和日志表等业务语义继续沿用。

## 当前文档分工

- 当前功能、路由、权限和文件归属：`module-map.md`。
- 启动链路和分层职责：`architecture.md`。
- 本地开发命令：`development.md`。
- 测试分层和 CI/lint：`testing.md`。

## 后续收尾方向

- 持续清理与当前实现无关的历史描述。
- 持续收紧文档和真实目录、路由、测试、CI 的一致性。
- 在新增重要业务流时同步更新对应专题文档。
```

- [ ] **Step 3: Rewrite `test/contract/README.md`**

Use this content:

```markdown
# Contract Tests

契约测试用于约束接口兼容、协议布局和核心业务流。完整测试策略见 `../../docs/testing.md`。

## 运行

```bash
go test ./test/contract -count=1 -timeout 60s
```

## 当前重点

- 扁平 `api/*.go` 协议目录。
- OpenAPI `/api.json` 和 Swagger `/swagger/`。
- 统一响应 `code / message / data`。
- 登录、短信二验、权限和核心后台业务流。
- 商品、第三方对接、开放订单和后台订单主流程。

契约测试会启动测试态应用，使用 MySQL 测试库 `admin_test`、`miniredis` 和 mock 短信 sender。
```

- [ ] **Step 4: Rewrite `test/integration/README.md`**

Use this content:

```markdown
# Integration Tests

集成测试用于验证需要多组件联动或模拟外部服务的行为。完整测试策略见 `../../docs/testing.md`。

## 默认运行

```bash
go test ./test/integration -count=1 -timeout 60s
```

默认测试使用 `httptest.Server` 模拟第三方平台或云发卡，不依赖真实外部账号。

## 聚焦运行

订单 worker：

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

runtime smoke test：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

第三方平台 live 验证：

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_NAME='示例平台'
export SUPPLIER_LIVE_DOMAIN=example.com
export SUPPLIER_LIVE_BACKUP_DOMAIN=example.com
export SUPPLIER_LIVE_TOKEN_ID=1008612345
export SUPPLIER_LIVE_SECRET_KEY=secret
go test ./test/integration -run TestSupplierPlatformRefresh_LiveProviderBalance -count=1 -v
```
```

- [ ] **Step 5: Verify migration no longer owns current file lists**

Run:

```bash
rg -n "api/auth.go|api/open_order.go|api/order.go|api/settings_sms.go|api/settings_system.go|product_goods_channel_config.go" docs/migration.md
```

Expected: no matches. Current protocol file facts should live in `docs/module-map.md`.

- [ ] **Step 6: Commit Task 6**

Run:

```bash
git add docs/migration.md test/contract/README.md test/integration/README.md
git commit -m "docs: 收敛迁移与测试目录说明"
```

Expected: commit succeeds with only these three docs modified.

## Task 7: Cross-Document Consistency Pass

**Files:**
- Modify if needed: `README.md`
- Modify if needed: `docs/README.md`
- Modify if needed: `docs/overview.md`
- Modify if needed: `docs/architecture.md`
- Modify if needed: `docs/module-map.md`
- Modify if needed: `docs/development.md`
- Modify if needed: `docs/testing.md`
- Modify if needed: `docs/migration.md`
- Modify if needed: `docs/superpowers/README.md`
- Modify if needed: `test/contract/README.md`
- Modify if needed: `test/integration/README.md`

- [ ] **Step 1: Search for stale or missing paths**

Run:

```bash
rg -n "docs/superpowers/README.md|docs/README.md|api/settings.go|api/common.go|admin/ 子工程|internal/model/table|T[O]DO|T[B]D|尚未实现|未来接口" README.md docs test/contract/README.md test/integration/README.md
```

Expected: `docs/superpowers/README.md` and `docs/README.md` references are valid; no placeholder markers; any `尚未实现` or `未来接口` mention only appears in non-goal or boundary language.

- [ ] **Step 2: Verify every Markdown-referenced local doc exists**

Run:

```bash
for path in README.md docs/README.md docs/overview.md docs/architecture.md docs/module-map.md docs/development.md docs/testing.md docs/migration.md docs/superpowers/README.md platform_docs/README.md test/contract/README.md test/integration/README.md; do
  test -e "$path" || { echo "missing $path"; exit 1; }
done
```

Expected: no output and exit code 0.

- [ ] **Step 3: Verify API layout contract**

Run:

```bash
go test ./test/contract -run TestAPILayout -count=1 -timeout 60s
```

Expected: PASS.

- [ ] **Step 4: Verify CI workflow contract**

Run:

```bash
go test ./test/contract -run TestCIWorkflow -count=1 -timeout 60s
```

Expected: PASS.

- [ ] **Step 5: Commit consistency fixes if any**

If Step 1 or Step 2 required edits, run:

```bash
git add README.md docs test/contract/README.md test/integration/README.md
git commit -m "docs: 校准文档交叉引用"
```

Expected: commit succeeds only if there were consistency edits. If no files changed, skip this commit.

## Task 8: Final Verification

**Files:**
- No planned edits.

- [ ] **Step 1: Run full test suite**

Run:

```bash
go test ./... -count=1 -timeout 60s
```

Expected: PASS. If MySQL is unavailable, record the failure and exact reason.

- [ ] **Step 2: Run build**

Run:

```bash
go build ./...
```

Expected: PASS.

- [ ] **Step 3: Run lint**

Run:

```bash
golangci-lint run --timeout=5m
```

Expected: PASS. If `golangci-lint` is not installed, record that explicitly.

- [ ] **Step 4: Inspect final diff**

Run:

```bash
git status --short
git log --oneline -8
```

Expected: working tree is clean after all task commits; recent commits show the design, plan, and documentation rewrite commits.

- [ ] **Step 5: Prepare final summary**

Include:

- Changed files.
- Why each document belongs in the new information architecture.
- What was split or compressed.
- Which old files were kept and why.
- Key documentation constraints added.
- Validation commands run.
- Any validations not run and why.

## Self-Review Notes

- Spec coverage: all spec targets are represented by tasks: portal and superpowers entry in Task 1, README in Task 2, overview/architecture in Task 3, module-map in Task 4, development/testing in Task 5, migration/test README compression in Task 6, consistency and validation in Tasks 7-8.
- Placeholder scan: no placeholder markers or unspecified implementation steps are intentionally present.
- Scope check: this remains one documentation-system plan; it does not modify Go code, SQL, config, CI workflow, or tests.
