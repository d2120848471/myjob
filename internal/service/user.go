package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// UserService 定义员工（管理员）管理相关能力。
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
