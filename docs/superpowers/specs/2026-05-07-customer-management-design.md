# 客户管理 V1 设计

## 背景

当前系统已经具备后台员工登录、短信二次验证、后台权限、操作日志和统一响应能力，但还没有独立的客户账号域。客户不是后台员工，也不是开放订单 token 调用方，因此本次新增独立客户边界，避免污染现有 `/api/admin/auth` 和 `/api/open` 语义。

本设计覆盖客户侧认证闭环和后台客户完整管理。客户侧面向注册、登录和忘记密码；后台侧面向客户资料维护、启停、回收站和密码重置。

## 目标

- 新增客户侧认证接口：短信发送、注册、登录、忘记密码。
- 新增后台客户管理接口：新增、列表、详情、编辑、启停、软删除、回收站、恢复、重置登录密码、重置支付密码。
- 复用现有短信配置和短信 sender，但客户短信验证码按场景隔离。
- 建立独立客户 session 与 token 体系，不复用后台员工会话。
- 保持 `api/` 扁平协议目录和同 package 多文件拆分风格。
- 补齐契约测试、核心逻辑测试和稳定文档。

## 非目标

- 不做客户余额、授信、等级、联系人、备注等 CRM 字段。
- 不做客户侧 `me/logout` 接口；后续客户中心或客户下单接口接入时再补。
- 不做协议勾选后端校验，协议勾选只由前端限制。
- 不新增注册和找回密码的独立短信模板配置。
- 回收站不支持永久删除。

## 架构边界

客户侧 API 使用独立前缀：

```text
/api/customer/auth/*
  ├─ 发送短信验证码
  ├─ 注册
  ├─ 登录
  └─ 忘记密码
```

后台客户管理使用后台前缀：

```text
/api/admin/customers*
  ├─ 新增客户
  ├─ 列表 / 回收站 / 详情
  ├─ 编辑资料
  ├─ 启停
  ├─ 软删除 / 恢复
  └─ 重置登录密码 / 支付密码
```

建议文件归属：

```text
api/customer_auth.go
api/customer.go
internal/controller/customer/auth.go
internal/controller/admin/customer.go
internal/service/customer_auth.go
internal/service/customer.go
internal/logic/customer/auth.go
internal/logic/customer/auth_sms.go
internal/logic/customer/auth_session.go
internal/logic/customer/auth_validate.go
internal/logic/admin/customer.go
internal/logic/admin/customer_query.go
internal/logic/admin/customer_write.go
internal/logic/admin/customer_validate.go
internal/logic/admin/customer_mapper.go
internal/app/customer_session.go
internal/app/customer_sms.go
```

客户侧逻辑放到 `internal/logic/customer`，后台客户管理逻辑放到 `internal/logic/admin`。两者共享 `customer_user` 表，但不共享后台员工 `AuthLogic`。

## 数据模型

新增客户主表 `customer_user`：

```sql
CREATE TABLE customer_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  company_name VARCHAR(100) NOT NULL COMMENT '店铺或公司名称',
  phone VARCHAR(20) NOT NULL COMMENT '手机号',
  password_hash VARCHAR(255) NOT NULL COMMENT '登录密码哈希',
  pay_password_hash VARCHAR(255) NOT NULL COMMENT '支付密码哈希',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用，0禁用',
  is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
  last_login_ip VARCHAR(45) NULL COMMENT '最后登录IP',
  last_login_at DATETIME NULL COMMENT '最后登录时间',
  token_version INT NOT NULL DEFAULT 0 COMMENT '令牌版本',
  deleted_at DATETIME NULL COMMENT '删除时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_customer_user_phone (phone),
  KEY idx_customer_user_status_deleted (status, is_deleted, id),
  KEY idx_customer_user_company (company_name, id)
);
```

规则：

- `company_name` 必填，注册和后台新增都必须提交。
- `phone` 全局唯一；软删除后仍然占用手机号。
- 登录密码和支付密码都只存哈希。
- 支付密码必须是 6 位数字。
- 注册默认 `status=1`，但保留状态字段给后台启停。
- `token_version` 用于让旧客户 token 失效。

需要同步 MySQL 初始化 SQL、测试态 SQLite schema、entity/model 结构和 DAO 表清单。

## 客户侧接口

### 发送短信验证码

```text
POST /api/customer/auth/sms/send
```

入参：

- `phone`
- `scene`：`register` 或 `forgot_password`

行为：

- 校验手机号格式。
- `register` 场景下，手机号已存在或在回收站时返回已注册。
- `forgot_password` 场景下，手机号不存在、已删除或禁用时返回不可找回密码。
- 写入按 `scene + phone` 隔离的验证码缓存和发送锁。

### 注册

```text
POST /api/customer/auth/register
```

入参：

- `company_name`
- `phone`
- `sms_code`
- `password`
- `confirm_password`
- `pay_password`
- `confirm_pay_password`

行为：

- 校验短信验证码。
- 校验手机号唯一。
- 校验登录密码和确认登录密码一致。
- 校验支付密码和确认支付密码一致，并且支付密码是 6 位数字。
- 创建客户，注册成功后直接签发客户 token。

返回：

- `token`
- `customer`

### 登录

```text
POST /api/customer/auth/login
```

入参：

- `phone`
- `password`

行为：

- 校验客户存在、未删除、状态启用、密码正确。
- 更新最后登录 IP 和时间。
- 签发客户 token。

返回：

- `token`
- `customer`

### 忘记密码

```text
POST /api/customer/auth/forgot-password
```

入参：

- `phone`
- `sms_code`
- `password`
- `confirm_password`

行为：

- 校验短信验证码。
- 重置登录密码哈希。
- 提升 `token_version`，让旧 token 失效。

返回体为空。

## 后台接口

后台客户管理接口都需要后台 Bearer 鉴权和 `customer.manage` 权限。

```text
GET    /api/admin/customers
GET    /api/admin/customers/trash
GET    /api/admin/customers/{id}
POST   /api/admin/customers
PUT    /api/admin/customers/{id}
PATCH  /api/admin/customers/{id}/status
DELETE /api/admin/customers/{id}
PATCH  /api/admin/customers/{id}/restore
PATCH  /api/admin/customers/{id}/password
PATCH  /api/admin/customers/{id}/pay-password
```

后台能力：

- 列表支持 `page`、`page_size`、`keyword`、`status`；`keyword` 匹配公司名或手机号。
- 回收站列表支持 `page`、`page_size`、`keyword`。
- 新增客户由管理员输入公司名、手机号、登录密码、支付密码和状态，不走短信验证码。
- 编辑客户资料允许修改公司名、手机号和状态。
- 启停接口独立存在；禁用时提升 `token_version`。
- 软删除写入 `is_deleted/deleted_at`，并提升 `token_version`。
- 回收站只支持恢复，不支持永久删除。
- 重置登录密码由管理员输入新密码，不返回明文，并提升 `token_version`。
- 重置支付密码由管理员输入 6 位数字，不返回明文，不强制踢登录态。

## 短信设计

客户短信复用现有短信配置和短信 sender，但使用独立 Redis key：

```text
customer:sms:<scene>:<phone>
customer:sms:send_lock:<scene>:<phone>
customer:sms:attempts:<scene>:<phone>
```

规则：

- 验证码为 6 位数字。
- TTL 复用短信配置 `expire_minutes`。
- 发送锁复用短信配置 `interval_minutes`。
- 错误次数 key 复用验证码剩余 TTL，超过 5 次后删除验证码缓存。
- 发送失败时删除验证码缓存、发送锁和错误次数 key。
- 注册或找回密码成功后删除验证码缓存和错误次数 key。
- 不同场景验证码不能互用。

## 会话设计

客户会话独立于后台员工会话：

```text
customer:session:<jti>
customer:user:sessions:<customer_id>
```

客户 token claims：

```text
customer_id
token_version
jti
exp
```

失效策略：

- 登录密码重置、找回密码成功、后台禁用、后台软删除都提升 `token_version`。
- 支付密码重置不提升 `token_version`。
- 编辑公司名和手机号不踢当前登录态。

## 权限与审计

新增后台权限码：

```text
customer.manage
```

后台客户接口全部使用该权限。超级管理员天然可访问，普通后台员工需要用户组授权。

操作日志记录：

- 新增客户：记录公司名和手机号。
- 编辑客户：记录客户 ID。
- 启用/禁用：记录客户 ID。
- 软删除/恢复：记录客户 ID。
- 重置登录密码/支付密码：记录客户 ID，不记录密码明文。

## 测试设计

新增契约测试：

```text
customer_auth_contract_test.go
customer_admin_contract_test.go
```

覆盖：

- OpenAPI 暴露客户侧和后台客户接口。
- 统一响应结构仍为 `code / message / data`。
- 注册短信发送、验证码校验、注册成功返回 token。
- 登录成功返回 token；禁用、软删除、密码错误不能登录。
- 忘记密码短信和登录密码重置。
- 后台新增、列表、详情、编辑、启停、软删除、回收站、恢复。
- 后台重置登录密码后旧 token 失效。
- 后台重置支付密码不返回明文。

核心逻辑测试：

- 正常客户和回收站客户都占用手机号。
- 支付密码必须是 6 位数字。
- 短信发送失败清理验证码和发送锁。
- `scene + phone` 隔离。
- `token_version` 失效策略。

交付前验证命令：

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

## 文档同步

实现时需要同步：

- `README.md`
- `docs/README.md`
- `docs/module-map.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/testing.md`
- `test/contract/README.md`

同步内容：

- 客户管理业务域。
- 客户侧 `/api/customer/auth/*` 路由。
- 后台 `/api/admin/customers*` 路由与 `customer.manage` 权限。
- `customer_user` 表。
- 客户短信和客户会话 Redis key 语义。
- 客户认证和后台客户管理契约测试口径。

## 风险与约束

- 客户侧认证不能复用后台员工认证逻辑，否则权限和 session 语义会混杂。
- 手机号软删除后仍占用，能避免恢复冲突，但运营需要从回收站恢复旧客户而不是重新注册。
- 支付密码虽然是 6 位数字，也必须哈希落库。
- 后台重置密码不能在响应或日志中暴露明文。
- V1 不做客户侧 `me/logout`，后续接客户中心或下单鉴权时再补。
