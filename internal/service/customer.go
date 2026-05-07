package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// CustomerService 定义后台客户管理能力。
type CustomerService interface {
	List(ctx context.Context, req *adminapi.CustomerListReq) (*adminapi.CustomerListRes, error)
	Trash(ctx context.Context, req *adminapi.CustomerTrashReq) (*adminapi.CustomerTrashRes, error)
	Detail(ctx context.Context, req *adminapi.CustomerDetailReq) (*adminapi.CustomerDetailRes, error)
	Add(ctx context.Context, req *adminapi.CustomerCreateReq, actor entity.AdminUser, ip string) (*adminapi.CustomerCreateRes, error)
	Edit(ctx context.Context, req *adminapi.CustomerUpdateReq, actor entity.AdminUser, ip string) (*adminapi.CustomerUpdateRes, error)
	Status(ctx context.Context, req *adminapi.CustomerStatusReq, actor entity.AdminUser, ip string) (*adminapi.CustomerStatusRes, error)
	Delete(ctx context.Context, req *adminapi.CustomerDeleteReq, actor entity.AdminUser, ip string) (*adminapi.CustomerDeleteRes, error)
	Restore(ctx context.Context, req *adminapi.CustomerRestoreReq, actor entity.AdminUser, ip string) (*adminapi.CustomerRestoreRes, error)
	ResetPassword(ctx context.Context, req *adminapi.CustomerPasswordResetReq, actor entity.AdminUser, ip string) (*adminapi.CustomerPasswordResetRes, error)
	ResetPayPassword(ctx context.Context, req *adminapi.CustomerPayPasswordResetReq, actor entity.AdminUser, ip string) (*adminapi.CustomerPayPasswordResetRes, error)
}
