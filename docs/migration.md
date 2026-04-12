# 迁移说明

## 迁移目标

把旧 `admin/` 子目录里的后台实现，迁移成仓库根主应用，形成统一的 GoFrame
工程目录和更稳定的职责边界。

## 迁移策略

### 第一阶段：冻结旧实现能力面

以旧 `admin/` 为兼容基线，盘点：

- 路由
- 响应结构
- 配置语义
- SQL 表结构
- Redis key 规则

### 第二阶段：在仓库根建立新骨架

已经完成：

- `main.go` + `internal/cmd`
- `internal/bootstrap` 启动装配
- `internal/controller/admin`
- `internal/logic/admin`
- `internal/service`
- `internal/library/*`
- `manifest/config`、`manifest/sql`、`test/contract`

### 第三阶段：按业务域迁移

当前已完成以下业务域落地：

1. `auth`
2. `user`
3. `group / rbac`
4. `subject`
5. `sms config`
6. `logs`

### 第四阶段：基础能力替换

已经把旧工程里的关键基础能力迁到独立目录：

- 鉴权：`internal/library/auth`
- 短信：`internal/library/sms`
- 审计：`internal/library/audit`
- IP 归属地：`internal/library/region`

## 旧目录的当前状态

`admin/` 目录目前仍保留在仓库里，作用只有两个：

1. 迁移对照参考
2. 回归核对基线

它不再是当前主运行入口。

## 后续可继续推进的收尾点

- 根据实际表结构进一步用 `gf gen dao` 收紧 DAO 自动生成产物
- 在确认不再需要对照后归档或删除旧 `admin/`
- 继续补强 `test/integration` 的真实依赖覆盖面
