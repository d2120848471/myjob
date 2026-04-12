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

- MySQL：`127.0.0.1:3307`
- Redis：`127.0.0.1:6380`

### 2. 准备超级管理员引导参数

```bash
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
```

说明：

- 启动期会检查超级管理员手机号和密码
- 如果配置文件没有提供值，运行时必须从环境变量读取
- 这两个值缺失时，应用不会成功启动

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

## 常用命令

### 全量测试

```bash
go test ./... -count=1 -timeout 60s
```

### 契约测试

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 集成 smoke test

```bash
export MYJOB_RUN_INTEGRATION=1
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go test ./test/integration -count=1 -timeout 60s
```

### 构建

```bash
go build ./...
```

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
  super_admin_phone: "${SUPER_ADMIN_PHONE}"
  super_admin_password: "${SUPER_ADMIN_PASSWORD}"
```

## SQL 与初始化

- `manifest/sql/001_schema.sql`：数据库结构
- `manifest/sql/002_seed_menu.sql`：菜单与权限种子
- `manifest/sql/003_seed_admin.sql.tmpl`：超级管理员 SQL 模板
- `manifest/sql/004_seed_config.sql`：系统配置初始值

如需生成超级管理员初始化 SQL：

```bash
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_BCRYPT_HASH='$2a$10$exampleexampleexampleexampleexampleexampleexampleexample'
./hack/bootstrap-admin.sh
```

生成结果会写入 `manifest/sql/003_seed_admin.sql`。

## DAO 生成

当前使用封装脚本统一生成 DAO：

```bash
export GF_DAO_LINK='mysql:root:root123456@tcp(127.0.0.1:3307)/admin?charset=utf8mb4&parseTime=true&loc=Local'
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
