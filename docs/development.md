# 开发说明

## 本地依赖

- Go 1.26+
- Docker / Docker Compose
- MySQL 8.4
- Redis 7

## 常用命令

### 启动依赖

```bash
docker compose up -d mysql redis
```

### 启动服务

```bash
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go run .
```

### 运行测试

```bash
go test ./... -count=1 -timeout 60s
```

### 构建

```bash
go build ./...
```

## 配置规则

配置文件位于 `manifest/config/`：

- `config.example.yaml`：示例模板
- `config.local.yaml`：本地运行模板

配置加载规则：

1. 优先读取 `ADMIN_CONFIG`
2. 否则读取 `manifest/config/config.local.yaml`
3. 如果本地配置不存在，再回退到 `manifest/config/config.example.yaml`

同时支持环境变量占位，例如：

```yaml
bootstrap:
  super_admin_phone: "${SUPER_ADMIN_PHONE}"
  super_admin_password: "${SUPER_ADMIN_PASSWORD}"
```

## SQL 与初始化

- `manifest/sql/001_schema.sql`：数据库结构
- `manifest/sql/002_seed_menu.sql`：菜单基础数据
- `manifest/sql/003_seed_admin.sql.tmpl`：超级管理员模板
- `manifest/sql/004_seed_config.sql`：系统配置初始值

如果需要生成超级管理员初始化 SQL：

```bash
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_BCRYPT_HASH='$2a$10$exampleexampleexampleexampleexampleexampleexampleexample'
./hack/bootstrap-admin.sh
```

## DAO 生成

当前项目已经预留 GoFrame DAO 生成脚本：

```bash
export GF_DAO_LINK='mysql:root:root123456@tcp(127.0.0.1:3307)/admin?charset=utf8mb4&parseTime=true&loc=Local'
./hack/gen-dao.sh
```

脚本会把生成结果落到：

- `internal/dao`
- `internal/model/do`
- `internal/model/entity`
- `internal/model/table`（表元数据）

## 开发约束

- controller 不直接访问 DAO
- logic 不直接暴露 HTTP 请求结构给更深层
- 跨模块能力优先放到 `internal/library`
- 关键流程保留简体中文注释，尤其是启动、鉴权、短信、审计、配置兜底
