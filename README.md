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
3. `manifest/config/config.example.yaml`

如需切换环境：

```bash
export ADMIN_CONFIG=manifest/config/config.local.yaml
```

### 3. 启动服务

```bash
go run .
```

默认监听地址来自 `manifest/config/*.yaml`，当前示例配置为 `:8080`。

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
