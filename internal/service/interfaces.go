package service

import (
	"context"

	authapi "myjob/api/auth"
	configapi "myjob/api/config"
	groupapi "myjob/api/group"
	logapi "myjob/api/log"
	subjectapi "myjob/api/subject"
	userapi "myjob/api/user"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type AuthService interface {
	Login(ctx context.Context, req authapi.LoginReq, ip string) (map[string]any, *modelruntime.APIError)
	LoginSMSSend(ctx context.Context, req authapi.LoginSMSSendReq) (map[string]any, *modelruntime.APIError)
	LoginSMSVerify(ctx context.Context, req authapi.LoginSMSVerifyReq) (map[string]any, *modelruntime.APIError)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (map[string]any, *modelruntime.APIError)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (map[string]any, *modelruntime.APIError)
}

type UserService interface {
	List(ctx context.Context, req userapi.ListReq) (map[string]any, *modelruntime.APIError)
	Trash(ctx context.Context, req userapi.ListReq) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req userapi.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req userapi.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Delete(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Restore(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Status(ctx context.Context, id int64, req userapi.StatusReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Notify(ctx context.Context, id int64, req userapi.NotifyReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	SetBusiness(ctx context.Context, req userapi.BusinessReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	CancelBusiness(ctx context.Context, req userapi.BusinessReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type GroupService interface {
	List(ctx context.Context, req groupapi.ListReq) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req groupapi.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req groupapi.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Delete(ctx context.Context, id int64, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Status(ctx context.Context, id int64, req groupapi.StatusReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	AuthGet(ctx context.Context, id int64) (map[string]any, *modelruntime.APIError)
	AuthSave(ctx context.Context, id int64, req groupapi.AuthSaveReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	MenuTree(ctx context.Context) (any, *modelruntime.APIError)
}

type SubjectService interface {
	List(ctx context.Context) (map[string]any, *modelruntime.APIError)
	Add(ctx context.Context, req subjectapi.AddReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
	Edit(ctx context.Context, id int64, req subjectapi.EditReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type SMSConfigService interface {
	Get(ctx context.Context) (configapi.SMSConfigGetRes, *modelruntime.APIError)
	Save(ctx context.Context, req configapi.SMSConfigSaveReq, actor entity.AdminUser, ip string) (map[string]any, *modelruntime.APIError)
}

type AuditLogService interface {
	OperationList(ctx context.Context, req logapi.ListReq) (map[string]any, *modelruntime.APIError)
	LoginList(ctx context.Context, req logapi.ListReq) (map[string]any, *modelruntime.APIError)
}
