package adminlogic

import (
	"context"

	"fmt"
	authlib "myjob/internal/library/auth"
	"net/http"
	"strings"

	groupapi "myjob/api/group"
	"myjob/internal/kernel"
	"myjob/internal/model/dto/admin"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gdb"
)

type GroupLogic struct{ core *kernel.Core }

func (l *GroupLogic) List(ctx context.Context, req groupapi.ListReq) (map[string]any, *modelruntime.APIError) {
	page, pageSize := kernel.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group`)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组列表查询失败")
	}
	items := make([]kernel.GroupListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `SELECT g.id, g.name, g.description, g.status, (SELECT COUNT(*) FROM admin_user u WHERE u.group_id = g.id) AS user_count FROM admin_group g ORDER BY g.id DESC LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组列表查询失败")
	}
	return map[string]any{"list": items, "pagination": admin.Pagination{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

func (l *GroupLogic) Add(ctx context.Context, req groupapi.AddReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(http.StatusBadRequest, 400, "请输入用户组名称")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group WHERE name = ?`, req.Name)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(http.StatusConflict, 409, "用户组名称已存在")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_group (name, description, status, created_at, updated_at) VALUES (?, ?, 1, ?, ?)`, req.Name, strings.TrimSpace(req.Description), l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加用户组：%s", req.Name), ip)
	return map[string]any{"id": id}, nil
}

func (l *GroupLogic) Edit(ctx context.Context, id int64, req groupapi.EditReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, apiErr(http.StatusBadRequest, 400, "请输入用户组名称")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group WHERE name = ? AND id <> ?`, req.Name, id)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(http.StatusConflict, 409, "用户组名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_group SET name = ?, description = ?, updated_at = ? WHERE id = ?`, req.Name, strings.TrimSpace(req.Description), l.core.Now(), id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑用户组：%s", req.Name), ip)
	return map[string]any{}, nil
}

func (l *GroupLogic) Delete(ctx context.Context, id int64, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	userCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE group_id = ?`, id)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组校验失败")
	}
	if userCount.Int() > 0 {
		return nil, apiErr(http.StatusConflict, 409, fmt.Sprintf("该用户组下还有 %d 名员工，请先转移", userCount.Int()))
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM admin_group_menu WHERE group_id = ?`, id); txErr != nil {
			return txErr
		}
		if _, txErr := tx.Exec(`DELETE FROM admin_group WHERE id = ?`, id); txErr != nil {
			return txErr
		}
		return nil
	}); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组删除失败")
	}
	_ = l.core.Redis().Del(ctx, authlib.PermissionCacheKey(id)).Err()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除用户组：%d", id), ip)
	return map[string]any{}, nil
}

func (l *GroupLogic) Status(ctx context.Context, id int64, req groupapi.StatusReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(http.StatusBadRequest, 400, "状态错误")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE admin_group SET status = ?, updated_at = ? WHERE id = ?`, req.Status, l.core.Now(), id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "用户组状态更新失败")
	}
	_ = l.core.Redis().Del(ctx, authlib.PermissionCacheKey(id)).Err()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换用户组状态：%d -> %d", id, req.Status), ip)
	return map[string]any{}, nil
}

func (l *GroupLogic) AuthGet(ctx context.Context, id int64) (map[string]any, *modelruntime.APIError) {
	arr, err := l.core.DB().GetCore().GetArray(ctx, `SELECT gm.menu_id FROM admin_group_menu gm JOIN admin_menu m ON m.id = gm.menu_id WHERE gm.group_id = ? AND m.status = 1 AND m.super_only = 0 ORDER BY gm.menu_id ASC`, id)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "授权查询失败")
	}
	ids := make([]int64, 0, len(arr))
	for _, item := range arr {
		ids = append(ids, item.Int64())
	}
	return map[string]any{"menu_ids": ids}, nil
}

func (l *GroupLogic) AuthSave(ctx context.Context, id int64, req groupapi.AuthSaveReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	allowed := make([]int64, 0, len(req.MenuIDs))
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM admin_group_menu WHERE group_id = ?`, id); txErr != nil {
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
			if _, txErr := tx.Exec(`INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (?, ?, ?)`, id, menuID, l.core.Now()); txErr != nil {
				return txErr
			}
		}
		return nil
	}); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "授权保存失败")
	}
	_ = l.core.Redis().Del(ctx, authlib.PermissionCacheKey(id)).Err()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("保存用户组授权：%d，共 %d 个权限", id, len(allowed)), ip)
	return map[string]any{}, nil
}

func (l *GroupLogic) MenuTree(ctx context.Context) (any, *modelruntime.APIError) {
	items := make([]kernel.AdminMenu, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, parent_id, name, code, menu_type, menu_level, status, super_only, sort FROM admin_menu WHERE status = 1 AND super_only = 0 ORDER BY sort ASC, id ASC`); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "菜单树查询失败")
	}
	return kernel.BuildMenuTree(items, 0), nil
}
