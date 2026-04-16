package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
)

// Notify 设置员工余额变动通知开关，并写入操作日志。
func (l *UserLogic) Notify(ctx context.Context, req *adminapi.UserNotifyReq, actor app.AdminUser, ip string) (*adminapi.UserNotifyRes, error) {
	if req.BalanceNotify != 0 && req.BalanceNotify != 1 {
		return nil, apiErr(consts.CodeBadRequest, "余额通知值错误")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE admin_user SET balance_notify = ?, updated_at = ? WHERE id = ?`, req.BalanceNotify, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "余额通知更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换余额通知：%d", req.ID), ip)
	return &adminapi.UserNotifyRes{}, nil
}
