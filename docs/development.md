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
