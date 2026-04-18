# 开发说明

## 本地依赖

- Go 1.26+
- Docker / Docker Compose
- MySQL 8.4
- Redis 7

## 本地启动步骤

### 1. 启动数据库与缓存

```bash
docker compose up -d mysql redis
```

默认映射端口：

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6380`

### 2. 确认本地默认超管

- 用户名：`admin`
- 手机号：`15881767197`
- 密码：`abc123`

说明：

- 启动期会使用 `manifest/config/config.local.yaml` 中的固定超管凭证
- 本地默认链路不再要求每次手工设置 `SUPER_ADMIN_PHONE` / `SUPER_ADMIN_PASSWORD`
- 如果你有自定义配置文件，仍然可以显式写自己的 `bootstrap.super_admin_*`

### 3. 选择配置文件

配置加载顺序：

1. `ADMIN_CONFIG`
2. 默认回退到 `manifest/config/config.local.yaml`

例如：

```bash
export ADMIN_CONFIG=manifest/config/config.local.yaml
```

### 4. 启动应用

```bash
go run .
```

### 5. 重建本地 MySQL 并校验注释

当你修改了 MySQL 表结构、表注释或字段注释时，推荐直接删卷重建本地库，再启动应用回灌默认种子：

```bash
docker compose down -v
docker compose up -d mysql redis
go run .
```

然后查询 `information_schema`，确认表注释和字段注释都已落库：

```bash
docker exec $(docker compose ps -q mysql) mysql -uroot -proot123456 -D admin -e "SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA='admin' AND TABLE_COMMENT = '';"

docker exec $(docker compose ps -q mysql) mysql -uroot -proot123456 -D admin -e "SELECT TABLE_NAME, COLUMN_NAME, COLUMN_COMMENT FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='admin' AND COLUMN_COMMENT = '';"
```

## 常用命令

### 全量测试

```bash
go test ./... -count=1 -timeout 60s
```

全量测试和契约测试都会通过 `NewTestCore()` 使用 MySQL 测试库 `admin_test`。
测试启动前会自动建库并清空旧表，因此不会污染日常运行使用的 `admin`。

### 契约测试

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 集成测试

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

### 第三方平台 live 验证

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_NAME='木木（星权益未税）'
export SUPPLIER_LIVE_DOMAIN=xqy.api.xqy1.cn
export SUPPLIER_LIVE_BACKUP_DOMAIN=xqy.api.xqy1.cn
export SUPPLIER_LIVE_TOKEN_ID=74
export SUPPLIER_LIVE_SECRET_KEY='***'
go test ./test/integration -run TestSupplierPlatformRefresh_LiveProviderBalance -count=1 -v
```

### 构建

```bash
go build ./...
```

### Lint（与 CI 对齐）

仓库 CI 会执行 `golangci-lint`（见 `.golangci.yml`），当前启用的核心检查包括：

- `govet`
- `staticcheck`
- `ineffassign`
- `unused`
- `typecheck`

本地建议执行：

```bash
golangci-lint run --timeout=5m
```

CI 里会使用增量参数（`--new-from-rev=origin/main`）减少无关历史问题的干扰。

## 配置说明

当前本地配置文件位于 `manifest/config/config.local.yaml`，主要包含：

- `server.address`
- `database.driver` / `database.dsn`
- `redis.addr` / `redis.password` / `redis.db`
- `auth.jwt_secret`
- `auth.access_token_ttl_minutes`
- `auth.temp_login_ttl_minutes`
- `bootstrap.super_admin_*`
- `sms.provider`
- `audit.async` / `audit.buffer_size`

配置文件支持环境变量展开，例如：

```yaml
bootstrap:
  super_admin_phone: "15881767197"
  super_admin_password: "abc123"
```

## SQL 与初始化

- `manifest/sql/001_schema.sql`：数据库结构
- `manifest/sql/002_seed_menu.sql`：菜单与权限种子
- `manifest/sql/003_seed_admin.sql.tmpl`：超级管理员 SQL 模板
- `manifest/sql/004_seed_config.sql`：系统配置初始值
- `manifest/sql/005_supplier_platform.sql`：第三方平台类型、账号和余额日志结构
- `manifest/sql/006_product_goods_channel_binding.sql`：商品渠道绑定表结构

补充说明：

- Docker 首次建库时，MySQL 会执行 `manifest/sql/*.sql`
- 应用启动时，`internal/app/schema.go` 会执行内置建表语句并补齐缺失表
- 因此修改 MySQL schema、表注释或字段注释时，必须同步 `manifest/sql/*.sql` 和 `internal/app/schema.go`

如需生成超级管理员初始化 SQL：

```bash
export SUPER_ADMIN_PHONE=15881767197
export SUPER_ADMIN_BCRYPT_HASH='$2a$10$exampleexampleexampleexampleexampleexampleexampleexample'
./hack/bootstrap-admin.sh
```

生成结果会写入 `manifest/sql/003_seed_admin.sql`。

## DAO 生成

当前使用封装脚本统一生成 DAO：

```bash
export GF_DAO_LINK='mysql:root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local'
./hack/gen-dao.sh
```

脚本会把产物写入：

- `internal/dao`
- `internal/model/do`
- `internal/model/entity`
- `internal/model/table`

## 日常开发约束

- controller 不直接访问 DAO
- logic 不把 HTTP 请求结构继续下传到更深层
- 跨模块基础能力优先收口到 `internal/library`
- 关键流程保留简体中文注释，尤其是启动、鉴权、短信、审计、配置兜底
- 文档描述以当前真实代码行为为准，不提前描述尚未实现的能力
