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
- 订单履约：开放下单/查单、充值账号风控拦截、云发卡异步提交与轮询、后台订单记录列表和基础统计。

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
go test ./... -count=1 -timeout 120s
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
- `docs/superpowers/README.md`：阶段性规格和实施计划目录约定。
- `platform_docs/README.md`：第三方渠道原始接口资料总览。
