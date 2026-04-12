# 架构说明

## 启动链路

```text
main.go
  -> internal/cmd/root.go
    -> bootstrap.NewApplicationFromEnv()
      -> model/config.LoadFromGoFrame()
      -> app.NewCoreFromEnv()
      -> bootstrap.assemble()
      -> ghttp.Server.Run()
```

其中：

- `main.go` 只负责交给 GoFrame 命令入口
- `internal/cmd` 只负责启动命令，不承载业务逻辑
- `internal/bootstrap` 负责把 controller、logic、middleware、app 组装起来
- `internal/app` 负责运行时资源、启动期数据准备和跨模块共用的数据访问封装

## 分层职责

### `api/admin/v1`

放版本化请求和响应协议结构，并通过 `g.Meta` 同时声明路由、方法、标签和文档信息。

### `internal/controller/admin`

只处理 HTTP 协议适配：

- 承接 GoFrame 标准 `Req/Res + error`
- 调用 `service` 接口
- 不直接写响应，由中间件统一输出 JSON 包裹

控制器不直接访问 DAO，也不直接拼 SQL。

### `internal/service`

只放接口定义，作为跨模块依赖边界。

### `internal/logic/admin`

负责业务编排：

- 参数合法性判断
- 业务流程控制
- 调用 `app` / `library`
- 组织返回数据

### `internal/app`

这是当前 GoFrame 官方化后的运行时核心层，主要承担两类职责：

1. 运行时资源管理：GoFrame 配置、数据库、Redis、短信发送器、审计写入器、区域解析器
2. 数据访问与通用业务辅助：会话签发、权限读取、种子初始化、公共查询

后续如果 DAO 自动生成进一步稳定，可以把 `app` 中更细的查询逐步继续下沉。

### `internal/dao`、`internal/model/do`、`internal/model/entity`

承担 GoFrame 数据访问基础目录：

- `dao`：表级入口和查询模型装配
- `model/do`：写入 / 条件对象
- `model/entity`：数据库实体映射

### `internal/library`

放跨业务模块的基础能力：

- `auth`：JWT、会话 Redis key 规则
- `sms`：短信供应商抽象和实现
- `audit`：审计日志写入器
- `region`：IP 归属地解析

## 请求流转

```text
HTTP Request
  -> middleware.AuthGuard
  -> controller/admin
  -> service interface
  -> logic/admin
  -> app / dao / library
  -> MySQL / Redis / provider
```

统一响应由 `middleware.Response` 在出口生成，格式固定为 `code / message / data`。

## 基础设施说明

### 数据库

- 默认运行时数据库是 MySQL
- 测试里使用 SQLite 临时文件做快速契约验证
- 初始化 SQL 放在 `manifest/sql/`

### 缓存与会话

- Redis 负责登录会话、临时登录态、短信验证码和权限缓存
- `internal/library/auth` 统一维护相关 key 规则

### 短信

- 运行态默认使用 `aliyun` provider
- `mock` 只保留给测试场景和 `NewTestCore` 这类测试初始化逻辑
- provider 选择由 `manifest/config/config.local.yaml` 或 `ADMIN_CONFIG` 指向的真实配置控制

### 审计

- 操作日志写入支持同步 / 异步两种模式
- 默认是否异步由 `audit.async` 控制

### OpenAPI

- 服务默认暴露 `/api.json` 与 `/swagger/`
- OpenAPI 公共响应壳与 Bearer 鉴权方案在启动时统一注册
