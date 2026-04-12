package adminlogic

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	subjectapi "myjob/api/subject"
	"myjob/internal/kernel"
	modelruntime "myjob/internal/model/runtime"
)

type SubjectLogic struct{ core *kernel.Core }

func (l *SubjectLogic) List(ctx context.Context) (map[string]any, *modelruntime.APIError) {
	items := make([]kernel.AdminSubject, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, name, has_tax, created_at, updated_at FROM admin_subject ORDER BY id DESC`); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "主体列表查询失败")
	}
	return map[string]any{"list": items}, nil
}

func (l *SubjectLogic) Add(ctx context.Context, req subjectapi.AddReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		return nil, apiErr(http.StatusBadRequest, 400, "主体参数错误")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE name = ?`, req.Name)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "主体查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(http.StatusConflict, 409, "主体名称已存在")
	}
	result, err := l.core.DB().Exec(ctx, `INSERT INTO admin_subject (name, has_tax, created_at, updated_at) VALUES (?, ?, ?, ?)`, req.Name, req.HasTax, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "主体新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("添加主体：%s", req.Name), ip)
	return map[string]any{"id": id}, nil
}

func (l *SubjectLogic) Edit(ctx context.Context, id int64, req subjectapi.EditReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		return nil, apiErr(http.StatusBadRequest, 400, "主体参数错误")
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE name = ? AND id <> ?`, req.Name, id)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "主体查询失败")
	}
	if exists.Int() > 0 {
		return nil, apiErr(http.StatusConflict, 409, "主体名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `UPDATE admin_subject SET name = ?, has_tax = ?, updated_at = ? WHERE id = ?`, req.Name, req.HasTax, l.core.Now(), id); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "主体编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑主体：%s", req.Name), ip)
	return map[string]any{}, nil
}
