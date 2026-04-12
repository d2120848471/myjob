# 模块职责地图

## 业务域到实现目录

### 认证与会话

- 协议：`api/auth.go`
- controller：`internal/controller/admin/auth.go`、`internal/controller/admin/session.go`
- service：`internal/service/interfaces.go` 中的 `AuthService`
- logic：`internal/logic/admin/auth.go`
- 路由前缀：`/api/admin/auth/*`
- 主要能力：
  - 账号密码登录
  - 条件短信二验
  - 发送验证码 / 校验验证码
  - 获取当前登录信息
  - 退出登录

### 员工管理

- 协议：`api/user.go`
- controller：`internal/controller/admin/user.go`
- service：`UserService`
- logic：`internal/logic/admin/user.go`
- 路由前缀：`/api/admin/users*`
- 主要能力：
  - 员工列表
  - 回收站
  - 新增 / 编辑 / 删除 / 恢复
  - 启停
  - 余额通知开关
  - 批量设置 / 取消商务

### 用户组与授权

- 协议：`api/group.go`
- controller：`internal/controller/admin/group.go`
- service：`GroupService`
- logic：`internal/logic/admin/group.go`
- 路由前缀：`/api/admin/groups*`、`/api/admin/menus/tree`
- 主要能力：
  - 用户组列表、增删改、状态切换
  - 菜单树
  - 保存用户组菜单授权
- 权限边界：
  - 用户组授权只允许分配 `super_only = 0` 的菜单
  - 菜单树只返回 `super_only = 0` 的菜单项

### 主体管理

- 协议：`api/subject.go`
- controller：`internal/controller/admin/subject.go`
- service：`SubjectService`
- logic：`internal/logic/admin/subject.go`
- 路由前缀：`/api/admin/subjects*`
- 主要能力：
  - 主体列表
  - 新增主体
  - 编辑主体

### 品牌管理

- 协议：`api/brand.go`
- controller：`internal/controller/admin/brand.go`
- service：`BrandService`
- logic：`internal/logic/admin/brand.go`
- 路由前缀：`/api/admin/brands*`
- 主要能力：
  - 一级品牌分页与搜索
  - 品牌子级懒加载（支持三级）
  - 新增 / 编辑 / 删除
  - 同级排序与显隐切换
  - 本地图片上传
- 权限边界：
  - 品牌接口统一要求 `product.brand`

### 行业管理

- 协议：`api/industry.go`
- controller：`internal/controller/admin/industry.go`
- service：`IndustryService`
- logic：`internal/logic/admin/industry.go`
- 路由前缀：`/api/admin/industries*`
- 主要能力：
  - 行业列表、增删改、排序
  - 品牌选择器
  - 行业品牌关联增删排序
  - 行业删除 / 品牌删除前的关联校验
- 权限边界：
  - 行业接口统一要求 `product.industry`

### 短信配置

- 协议：`api/settings.go`
- controller：`internal/controller/admin/settings.go`
- service：`SMSConfigService`
- logic：`internal/logic/admin/config.go`
- 路由前缀：`/api/admin/settings/sms`
- 主要能力：
  - 读取脱敏后的短信配置状态
  - 保存阿里云 AccessKey、签名、模板和验证码时效配置
- 权限边界：
  - 该模块是 super-only 接口
  - 普通用户组不会获得 `config.sms` 菜单权限

### 审计日志

- 协议：`api/log.go`
- controller：`internal/controller/admin/operation_log.go`、`internal/controller/admin/login_log.go`
- service：`AuditLogService`
- logic：`internal/logic/admin/log.go`
- 路由前缀：`/api/admin/logs/*`
- 主要能力：
  - 操作日志分页查询
  - 登录日志分页查询
  - 支持管理员 ID、时间范围过滤
  - 操作日志额外支持关键字过滤

## 目录地图

### `api`

对外协议目录，当前是扁平文件结构：

- `auth.go`
- `common.go`
- `brand.go`
- `group.go`
- `industry.go`
- `log.go`
- `settings.go`
- `subject.go`
- `user.go`

### `internal/bootstrap`

应用装配层，负责：

- 创建 controller / service / logic 组合
- 绑定 `/api/admin` 路由
- 注册统一响应中间件和鉴权中间件
- 注册 OpenAPI / Swagger

### `internal/controller/admin`

HTTP 协议适配层，不直接写业务规则。

### `internal/service`

模块接口边界层。

### `internal/logic/admin`

业务编排层，当前按业务域拆分为：

- `auth.go`
- `user.go`
- `group.go`
- `subject.go`
- `brand.go`
- `industry.go`
- `config.go`
- `log.go`

### `internal/app`

运行时核心层，负责配置、依赖初始化、种子引导、公共查询和通用辅助能力。

### `internal/library`

基础能力库：

- `auth`
- `sms`
- `audit`
- `region`

### `internal/model`

当前已使用的模型子目录：

- `config`：配置结构体
- `do`：写入 / 条件对象
- `dto/admin`：分页 DTO 等内部传输对象
- `entity`：数据库实体与查询结果结构
- `runtime`：登录态、短信配置等运行时模型

### `manifest/config`

运行时配置目录，当前主要是 `config.local.yaml`。

### `manifest/sql`

数据库 schema、菜单种子、超级管理员模板和系统配置种子。

### `test/contract`

契约测试目录，覆盖接口兼容和核心业务流。

### `test/integration`

集成测试目录，当前只有 runtime smoke test。

## 当前路由与权限摘要

| 模块 | 路由前缀 | 进入条件 |
| --- | --- | --- |
| 认证登录 / 短信发送 / 短信验证 | `/api/admin/auth/login`、`/api/admin/auth/sms/*` | 无需登录 |
| 会话信息 / 退出登录 | `/api/admin/auth/me`、`/api/admin/auth/session` | 需要登录 |
| 员工管理 | `/api/admin/users*` | `admin.list` |
| 用户组与授权 | `/api/admin/groups*`、`/api/admin/menus/tree` | `admin.department` |
| 主体管理 | `/api/admin/subjects*` | `subject.manage` |
| 品牌管理 | `/api/admin/brands*` | `product.brand` |
| 行业管理 | `/api/admin/industries*` | `product.industry` |
| 短信配置 | `/api/admin/settings/sms` | super-only |
| 操作日志 | `/api/admin/logs/operations` | `admin.action` |
| 登录日志 | `/api/admin/logs/logins` | `admin.loginlog` |
