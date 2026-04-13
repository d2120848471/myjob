# MyJob Admin Backend

MyJob Admin Backend 是当前仓库根目录下运行的 GoFrame 单体后台项目。
当前代码已经不再依赖历史 `admin/` 子工程；仓库根就是唯一需要维护、启动和发布的后端入口。

## 当前状态

- 主入口在仓库根：`main.go` -> `internal/cmd` -> `internal/bootstrap`
- HTTP 接口统一挂在 `/api/admin`
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

### 2. 准备运行时环境变量

```bash
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
```

> 如果配置文件里没有写死超级管理员手机号和密码，运行时必须提供这两个环境变量；否则启动期引导会失败。

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

### 5. 运行验证

```bash
go test ./... -count=1 -timeout 60s
go build ./...
```

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
│   ├── settings.go
│   ├── subject.go
│   └── user.go
├── docs/
├── hack/
├── internal/
│   ├── app
│   ├── bootstrap
│   ├── cmd
│   ├── consts
│   ├── controller/admin
│   ├── dao
│   ├── library
│   ├── logic/admin
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
当前协议目录已经拍平成仓库根下的 `api/*.go`，不再使用历史的 历史嵌套协议包路径。

### `internal/bootstrap`

负责把配置、运行时 `Core`、控制器、服务、路由和中间件组装成可运行应用。

### `internal/controller/admin`

负责 HTTP 协议适配，方法签名统一是 GoFrame 标准 `Req/Res + error`。

### `internal/service`

定义 controller 依赖的服务接口，是模块边界层。

### `internal/logic/admin`

负责业务编排和规则控制，调用 `app`、`library` 与数据库访问能力。

### `internal/app`

负责运行时核心能力，包括：

- 配置加载
- MySQL / Redis 初始化
- 会话签发与校验
- 菜单与种子初始化
- 短信配置读取
- IP 归属地解析
- 审计写入辅助

### `internal/library`

跨模块基础能力库：

- `auth`：JWT、Redis key、会话存储
- `sms`：短信 sender 抽象、mock sender、阿里云 sender
- `audit`：审计日志写入器
- `region`：IP 归属地解析与脱敏工具

### `manifest/config`

当前只保留本地真实开发配置 `config.local.yaml`。

### `manifest/sql`

初始化 SQL 资源目录，包含 schema、菜单、超级管理员模板和系统配置种子。

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
- 员工、用户组、主体、短信配置、日志等核心流程

### `test/integration`

当前只有一个显式环境开关控制的 runtime smoke test。
它验证应用能按真实配置启动，并完成一次后台登录请求；不是完整的外部依赖闭环回归集。

## 文档索引

- `docs/overview.md`：项目定位与当前功能状态
- `docs/architecture.md`：启动链路、请求流、认证流、权限与运行时结构
- `docs/module-map.md`：模块、接口前缀、权限与目录地图
- `docs/development.md`：开发依赖、配置、脚本和日常命令
- `docs/testing.md`：测试分层、当前覆盖范围与执行命令
- `docs/migration.md`：从历史后台迁到当前根应用结构的迁移背景
