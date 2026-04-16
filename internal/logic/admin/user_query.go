package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

// List 分页查询未删除的员工列表。
func (l *UserLogic) List(ctx context.Context, req *adminapi.UserListReq) (*adminapi.UserListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 0`)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工列表查询失败")
	}
	items := make([]app.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 0
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "员工列表查询失败")
	}
	return &adminapi.UserListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

// Trash 分页查询已删除的员工列表（回收站）。
func (l *UserLogic) Trash(ctx context.Context, req *adminapi.UserTrashReq) (*adminapi.UserTrashRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE is_deleted = 1`)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "回收站查询失败")
	}
	items := make([]app.UserListItem, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &items, `
SELECT u.id, u.username, u.real_name, u.phone, u.group_id, COALESCE(g.name, '超级管理员') AS group_name,
       u.status, u.balance_notify, u.is_business
FROM admin_user u
LEFT JOIN admin_group g ON g.id = u.group_id
WHERE u.is_deleted = 1
ORDER BY u.id DESC
LIMIT ? OFFSET ?
`, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "回收站查询失败")
	}
	return &adminapi.UserTrashRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}
