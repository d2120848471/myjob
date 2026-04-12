# 模块职责地图

## 业务域与服务接口

| 业务域 | service 接口 | controller | logic | 说明 |
| --- | --- | --- | --- | --- |
| 登录鉴权 | `AuthService` | `internal/controller/admin/auth.go` | `internal/logic/admin/auth.go` | 账号登录、短信二验、获取 me、退出登录 |
| 员工管理 | `UserService` | `internal/controller/admin/user.go` | `internal/logic/admin/user.go` | 用户列表、增改删、恢复、状态、业务账号设置 |
| 用户组 / 授权 | `GroupService` | `internal/controller/admin/group.go` | `internal/logic/admin/group.go` | 用户组 CRUD、菜单授权、菜单树 |
| 主体管理 | `SubjectService` | `internal/controller/admin/subject.go` | `internal/logic/admin/subject.go` | 主体列表、新增、编辑 |
| 会话管理 | `AuthService` | `internal/controller/admin/session.go` | `internal/logic/admin/auth.go` | 当前登录态与退出登录 |
| 短信配置 | `SMSConfigService` | `internal/controller/admin/settings.go` | `internal/logic/admin/config.go` | 短信配置读取、保存与脱敏展示 |
| 审计日志 | `AuditLogService` | `internal/controller/admin/operation_log.go` / `internal/controller/admin/login_log.go` | `internal/logic/admin/log.go` | 操作日志、登录日志查询 |

## 基础目录职责

| 目录 | 主要职责 | 备注 |
| --- | --- | --- |
| `api/admin/v1` | 请求 / 响应协议定义 | 版本化 GoFrame 严格路由协议，不承载业务逻辑 |
| `internal/bootstrap` | 应用装配、路由绑定、运行入口对接 | 启动链路核心 |
| `internal/cmd` | GoFrame 命令入口 | 根应用启动命令 |
| `internal/consts` | 后台通用常量 | 状态值、通用枚举 |
| `internal/controller/admin` | HTTP 协议层 | 标准 `Req/Res + error` 控制器 |
| `internal/service` | 模块接口边界 | 供 controller 依赖 |
| `internal/logic/admin` | 业务编排层 | 组合 app 与 library |
| `internal/app` | 运行时核心与公共数据能力 | GoFrame 配置 / DB / Redis 入口收口层 |
| `internal/dao` | GoFrame DAO 模型装配 | 为进一步自动生成预留统一入口 |
| `internal/model/do` | 数据写入 / 条件对象 | 贴近持久化层 |
| `internal/model/entity` | 数据库实体结构 | 贴近表结构 |
| `internal/middleware` | 中间件 | 统一响应、鉴权与授权 |
| `internal/library` | 跨模块基础能力 | auth / sms / audit / region |
| `manifest/config` | 配置模板 | 环境变量展开后加载 |
| `manifest/sql` | SQL 初始化资源 | schema / seed / config |
| `hack` | 辅助脚本 | 初始化 SQL、DAO 生成 |
| `test/contract` | 接口兼容与核心业务流测试 | 默认 CI 风格测试入口 |
| `test/integration` | 真实依赖联动测试 | 通过环境开关显式启用 |
| `test/fixture` | 测试说明与样例资源目录 | 放共享伪数据或样例配置 |

## 历史目录映射

| 旧位置 | 新位置 | 说明 |
| --- | --- | --- |
| `admin/cmd/admin/main.go` | `main.go` + `internal/cmd/root.go` | 根目录成为唯一主入口 |
| `admin/internal/app/application.go` | `internal/bootstrap/application.go` | 启动装配拆出 |
| `admin/internal/app/routes.go` | `internal/bootstrap/application.go` | 路由绑定收口 |
| `admin/internal/app/*handlers.go` | `internal/controller/admin` + `internal/logic/admin` | 协议层与业务层拆分 |
| `admin/utility/ipx` | `internal/library/region` | 基础库归位 |

## 当前接口入口

| 分类 | 路径前缀 | 说明 |
| --- | --- | --- |
| 认证 | `/api/admin/auth/*` | 登录、短信二验 |
| 会话 | `/api/admin/auth/*` | `me`、退出登录 |
| 员工 | `/api/admin/users*` | 员工 CRUD、状态、业务设置 |
| 用户组 | `/api/admin/groups*` | 用户组 CRUD 与权限 |
| 主体 | `/api/admin/subjects*` | 主体管理 |
| 设置 | `/api/admin/settings/*` | 短信配置 |
| 日志 | `/api/admin/logs/*` | 操作日志、登录日志 |
