package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// CreateRule 新增充值账号风控规则，并记录操作日志。
func (l *RechargeRiskLogic) CreateRule(ctx context.Context, req *adminapi.RechargeRiskRuleCreateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleCreateRes, error) {
	normalized, err := normalizeRechargeRiskRuleInput(req.Account, req.GoodsKeyword, req.Reason, req.Status)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if err = l.ensureRechargeRiskRuleUnique(ctx, normalized.Account, normalized.GoodsKeyword, 0); err != nil {
		return nil, err
	}
	now := l.core.Now()
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO recharge_risk_rule (
    account, goods_keyword, reason, status, hit_count,
    created_by_id, created_by_name, updated_by_id, updated_by_name,
    is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, 0, ?, ?, ?, ?, 0, ?, ?)
	`, normalized.Account, normalized.GoodsKeyword, normalized.Reason, normalized.Status, actor.ID, actor.RealName, actor.ID, actor.RealName, now, now)
	if err != nil {
		if isDuplicateDBError(err) {
			return nil, apiErr(consts.CodeConflict, "相同充值账号和关键词的风控规则已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "风控规则新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增充值风控规则：%s / %s", normalized.Account, normalized.GoodsKeyword), ip)
	return &adminapi.RechargeRiskRuleCreateRes{ID: id}, nil
}

// UpdateRule 编辑充值账号风控规则，并保留累计拦截次数。
func (l *RechargeRiskLogic) UpdateRule(ctx context.Context, req *adminapi.RechargeRiskRuleUpdateReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleUpdateRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	if _, err := l.getRechargeRiskRule(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	normalized, err := normalizeRechargeRiskRuleInput(req.Account, req.GoodsKeyword, req.Reason, req.Status)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if err = l.ensureRechargeRiskRuleUnique(ctx, normalized.Account, normalized.GoodsKeyword, req.ID); err != nil {
		return nil, err
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET account = ?, goods_keyword = ?, reason = ?, status = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
	WHERE id = ? AND is_deleted = 0
	`, normalized.Account, normalized.GoodsKeyword, normalized.Reason, normalized.Status, actor.ID, actor.RealName, l.core.Now(), req.ID); err != nil {
		if isDuplicateDBError(err) {
			return nil, apiErr(consts.CodeConflict, "相同充值账号和关键词的风控规则已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "风控规则编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑充值风控规则：%d", req.ID), ip)
	return &adminapi.RechargeRiskRuleUpdateRes{}, nil
}

// UpdateRuleStatus 启用或停用充值账号风控规则。
func (l *RechargeRiskLogic) UpdateRuleStatus(ctx context.Context, req *adminapi.RechargeRiskRuleStatusReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleStatusRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	if req.Status != rechargeRiskStatusDisabled && req.Status != rechargeRiskStatusEnabled {
		return nil, apiErr(consts.CodeBadRequest, "状态值错误")
	}
	if _, err := l.getRechargeRiskRule(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET status = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, req.Status, actor.ID, actor.RealName, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则状态修改失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("修改充值风控规则状态：%d -> %s", req.ID, rechargeRiskStatusText(req.Status)), ip)
	return &adminapi.RechargeRiskRuleStatusRes{}, nil
}

// DeleteRule 软删除充值账号风控规则，并归档关键词以避免唯一约束阻塞后续重建。
func (l *RechargeRiskLogic) DeleteRule(ctx context.Context, req *adminapi.RechargeRiskRuleDeleteReq, actor entity.AdminUser, ip string) (*adminapi.RechargeRiskRuleDeleteRes, error) {
	if req.ID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "规则ID不能为空")
	}
	rule, err := l.getRechargeRiskRule(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "风控规则不存在")
	}
	now := l.core.Now()
	if _, err = l.core.DB().Exec(ctx, `
UPDATE recharge_risk_rule
SET goods_keyword = ?, is_deleted = 1, deleted_at = ?, updated_by_id = ?, updated_by_name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, archivedRechargeRiskKeyword(rule.GoodsKeyword, req.ID), now, actor.ID, actor.RealName, now, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除充值风控规则：%d -> %s / %s", req.ID, rule.Account, rule.GoodsKeyword), ip)
	return &adminapi.RechargeRiskRuleDeleteRes{}, nil
}

func (l *RechargeRiskLogic) getRechargeRiskRule(ctx context.Context, id int64) (entity.RechargeRiskRule, error) {
	row := entity.RechargeRiskRule{}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT id, account, goods_keyword, reason, status, hit_count, created_by_id, created_by_name,
       updated_by_id, updated_by_name, is_deleted, deleted_at, created_at, updated_at
FROM recharge_risk_rule
WHERE id = ? AND is_deleted = 0
`, id)
	return row, err
}
