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

## 新增后台业务域示例流程

以下流程适用于新增一个相对独立的后台管理域，例如“活动配置”“渠道分组”这类有独立列表、表单和权限点的能力。若只是给现有业务域加字段或动作，优先在现有域文件内按职责补充，不要新建无意义目录。

### 1. 先确认边界

- 在 `docs/module-map.md` 找是否已有相近业务域。
- 如果只是商品、订单、第三方平台、设置等既有域的子能力，优先沿用对应 package 和文件前缀。
- 只有当它有稳定的协议、权限、数据和业务编排边界时，才按新业务域处理。

### 2. 补协议和接口

- 在 `api/` 新增或扩展扁平协议文件，例如 `api/example.go`。
- 只放请求、响应、列表项、枚举和路由协议元信息。
- 在 `internal/service/` 新增接口文件，例如 `example.go`，只定义 service 接口。
- 导出 Req / Res / Item / Enum 和 service 接口都要写 Go 风格注释。

### 3. 补 controller 和 logic

- 在 `internal/controller/admin/` 新增 controller 文件，例如 `example.go`。
- controller 只做参数接收、调用 service、返回统一响应，不写业务判断。
- 在 `internal/logic/admin/` 新增同 package 多文件实现。推荐起步：
  - `example.go`：只放 `ExampleLogic` 声明和构造。
  - `example_query.go`：列表、详情、选项读取。
  - `example_write.go`：新增、编辑、删除、启停。
  - `example_validate.go`：参数和业务规则校验。
  - `example_mapper.go`：实体到响应结构映射。
- 若当前只需要一两个简单方法，可以少建文件；后续职责变多时再按职责拆开。

### 4. 挂路由、权限和种子

- 在 `internal/bootstrap/application.go` 装配 controller，并按权限边界挂到 `/api/admin`。
- 如果新增权限点，同步 `manifest/sql/002_seed_menu.sql` 和启动期 seed 逻辑。
- 如果新增表，同步 `manifest/sql/*.sql` 与 `internal/app/schema.go`。
- 配置项变化同步 `manifest/config/config.local.yaml`、`internal/model/config` 和本文档配置项列表。

### 5. 补测试和文档

- 协议、路由、权限或文件布局变化，优先补 `test/contract`。
- 业务规则、边界条件或事务分支变化，优先补包内单测或 `test/integration`。
- 同步 `docs/module-map.md` 的业务域映射、核心业务流速览和路由权限摘要。
- 同步 `docs/testing.md` 中新增的聚焦验证命令。

### 6. 提交前自检

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

若 `golangci-lint` 本机未安装，最终说明要写明未跑原因；CI 仍会按 `.github/workflows/ci.yml` 执行 lint。

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

## 供应商商品详情同步能力

供应商商品详情同步通过 `internal/library/supplierplatform/provider` 暴露 `ProductInfoProvider`。新增供应商适配器时先补 `BuildProductInfoRequest` 和 `ParseProductInfoResponse` 单测，再注册到 `LookupProductInfo`。业务侧只消费标准化的 `ProductInfoResult`，不得在商品逻辑层直接解析第三方原始响应。

商品渠道同步入口在 `internal/logic/admin/product_goods_channel_sync.go`。同步进价时必须复用 `computeChannelCostSnapshot`，保证商品税态和渠道税态的加税、扣税规则与人工维护绑定一致；自动改价利润字段只保留用户配置，不在同步流程中重写。

### 供应商商品推送订阅

商品变动推送 provider 需要实现 `ProductSubscriptionProvider` 和 `ProductChangePushProvider`。回调 URL 统一使用 `/api/open/supplier-platforms/{providerCode}/{platformAccountId}/product-change-callback`，通过平台账号 ID 找密钥验签。新增渠道绑定后的订阅失败只能写入 `supplier_product_subscription`，不得阻断本地绑定保存。

## 订单金额快照

订单提交会按实际选中的渠道绑定规则写入 `unit_price / order_amount / cost_amount / profit_amount`，并在 `external_order_attempt` 保存渠道主体快照。历史订单如需再次修复，应先按明确订单号或历史时间窗口预览影响范围，再在维护窗口执行一次性 SQL；不要对实时处理中订单做批量回算。

`external_order_attempt` 的渠道主体快照列有启动兜底补列逻辑，并设置了较短的 MySQL `lock_wait_timeout`。生产大表仍建议先在维护窗口执行显式 DDL，再启动新版本，避免启动期等待元数据锁。

## 日常开发约束

- controller 不直接访问 DAO。
- logic 不把 HTTP 请求结构继续下传到更深层。
- 跨模块基础能力优先收口到 `internal/library`。
- 关键流程保留简体中文注释。
- 文档描述以当前真实代码行为为准。
