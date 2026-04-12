package service

import (
	"context"

	authv1 "myjob/api/admin/auth/v1"
	configv1 "myjob/api/admin/config/v1"
	groupv1 "myjob/api/admin/group/v1"
	logv1 "myjob/api/admin/log/v1"
	subjectv1 "myjob/api/admin/subject/v1"
	userv1 "myjob/api/admin/user/v1"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type AuthService interface {
	Login(ctx context.Context, req authv1.LoginReq, ip string) (map[string]any, *modelruntime.APIError)
	LoginSMSSend(ctx context.Context, req authv1.LoginSMSSendReq) (map[string]any, *modelruntime.APIError)
	LoginSMSVerify(ctx context.Context, req authv1.LoginSMSVerifyReq) (map[string]any, *modelruntime.APIError)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (map[string]any, *modelruntime.APIError)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (map[string]any, *modelruntime.APIError)
}

type UserService interface {
	List(ctx context.Context, req userv1.ListReq) (map[string]any, *modelruntime.APIError)
	Trash(ctx context.Context, req userv1.ListReq) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req userv1.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req userv1.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Delete(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Restore(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Status(ctx context.Context, id int64, req userv1.StatusReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Notify(ctx context.Context, id int64, req userv1.NotifyReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	SetBusiness(ctx context.Context, req userv1.BusinessReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	CancelBusiness(ctx context.Context, req userv1.BusinessReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type GroupService interface {
	List(ctx context.Context, req groupv1.ListReq) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req groupv1.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req groupv1.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Delete(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Status(ctx context.Context, id int64, req groupv1.StatusReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	AuthGet(ctx context.Context, id int64) (map[string]any, *modelruntime.APIError)
	AuthSave(ctx context.Context, id int64, req groupv1.AuthSaveReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	MenuTree(ctx context.Context) (any, *modelruntime.APIError)
}

type SubjectService interface {
	List(ctx context.Context) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req subjectv1.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req subjectv1.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type SMSConfigService interface {
	Get(ctx context.Context) (configv1.SMSConfigGetRes, *modelruntime.APIError)
	Save(ctx context.Context, req configv1.SMSConfigSaveReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type AuditLogService interface {
	OperationList(ctx context.Context, req logv1.ListReq) (map[string]any, *modelruntime.APIError)
	LoginList(ctx context.Context, req logv1.ListReq) (map[string]any, *modelruntime.APIError)
}
