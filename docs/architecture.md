# 架构说明

## 启动链路

当前实际启动顺序如下：

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
      -> ghttp.Server.Run()
```

补充说明：

- `main.go` 只负责进入 GoFrame 命令入口
- `internal/cmd` 负责启动命令，不承载业务逻辑
- `internal/app` 负责配置解析、MySQL / Redis 初始化、种子初始化和运行时资源装配
- `internal/bootstrap` 负责路由、中间件、controller、service、logic 的组合

## 分层职责

### `api`

放请求/响应协议结构，使用 `g.Meta` 声明路径、方法、标签和 OpenAPI 信息。
当前已经统一使用扁平 `api/*.go` 结构。

### `internal/controller/admin`

负责 HTTP 协议适配：

- 接收 GoFrame 标准 `Req/Res + error`
- 从上下文中读取认证用户
- 调用 `service` 接口
- 不直接拼接统一响应壳

### `internal/service`

只定义模块接口，是 controller 依赖的抽象边界。

### `internal/logic/admin`

负责业务编排和规则控制，包括：

- 参数校验
- 状态校验
- 权限相关业务判断
- 数据访问编排
- 审计写入

### `internal/app`

`internal/app` 是运行时核心层，当前主要承担：

1. 配置、数据库、Redis、短信 sender、区域解析器、审计写入器的初始化
2. 启动期 schema / seed / bootstrap 逻辑
3. 会话签发、权限读取、短信配置缓存、公共查询和辅助工具

### `internal/library`

封装跨模块基础能力：

- `auth`：JWT、session、Redis key 规则
- `sms`：sender 抽象、mock sender、阿里云 sender
- `audit`：同步 / 异步写入器
- `region`：IP 归属地解析与脱敏工具

## 请求处理链路

```text
HTTP Request
  -> middleware.Response
  -> middleware.AuthGuard (按路由分组决定是否启用)
  -> controller/admin
  -> service interface
  -> logic/admin
  -> app / dao / library
  -> MySQL / Redis / provider
```

补充说明：

- `middleware.Response` 统一在出口包装 `code / message / data`
- `middleware.AuthGuard` 负责解析 Bearer token、加载用户并校验权限
- 若路由声明 `superOnly=true`，则只有 `group_id = 0` 的超级管理员可访问

## 认证与短信流程

### 账号密码登录

```text
POST /api/admin/auth/login
  -> 校验用户名和密码
  -> 校验用户状态
  -> 判断是否需要短信二验
       -> 不需要：直接签发 token
       -> 需要：写入临时登录态，返回 login_token
```

当前短信二验触发条件：

- 用户没有 `last_login_ip`，返回 `reason=first_login`
- 本次 IP 与 `last_login_ip` 不同，返回 `reason=ip_changed`

### 发送验证码

```text
POST /api/admin/auth/sms/send
  -> 校验 login_token
  -> 读取短信配置
  -> 写入验证码缓存
  -> 写入发送锁
  -> 调用 sender 发送短信
       -> 失败：回滚验证码缓存和发送锁
       -> 成功：等待用户提交验证码
```

### 验证验证码

```text
POST /api/admin/auth/sms/verify
  -> 校验 login_token + sms_code
  -> 校验验证码缓存
  -> 错误时累计 attempts
  -> 超过 5 次后清理临时登录态和验证码
  -> 成功后签发正式 token
```

## 权限模型

当前权限模型分两层：

1. 是否已登录
2. 是否满足权限码或 super-only 要求

具体规则：

- 超级管理员：`group_id = 0`
- 普通用户：通过用户组读取菜单权限码
- 菜单树和用户组授权只暴露 `super_only = 0` 的菜单
- 短信配置接口和系统参数配置接口都独立走 super-only 保护，不依赖普通权限码
- 系统参数配置当前内置 `finance`、`integration` 两组，支持单组读取和多组批量保存

## 运行时依赖

### MySQL

- 默认真实运行数据库是 MySQL
- 本地开发默认连接 `127.0.0.1:3307/admin`
- 初始化 SQL 位于 `manifest/sql/`

### Redis

- 用于会话、临时登录态、短信验证码、短信发送锁和权限缓存
- 本地开发默认连接 `127.0.0.1:6380`

### 短信 provider

- 运行态默认 `aliyun`
- 测试态默认 `mock`
- provider 选择来自配置文件中的 `sms.provider`

### 审计写入

- 写入器支持同步和异步模式
- 是否异步由 `audit.async` 控制
- 本地测试初始化时默认关闭异步，减少不确定性

## OpenAPI 与文档

- OpenAPI JSON 默认路径：`/api.json`
- Swagger UI 默认路径：`/swagger/`
- Bearer 鉴权方案和公共响应壳由启动期统一注册

## 测试时的运行形态

- 契约测试通过 `NewTestApplication()` 启动应用
- `NewTestCore()` 会使用临时 SQLite 文件 + `miniredis`
- 契约测试默认不会调用真实阿里云短信发送
- 集成测试才会按真实配置尝试启动应用
