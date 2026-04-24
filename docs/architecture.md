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
