package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// Add 新增购买数量限制策略，并写入操作日志。
func (l *PurchaseLimitLogic) Add(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyCreateRes, error) {
	normalized, err := normalizePurchaseLimitInput(req.Name, req.LimitType, req.PeriodType, req.Period, req.LimitNums, req.LimitTimes)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	// 新增弹层不承载状态切换，后端默认开通，后续启停统一走独立状态接口。
	defaultStatus := consts.StatusEnabled
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_purchase_limit_strategy (
    name, limit_type, period_type, period, limit_nums, limit_times, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, normalized.Name, normalized.LimitType, normalized.PeriodType, normalized.Period, normalized.LimitNums, normalized.LimitTimes, defaultStatus, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品购买数量限制策略：%s", normalized.Name), ip)
	return &adminapi.PurchaseLimitStrategyCreateRes{ID: id}, nil
}

// Edit 编辑购买数量限制策略，并写入操作日志。
func (l *PurchaseLimitLogic) Edit(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyUpdateRes, error) {
	if _, err := l.getStrategy(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	normalized, err := normalizePurchaseLimitInput(req.Name, req.LimitType, req.PeriodType, req.Period, req.LimitNums, req.LimitTimes)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	// 编辑弹层只更新策略内容，不顺带改状态，避免未点启停却把策略误关停。
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_purchase_limit_strategy
SET name = ?, limit_type = ?, period_type = ?, period = ?, limit_nums = ?, limit_times = ?, updated_at = ?
WHERE id = ?
`, normalized.Name, normalized.LimitType, normalized.PeriodType, normalized.Period, normalized.LimitNums, normalized.LimitTimes, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品购买数量限制策略：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.PurchaseLimitStrategyUpdateRes{}, nil
}

// Delete 删除购买数量限制策略（要求未被商品引用），并写入操作日志。
func (l *PurchaseLimitLogic) Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyDeleteRes, error) {
	strategy, err := l.getStrategy(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	goodsRefCount, err := countActiveGoodsReference(ctx, l.core.DB(), "purchase_limit_strategy_id", req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略删除校验失败")
	}
	if goodsRefCount > 0 {
		return nil, apiErr(consts.CodeConflict, "该商品购买数量限制策略已被商品引用，请先处理关联商品")
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_purchase_limit_strategy WHERE id = ?`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品购买数量限制策略：%d -> %s", req.ID, strategy.Name), ip)
	return &adminapi.PurchaseLimitStrategyDeleteRes{}, nil
}

// Status 切换购买数量限制策略启用/禁用状态，并写入操作日志。
func (l *PurchaseLimitLogic) Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyStatusRes, error) {
	strategy, err := l.getStrategy(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_purchase_limit_strategy
SET status = ?, updated_at = ?
WHERE id = ?
`, req.Status, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略状态更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换商品购买数量限制策略状态：%d -> %s -> %d", req.ID, strategy.Name, req.Status), ip)
	return &adminapi.PurchaseLimitStrategyStatusRes{}, nil
}
