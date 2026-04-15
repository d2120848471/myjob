package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"

	"github.com/gogf/gf/v2/database/gdb"
)

// GroupLogic 提供用户组与授权管理相关业务能力。
type GroupLogic struct{ core *app.Core }

// List 分页查询用户组列表。
func (l *GroupLogic) List(ctx context.Context, req *adminapi.GroupListReq) (*adminapi.GroupListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group`)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组列表查询失败")
	}
	items := make([]app.GroupListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT g.id, g.name, g.description, g.status, (SELECT COUNT(*) FROM admin_user u WHERE u.group_id = g.id) AS user_count FROM admin_group g ORDER BY g.id DESC LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组列表查询失败")
	}
	return &adminapi.GroupListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

// Add 新增用户组，并写入操作日志。
func (l *GroupLogic) Add(ctx context.Context, req *adminapi.GroupCreateReq, actor app.AdminUser, ip string) (*adminapi.GroupCreateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(consts.CodeBadRequest, "请输入用户组名称")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group WHERE name = ?`, req.Name)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "用户组名称已存在")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_group (name, description, status, created_at, updated_at) VALUES (?, ?, 1, ?, ?)`, req.Name, strings.TrimSpace(req.Description), l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加用户组：%s", req.Name), ip)
	return &adminapi.GroupCreateRes{ID: id}, nil
}

// Edit 编辑用户组信息，并写入操作日志。
func (l *GroupLogic) Edit(ctx context.Context, req *adminapi.GroupUpdateReq, actor app.AdminUser, ip string) (*adminapi.GroupUpdateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(consts.CodeBadRequest, "请输入用户组名称")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group WHERE name = ? AND id <> ?`, req.Name, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "用户组名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_group SET name = ?, description = ?, updated_at = ? WHERE id = ?`, req.Name, strings.TrimSpace(req.Description), l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑用户组：%s", req.Name), ip)
	return &adminapi.GroupUpdateRes{}, nil
}

// Delete 删除用户组（需保证组内无员工），并清理授权与权限缓存。
func (l *GroupLogic) Delete(ctx context.Context, req *adminapi.GroupDeleteReq, actor app.AdminUser, ip string) (*adminapi.GroupDeleteRes, error) {
	userCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE group_id = ?`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组校验失败")
	}
	if userCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, fmt.Sprintf("该用户组下还有 %d 名员工，请先转移", userCount.Int()))
	}
	// 删除用户组与其授权关系需在同一事务中完成，避免出现孤儿授权数据。
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM admin_group_menu WHERE group_id = ?`, req.ID); txErr != nil {
			return txErr
		}
		if _, txErr := tx.Exec(`DELETE FROM admin_group WHERE id = ?`, req.ID); txErr != nil {
			return txErr
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组删除失败")
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.PermissionCacheKey(req.ID))
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除用户组：%d", req.ID), ip)
	return &adminapi.GroupDeleteRes{}, nil
}

// Status 切换用户组启用/禁用状态，并清理权限缓存。
func (l *GroupLogic) Status(ctx context.Context, req *adminapi.GroupStatusReq, actor app.AdminUser, ip string) (*adminapi.GroupStatusRes, error) {
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE admin_group SET status = ?, updated_at = ? WHERE id = ?`, req.Status, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "用户组状态更新失败")
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.PermissionCacheKey(req.ID))
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换用户组状态：%d -> %d", req.ID, req.Status), ip)
	return &adminapi.GroupStatusRes{}, nil
}

// AuthGet 获取用户组当前已授权的菜单 id 列表（过滤超级管理员专属权限点）。
func (l *GroupLogic) AuthGet(ctx context.Context, req *adminapi.GroupPermissionsGetReq) (*adminapi.GroupPermissionsGetRes, error) {
	arr, err := l.core.DB().GetCore().GetArray(ctx, `SELECT gm.menu_id FROM admin_group_menu gm JOIN admin_menu m ON m.id = gm.menu_id WHERE gm.group_id = ? AND m.status = 1 AND m.super_only = 0 ORDER BY gm.menu_id ASC`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "授权查询失败")
	}
	ids := make([]int64, 0, len(arr))
	for _, item := range arr {
		ids = append(ids, item.Int64())
	}
	return &adminapi.GroupPermissionsGetRes{MenuIDs: ids}, nil
}

// AuthSave 保存用户组授权菜单，并清理权限缓存。
func (l *GroupLogic) AuthSave(ctx context.Context, req *adminapi.GroupPermissionsSaveReq, actor app.AdminUser, ip string) (*adminapi.GroupPermissionsSaveRes, error) {
	allowed := make([]int64, 0, len(req.MenuIDs))
	// 删除旧授权并写入新授权需同事务完成，避免中间态导致权限瞬时异常。
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM admin_group_menu WHERE group_id = ?`, req.ID); txErr != nil {
			return txErr
		}
		if len(req.MenuIDs) > 0 {
			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(req.MenuIDs)), ",")
			args := make([]any, 0, len(req.MenuIDs))
			for _, menuID := range req.MenuIDs {
				args = append(args, menuID)
			}
			rows, err := tx.GetAll(`SELECT id FROM admin_menu WHERE status = 1 AND super_only = 0 AND id IN (`+placeholders+`) ORDER BY sort ASC, id ASC`, args...)
			if err != nil {
				return err
			}
			allowed = allowed[:0]
			for _, item := range rows {
				allowed = append(allowed, item["id"].Int64())
			}
		}
		for _, menuID := range allowed {
			if _, txErr := tx.Exec(`INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (?, ?, ?)`, req.ID, menuID, l.core.Now()); txErr != nil {
				return txErr
			}
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "授权保存失败")
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.PermissionCacheKey(req.ID))
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("保存用户组授权：%d，共 %d 个权限", req.ID, len(allowed)), ip)
	return &adminapi.GroupPermissionsSaveRes{}, nil
}

// MenuTree 返回可授权的菜单树（过滤超级管理员专属权限点）。
func (l *GroupLogic) MenuTree(ctx context.Context, _ *adminapi.MenuTreeReq) (*adminapi.MenuTreeRes, error) {
	items := make([]app.AdminMenu, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, parent_id, name, code, menu_type, menu_level, status, super_only, sort FROM admin_menu WHERE status = 1 AND super_only = 0 ORDER BY sort ASC, id ASC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "菜单树查询失败")
	}
	return &adminapi.MenuTreeRes{List: app.BuildMenuTree(items, 0)}, nil
}
