package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

// SubjectLogic 提供主体配置管理相关业务能力。
type SubjectLogic struct{ core *app.Core }

// List 查询主体列表。
func (l *SubjectLogic) List(ctx context.Context, _ *adminapi.SubjectListReq) (*adminapi.SubjectListRes, error) {
	items := make([]app.AdminSubject, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, name, has_tax, created_at, updated_at FROM admin_subject ORDER BY id DESC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "主体列表查询失败")
	}
	return &adminapi.SubjectListRes{List: items}, nil
}

// Add 新增主体，并写入操作日志。
func (l *SubjectLogic) Add(ctx context.Context, req *adminapi.SubjectCreateReq, actor app.AdminUser, ip string) (*adminapi.SubjectCreateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		return nil, apiErr(consts.CodeBadRequest, "主体参数错误")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE name = ?`, req.Name)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "主体查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "主体名称已存在")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_subject (name, has_tax, created_at, updated_at) VALUES (?, ?, ?, ?)`, req.Name, req.HasTax, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "主体新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加主体：%s", req.Name), ip)
	return &adminapi.SubjectCreateRes{ID: id}, nil
}

// Edit 编辑主体信息，并写入操作日志。
func (l *SubjectLogic) Edit(ctx context.Context, req *adminapi.SubjectUpdateReq, actor app.AdminUser, ip string) (*adminapi.SubjectUpdateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		return nil, apiErr(consts.CodeBadRequest, "主体参数错误")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE name = ? AND id <> ?`, req.Name, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "主体查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "主体名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_subject SET name = ?, has_tax = ?, updated_at = ? WHERE id = ?`, req.Name, req.HasTax, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "主体编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑主体：%s", req.Name), ip)
	return &adminapi.SubjectUpdateRes{}, nil
}
