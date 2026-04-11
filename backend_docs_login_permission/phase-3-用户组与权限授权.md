# Phase 3｜用户组与权限授权

## 1. 本期目标

本期完成“部门管理（用户组 & 授权）”闭环，让系统真正具备**按用户组控制后台页面访问**的能力。

本期结束后必须具备：

- 用户组列表
- 添加/编辑/删除用户组
- 用户组状态切换
- 获取完整权限菜单树
- 获取用户组当前授权
- 保存用户组授权
- 授权变更后普通员工下次请求自动生效

## 2. 本期范围

### 2.1 接口范围

- `GET /api/admin/group/list`
- `POST /api/admin/group/add`
- `PUT /api/admin/group/:id`
- `DELETE /api/admin/group/:id`
- `PUT /api/admin/group/:id/status`
- `GET /api/admin/menu/tree`
- `GET /api/admin/group/:id/auth`
- `PUT /api/admin/group/:id/auth`

### 2.2 表范围

- `admin_group`
- `admin_menu`
- `admin_group_menu`
- `admin_user`
- `admin_operation_log`

## 3. 目录新增清单

```text
api/admin/group/v1/
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── auth_get.go
└── auth_save.go

api/admin/menu/v1/
└── tree.go

internal/controller/admin/group/
├── group.go
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── auth_get.go
└── auth_save.go

internal/controller/admin/menu/
├── menu.go
└── tree.go

internal/service/
├── group.go
├── menu.go
└── permission.go

internal/logic/admin/group/
├── group.go
├── list.go
├── add.go
├── edit.go
├── delete.go
├── status.go
├── auth_get.go
└── auth_save.go

internal/logic/admin/menu/
├── menu.go
└── tree.go

internal/logic/admin/permission/
├── permission.go
├── cache.go
└── refresh.go
```

## 4. API 对象定义

### 4.1 用户组列表

```go
type GroupListReq struct {
    g.Meta `path:"/group/list" method:"get" tags:"用户组管理" summary:"用户组列表" permission:"admin.department"`
    Page     int `p:"page" d:"1" v:"min:1#页码最小为1"`
    PageSize int `p:"page_size" d:"20" v:"between:1,100#每页数量范围错误"`
}

type GroupListItem struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      int    `json:"status"`
    UserCount   int    `json:"user_count"`
}
```

### 4.2 添加 / 编辑用户组

```go
type GroupAddReq struct {
    g.Meta `path:"/group/add" method:"post" tags:"用户组管理" summary:"添加用户组" permission:"admin.department"`
    Name        string `json:"name" v:"required#请输入用户组名称"`
    Description string `json:"description"`
}

type GroupEditReq struct {
    g.Meta `path:"/group/{id}" method:"put" tags:"用户组管理" summary:"编辑用户组" permission:"admin.department"`
    ID          int64  `p:"id" v:"required|min:1#用户组ID错误"`
    Name        string `json:"name" v:"required#请输入用户组名称"`
    Description string `json:"description"`
}
```

### 4.3 删除 / 状态切换

```go
type GroupDeleteReq struct {
    g.Meta `path:"/group/{id}" method:"delete" tags:"用户组管理" summary:"删除用户组" permission:"admin.department"`
    ID int64 `p:"id" v:"required|min:1#用户组ID错误"`
}

type GroupStatusReq struct {
    g.Meta `path:"/group/{id}/status" method:"put" tags:"用户组管理" summary:"切换用户组状态" permission:"admin.department"`
    ID     int64 `p:"id" v:"required|min:1#用户组ID错误"`
    Status int   `json:"status" v:"required|in:0,1#状态不能为空|状态错误"`
}
```

### 4.4 权限树与授权

```go
type MenuTreeReq struct {
    g.Meta `path:"/menu/tree" method:"get" tags:"权限管理" summary:"获取完整权限菜单树" permission:"admin.department"`
}

type MenuNode struct {
    ID       int64       `json:"id"`
    Name     string      `json:"name"`
    Code     string      `json:"code"`
    Children []*MenuNode `json:"children"`
}

type GroupAuthGetReq struct {
    g.Meta `path:"/group/{id}/auth" method:"get" tags:"权限管理" summary:"获取用户组已授权菜单" permission:"admin.department"`
    ID int64 `p:"id" v:"required|min:1#用户组ID错误"`
}

type GroupAuthGetRes struct {
    MenuIDs []int64 `json:"menu_ids"`
}

type GroupAuthSaveReq struct {
    g.Meta `path:"/group/{id}/auth" method:"put" tags:"权限管理" summary:"保存用户组授权" permission:"admin.department"`
    ID      int64   `p:"id" v:"required|min:1#用户组ID错误"`
    MenuIDs []int64 `json:"menu_ids"`
}
```

## 5. Service 设计

### 5.1 `internal/service/group.go`

```go
type IGroup interface {
    List(ctx context.Context, in adminDto.GroupListInput) (out adminDto.GroupListOutput, err error)
    Add(ctx context.Context, in adminDto.GroupAddInput, operator contextx.AdminContext) error
    Edit(ctx context.Context, in adminDto.GroupEditInput, operator contextx.AdminContext) error
    Delete(ctx context.Context, groupID int64, operator contextx.AdminContext) error
    ChangeStatus(ctx context.Context, groupID int64, status int, operator contextx.AdminContext) error
    GetAuth(ctx context.Context, groupID int64) (menuIDs []int64, err error)
    SaveAuth(ctx context.Context, groupID int64, menuIDs []int64, operator contextx.AdminContext) error
}
```

### 5.2 `internal/service/menu.go`

```go
type IMenu interface {
    Tree(ctx context.Context) (nodes []*adminDto.MenuTreeNode, err error)
    GetAllEnabledMenuCodes(ctx context.Context) ([]string, error)
}
```

## 6. 用户组业务规则

## 6.1 列表

返回：

- `id`
- `name`
- `description`
- `status`
- `user_count`

`user_count` 建议统计该组全部员工数量，而不是只统计活跃员工。  
这样删除前展示更贴近真实挂载情况，也有利于排查恢复员工场景。

## 6.2 添加用户组

规则：

- 名称非空
- 名称唯一
- 默认 `status=1`
- 写操作日志：`添加用户组：{name}`

## 6.3 编辑用户组

规则：

- 用户组必须存在
- 新名称仍需唯一
- 写操作日志

## 6.4 删除用户组

规则：

- 如果该组下还有员工，不允许删除
- **建议统计该组全部员工（含回收站）**，避免删除后软删除员工恢复时 `group_id` 悬空
- 无员工时才允许删除
- 删除时同时删除 `admin_group_menu` 中该组授权
- 删除后清理权限缓存
- 写操作日志

### 6.4.1 删除逻辑骨架

```go
func (s *sGroup) Delete(ctx context.Context, groupID int64, operator contextx.AdminContext) error {
    group, err := s.mustGetGroup(ctx, groupID)
    if err != nil {
        return err
    }

    count, err := dao.AdminUser.Ctx(ctx).
        Where("group_id", groupID).
        Count()
    if err != nil {
        return err
    }
    if count > 0 {
        return gerror.Newf("该用户组下还有 %d 名员工，请先转移", count)
    }

    err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        if _, err = dao.AdminGroup.Ctx(ctx).TX(tx).WherePri(groupID).Delete(); err != nil {
            return err
        }
        if _, err = dao.AdminGroupMenu.Ctx(ctx).TX(tx).Where("group_id", groupID).Delete(); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return err
    }

    _ = service.Permission().RefreshGroupPermissionCache(ctx, groupID)
    _ = service.Log().WriteOperation(ctx, operator, fmt.Sprintf("删除用户组：%s", group.Name))
    return nil
}
```

## 7. 用户组状态切换

规则：

- 正常 ↔ 禁用
- 禁用用户组后：
  - 该组权限缓存要刷新
  - 组内员工下次请求时应被权限链识别为不可用
- 有两种落地方式：
  1. 中间件取权限时，额外校验用户组状态
  2. 或在绑定用户上下文时直接校验组状态

本项目建议采用 **中间件统一校验用户组状态**：

- 超管跳过
- 普通员工若所属组 `status=0` → 返回 403

## 8. 菜单树实现

### 8.1 数据来源

- 只读取 `admin_menu.status = 1`
- 排序规则：`sort ASC, id ASC`

### 8.2 树构建规则

- `parent_id = 0` 为一级模块
- 一级节点挂二级页面
- 二级节点再挂三级按钮
- 没有子节点的直接返回空数组 `[]`

### 8.3 树构建逻辑骨架

```go
func (s *sMenu) Tree(ctx context.Context) ([]*adminDto.MenuTreeNode, error) {
    list, err := dao.AdminMenu.Ctx(ctx).
        Where("status", 1).
        Order("sort asc, id asc").
        All()
    if err != nil {
        return nil, err
    }

    // 1. 全量转 node map[id]*MenuTreeNode
    // 2. 再按 parent_id 组装树
    // 3. 返回 parent_id=0 的节点数组
}
```

## 9. 获取用户组已授权菜单

实现：

- 查询 `admin_group_menu` where `group_id = ?`
- 返回 `menu_id` 数组
- 不需要返回权限码，由前端根据菜单树勾选

## 10. 保存授权

### 10.1 规则

1. 用户组必须存在
2. `menu_ids` 中所有 ID 必须存在于 `admin_menu`
3. 允许提交空数组，表示清空授权
4. 使用事务：
   - 删除旧授权
   - 批量插入新授权
5. 提交后清理 `admin:perm:group:{group_id}`
6. 写操作日志

### 10.2 保存授权逻辑骨架

```go
func (s *sGroup) SaveAuth(ctx context.Context, groupID int64, menuIDs []int64, operator contextx.AdminContext) error {
    if _, err := s.mustGetGroup(ctx, groupID); err != nil {
        return err
    }

    if err := s.validateMenuIDs(ctx, menuIDs); err != nil {
        return err
    }

    err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        if _, err := dao.AdminGroupMenu.Ctx(ctx).TX(tx).
            Where("group_id", groupID).
            Delete(); err != nil {
            return err
        }

        if len(menuIDs) == 0 {
            return nil
        }

        data := make([]g.Map, 0, len(menuIDs))
        for _, menuID := range menuIDs {
            data = append(data, g.Map{
                "group_id": groupID,
                "menu_id":  menuID,
            })
        }

        _, err := dao.AdminGroupMenu.Ctx(ctx).TX(tx).Insert(data)
        return err
    })
    if err != nil {
        return err
    }

    _ = service.Permission().RefreshGroupPermissionCache(ctx, groupID)
    _ = service.Log().WriteOperation(ctx, operator,
        fmt.Sprintf("保存用户组授权：group_id=%d，共 %d 个权限", groupID, len(menuIDs)),
    )
    return nil
}
```

## 11. 权限缓存实现

### 11.1 读取逻辑

普通员工权限读取顺序：

1. 先取 Redis：`admin:perm:group:{group_id}`
2. 缓存 miss：
   - 查 `admin_group` 状态
   - 查 `admin_group_menu`
   - 联查 `admin_menu.code`
   - 写回 Redis

### 11.2 刷新时机

以下操作后必须删除组权限缓存：

- 保存授权
- 删除用户组
- 修改用户组状态

### 11.3 `internal/logic/admin/permission/cache.go`

```go
// RefreshGroupPermissionCache 删除用户组权限缓存。
// 注意：这里使用删除而不是直接重建，避免并发写入时覆盖最新授权。
func (s *sPermission) RefreshGroupPermissionCache(ctx context.Context, groupID int64) error {
    key := fmt.Sprintf(consts.CacheKeyAdminPermGroup, groupID)
    _, err := g.Redis().Del(ctx, key)
    return err
}
```

## 12. 中间件权限判定补齐

Phase 1 的权限中间件骨架在本期补齐完整逻辑。

### 12.1 判定步骤

1. 读取请求 handler 的 `permission` 元数据
2. 读取当前用户上下文
3. `group_id=0` 直接放行
4. 若 `superOnly=true` 且非超管 → 403
5. 若 `permission` 为空 → 只要已登录即可
6. 若用户组被禁用 → 403
7. 读取权限码列表并判断是否包含当前接口所需权限码
8. 不包含则 403

### 12.2 权限检查骨架

```go
func (s *sPermission) CheckRequestPermission(r *ghttp.Request) error {
    handler := r.GetServeHandler()
    if handler == nil {
        return nil
    }
    if handler.GetMetaTag("noAuth") == "true" {
        return nil
    }

    adminCtx := contextx.MustFromRequest(r)
    if adminCtx.GroupID == 0 {
        return nil
    }

    if handler.GetMetaTag("superOnly") == "true" {
        return gerror.New("无访问权限")
    }

    permissionCode := handler.GetMetaTag("permission")
    if permissionCode == "" {
        return nil
    }

    perms, err := s.GetPermissionCodes(r.GetCtx(), adminCtx.AdminID, adminCtx.GroupID)
    if err != nil {
        return err
    }
    if !gstr.InArray(perms, permissionCode) {
        return gerror.New("无访问权限")
    }
    return nil
}
```

## 13. 路由与权限码映射建议

### 13.1 `admin.department`

用于：

- 用户组列表
- 添加/编辑/删除用户组
- 用户组状态
- 获取菜单树
- 获取授权
- 保存授权

### 13.2 `admin.list`

用于：

- 员工管理
- 主体配置

### 13.3 `admin.action`

用于：

- 操作日志列表（Phase 4）

### 13.4 `admin.loginlog`

用于：

- 登录日志列表（Phase 4）

## 14. 本期测试清单

### 14.1 用户组

- 新增用户组成功
- 用户组名称重复失败
- 编辑用户组成功
- 删除空用户组成功
- 删除带员工用户组失败
- 切换用户组状态成功

### 14.2 权限树

- 菜单树结构正确
- 菜单层级正确
- 排序正确
- 空 children 返回 `[]`

### 14.3 授权

- 获取授权成功
- 保存授权成功
- 清空授权成功
- 提交不存在 menu_id 失败
- 保存授权后缓存删除成功

### 14.4 权限中间件

- 超管无视权限码放行
- 普通员工缺权限返回 403
- 普通员工有权限放行
- 禁用用户组返回 403

## 15. 本期完成标准

- 用户组 CRUD 可用
- 权限树可用
- 授权保存可用
- 授权修改后普通员工下次请求自动生效
- 权限逻辑没有写死在 controller 中
- 不存在一个 1000 行的 `permission.go`
