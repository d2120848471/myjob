# Phase 4｜短信配置与审计日志

## 1. 本期目标

本期补齐：

- 阿里云短信配置维护
- 操作日志列表查询
- 登录日志列表查询
- 异步操作日志写入 worker
- IP 归属地离线解析接入

本期结束后，整个登录与权限系统应达到“可配置、可追踪、可审计”的生产可交付状态。

## 2. 本期范围

### 2.1 接口范围

- `GET /api/admin/config/sms`
- `PUT /api/admin/config/sms`
- `GET /api/admin/log/operation`
- `GET /api/admin/log/login`

### 2.2 表范围

- `system_config`
- `admin_operation_log`
- `admin_login_log`
- `admin_user`（人员筛选辅助）

### 2.3 资源范围

- `resource/ipdb/ip_region.xdb`

## 3. 目录新增清单

```text
api/admin/config/v1/
├── sms_get.go
└── sms_save.go

api/admin/log/v1/
├── operation_list.go
└── login_list.go

internal/controller/admin/config/
├── config.go
├── sms_get.go
└── sms_save.go

internal/controller/admin/log/
├── log.go
├── operation_list.go
└── login_list.go

internal/service/
├── config.go
└── log.go

internal/logic/admin/config/
├── config.go
├── sms_get.go
└── sms_save.go

internal/logic/admin/log/
├── operation.go
├── operation_write.go
├── operation_list.go
├── login.go
├── login_write.go
└── login_list.go

internal/bootstrap/
└── op_log_worker.go

utility/ipx/
└── region.go
```

## 4. 短信配置接口设计

### 4.1 获取短信配置 `GET /api/admin/config/sms`

权限：

- `superOnly:"true"`

返回：

```json
{
  "code": 0,
  "data": {
    "access_key": "LTAI****",
    "access_key_secret": "chxU****",
    "sign_name": "玖权益",
    "template_code": "SMS_308586082",
    "expire_minutes": 30,
    "interval_minutes": 1
  }
}
```

### 4.2 保存短信配置 `PUT /api/admin/config/sms`

权限：

- `superOnly:"true"`

入参：

```go
type SMSConfigSaveReq struct {
    g.Meta `path:"/config/sms" method:"put" tags:"系统设置" summary:"保存短信配置" superOnly:"true"`
    AccessKey         string `json:"access_key" v:"required#请输入AccessKey"`
    AccessKeySecret   string `json:"access_key_secret" v:"required#请输入AccessKeySecret"`
    SignName          string `json:"sign_name" v:"required#请输入签名"`
    TemplateCode      string `json:"template_code" v:"required#请输入模板编号"`
    ExpireMinutes     int    `json:"expire_minutes" v:"required|between:1,60#请输入有效期|有效期范围1~60分钟"`
    IntervalMinutes   int    `json:"interval_minutes" v:"required|between:1,10#请输入发送间隔|发送间隔范围1~10分钟"`
}
```

## 5. Config Service 设计

### 5.1 `internal/service/config.go`

```go
type IConfig interface {
    GetSMSConfig(ctx context.Context) (out adminDto.SMSConfigOutput, err error)
    SaveSMSConfig(ctx context.Context, in adminDto.SMSConfigInput, operator contextx.AdminContext) error
}
```

### 5.2 DTO 设计

```go
type SMSConfigInput struct {
    AccessKey       string
    AccessKeySecret string
    SignName        string
    TemplateCode    string
    ExpireMinutes   int
    IntervalMinutes int
}

type SMSConfigOutput struct {
    AccessKey       string
    AccessKeySecret string
    SignName        string
    TemplateCode    string
    ExpireMinutes   int
    IntervalMinutes int
}
```

## 6. 短信配置业务规则落实

### 6.1 获取配置

处理顺序：

1. 先读缓存 `admin:config:sms`
2. miss 后从 `system_config` 批量读取
3. 拼装 DTO
4. 对 `access_key_secret` 做脱敏
5. 回写缓存

### 6.2 保存配置

处理顺序：

1. 参数校验
2. 事务 upsert 6 个 key
3. 删除 `admin:config:sms` 缓存
4. 写操作日志：只写配置项名称，不写真实密钥

### 6.3 脱敏规则

- `access_key_secret`：前 4 位 + `****`
- `access_key` 也建议脱敏返回，避免完整暴露

### 6.4 保存逻辑骨架

```go
func (s *sConfig) SaveSMSConfig(ctx context.Context, in adminDto.SMSConfigInput, operator contextx.AdminContext) error {
    kvs := map[string]string{
        "sms_access_key":         in.AccessKey,
        "sms_access_key_secret":  in.AccessKeySecret,
        "sms_sign_name":          in.SignName,
        "sms_template_code":      in.TemplateCode,
        "sms_expire_minutes":     gconv.String(in.ExpireMinutes),
        "sms_interval_minutes":   gconv.String(in.IntervalMinutes),
    }

    err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        for key, value := range kvs {
            _, err := dao.SystemConfig.Ctx(ctx).TX(tx).Insert(
                g.Map{
                    "config_key":   key,
                    "config_value": value,
                    "description":  s.configDescription(key),
                },
                "config_key",
            )
            if err != nil {
                return err
            }
        }
        return nil
    })
    if err != nil {
        return err
    }

    _ = g.Redis().Del(ctx, consts.CacheKeyAdminSMSConfig)
    _ = service.Log().WriteOperation(ctx, operator,
        fmt.Sprintf("更新短信配置：签名=%s，模板=%s，有效期=%d，间隔=%d",
            in.SignName, in.TemplateCode, in.ExpireMinutes, in.IntervalMinutes,
        ),
    )
    return nil
}
```

> 注：实际 upsert SQL 由 Codex 按 GoFrame ORM 的最终写法调整，但要求必须保证 6 个配置项原子提交。

## 7. 操作日志写入模型

## 7.1 设计目标

需求要求“操作日志异步写入，不阻塞业务请求”，因此本期必须把之前埋下的 `WriteOperation()` 调用真正接到异步 worker 上。

## 7.2 推荐实现

### 7.2.1 结构

```text
业务 logic
  ↓
service.Log().WriteOperation(...)
  ↓
写入 buffered channel
  ↓
后台 worker 批量 / 单条落库
```

### 7.2.2 降级策略

当 channel 已满时：

- 不丢日志
- 降级为同步落库
- 同时打 warning 日志

这样既满足“尽量异步”，也避免审计日志静默丢失。

## 7.3 `internal/bootstrap/op_log_worker.go`

```go
package bootstrap

type OperationLogJob struct {
    AdminID     int64
    AdminName   string
    Description string
    IP          string
}

var opLogChan chan OperationLogJob

func InitOperationLogWorker(ctx context.Context) {
    opLogChan = make(chan OperationLogJob, 1000)

    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case job := <-opLogChan:
                if err := service.Log().PersistOperation(ctx, adminDto.OperationLogInput{
                    AdminID:     job.AdminID,
                    AdminName:   job.AdminName,
                    Description: job.Description,
                    IP:          job.IP,
                }); err != nil {
                    g.Log().Warning(ctx, "persist operation log failed:", err)
                }
            }
        }
    }()
}
```

## 7.4 `internal/service/log.go`

```go
type ILog interface {
    // WriteOperation 写操作日志，优先异步入队。
    WriteOperation(ctx context.Context, operator contextx.AdminContext, description string) error

    // PersistOperation 真正执行落库，仅 worker 或降级同步路径调用。
    PersistOperation(ctx context.Context, in adminDto.OperationLogInput) error

    // WriteLogin 写登录日志。
    WriteLogin(ctx context.Context, in adminDto.LoginLogInput) error

    // ListOperation 查询操作日志。
    ListOperation(ctx context.Context, in adminDto.OperationLogListInput) (out adminDto.OperationLogListOutput, err error)

    // ListLogin 查询登录日志。
    ListLogin(ctx context.Context, in adminDto.LoginLogListInput) (out adminDto.LoginLogListOutput, err error)
}
```

### 7.4.1 `WriteOperation` 逻辑骨架

```go
func (s *sLog) WriteOperation(ctx context.Context, operator contextx.AdminContext, description string) error {
    job := bootstrap.OperationLogJob{
        AdminID:     operator.AdminID,
        AdminName:   operator.RealName,
        Description: description,
        IP:          operator.IP,
    }

    select {
    case bootstrap.OperationLogChan() <- job:
        return nil
    default:
        // 队列满时不丢日志，降级同步写库。
        return s.PersistOperation(ctx, adminDto.OperationLogInput{
            AdminID:     job.AdminID,
            AdminName:   job.AdminName,
            Description: job.Description,
            IP:          job.IP,
        })
    }
}
```

## 8. 操作日志列表查询

### 8.1 接口对象

```go
type OperationLogListReq struct {
    g.Meta `path:"/log/operation" method:"get" tags:"审计日志" summary:"操作日志列表" permission:"admin.action"`
    AdminID   int64  `p:"admin_id"`
    Keyword   string `p:"keyword"`
    StartTime string `p:"start_time"`
    EndTime   string `p:"end_time"`
    Page      int    `p:"page" d:"1" v:"min:1#页码最小为1"`
    PageSize  int    `p:"page_size" d:"20" v:"between:1,100#每页数量范围错误"`
}
```

### 8.2 查询规则

- 默认 `created_at DESC`
- `admin_id` 精确过滤
- `keyword` 按 `description LIKE`
- 时间范围用 `created_at >= start` 和 `created_at <= end`

### 8.3 查询逻辑骨架

```go
func (s *sLog) ListOperation(ctx context.Context, in adminDto.OperationLogListInput) (out adminDto.OperationLogListOutput, err error) {
    model := dao.AdminOperationLog.Ctx(ctx)

    if in.AdminID > 0 {
        model = model.Where("admin_id", in.AdminID)
    }
    if in.Keyword != "" {
        model = model.WhereLike("description", "%"+in.Keyword+"%")
    }
    if in.StartTime != "" {
        model = model.WhereGTE("created_at", in.StartTime)
    }
    if in.EndTime != "" {
        model = model.WhereLTE("created_at", in.EndTime)
    }

    total, err := model.Count()
    if err != nil {
        return out, err
    }

    list, err := model.Order("created_at desc").
        Page(in.Page, in.PageSize).
        All()
    if err != nil {
        return out, err
    }

    // 转 DTO 输出
    return out, nil
}
```

## 9. 登录日志列表查询

### 9.1 接口对象

```go
type LoginLogListReq struct {
    g.Meta `path:"/log/login" method:"get" tags:"审计日志" summary:"登录日志列表" permission:"admin.loginlog"`
    AdminID   int64  `p:"admin_id"`
    StartTime string `p:"start_time"`
    EndTime   string `p:"end_time"`
    Page      int    `p:"page" d:"1" v:"min:1#页码最小为1"`
    PageSize  int    `p:"page_size" d:"20" v:"between:1,100#每页数量范围错误"`
}
```

### 9.2 查询规则

- 默认 `created_at DESC`
- `admin_id` 精确过滤
- 支持开始时间 / 结束时间范围

## 10. IP 归属地离线解析接入

### 10.1 设计要求

需求明确要求离线 IP 库解析，解析失败时允许留空。  
因此实现必须满足：

- 启动时加载离线 IP 数据文件
- 查询失败不影响主流程
- 登录日志写入时尽量补 `ip_region`
- 解析失败时 `ip_region=""`

### 10.2 推荐抽象

```go
type RegionResolver interface {
    Resolve(ip string) (string, error)
}
```

### 10.3 `utility/ipx/region.go`

```go
package ipx

type RegionResolver interface {
    Resolve(ip string) (string, error)
}

type EmptyResolver struct{}

func (r *EmptyResolver) Resolve(ip string) (string, error) {
    return "", nil
}
```

> Codex 落地时，把真实离线 IP 库实现放到 `resource/ipdb/ip_region.xdb` 读取逻辑中。  
> 这里先要求保留接口与调用点，不把第三方库调用硬编码到登录 logic 里。

## 11. 登录日志写入补齐

Phase 1 已埋写入点，本期要求正式补齐：

1. 登录成功时调用 `WriteLogin()`
2. 由 `WriteLogin()` 执行同步落库
3. 调 `RegionResolver.Resolve(ip)` 获取地区
4. 解析失败时继续写空字符串

### 11.1 写登录日志骨架

```go
func (s *sLog) WriteLogin(ctx context.Context, in adminDto.LoginLogInput) error {
    region, _ := s.regionResolver.Resolve(in.IP)

    _, err := dao.AdminLoginLog.Ctx(ctx).Insert(do.AdminLoginLog{
        AdminId:   in.AdminID,
        AdminName: in.AdminName,
        Ip:        in.IP,
        IpRegion:  region,
    })
    return err
}
```

## 12. 短信发送链路与配置联动收口

本期要把 Phase 1 中的 `getSMSConfig()` 真正接到配置服务上：

- 优先读 Redis 缓存
- 缓存 miss 回源 `system_config`
- 若必需配置为空：
  - 返回 `短信配置未完成，请联系超级管理员`
- `sms_expire_minutes` 和 `sms_interval_minutes` 做 string -> int 转换并校验

## 13. 本期测试清单

### 13.1 短信配置

- 超管获取短信配置成功
- 普通员工获取短信配置失败
- 超管保存短信配置成功
- 配置项为空失败
- 有效期超范围失败
- 间隔超范围失败
- `access_key_secret` 返回脱敏

### 13.2 操作日志

- 写员工新增操作日志成功
- 写员工删除操作日志成功
- 队列正常入队成功
- 队列满时同步降级成功
- 按人员筛选成功
- 按关键词筛选成功
- 按时间范围筛选成功

### 13.3 登录日志

- 登录成功写登录日志
- IP 地区解析成功
- IP 地区解析失败不影响登录
- 按人员/时间范围筛选成功

## 14. 本期完成标准

- 短信配置接口可用
- 操作日志列表可用
- 登录日志列表可用
- 异步操作日志 worker 可用
- 登录 IP 地区解析接入点完成
- AccessKeySecret 不会出现在明文响应和操作日志中
