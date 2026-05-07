package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

// List 分页查询未删除客户列表。
func (l *CustomerLogic) List(ctx context.Context, req *adminapi.CustomerListReq) (*adminapi.CustomerListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	where, args := customerListWhere(0, strings.TrimSpace(req.Keyword), req.Status)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM customer_user `+where, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户列表查询失败")
	}
	rows := make([]customerListRow, 0)
	queryArgs := append(args, pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, company_name, phone, status,
       COALESCE(last_login_ip, '') AS last_login_ip,
       last_login_at, created_at, updated_at
FROM customer_user `+where+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户列表查询失败")
	}
	return &adminapi.CustomerListRes{List: mapCustomerListRows(rows), Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

// Trash 分页查询客户回收站。
func (l *CustomerLogic) Trash(ctx context.Context, req *adminapi.CustomerTrashReq) (*adminapi.CustomerTrashRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	where, args := customerListWhere(1, strings.TrimSpace(req.Keyword), nil)
	totalVal, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM customer_user `+where, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户回收站查询失败")
	}
	rows := make([]customerListRow, 0)
	queryArgs := append(args, pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, company_name, phone, status,
       COALESCE(last_login_ip, '') AS last_login_ip,
       last_login_at, created_at, updated_at
FROM customer_user `+where+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户回收站查询失败")
	}
	return &adminapi.CustomerTrashRes{List: mapCustomerListRows(rows), Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: totalVal.Int()}}, nil
}

// Detail 读取客户详情。
func (l *CustomerLogic) Detail(ctx context.Context, req *adminapi.CustomerDetailReq) (*adminapi.CustomerDetailRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	return mapCustomerDetail(customer), nil
}

func customerListWhere(isDeleted int, keyword string, status *int) (string, []any) {
	conditions := []string{"WHERE is_deleted = ?"}
	args := []any{isDeleted}
	if keyword != "" {
		conditions = append(conditions, "(company_name LIKE ? OR phone LIKE ?)")
		like := fmt.Sprintf("%%%s%%", keyword)
		args = append(args, like, like)
	}
	if status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *status)
	}
	return strings.Join(conditions, " AND "), args
}
