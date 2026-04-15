package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// ProductTemplateService 定义商品模板管理相关能力。
type ProductTemplateService interface {
	List(ctx context.Context, req *adminapi.ProductTemplateListReq) (*adminapi.ProductTemplateListRes, error)
	Add(ctx context.Context, req *adminapi.ProductTemplateCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateCreateRes, error)
	Edit(ctx context.Context, req *adminapi.ProductTemplateUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateDeleteRes, error)
	BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateBatchDeleteRes, error)
	ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (*adminapi.ProductTemplateValidateTypeListRes, error)
}
