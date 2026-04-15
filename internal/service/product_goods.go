package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// ProductGoodsService 定义商品管理相关能力。
type ProductGoodsService interface {
	List(ctx context.Context, req *adminapi.ProductGoodsListReq) (*adminapi.ProductGoodsListRes, error)
	Detail(ctx context.Context, req *adminapi.ProductGoodsDetailReq) (*adminapi.ProductGoodsDetailRes, error)
	FormOptions(ctx context.Context, req *adminapi.ProductGoodsFormOptionsReq) (*adminapi.ProductGoodsFormOptionsRes, error)
	Add(ctx context.Context, req *adminapi.ProductGoodsCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsCreateRes, error)
	Edit(ctx context.Context, req *adminapi.ProductGoodsUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsDeleteRes, error)
	Status(ctx context.Context, req *adminapi.ProductGoodsStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsStatusRes, error)
}
