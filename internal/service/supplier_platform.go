package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// SupplierPlatformService 定义第三方对接平台账号与余额刷新相关能力。
type SupplierPlatformService interface {
	TypeList(ctx context.Context, req *adminapi.SupplierPlatformTypeListReq) (*adminapi.SupplierPlatformTypeListRes, error)
	List(ctx context.Context, req *adminapi.SupplierPlatformListReq) (*adminapi.SupplierPlatformListRes, error)
	Detail(ctx context.Context, req *adminapi.SupplierPlatformDetailReq) (*adminapi.SupplierPlatformDetailRes, error)
	Add(ctx context.Context, req *adminapi.SupplierPlatformCreateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformCreateRes, error)
	Edit(ctx context.Context, req *adminapi.SupplierPlatformUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.SupplierPlatformDeleteReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformDeleteRes, error)
	RefreshBalance(ctx context.Context, req *adminapi.SupplierPlatformRefreshBalanceReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformRefreshBalanceRes, error)
}
