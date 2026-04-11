# Phase 1｜基础骨架与登录鉴权

## 1. 本期目标

本期必须完成一个**可登录、可退出、可鉴权、可记录登录日志**的后端骨架，并把“首次登录 / IP 变更时短信二次验证”的登录链路完整打通。

本期结束后，系统至少要具备：

- 启动服务能力
- 超管账号登录能力
- JWT + Redis 会话校验
- `/api/admin/me`
- `/api/admin/logout`
- 统一鉴权中间件
- 登录日志写入
- 普通接口的权限校验骨架

## 2. 本期范围

### 2.1 接口范围

- `POST /api/admin/login`
- `POST /api/admin/login/sms/send`
- `POST /api/admin/login/sms/verify`
- `POST /api/admin/me`
- `POST /api/admin/logout`

### 2.2 表范围

- `admin_user`
- `admin_menu`
- `admin_login_log`
- `system_config`

### 2.3 Redis Key 范围

- `admin:session:{jti}`
- `admin:user:sessions:{user_id}`
- `admin:login:tmp:{login_token}`
- `sms:login:{user_id}`
- `sms:login:send_lock:{user_id}`
- `admin:perm:group:{group_id}`
- `admin:config:sms`

## 3. 目录新增清单

```text
api/admin/auth/v1/
├── login.go
├── login_sms_send.go
├── login_sms_verify.go
├── me.go
└── logout.go

internal/controller/admin/auth/
├── auth.go
├── login.go
├── login_sms_send.go
├── login_sms_verify.go
├── me.go
└── logout.go

internal/service/
└── auth.go

internal/logic/admin/auth/
├── auth.go
├── login.go
├── login_sms_send.go
├── login_sms_verify.go
├── me.go
└── logout.go

internal/logic/admin/log/
├── login.go
└── login_write.go

internal/logic/admin/permission/
├── permission.go
└── cache.go

internal/middleware/
├── auth_jwt.go
└── auth_permission.go

internal/model/dto/admin/
└── auth.go

internal/model/contextx/
└── admin.go

utility/
├── bcryptx/bcryptx.go
├── jwtx/jwtx.go
├── randx/code.go
└── ipx/
    ├── client_ip.go
    ├── mask.go
    └── region.go
```

## 4. API 对象定义

### 4.1 `api/admin/auth/v1/login.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq 第一步：账号密码登录。
type LoginReq struct {
    g.Meta `path:"/login" method:"post" tags:"登录认证" summary:"账号密码登录" noAuth:"true"`
    Username string `json:"username" v:"required#请输入用户名"`
    Password string `json:"password" v:"required#请输入密码"`
}

// LoginRes 第一步登录返回。
// need_sms_verify=false 时直接返回 token；否则返回 login_token。
type LoginRes struct {
    NeedSMSVerify bool         `json:"need_sms_verify"`
    LoginToken    string       `json:"login_token,omitempty"`
    Phone         string       `json:"phone,omitempty"`
    Reason        string       `json:"reason,omitempty"`
    Token         string       `json:"token,omitempty"`
    User          *LoginUser   `json:"user,omitempty"`
    Permissions   []string     `json:"permissions,omitempty"`
}

// LoginUser 登录态用户信息。
type LoginUser struct {
    ID         int64  `json:"id"`
    Username   string `json:"username"`
    RealName   string `json:"real_name"`
    GroupID    int64  `json:"group_id"`
    GroupName  string `json:"group_name"`
    IsBusiness int    `json:"is_business"`
}
```

### 4.2 `api/admin/auth/v1/login_sms_send.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginSMSSendReq 发送登录验证码。
type LoginSMSSendReq struct {
    g.Meta `path:"/login/sms/send" method:"post" tags:"登录认证" summary:"发送登录验证码" noAuth:"true"`
    LoginToken string `json:"login_token" v:"required#login_token不能为空"`
}

// LoginSMSSendRes 发送验证码返回。
type LoginSMSSendRes struct{}
```

### 4.3 `api/admin/auth/v1/login_sms_verify.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginSMSVerifyReq 验证登录验证码并完成登录。
type LoginSMSVerifyReq struct {
    g.Meta `path:"/login/sms/verify" method:"post" tags:"登录认证" summary:"验证登录验证码并完成登录" noAuth:"true"`
    LoginToken string `json:"login_token" v:"required#login_token不能为空"`
    SMSCode    string `json:"sms_code" v:"required|regex:^\d{6}$#请输入验证码|验证码格式错误"`
}

// LoginSMSVerifyRes 二次验证成功后的最终登录结果。
type LoginSMSVerifyRes struct {
    Token       string      `json:"token"`
    User        *LoginUser  `json:"user"`
    Permissions []string    `json:"permissions"`
}
```

### 4.4 `api/admin/auth/v1/me.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

// MeReq 获取当前登录用户信息。
type MeReq struct {
    g.Meta `path:"/me" method:"post" tags:"登录认证" summary:"获取当前登录用户信息"`
}

// MeRes 当前用户信息。
type MeRes struct {
    User        *LoginUser `json:"user"`
    Permissions []string   `json:"permissions"`
}
```

### 4.5 `api/admin/auth/v1/logout.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

// LogoutReq 退出登录。
type LogoutReq struct {
    g.Meta `path:"/logout" method:"post" tags:"登录认证" summary:"退出登录"`
}

// LogoutRes 退出登录返回。
type LogoutRes struct{}
```

## 5. DTO 与 Service 设计

### 5.1 `internal/model/dto/admin/auth.go`

```go
package admin

// LoginInput 第一步登录输入。
type LoginInput struct {
    Username string
    Password string
    LoginIP  string
}

// LoginOutput 第一步登录输出。
type LoginOutput struct {
    NeedSMSVerify bool
    LoginToken    string
    Phone         string
    Reason        string
    Token         string
    User          LoginUserVO
    Permissions   []string
}

// VerifyLoginSMSInput 第二步登录输入。
type VerifyLoginSMSInput struct {
    LoginToken string
    SMSCode    string
    LoginIP    string
}

// LoginUserVO 登录态用户对象。
type LoginUserVO struct {
    ID         int64
    Username   string
    RealName   string
    GroupID    int64
    GroupName  string
    IsBusiness int
}
```

### 5.2 `internal/service/auth.go`

```go
package service

import (
    "context"
    adminDto "project/internal/model/dto/admin"
)

type IAuth interface {
    // Login 执行账号密码登录第一步。
    Login(ctx context.Context, in adminDto.LoginInput) (out adminDto.LoginOutput, err error)

    // SendLoginSMS 发送登录验证码。
    SendLoginSMS(ctx context.Context, loginToken string) error

    // VerifyLoginSMS 验证验证码并完成登录。
    VerifyLoginSMS(ctx context.Context, in adminDto.VerifyLoginSMSInput) (out adminDto.LoginOutput, err error)

    // GetMe 获取当前登录用户信息与权限。
    GetMe(ctx context.Context, adminID int64) (out adminDto.LoginOutput, err error)

    // Logout 注销当前 token。
    Logout(ctx context.Context, adminID int64, jti string) error
}

var localAuth IAuth

func RegisterAuth(i IAuth) {
    localAuth = i
}

func Auth() IAuth {
    if localAuth == nil {
        panic("service IAuth not registered")
    }
    return localAuth
}
```

## 6. 登录流程拆分

### 6.1 第一步 `/login`

处理顺序：

1. 根据 `username` 查 `admin_user`
2. 统一处理“账号不存在 / 已删除 / 密码错误”为 `账号或密码错误`
3. `status=0` 返回 `账号已被禁用，请联系管理员`
4. `phone` 为空返回 `请联系管理员配置手机号`
5. 判断是否需要短信验证：
   - `last_login_ip IS NULL` → `first_login`
   - `current_ip != last_login_ip` → `ip_changed`
6. 不需要短信：
   - 直接签发正式 token
   - 写 Redis Session
   - 查询权限码
   - 更新 `last_login_ip`、`last_login_at`
   - 写登录日志
7. 需要短信：
   - 生成 `login_token`
   - Redis 记录临时状态，TTL=5 分钟
   - 返回 `need_sms_verify=true`

### 6.2 第二步 `/login/sms/send`

处理顺序：

1. 校验 `login_token`
2. 校验发送间隔锁
3. 读取短信配置
4. 生成 6 位数字验证码
5. 写 `sms:login:{user_id}`
6. 写 `sms:login:send_lock:{user_id}`
7. 调短信发送器发送

### 6.3 第三步 `/login/sms/verify`

处理顺序：

1. 校验 `login_token`
2. 读取 `sms:login:{user_id}`
3. 验证失败次数 +1
4. 超过 5 次删除 `login_token`
5. 验证通过后删除短信验证码
6. 调用统一 `completeLogin`：
   - 签发 token
   - 写 session
   - 查询权限
   - 更新最后登录 IP/时间
   - 写登录日志

## 7. 临时登录态 Redis 结构

### 7.1 `admin:login:tmp:{login_token}`

建议保存 JSON：

```json
{
  "user_id": 12,
  "username": "admin02",
  "login_ip": "1.2.3.4",
  "reason": "first_login",
  "failed_count": 0,
  "created_at": "2026-04-11 10:00:00"
}
```

### 7.2 验证码 Key

- `sms:login:{user_id}` → `123456`
- TTL：读取系统配置，默认 30 分钟

### 7.3 发送锁 Key

- `sms:login:send_lock:{user_id}` → `1`
- TTL：读取系统配置，默认 1 分钟

## 8. 会话设计

### 8.1 JWT Claims

```go
package jwtx

type AdminClaims struct {
    UserID       int64  `json:"user_id"`
    GroupID      int64  `json:"group_id"`
    TokenVersion uint32 `json:"token_version"`
    JTI          string `json:"jti"`
    jwt.RegisteredClaims
}
```

### 8.2 Redis Session 值

`admin:session:{jti}` 建议保存 JSON：

```json
{
  "user_id": 12,
  "group_id": 3,
  "token_version": 5,
  "login_at": "2026-04-11 10:00:00"
}
```

### 8.3 用户活跃会话集合

`admin:user:sessions:{user_id}` 保存用户全部 `jti`，用于：

- 退出单个 token
- 删除用户时清理全部 token
- 禁用用户时清理全部 token

## 9. 核心代码骨架

### 9.1 `internal/logic/admin/auth/auth.go`

```go
package auth

import "project/internal/service"

type sAuth struct{}

func New() *sAuth {
    return &sAuth{}
}

func init() {
    service.RegisterAuth(New())
}
```

### 9.2 `internal/logic/admin/auth/login.go`

```go
package auth

import (
    "context"
    "fmt"
    "time"

    "github.com/gogf/gf/v2/errors/gerror"

    "project/internal/dao"
    adminDto "project/internal/model/dto/admin"
    "project/internal/model/do"
    "project/utility/bcryptx"
    "project/utility/ipx"
    "project/utility/jwtx"
)

// Login 执行账号密码登录第一步。
func (s *sAuth) Login(ctx context.Context, in adminDto.LoginInput) (out adminDto.LoginOutput, err error) {
    user, err := dao.AdminUser.Ctx(ctx).
        Where(do.AdminUser{
            Username:  in.Username,
            IsDeleted: 0,
        }).
        One()
    if err != nil {
        return out, err
    }
    if user.IsEmpty() {
        return out, gerror.New("账号或密码错误")
    }

    entity := user.MapToStruct()
    if entity.Status == 0 {
        return out, gerror.New("账号已被禁用，请联系管理员")
    }
    if entity.Phone == "" {
        return out, gerror.New("请联系管理员配置手机号")
    }
    if !bcryptx.Verify(in.Password, entity.PasswordHash) {
        return out, gerror.New("账号或密码错误")
    }

    // 首次登录或 IP 变更时，返回 login_token，由前端进入短信验证第二步。
    if entity.LastLoginIp == "" || entity.LastLoginIp != in.LoginIP {
        loginToken := jwtx.NewTempToken()
        reason := "ip_changed"
        if entity.LastLoginIp == "" {
            reason = "first_login"
        }
        err = s.saveTempLoginState(ctx, loginToken, entity.Id, entity.Username, in.LoginIP, reason)
        if err != nil {
            return out, err
        }
        out.NeedSMSVerify = true
        out.LoginToken = loginToken
        out.Phone = ipx.MaskPhone(entity.Phone)
        out.Reason = reason
        return out, nil
    }

    return s.completeLogin(ctx, entity.Id, in.LoginIP)
}
```

> 说明：上面是骨架代码，`MapToStruct()` 由 Codex 按实际 DAO 生成结果替换成明确实体转换代码，不允许直接复制后不编译检查。

### 9.3 `internal/logic/admin/auth/login_sms_send.go`

```go
// SendLoginSMS 发送登录验证码。
// 仅在 login_token 合法、且发送间隔允许时才发起短信发送。
func (s *sAuth) SendLoginSMS(ctx context.Context, loginToken string) error {
    state, err := s.getTempLoginState(ctx, loginToken)
    if err != nil {
        return err
    }

    cfg, err := s.getSMSConfig(ctx)
    if err != nil {
        return err
    }

    if ok, err := s.checkSendInterval(ctx, state.UserID, cfg.IntervalMinutes); err != nil {
        return err
    } else if !ok {
        return gerror.NewCode(gcode.New(429, "请稍后再试", nil), "请稍后再试")
    }

    code := randx.Numeric6()
    if err = s.cacheSMSCode(ctx, state.UserID, code, cfg.ExpireMinutes); err != nil {
        return err
    }
    if err = s.cacheSMSSendLock(ctx, state.UserID, cfg.IntervalMinutes); err != nil {
        return err
    }

    // 真正的阿里云调用通过 sender 接口完成，避免业务逻辑直接耦合 SDK。
    return s.smsSender.SendLoginCode(ctx, state.Phone, code, cfg)
}
```

### 9.4 `internal/logic/admin/auth/login_sms_verify.go`

```go
// VerifyLoginSMS 验证验证码并完成登录。
func (s *sAuth) VerifyLoginSMS(ctx context.Context, in adminDto.VerifyLoginSMSInput) (out adminDto.LoginOutput, err error) {
    state, err := s.getTempLoginState(ctx, in.LoginToken)
    if err != nil {
        return out, err
    }

    ok, remain, err := s.checkSMSCode(ctx, state.UserID, in.SMSCode)
    if err != nil {
        return out, err
    }
    if !ok {
        _ = s.incrTempLoginFailCount(ctx, in.LoginToken)
        if remain <= 0 {
            _ = s.deleteTempLoginState(ctx, in.LoginToken)
            return out, gerror.New("验证码错误，请重新输入账号密码")
        }
        return out, gerror.Newf("验证码错误，剩余 %d 次机会", remain)
    }

    _ = s.deleteSMSCode(ctx, state.UserID)
    _ = s.deleteTempLoginState(ctx, in.LoginToken)
    return s.completeLogin(ctx, state.UserID, in.LoginIP)
}
```

### 9.5 `internal/logic/admin/auth/me.go`

```go
// GetMe 获取当前用户信息与权限。
// 说明：权限不以 JWT 内数据为准，而是实时从缓存/数据库获取。
func (s *sAuth) GetMe(ctx context.Context, adminID int64) (out adminDto.LoginOutput, err error) {
    user, err := s.mustGetUserByID(ctx, adminID)
    if err != nil {
        return out, err
    }
    perms, err := s.getPermissions(ctx, user.Id, user.GroupId)
    if err != nil {
        return out, err
    }

    out.User = s.toLoginUserVO(user)
    out.Permissions = perms
    return out, nil
}
```

### 9.6 `internal/logic/admin/auth/logout.go`

```go
// Logout 删除当前 jti 对应会话。
func (s *sAuth) Logout(ctx context.Context, adminID int64, jti string) error {
    if err := s.deleteSession(ctx, jti); err != nil {
        return err
    }
    return s.removeUserSessionJTI(ctx, adminID, jti)
}
```

## 10. 中间件设计

### 10.1 `internal/middleware/auth_jwt.go`

```go
package middleware

import (
    "net/http"

    "github.com/gogf/gf/v2/net/ghttp"
    "project/internal/service"
)

// AuthJWT 统一执行后台登录态校验。
func AuthJWT(r *ghttp.Request) {
    handler := r.GetServeHandler()
    if handler != nil && handler.GetMetaTag("noAuth") == "true" {
        r.Middleware.Next()
        return
    }

    // 1. 从 Authorization: Bearer xxx 提取 token
    // 2. 解析 JWT，拿到 user_id / group_id / token_version / jti
    // 3. 校验 admin:session:{jti} 是否存在
    // 4. 校验用户是否存在、是否禁用、是否删除
    // 5. 校验 token_version 是否与数据库一致
    // 6. 将当前登录用户写入 request context

    if err := service.Permission().BindRequestAdminContext(r); err != nil {
        r.Response.WriteStatus(http.StatusUnauthorized)
        r.ExitAll()
        return
    }

    r.Middleware.Next()
}
```

### 10.2 `internal/middleware/auth_permission.go`

```go
package middleware

import (
    "net/http"

    "github.com/gogf/gf/v2/net/ghttp"
    "project/internal/service"
)

// AuthPermission 统一执行后台菜单权限校验。
func AuthPermission(r *ghttp.Request) {
    handler := r.GetServeHandler()
    if handler == nil || handler.GetMetaTag("noAuth") == "true" {
        r.Middleware.Next()
        return
    }

    if err := service.Permission().CheckRequestPermission(r); err != nil {
        r.Response.WriteStatus(http.StatusForbidden)
        r.ExitAll()
        return
    }

    r.Middleware.Next()
}
```

## 11. 权限服务骨架

### 11.1 规则

- `group_id=0` 直接放行
- `superOnly=true` 时仅允许 `group_id=0`
- `permission=""` 时只需登录，不需菜单权限
- 普通员工权限来源于缓存/数据库

### 11.2 `internal/service/permission.go`

```go
type IPermission interface {
    BindRequestAdminContext(r *ghttp.Request) error
    CheckRequestPermission(r *ghttp.Request) error
    GetPermissionCodes(ctx context.Context, adminID int64, groupID int64) ([]string, error)
    RefreshGroupPermissionCache(ctx context.Context, groupID int64) error
}
```

## 12. 路由注册

### 12.1 `internal/bootstrap/route.go`

```go
package bootstrap

import (
    "github.com/gogf/gf/v2/net/ghttp"

    authCtl "project/internal/controller/admin/auth"
    "project/internal/middleware"
)

func BindRoute(s *ghttp.Server) {
    s.Use(middleware.Recover, middleware.Response)

    s.Group("/api/admin", func(group *ghttp.RouterGroup) {
        group.Middleware(middleware.AuthJWT, middleware.AuthPermission)
        group.Bind(
            authCtl.New(),
        )
    })
}
```

> 这里统一把 `/api/admin` 都挂到鉴权链上，再通过 `noAuth` 元数据让登录相关接口跳过校验。

## 13. 关键辅助实现要求

### 13.1 `utility/bcryptx/bcryptx.go`

必须至少提供：

- `Hash(password string) (string, error)`
- `Verify(password, hash string) bool`

### 13.2 `utility/jwtx/jwtx.go`

必须至少提供：

- `NewAdminToken(userID, groupID, tokenVersion, expireAt)`
- `ParseAdminToken(tokenString)`
- `NewTempToken()`（用于 login_token）

### 13.3 `utility/ipx/client_ip.go`

必须考虑：

- `X-Forwarded-For`
- `X-Real-IP`
- 无代理时回落 remote addr

## 14. 登录日志写入

### 14.1 写入时机

以下两种情况必须写登录日志：

- 不需要短信验证，第一步直接登录成功
- 需要短信验证，第二步验证通过后登录成功

### 14.2 写入字段

- `admin_id`
- `admin_name`
- `ip`
- `ip_region`
- `created_at`

### 14.3 IP 地区解析

本期只保留解析接口与调用点，真正离线库接入在 Phase 4 补齐。

## 15. 本期测试清单

### 15.1 正向

- 超管正常登录
- 非首次登录且 IP 不变直接登录
- 首次登录触发短信验证
- IP 变化触发短信验证
- `/me` 返回权限集
- `/logout` 成功

### 15.2 反向

- 用户名不存在
- 密码错误
- 账号已禁用
- 账号已删除
- 手机号为空
- login_token 过期
- 验证码错误
- 错误次数超过 5 次
- token 过期
- token 被删除
- token_version 不一致

## 16. 本期完成标准

- 所有 Phase 1 接口均可访问
- 登录链路闭环
- 会话失效机制可用
- 中间件可正确区分 401 / 403
- 登录日志可写入数据库
- 代码中不存在 500+ 行的 `auth.go`
