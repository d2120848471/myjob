package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// BrandController 提供品牌管理相关 HTTP handler。
type BrandController struct{ svc service.BrandService }

// NewBrand 创建 BrandController。
func NewBrand(svc service.BrandService) *BrandController { return &BrandController{svc: svc} }

// List 返回一级品牌分页列表。
func (c *BrandController) List(ctx context.Context, req *adminapi.BrandListReq) (res *adminapi.BrandListRes, err error) {
	return c.svc.List(ctx, req)
}

// Children 返回指定品牌的直接子品牌列表。
func (c *BrandController) Children(ctx context.Context, req *adminapi.BrandChildrenReq) (res *adminapi.BrandChildrenRes, err error) {
	return c.svc.Children(ctx, req)
}

// Create 新增品牌（支持新增子品牌），并记录操作人与客户端 IP。
func (c *BrandController) Create(ctx context.Context, req *adminapi.BrandCreateReq) (res *adminapi.BrandCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑品牌信息，并记录操作人与客户端 IP。
func (c *BrandController) Update(ctx context.Context, req *adminapi.BrandUpdateReq) (res *adminapi.BrandUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Delete 删除品牌，并记录操作人与客户端 IP。
func (c *BrandController) Delete(ctx context.Context, req *adminapi.BrandDeleteReq) (res *adminapi.BrandDeleteRes, err error) {
	return c.svc.Delete(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Sort 调整同级品牌排序，并记录操作人与客户端 IP。
func (c *BrandController) Sort(ctx context.Context, req *adminapi.BrandSortReq) (res *adminapi.BrandSortRes, err error) {
	return c.svc.Sort(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Visibility 切换品牌显隐状态，并记录操作人与客户端 IP。
func (c *BrandController) Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq) (res *adminapi.BrandVisibilityRes, err error) {
	return c.svc.Visibility(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Upload 上传品牌图片（icon/资质图），并记录操作人与客户端 IP。
func (c *BrandController) Upload(ctx context.Context, req *adminapi.BrandUploadReq) (res *adminapi.BrandUploadRes, err error) {
	return c.svc.Upload(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
