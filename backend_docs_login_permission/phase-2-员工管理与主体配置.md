# Phase 2｜员工管理与主体配置

## 1. 本期目标

本期要把“员工管理 3 个 Tab”中的前两个核心域补齐：

1. 管理员列表
2. 回收站
3. 主体配置

本期结束后必须具备：

- 员工列表分页
- 添加员工
- 编辑员工
- 删除员工（软删除）
- 状态切换
- 余额阈值通知切换
- 批量设置/取消商务
- 回收站列表
- 恢复员工
- 主体配置列表/新增/编辑

## 2. 本期范围

### 2.1 接口范围

#### 员工管理

- `GET /api/admin/user/list`
- `POST /api/admin/user/add`
- `PUT /api/admin/user/:id`
- `DELETE /api/admin/user/:id`
- `PUT /api/admin/user/:id/status`
- `PUT /api/admin/user/:id/notify`
- `POST /api/admin/user/setBusiness`
- `POST /api/admin/user/cancelBusiness`
- `GET /api/admin/user/trash`
- `PUT /api/admin/user/:id/restore`

#### 主体配置

- `GET /api/admin/subject/list`
- `POST /api/admin/subject/add`
- `PUT /api/admin/subject/:id`

### 2.2 表范围

- `admin_user`
- `admin_subject`
- `admin_group`
- `admin_operation_log`（先写调用点，异步化在 Phase 4 完整收口）

## 3. 需求冲突修正规则

需求页面字段明确要求添加/编辑员工必须带手机号，但接口示例遗漏了 `phone`。  
本期按以下接口入参落地，不再沿用遗漏版示例：

### 3.1 添加员工

```json
{
  "username": "admin02",
  "password": "Admin_1",
  "real_name": "张三",
  "phone": "13800001111",
  "group_id": 2
}
```

### 3.2 编辑员工

```json
{
  "password": "Admin_2",
  "real_name": "李四",
  "phone": "13800002222",
  "group_id": 3
}
```

## 4. 目录新增清单

```text
api/admin/user/v1/
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── notify.go
├── set_business.go
├── cancel_business.go
├── trash.go
└── restore.go

api/admin/subject/v1/
├── list.go
├── add.go
└── edit.go

internal/controller/admin/user/
├── user.go
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── notify.go
├── set_business.go
├── cancel_business.go
├── trash.go
└── restore.go

internal/controller/admin/subject/
├── subject.go
├── list.go
├── add.go
└── edit.go

internal/service/
├── user.go
└── subject.go

internal/logic/admin/user/
├── user.go
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── notify.go
├── batch_business.go
├── trash.go
└── restore.go

internal/logic/admin/subject/
├── subject.go
├── list.go
├── add.go
└── edit.go
```

## 5. API 对象定义

### 5.1 员工列表 `api/admin/user/v1/list.go`

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

type ListReq struct {
    g.Meta `path:"/user/list" method:"get" tags:"员工管理" summary:"员工列表" permission:"admin.list"`
    Page     int `p:"page" d:"1" v:"min:1#页码最小为1"`
    PageSize int `p:"page_size" d:"20" v:"between:1,100#每页数量范围错误"`
}

type ListItem struct {
    ID            int64  `json:"id"`
    Username      string `json:"username"`
    RealName      string `json:"real_name"`
    IsBusiness    int    `json:"is_business"`
    GroupID       int64  `json:"group_id"`
    GroupName     string `json:"group_name"`
    Status        int    `json:"status"`
    BalanceNotify int    `json:"balance_notify"`
    Phone         string `json:"phone"`
}

type ListRes struct {
    List       []*ListItem `json:"list"`
    Pagination struct {
        Page     int `json:"page"`
        PageSize int `json:"page_size"`
        Total    int `json:"total"`
    } `json:"pagination"`
}
```

### 5.2 添加员工 `api/admin/user/v1/add.go`

```go
type AddReq struct {
    g.Meta `path:"/user/add" method:"post" tags:"员工管理" summary:"添加员工" permission:"admin.list"`
    Username        string `json:"username" v:"required|regex:^[A-Za-z][A-Za-z0-9]{5,9}$#请输入用户名|用户名格式错误"`
    ConfirmUsername string `json:"confirm_username" v:"required#请重复输入用户名"`
    Password        string `json:"password" v:"required|regex:^[A-Za-z][A-Za-z0-9_]{5,9}$#请输入密码|密码格式错误"`
    ConfirmPassword string `json:"confirm_password" v:"required#请重复输入密码"`
    RealName        string `json:"real_name" v:"required#请输入真实姓名"`
    Phone           string `json:"phone" v:"required|regex:^1\d{10}$#请输入手机号|手机号格式错误"`
    GroupID         int64  `json:"group_id" v:"required|min:1#请选择用户组|用户组错误"`
}
```

### 5.3 编辑员工 `api/admin/user/v1/edit.go`

```go
type EditReq struct {
    g.Meta `path:"/user/{id}" method:"put" tags:"员工管理" summary:"编辑员工" permission:"admin.list"`
    ID              int64  `p:"id" v:"required|min:1#员工ID错误"`
    Password        string `json:"password" v:"regex:^$|^[A-Za-z][A-Za-z0-9_]{5,9}$#密码格式错误"`
    ConfirmPassword string `json:"confirm_password"`
    RealName        string `json:"real_name" v:"required#请输入真实姓名"`
    Phone           string `json:"phone" v:"required|regex:^1\d{10}$#请输入手机号|手机号格式错误"`
    GroupID         int64  `json:"group_id" v:"required|min:1#请选择用户组|用户组错误"`
}
```

### 5.4 删除、状态、通知、恢复

```go
type DeleteReq struct {
    g.Meta `path:"/user/{id}" method:"delete" tags:"员工管理" summary:"删除员工" permission:"admin.list"`
    ID int64 `p:"id" v:"required|min:1#员工ID错误"`
}

type ChangeStatusReq struct {
    g.Meta `path:"/user/{id}/status" method:"put" tags:"员工管理" summary:"切换员工状态" permission:"admin.list"`
    ID     int64 `p:"id" v:"required|min:1#员工ID错误"`
    Status int   `json:"status" v:"required|in:0,1#状态不能为空|状态错误"`
}

type ChangeNotifyReq struct {
    g.Meta `path:"/user/{id}/notify" method:"put" tags:"员工管理" summary:"切换余额阈值通知" permission:"admin.list"`
    ID            int64 `p:"id" v:"required|min:1#员工ID错误"`
    BalanceNotify int   `json:"balance_notify" v:"required|in:0,1#通知状态不能为空|通知状态错误"`
}

type RestoreReq struct {
    g.Meta `path:"/user/{id}/restore" method:"put" tags:"员工管理" summary:"恢复员工" permission:"admin.list"`
    ID int64 `p:"id" v:"required|min:1#员工ID错误"`
}
```

### 5.5 批量商务标记

```go
type SetBusinessReq struct {
    g.Meta `path:"/user/setBusiness" method:"post" tags:"员工管理" summary:"批量设置商务" permission:"admin.list"`
    IDs []int64 `json:"ids" v:"required#请选择员工"`
}

type CancelBusinessReq struct {
    g.Meta `path:"/user/cancelBusiness" method:"post" tags:"员工管理" summary:"批量取消商务" permission:"admin.list"`
    IDs []int64 `json:"ids" v:"required#请选择员工"`
}
```

### 5.6 主体配置

```go
type SubjectListReq struct {
    g.Meta `path:"/subject/list" method:"get" tags:"主体配置" summary:"主体列表" permission:"admin.list"`
}

type SubjectAddReq struct {
    g.Meta `path:"/subject/add" method:"post" tags:"主体配置" summary:"添加主体" permission:"admin.list"`
    Name   string `json:"name" v:"required#请输入主体名称"`
    HasTax int    `json:"has_tax" v:"required|in:0,1#请选择含税状态|含税状态错误"`
}

type SubjectEditReq struct {
    g.Meta `path:"/subject/{id}" method:"put" tags:"主体配置" summary:"编辑主体" permission:"admin.list"`
    ID     int64  `p:"id" v:"required|min:1#主体ID错误"`
    Name   string `json:"name" v:"required#请输入主体名称"`
    HasTax int    `json:"has_tax" v:"required|in:0,1#请选择含税状态|含税状态错误"`
}
```

## 6. User Service 设计

### 6.1 `internal/service/user.go`

```go
type IUser interface {
    List(ctx context.Context, in adminDto.UserListInput) (out adminDto.UserListOutput, err error)
    Add(ctx context.Context, in adminDto.UserAddInput) error
    Edit(ctx context.Context, in adminDto.UserEditInput) error
    Delete(ctx context.Context, userID int64, operator contextx.AdminContext) error
    ChangeStatus(ctx context.Context, userID int64, status int, operator contextx.AdminContext) error
    ChangeNotify(ctx context.Context, userID int64, notify int, operator contextx.AdminContext) error
    SetBusiness(ctx context.Context, ids []int64, operator contextx.AdminContext) error
    CancelBusiness(ctx context.Context, ids []int64, operator contextx.AdminContext) error
    Trash(ctx context.Context, in adminDto.UserTrashInput) (out adminDto.UserTrashOutput, err error)
    Restore(ctx context.Context, userID int64, operator contextx.AdminContext) error
}
```

### 6.2 `internal/service/subject.go`

```go
type ISubject interface {
    List(ctx context.Context) (out adminDto.SubjectListOutput, err error)
    Add(ctx context.Context, in adminDto.SubjectAddInput, operator contextx.AdminContext) error
    Edit(ctx context.Context, in adminDto.SubjectEditInput, operator contextx.AdminContext) error
}
```

## 7. 核心业务规则落实

## 7.1 员工列表

查询条件：

- `is_deleted = 0`
- 默认按 `id DESC`
- 分页默认 20

返回字段：

- `group_name` 通过 `admin_group` 左联或二次批量映射获取
- 超管 `group_id=0` 时，`group_name` 返回空字符串或 `超级管理员`
- `id=1` 删除按钮由前端控制，但后端也必须做保护

## 7.2 添加员工

处理顺序：

1. `username == confirm_username`
2. `password == confirm_password`
3. 校验用户名唯一
4. 校验用户组存在且状态正常
5. 密码 bcrypt 加密
6. 插入 `admin_user`
7. 写操作日志：`添加员工：{username}，用户组：{groupName}`

### 7.2.1 添加员工逻辑骨架

```go
// Add 添加员工。
func (s *sUser) Add(ctx context.Context, in adminDto.UserAddInput) error {
    if in.Username != in.ConfirmUsername {
        return gerror.New("两次输入的用户名不一致")
    }
    if in.Password != in.ConfirmPassword {
        return gerror.New("两次输入的密码不一致")
    }

    exists, err := s.existsActiveUsername(ctx, in.Username)
    if err != nil {
        return err
    }
    if exists {
        return gerror.New("用户名已存在")
    }

    group, err := s.mustGetEnabledGroup(ctx, in.GroupID)
    if err != nil {
        return err
    }

    hash, err := bcryptx.Hash(in.Password)
    if err != nil {
        return err
    }

    _, err = dao.AdminUser.Ctx(ctx).Insert(do.AdminUser{
        Username:      in.Username,
        PasswordHash:  hash,
        RealName:      in.RealName,
        Phone:         in.Phone,
        GroupId:       in.GroupID,
        Status:        1,
        BalanceNotify: 0,
        IsBusiness:    0,
        IsDeleted:     0,
        TokenVersion:  0,
    })
    if err != nil {
        return err
    }

    // 这里只调用操作日志 service，具体异步落库在 Phase 4 完成。
    _ = service.Log().WriteOperation(ctx, contextx.GetAdmin(ctx),
        fmt.Sprintf("添加员工：%s，用户组：%s", in.Username, group.Name),
    )
    return nil
}
```

## 7.3 编辑员工

规则：

- 用户名只读，不允许改
- 密码为空 → 不修改
- 密码不为空 → 必须和确认密码一致，并重新 bcrypt
- 手机号必填
- 用户组必须存在且状态正常

### 7.3.1 编辑员工逻辑骨架

```go
// Edit 编辑员工。
func (s *sUser) Edit(ctx context.Context, in adminDto.UserEditInput) error {
    user, err := s.mustGetActiveUser(ctx, in.ID)
    if err != nil {
        return err
    }
    if user.Id == 1 && in.GroupID == 0 {
        // 超管仍允许改姓名/手机号，但不允许把超级管理员做普通组迁移。
    }

    updateData := do.AdminUser{
        RealName: in.RealName,
        Phone:    in.Phone,
        GroupId:  in.GroupID,
    }

    if in.Password != "" {
        if in.Password != in.ConfirmPassword {
            return gerror.New("两次输入的密码不一致")
        }
        hash, err := bcryptx.Hash(in.Password)
        if err != nil {
            return err
        }
        updateData.PasswordHash = hash
        // 改密码后建议 token_version+1，强制旧登录态失效。
        updateData.TokenVersion = user.TokenVersion + 1
    }

    _, err = dao.AdminUser.Ctx(ctx).
        WherePri(in.ID).
        Where("is_deleted", 0).
        Update(updateData)
    if err != nil {
        return err
    }

    if in.Password != "" {
        _ = s.invalidateAllUserSessions(ctx, in.ID)
    }
    return nil
}
```

## 7.4 删除员工（软删除）

规则：

- `id=1` 绝对禁止删除
- 事务内完成：
  - 查询员工
  - 生成新用户名 `原用户名_时间戳`
  - `is_deleted=1`
  - `status=0`
  - `deleted_at=NOW()`
  - `token_version=token_version+1`
- 事务提交后清理全部 session
- 写操作日志

### 7.4.1 删除逻辑骨架

```go
// Delete 软删除员工。
func (s *sUser) Delete(ctx context.Context, userID int64, operator contextx.AdminContext) error {
    user, err := s.mustGetActiveUser(ctx, userID)
    if err != nil {
        return err
    }
    if user.Id == 1 {
        return gerror.New("超级管理员不允许删除")
    }

    deletedUsername := fmt.Sprintf("%s_%d", user.Username, gtime.Timestamp())

    err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        _, err = dao.AdminUser.Ctx(ctx).TX(tx).
            WherePri(userID).
            Where("is_deleted", 0).
            Update(do.AdminUser{
                Username:     deletedUsername,
                Status:       0,
                IsDeleted:    1,
                DeletedAt:    gtime.Now(),
                TokenVersion: user.TokenVersion + 1,
            })
        return err
    })
    if err != nil {
        return err
    }

    _ = s.invalidateAllUserSessions(ctx, userID)
    _ = service.Log().WriteOperation(ctx, operator, fmt.Sprintf("删除员工：%s", user.Username))
    return nil
}
```

## 7.5 切换员工状态

规则：

- `id=1` 不允许切状态
- 正常/禁用即时生效
- 禁用后递增 `token_version` 并清理全部 session
- 启用不需要签发新 token，用户重新登录即可

## 7.6 切换余额阈值通知

规则：

- 仅更新 `balance_notify`
- 记录操作日志

## 7.7 批量设置/取消商务

规则：

- `ids` 不得为空
- 对不存在/已删除用户过滤
- 超管可参与批量标记，但不建议前端默认勾选
- 一次 SQL 批量更新，不逐条循环写库

### 7.7.1 批量逻辑骨架

```go
func (s *sUser) SetBusiness(ctx context.Context, ids []int64, operator contextx.AdminContext) error {
    if len(ids) == 0 {
        return gerror.New("请选择员工")
    }
    _, err := dao.AdminUser.Ctx(ctx).
        WhereIn("id", ids).
        Where("is_deleted", 0).
        Update(do.AdminUser{IsBusiness: 1})
    if err != nil {
        return err
    }
    _ = service.Log().WriteOperation(ctx, operator, fmt.Sprintf("批量设置商务，员工ID：%v", ids))
    return nil
}
```

## 7.8 回收站列表

查询条件：

- `is_deleted = 1`
- 默认倒序
- 仅返回 `id / username / real_name`

## 7.9 恢复员工

规则：

1. 仅允许恢复 `is_deleted=1`
2. 将用户名从 `原用户名_时间戳` 还原
3. 若还原后用户名已被占用，则返回冲突
4. 恢复后：
   - `is_deleted=0`
   - `status=0`
   - `deleted_at=NULL`
   - `username=原用户名`
5. 恢复完成后不恢复旧 token；用户需重新登录

### 7.9.1 解析原用户名

由于需求已限定用户名只能“字母开头 + 字母数字”，**不允许下划线**，因此删除时追加 `_时间戳` 后，恢复时按最后一个 `_` 可靠拆分即可。

### 7.9.2 恢复逻辑骨架

```go
func (s *sUser) Restore(ctx context.Context, userID int64, operator contextx.AdminContext) error {
    user, err := s.mustGetDeletedUser(ctx, userID)
    if err != nil {
        return err
    }

    originUsername, err := s.parseOriginUsername(user.Username)
    if err != nil {
        return err
    }

    exists, err := s.existsActiveUsername(ctx, originUsername)
    if err != nil {
        return err
    }
    if exists {
        return gerror.New("用户名已被占用，无法恢复")
    }

    _, err = dao.AdminUser.Ctx(ctx).
        WherePri(userID).
        Where("is_deleted", 1).
        Update(do.AdminUser{
            Username:  originUsername,
            IsDeleted: 0,
            Status:    0,
            DeletedAt: nil,
        })
    if err != nil {
        return err
    }

    _ = service.Log().WriteOperation(ctx, operator, fmt.Sprintf("恢复员工：%s", originUsername))
    return nil
}
```

## 8. 主体配置实现

### 8.1 列表

- 直接查询全部
- 按 `id DESC` 或 `id ASC` 均可，建议 `id DESC`

### 8.2 添加主体

- 名称非空
- 名称唯一
- `has_tax in (0,1)`

### 8.3 编辑主体

- 主体必须存在
- 修改后名称仍需唯一
- 不提供删除接口

### 8.4 主体逻辑骨架

```go
// Add 添加主体。
func (s *sSubject) Add(ctx context.Context, in adminDto.SubjectAddInput, operator contextx.AdminContext) error {
    exists, err := s.existsName(ctx, in.Name, 0)
    if err != nil {
        return err
    }
    if exists {
        return gerror.New("主体名称已存在")
    }

    _, err = dao.AdminSubject.Ctx(ctx).Insert(do.AdminSubject{
        Name:   in.Name,
        HasTax: in.HasTax,
    })
    if err != nil {
        return err
    }

    _ = service.Log().WriteOperation(ctx, operator,
        fmt.Sprintf("添加主体：%s，是否含税：%d", in.Name, in.HasTax),
    )
    return nil
}
```

## 9. 权限元数据映射

本期接口统一使用 `permission:"admin.list"`，原因：

- 员工管理页面主权限码对应 `admin.list`
- 主体配置属于员工管理页面下的附属 Tab
- 当前需求中没有更细的主体配置独立权限码

## 10. 本期测试清单

### 10.1 员工管理

- 新增员工成功
- 用户名重复失败
- 两次用户名不一致失败
- 两次密码不一致失败
- 用户组不存在失败
- 编辑员工不填密码不改密码
- 编辑员工填密码后旧 token 失效
- 删除员工进入回收站
- 删除超管失败
- 禁用超管失败
- 恢复员工成功
- 恢复时用户名冲突失败
- 切换通知成功
- 批量商务成功

### 10.2 主体配置

- 新增主体成功
- 主体名称重复失败
- 编辑主体成功
- 主体不存在失败

## 11. 本期完成标准

- 员工三类状态：正常、禁用、已删除 可正确流转
- 删除与恢复闭环
- 主体配置闭环
- 所有写操作都已接入操作日志调用点
- 不存在一个 800 行的 `user.go`
