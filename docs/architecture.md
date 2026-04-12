# 架构说明

## 启动链路

```text
main.go
  -> internal/cmd/root.go
    -> bootstrap.NewApplicationFromEnv()
      -> model/config.Load()
      -> kernel.NewCoreFromConfig()
      -> bootstrap.assemble()
      -> ghttp.Server.Run()
```

其中：

- `main.go` 只负责交给 GoFrame 命令入口
- `internal/cmd` 只负责启动命令，不承载业务逻辑
- `internal/bootstrap` 负责把 controller、logic、middleware、kernel 组装起来
- `internal/kernel` 负责运行时资源、启动期数据准备和跨模块共用的数据访问封装

## 分层职责

### `api/*`

放请求和响应协议结构，保证 HTTP 协议类型不会直接渗透到更深层。

### `internal/controller/admin`

只处理 HTTP 协议适配：

- 解析参数
- 调用 `service` 接口
- 返回统一 JSON 包裹

控制器不直接访问 DAO，也不直接拼 SQL。

### `internal/service`

只放接口定义，作为跨模块依赖边界。

### `internal/logic/admin`

负责业务编排：

- 参数合法性判断
- 业务流程控制
- 调用 `kernel` / `library`
- 组织返回数据

### `internal/kernel`

这是本次迁移里的“应用内核层”，主要承担两类职责：

1. 运行时资源管理：数据库、Redis、短信发送器、审计写入器、区域解析器
2. 迁移期的数据访问与通用业务辅助：会话签发、权限读取、种子初始化、公共查询

后续如果 DAO 自动生成进一步稳定，可以把 `kernel` 中更细的查询逐步继续下沉。

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
- `response`：统一响应包裹

## 请求流转

```text
HTTP Request
  -> middleware.AuthGuard
  -> controller/admin
  -> service interface
  -> logic/admin
  -> kernel / dao / library
  -> MySQL / Redis / provider
```

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
