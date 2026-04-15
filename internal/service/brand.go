package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// BrandService 定义品牌管理相关能力。
type BrandService interface {
	List(ctx context.Context, req *adminapi.BrandListReq) (*adminapi.BrandListRes, error)
	Children(ctx context.Context, req *adminapi.BrandChildrenReq) (*adminapi.BrandChildrenRes, error)
	Add(ctx context.Context, req *adminapi.BrandCreateReq, actor entity.AdminUser, ip string) (*adminapi.BrandCreateRes, error)
	Edit(ctx context.Context, req *adminapi.BrandUpdateReq, actor entity.AdminUser, ip string) (*adminapi.BrandUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.BrandDeleteReq, actor entity.AdminUser, ip string) (*adminapi.BrandDeleteRes, error)
	Sort(ctx context.Context, req *adminapi.BrandSortReq, actor entity.AdminUser, ip string) (*adminapi.BrandSortRes, error)
	Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq, actor entity.AdminUser, ip string) (*adminapi.BrandVisibilityRes, error)
	Upload(ctx context.Context, req *adminapi.BrandUploadReq, actor entity.AdminUser, ip string) (*adminapi.BrandUploadRes, error)
}
