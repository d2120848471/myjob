package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// Update 更新指定商品的渠道配置，并刷新绑定摘要。
func (l *ProductGoodsChannelConfigLogic) Update(ctx context.Context, req *adminapi.ProductGoodsChannelConfigUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelConfigUpdateRes, error) {
	goods, err := l.getActiveGoodsForChannelConfig(ctx, req.GoodsID)
	if err != nil {
		return nil, err
	}
	if goods.SupplyType != "channel" {
		return nil, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}
	if err := l.ensureGoodsChannelConfigRow(ctx, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道配置更新失败")
	}

	routeMode := strings.TrimSpace(req.RouteMode)
	if err := validateGoodsChannelRouteMode(routeMode); err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	timeoutMinutes, err := normalizeGoodsChannelAttemptTimeout(req.AttemptTimeoutEnabled, req.AttemptTimeoutMinutes)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	maxLossAmount, err := normalizeGoodsChannelMaxLossAmount(req.AllowLoss, req.MaxLossAmount)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	var maxLossAny any = nil
	if maxLossAmount != nil {
		maxLossAny = *maxLossAmount
	}

	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_config
SET smart_replenish_enabled = ?,
    attempt_timeout_enabled = ?,
    attempt_timeout_minutes = ?,
    route_mode = ?,
    sync_cost_enabled = ?,
    sync_goods_name_enabled = ?,
    allow_loss = ?,
    max_loss_amount = ?,
    is_bundle = ?,
    updated_at = ?
WHERE goods_id = ?
`, boolToInt(req.SmartReplenishEnabled),
		boolToInt(req.AttemptTimeoutEnabled),
		timeoutMinutes,
		routeMode,
		boolToInt(req.SyncCostEnabled),
		boolToInt(req.SyncGoodsNameEnabled),
		boolToInt(req.AllowLoss),
		maxLossAny,
		boolToInt(req.IsBundle),
		l.core.Now(),
		req.GoodsID,
	); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道配置更新失败")
	}

	if err := l.refreshGoodsChannelSummary(ctx, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品渠道配置更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("更新商品渠道配置：%d", req.GoodsID), ip)
	return &adminapi.ProductGoodsChannelConfigUpdateRes{}, nil
}
