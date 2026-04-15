package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// IndustryService 定义行业管理与行业-品牌关联相关能力。
type IndustryService interface {
	List(ctx context.Context, req *adminapi.IndustryListReq) (*adminapi.IndustryListRes, error)
	Add(ctx context.Context, req *adminapi.IndustryCreateReq, actor entity.AdminUser, ip string) (*adminapi.IndustryCreateRes, error)
	Edit(ctx context.Context, req *adminapi.IndustryUpdateReq, actor entity.AdminUser, ip string) (*adminapi.IndustryUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.IndustryDeleteReq, actor entity.AdminUser, ip string) (*adminapi.IndustryDeleteRes, error)
	Sort(ctx context.Context, req *adminapi.IndustrySortReq, actor entity.AdminUser, ip string) (*adminapi.IndustrySortRes, error)
	BrandSelector(ctx context.Context, req *adminapi.IndustryBrandSelectorReq) (*adminapi.IndustryBrandSelectorRes, error)
	BrandList(ctx context.Context, req *adminapi.IndustryBrandListReq) (*adminapi.IndustryBrandListRes, error)
	BrandAdd(ctx context.Context, req *adminapi.IndustryBrandAddReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandAddRes, error)
	BrandDelete(ctx context.Context, req *adminapi.IndustryBrandDeleteReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandDeleteRes, error)
	BrandSort(ctx context.Context, req *adminapi.IndustryBrandSortReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandSortRes, error)
}
