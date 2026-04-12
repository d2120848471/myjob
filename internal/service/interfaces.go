package service

import (
	"context"

	v1 "myjob/api/admin/v1"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type AuthService interface {
	Login(ctx context.Context, req *v1.AuthLoginReq, ip string) (*v1.AuthLoginRes, error)
	LoginSMSSend(ctx context.Context, req *v1.AuthSMSSendReq) (*v1.AuthSMSSendRes, error)
	LoginSMSVerify(ctx context.Context, req *v1.AuthSMSVerifyReq) (*v1.AuthSMSVerifyRes, error)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*v1.AuthMeRes, error)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*v1.AuthSessionDeleteRes, error)
}

type UserService interface {
	List(ctx context.Context, req *v1.UserListReq) (*v1.UserListRes, error)
	Trash(ctx context.Context, req *v1.UserTrashReq) (*v1.UserTrashRes, error)
	Add(ctx context.Context, req *v1.UserCreateReq, actor entity.AdminUser, ip string) (*v1.UserCreateRes, error)
	Edit(ctx context.Context, req *v1.UserUpdateReq, actor entity.AdminUser, ip string) (*v1.UserUpdateRes, error)
	Delete(ctx context.Context, req *v1.UserDeleteReq, actor entity.AdminUser, ip string) (*v1.UserDeleteRes, error)
	Restore(ctx context.Context, req *v1.UserRestoreReq, actor entity.AdminUser, ip string) (*v1.UserRestoreRes, error)
	Status(ctx context.Context, req *v1.UserStatusReq, actor entity.AdminUser, ip string) (*v1.UserStatusRes, error)
	Notify(ctx context.Context, req *v1.UserNotifyReq, actor entity.AdminUser, ip string) (*v1.UserNotifyRes, error)
	SetBusiness(ctx context.Context, req *v1.UserBusinessAssignReq, actor entity.AdminUser, ip string) (*v1.UserBusinessAssignRes, error)
	CancelBusiness(ctx context.Context, req *v1.UserBusinessCancelReq, actor entity.AdminUser, ip string) (*v1.UserBusinessCancelRes, error)
}

type GroupService interface {
	List(ctx context.Context, req *v1.GroupListReq) (*v1.GroupListRes, error)
	Add(ctx context.Context, req *v1.GroupCreateReq, actor entity.AdminUser, ip string) (*v1.GroupCreateRes, error)
	Edit(ctx context.Context, req *v1.GroupUpdateReq, actor entity.AdminUser, ip string) (*v1.GroupUpdateRes, error)
	Delete(ctx context.Context, req *v1.GroupDeleteReq, actor entity.AdminUser, ip string) (*v1.GroupDeleteRes, error)
	Status(ctx context.Context, req *v1.GroupStatusReq, actor entity.AdminUser, ip string) (*v1.GroupStatusRes, error)
	AuthGet(ctx context.Context, req *v1.GroupPermissionsGetReq) (*v1.GroupPermissionsGetRes, error)
	AuthSave(ctx context.Context, req *v1.GroupPermissionsSaveReq, actor entity.AdminUser, ip string) (*v1.GroupPermissionsSaveRes, error)
	MenuTree(ctx context.Context, req *v1.MenuTreeReq) (*v1.MenuTreeRes, error)
}

type SubjectService interface {
	List(ctx context.Context, req *v1.SubjectListReq) (*v1.SubjectListRes, error)
	Add(ctx context.Context, req *v1.SubjectCreateReq, actor entity.AdminUser, ip string) (*v1.SubjectCreateRes, error)
	Edit(ctx context.Context, req *v1.SubjectUpdateReq, actor entity.AdminUser, ip string) (*v1.SubjectUpdateRes, error)
}

type SMSConfigService interface {
	Get(ctx context.Context, req *v1.SettingsSMSGetReq) (*v1.SettingsSMSGetRes, error)
	Save(ctx context.Context, req *v1.SettingsSMSSaveReq, actor entity.AdminUser, ip string) (*v1.SettingsSMSSaveRes, error)
}

type AuditLogService interface {
	OperationList(ctx context.Context, req *v1.OperationLogListReq) (*v1.OperationLogListRes, error)
	LoginList(ctx context.Context, req *v1.LoginLogListReq) (*v1.LoginLogListRes, error)
}
