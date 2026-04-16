package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// IndustryController 提供行业管理相关 HTTP handler。
type IndustryController struct{ svc service.IndustryService }

// NewIndustry 创建 IndustryController。
func NewIndustry(svc service.IndustryService) *IndustryController {
	return &IndustryController{svc: svc}
}

// List 返回行业分页列表。
func (c *IndustryController) List(ctx context.Context, req *adminapi.IndustryListReq) (res *adminapi.IndustryListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增行业，并记录操作人与客户端 IP。
func (c *IndustryController) Create(ctx context.Context, req *adminapi.IndustryCreateReq) (res *adminapi.IndustryCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑行业，并记录操作人与客户端 IP。
func (c *IndustryController) Update(ctx context.Context, req *adminapi.IndustryUpdateReq) (res *adminapi.IndustryUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除行业，并记录操作人与客户端 IP。
func (c *IndustryController) Delete(ctx context.Context, req *adminapi.IndustryDeleteReq) (res *adminapi.IndustryDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Sort 调整行业排序，并记录操作人与客户端 IP。
func (c *IndustryController) Sort(ctx context.Context, req *adminapi.IndustrySortReq) (res *adminapi.IndustrySortRes, err error) {
	return c.svc.Sort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BrandSelector 返回可供行业关联的一级品牌选择器列表。
func (c *IndustryController) BrandSelector(ctx context.Context, req *adminapi.IndustryBrandSelectorReq) (res *adminapi.IndustryBrandSelectorRes, err error) {
	return c.svc.BrandSelector(ctx, req)
}

// BrandList 返回行业已关联品牌列表。
func (c *IndustryController) BrandList(ctx context.Context, req *adminapi.IndustryBrandListReq) (res *adminapi.IndustryBrandListRes, err error) {
	return c.svc.BrandList(ctx, req)
}

// BrandAdd 给行业添加一级品牌关联，并记录操作人与客户端 IP。
func (c *IndustryController) BrandAdd(ctx context.Context, req *adminapi.IndustryBrandAddReq) (res *adminapi.IndustryBrandAddRes, err error) {
	return c.svc.BrandAdd(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BrandDelete 删除行业下的品牌关联，并记录操作人与客户端 IP。
func (c *IndustryController) BrandDelete(ctx context.Context, req *adminapi.IndustryBrandDeleteReq) (res *adminapi.IndustryBrandDeleteRes, err error) {
	return c.svc.BrandDelete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// BrandSort 调整行业内品牌排序，并记录操作人与客户端 IP。
func (c *IndustryController) BrandSort(ctx context.Context, req *adminapi.IndustryBrandSortReq) (res *adminapi.IndustryBrandSortRes, err error) {
	return c.svc.BrandSort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
