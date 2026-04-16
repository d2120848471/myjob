package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

// SetBusiness 批量设置员工为商务角色。
func (l *UserLogic) SetBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq, actor app.AdminUser, ip string) (*adminapi.UserBusinessAssignRes, error) {
	if err := l.handleBusiness(ctx, req.IDs, actor, ip, 1); err != nil {
		return nil, err
	}
	return &adminapi.UserBusinessAssignRes{}, nil
}

// CancelBusiness 批量取消员工商务角色。
func (l *UserLogic) CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq, actor app.AdminUser, ip string) (*adminapi.UserBusinessCancelRes, error) {
	if err := l.handleBusiness(ctx, req.IDs, actor, ip, 0); err != nil {
		return nil, err
	}
	return &adminapi.UserBusinessCancelRes{}, nil
}

func (l *UserLogic) handleBusiness(ctx context.Context, ids []int64, actor app.AdminUser, ip string, flag int) error {
	if len(ids) == 0 {
		return apiErr(consts.CodeBadRequest, "ID列表不能为空")
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+2)
	args = append(args, flag, l.core.Now())
	for _, id := range ids {
		args = append(args, id)
	}
	query := `UPDATE admin_user SET is_business = ?, updated_at = ? WHERE id IN (` + placeholders + `)`
	if _, err := l.core.DB().Exec(ctx, query, args...); err != nil {
		return apiErr(consts.CodeInternalError, "批量更新失败")
	}
	action := "批量取消商务"
	if flag == 1 {
		action = "批量设置商务"
	}
	l.core.WriteOperation(ctx, actor, action, ip)
	return nil
}
