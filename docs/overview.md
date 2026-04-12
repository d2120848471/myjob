# 项目概览

## 项目定位

当前仓库是一个运行在仓库根目录的 GoFrame 单体后台。
它的目标不是做微服务拆分，而是在保持既有后台语义的前提下，把后端整理成更清晰、可维护、可测试的目录结构。

当前根目录代码就是唯一主实现，历史 `admin/` 子工程不再作为运行入口。

## 当前运行形态

- 主进程入口：`main.go`
- 启动命令入口：`internal/cmd/root.go`
- 应用装配：`internal/bootstrap/application.go`
- 运行时核心：`internal/app`
- 配置优先级：`ADMIN_CONFIG` > `manifest/config/config.local.yaml`
- 默认本地依赖：MySQL 8.4、Redis 7

## 当前功能状态

### 认证与会话

当前认证链路已经是“账号密码 + 条件短信二验”的模式：

- 正常场景下，账号密码校验通过后直接签发 token
- 首次登录或登录 IP 变化时，登录接口返回临时登录态，要求继续走短信验证码验证
- 短信校验成功后再签发正式会话 token
- `me` 与退出登录复用同一套会话鉴权链路

### 权限与菜单

- 超级管理员固定使用 `group_id = 0`
- 普通后台账号通过用户组拿到权限码
- 菜单授权和菜单树默认过滤 `super_only = 1` 的菜单
- 短信配置接口属于 super-only 功能，不向普通用户组开放

### 后台业务域

当前已经落地的业务域包括：

- 登录鉴权与会话
- 员工管理
- 用户组与菜单授权
- 主体管理
- 短信配置
- 操作日志 / 登录日志

## 当前兼容约束

- 后台接口统一挂在 `/api/admin/*`
- 响应体统一为 `code / message / data`
- 继续支持 `ADMIN_CONFIG`、`SUPER_ADMIN_PHONE`、`SUPER_ADMIN_PASSWORD`
- 保留 MySQL、Redis、短信配置、菜单权限、日志表这套现有业务语义

## 非目标

当前项目不做这些事情：

- 不拆分微服务
- 不预埋无实际使用场景的平台化基础设施
- 不在文档里描述尚未实现的“未来接口”或“计划中的完整测试覆盖”

## 文档使用建议

- 想看怎么启动：先读 `docs/development.md`
- 想看架构和请求流：先读 `docs/architecture.md`
- 想看模块和权限关系：先读 `docs/module-map.md`
- 想看测试现状：先读 `docs/testing.md`
