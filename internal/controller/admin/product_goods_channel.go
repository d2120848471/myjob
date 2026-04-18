package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
)

// ChannelBindingList 返回指定商品的渠道绑定弹窗数据。
func (c *ProductGoodsController) ChannelBindingList(ctx context.Context, req *adminapi.ProductGoodsChannelBindingListReq) (res *adminapi.ProductGoodsChannelBindingListRes, err error) {
	return c.svc.ChannelBindingList(ctx, req)
}

// ChannelBindingFormOptions 返回商品渠道绑定表单的渠道账号、模板和枚举选项。
func (c *ProductGoodsController) ChannelBindingFormOptions(ctx context.Context, req *adminapi.ProductGoodsChannelBindingFormOptionsReq) (res *adminapi.ProductGoodsChannelBindingFormOptionsRes, err error) {
	return c.svc.ChannelBindingFormOptions(ctx, req)
}

// CreateChannelBinding 新增单条商品渠道绑定，并记录操作人与客户端 IP。
func (c *ProductGoodsController) CreateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq) (res *adminapi.ProductGoodsChannelBindingCreateRes, err error) {
	return c.svc.CreateChannelBinding(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateChannelBinding 编辑单条商品渠道绑定基础字段，并记录操作人与客户端 IP。
func (c *ProductGoodsController) UpdateChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq) (res *adminapi.ProductGoodsChannelBindingUpdateRes, err error) {
	return c.svc.UpdateChannelBinding(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// DeleteChannelBinding 软删除单条商品渠道绑定，并记录操作人与客户端 IP。
func (c *ProductGoodsController) DeleteChannelBinding(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq) (res *adminapi.ProductGoodsChannelBindingDeleteRes, err error) {
	return c.svc.DeleteChannelBinding(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// UpdateChannelBindingAutoPrice 编辑单条绑定的自动改价规则，并记录操作人与客户端 IP。
func (c *ProductGoodsController) UpdateChannelBindingAutoPrice(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq) (res *adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, err error) {
	return c.svc.UpdateChannelBindingAutoPrice(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
