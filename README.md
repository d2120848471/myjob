# MyJob Admin Backend

MyJob Admin Backend 是一个以仓库根为主应用入口的 GoFrame 单体后台项目，
目标是把原来那套手写后端迁到仓库根主应用中，整理成更适合企业协作的
标准目录、职责边界和工程化交付形态。

## 项目定位

- 运行入口在仓库根：`main.go` + `internal/cmd`
- HTTP 接口继续兼容既有后台路径，默认保持 `code / msg / data` 响应包裹
- 内部职责按 `api -> controller -> service -> logic -> kernel/dao/model -> library`
  分层，避免再回到“大一统 app 层”
- 当前仓库只保留一套主代码，运行入口和维护入口都在仓库根

## 快速开始

### 1. 准备依赖

```bash
docker compose up -d mysql redis
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
```

### 2. 检查配置

默认启动会优先读取下面的配置文件：

1. `ADMIN_CONFIG` 环境变量指定的路径
2. `manifest/config/config.local.yaml`

如需切换环境：

```bash
export ADMIN_CONFIG=manifest/config/config.local.yaml
```

### 3. 启动服务

```bash
go run .
```

默认监听地址来自 `manifest/config/config.local.yaml`，当前配置为 `:8080`。

### 4. 运行验证

```bash
go test ./... -count=1 -timeout 60s
go build ./...
```

## 目录总览

```text
.
├── README.md
├── docs/
├── api/
│   ├── auth
│   ├── user
│   ├── group
│   ├── subject
│   ├── config
│   └── log
├── internal/
│   ├── bootstrap
│   ├── cmd
│   ├── consts
│   ├── controller/admin
│   ├── dao
│   ├── kernel
│   ├── library
│   ├── logic/admin
│   ├── middleware
│   ├── model
│   └── service
├── manifest/
│   ├── config
│   └── sql
├── hack/
├── resource/
├── test/
│   ├── contract
│   ├── integration
│   └── fixture
```

## 目录职责说明

### 根目录文件

- `README.md`：项目总入口说明，包含启动方式、目录职责、常用命令和协作约定
- `main.go`：进程启动入口，只负责把执行权交给 `internal/cmd`
- `go.mod` / `go.sum`：Go 模块依赖定义和锁定文件
- `docker-compose.yml`：本地真实开发依赖的 MySQL / Redis 容器编排

### `api/`

`api/` 只放“对外协议结构”，也就是 controller 解析请求和组织响应时使用的
结构体定义，不写业务逻辑，不直接访问数据库。

- `api/auth`：登录、短信二验、退出登录、`me` 相关协议
- `api/user`：员工列表、新增、编辑、删除、恢复、状态切换、业务账号设置协议
- `api/group`：用户组管理、菜单授权、菜单树相关协议
- `api/subject`：主体列表、新增、编辑相关协议
- `api/config`：短信配置读写、脱敏展示相关协议
- `api/log`：操作日志、登录日志查询协议

### `internal/`

`internal/` 是真正的后端实现区，所有业务逻辑、运行时资源和内部能力都在这里。

- `internal/bootstrap`：应用装配层，负责加载配置、创建 `Core`、注册路由、组装 controller / logic / middleware
- `internal/cmd`：GoFrame 命令入口，负责进程启动，不承载业务逻辑
- `internal/consts`：状态值、权限相关固定值等通用常量
- `internal/controller/admin`：HTTP 协议层，负责参数解析、调用 service、返回统一 JSON 包裹
- `internal/service`：模块接口边界，controller 依赖这里暴露的抽象，不直接碰更深层实现
- `internal/logic/admin`：业务编排层，负责把权限、规则、流程、数据访问串起来
- `internal/kernel`：运行时核心层，管理 DB / Redis / sender / audit / region 等基础资源，并承接当前项目里的通用业务辅助能力
- `internal/dao`：GoFrame DAO 入口和表级查询模型装配
- `internal/model`：内部数据模型集合
  - `internal/model/config`：配置结构体定义
  - `internal/model/do`：持久化写入 / 条件对象
  - `internal/model/entity`：数据库实体映射
  - `internal/model/dto`：内部传输对象、分页对象等
  - `internal/model/runtime`：运行时上下文、统一响应、错误对象等
- `internal/middleware`：鉴权和授权中间件
- `internal/library`：跨模块基础能力库
  - `internal/library/auth`：JWT、Session、Redis key 规则
  - `internal/library/sms`：短信 provider 抽象和阿里云实现
  - `internal/library/audit`：操作日志异步 / 同步写入器
  - `internal/library/region`：IP 归属地解析与手机号/密钥脱敏
  - `internal/library/response`：统一 `code / msg / data` 响应输出

### `manifest/`

- `manifest/config`：运行时配置文件目录，当前只保留真实开发配置 `config.local.yaml`
- `manifest/sql`：初始化 SQL 资源，包含建表、菜单、管理员和系统配置初始化内容

### `hack/`

- `hack/bootstrap-admin.sh`：根据环境变量生成超级管理员初始化 SQL
- `hack/gen-dao.sh`：按当前数据库连接生成 DAO / DO / Entity / table 元数据

### `resource/`

- `resource/ipdb/ip_region.xdb`：IP 归属地解析库文件，供登录日志和操作日志写入地区信息时使用

### `test/`

`test/` 用来放需要集中管理的测试，而不是散落在业务目录里的包内测试。

- `test/contract`：接口契约和核心业务流测试，验证登录、短信、用户、用户组、主体、日志等主流程
- `test/integration`：真实依赖联动测试，验证 MySQL / Redis / 配置加载 / 运行时链路
- `test/fixture`：测试样例和共享说明文件

### `docs/`

`docs/` 用来承接 README 之外的详细说明，方便新同事在不同场景下快速定位。

- `docs/overview.md`：项目背景、目标和兼容约束
- `docs/architecture.md`：启动链路、分层说明、基础设施说明
- `docs/module-map.md`：业务模块与目录职责地图
- `docs/development.md`：开发配置、脚本、初始化和日常命令
- `docs/testing.md`：测试分层、运行方式和测试放置约定
- `docs/migration.md`：从旧实现迁移到当前根应用结构的过程说明

## 常用脚本与资源

- `hack/bootstrap-admin.sh`：从模板生成超级管理员初始化 SQL
- `hack/gen-dao.sh`：按当前数据库连接生成 GoFrame DAO / DO / Entity 及表元数据文件
- `manifest/sql/`：建表、菜单、管理员、配置初始化 SQL
- `resource/ipdb/ip_region.xdb`：IP 归属地解析库

## 文档索引

- `docs/overview.md`：项目简介与目标
- `docs/architecture.md`：启动链路、分层职责、关键基础能力
- `docs/module-map.md`：模块职责地图
- `docs/development.md`：开发约束、配置、脚本和日常命令
- `docs/testing.md`：测试分层和执行方式
- `docs/migration.md`：从旧实现迁到新根应用的迁移说明

## 迁移状态

- 当前仓库根已经是唯一主运行入口
- 旧 `admin/` 历史代码已从仓库删除，后续只维护新代码
