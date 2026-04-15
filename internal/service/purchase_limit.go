package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// PurchaseLimitService 定义购买数量限制策略管理相关能力。
type PurchaseLimitService interface {
	List(ctx context.Context, req *adminapi.PurchaseLimitStrategyListReq) (*adminapi.PurchaseLimitStrategyListRes, error)
	Add(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyCreateRes, error)
	Edit(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyDeleteRes, error)
	Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyStatusRes, error)
	Enums(ctx context.Context, req *adminapi.PurchaseLimitStrategyEnumsReq) (*adminapi.PurchaseLimitStrategyEnumsRes, error)
}
