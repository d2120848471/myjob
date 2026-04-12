package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type AuthService interface {
	Login(ctx context.Context, req *adminapi.AuthLoginReq, ip string) (*adminapi.AuthLoginRes, error)
	LoginSMSSend(ctx context.Context, req *adminapi.AuthSMSSendReq) (*adminapi.AuthSMSSendRes, error)
	LoginSMSVerify(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (*adminapi.AuthSMSVerifyRes, error)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthMeRes, error)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthSessionDeleteRes, error)
}

type UserService interface {
	List(ctx context.Context, req *adminapi.UserListReq) (*adminapi.UserListRes, error)
	Trash(ctx context.Context, req *adminapi.UserTrashReq) (*adminapi.UserTrashRes, error)
	Add(ctx context.Context, req *adminapi.UserCreateReq, actor entity.AdminUser, ip string) (*adminapi.UserCreateRes, error)
	Edit(ctx context.Context, req *adminapi.UserUpdateReq, actor entity.AdminUser, ip string) (*adminapi.UserUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.UserDeleteReq, actor entity.AdminUser, ip string) (*adminapi.UserDeleteRes, error)
	Restore(ctx context.Context, req *adminapi.UserRestoreReq, actor entity.AdminUser, ip string) (*adminapi.UserRestoreRes, error)
	Status(ctx context.Context, req *adminapi.UserStatusReq, actor entity.AdminUser, ip string) (*adminapi.UserStatusRes, error)
	Notify(ctx context.Context, req *adminapi.UserNotifyReq, actor entity.AdminUser, ip string) (*adminapi.UserNotifyRes, error)
	SetBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq, actor entity.AdminUser, ip string) (*adminapi.UserBusinessAssignRes, error)
	CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq, actor entity.AdminUser, ip string) (*adminapi.UserBusinessCancelRes, error)
}

type GroupService interface {
	List(ctx context.Context, req *adminapi.GroupListReq) (*adminapi.GroupListRes, error)
	Add(ctx context.Context, req *adminapi.GroupCreateReq, actor entity.AdminUser, ip string) (*adminapi.GroupCreateRes, error)
	Edit(ctx context.Context, req *adminapi.GroupUpdateReq, actor entity.AdminUser, ip string) (*adminapi.GroupUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.GroupDeleteReq, actor entity.AdminUser, ip string) (*adminapi.GroupDeleteRes, error)
	Status(ctx context.Context, req *adminapi.GroupStatusReq, actor entity.AdminUser, ip string) (*adminapi.GroupStatusRes, error)
	AuthGet(ctx context.Context, req *adminapi.GroupPermissionsGetReq) (*adminapi.GroupPermissionsGetRes, error)
	AuthSave(ctx context.Context, req *adminapi.GroupPermissionsSaveReq, actor entity.AdminUser, ip string) (*adminapi.GroupPermissionsSaveRes, error)
	MenuTree(ctx context.Context, req *adminapi.MenuTreeReq) (*adminapi.MenuTreeRes, error)
}

type SubjectService interface {
	List(ctx context.Context, req *adminapi.SubjectListReq) (*adminapi.SubjectListRes, error)
	Add(ctx context.Context, req *adminapi.SubjectCreateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectCreateRes, error)
	Edit(ctx context.Context, req *adminapi.SubjectUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectUpdateRes, error)
}

type SMSConfigService interface {
	Get(ctx context.Context, req *adminapi.SettingsSMSGetReq) (*adminapi.SettingsSMSGetRes, error)
	Save(ctx context.Context, req *adminapi.SettingsSMSSaveReq, actor entity.AdminUser, ip string) (*adminapi.SettingsSMSSaveRes, error)
}

type AuditLogService interface {
	OperationList(ctx context.Context, req *adminapi.OperationLogListReq) (*adminapi.OperationLogListRes, error)
	LoginList(ctx context.Context, req *adminapi.LoginLogListReq) (*adminapi.LoginLogListRes, error)
}
