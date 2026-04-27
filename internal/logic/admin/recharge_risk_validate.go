package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"myjob/internal/consts"
)

const (
	rechargeRiskStatusDisabled = 0
	rechargeRiskStatusEnabled  = 1

	rechargeRiskAccountMaxRunes      = 255
	rechargeRiskGoodsKeywordMaxRunes = 255
	rechargeRiskReasonMaxRunes       = 512
)

type normalizedRechargeRiskRuleInput struct {
	Account      string
	GoodsKeyword string
	Reason       string
	Status       int
}

func normalizeRechargeRiskRuleInput(account, goodsKeyword, reason string, status int) (normalizedRechargeRiskRuleInput, error) {
	account = strings.TrimSpace(account)
	goodsKeyword = strings.TrimSpace(goodsKeyword)
	reason = strings.TrimSpace(reason)
	if account == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("充值账号不能为空")
	}
	if goodsKeyword == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("匹配关键词不能为空")
	}
	if reason == "" {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("风控原因不能为空")
	}
	if len([]rune(account)) > rechargeRiskAccountMaxRunes {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("充值账号不能超过%d个字符", rechargeRiskAccountMaxRunes)
	}
	if len([]rune(goodsKeyword)) > rechargeRiskGoodsKeywordMaxRunes {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("匹配关键词不能超过%d个字符", rechargeRiskGoodsKeywordMaxRunes)
	}
	if len([]rune(reason)) > rechargeRiskReasonMaxRunes {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("风控原因不能超过%d个字符", rechargeRiskReasonMaxRunes)
	}
	if status != rechargeRiskStatusDisabled && status != rechargeRiskStatusEnabled {
		return normalizedRechargeRiskRuleInput{}, fmt.Errorf("状态值错误")
	}
	return normalizedRechargeRiskRuleInput{Account: account, GoodsKeyword: goodsKeyword, Reason: reason, Status: status}, nil
}

func normalizeRechargeRiskStatusFilter(value string) (int, bool, error) {
	switch strings.TrimSpace(value) {
	case "", "-1":
		return 0, false, nil
	case "0":
		return rechargeRiskStatusDisabled, true, nil
	case "1":
		return rechargeRiskStatusEnabled, true, nil
	default:
		return 0, false, fmt.Errorf("状态筛选值错误")
	}
}

func rechargeRiskStatusText(status int) string {
	if status == rechargeRiskStatusEnabled {
		return "启用"
	}
	return "停用"
}

func archivedRechargeRiskKeyword(keyword string, id int64) string {
	suffix := fmt.Sprintf("#deleted#%d", id)
	maxPrefixLength := rechargeRiskGoodsKeywordMaxRunes - len([]rune(suffix))
	if maxPrefixLength < 0 {
		maxPrefixLength = 0
	}
	return limitRechargeRiskRunes(strings.TrimSpace(keyword), maxPrefixLength) + suffix
}

func limitRechargeRiskRunes(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}

func (l *RechargeRiskLogic) ensureRechargeRiskRuleUnique(ctx context.Context, account, keyword string, excludeID int64) error {
	conditions := []string{"account = ?", "goods_keyword = ?", "is_deleted = 0"}
	args := []any{account, keyword}
	if excludeID > 0 {
		conditions = append(conditions, "id <> ?")
		args = append(args, excludeID)
	}
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_rule WHERE `+strings.Join(conditions, " AND "), args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "风控规则重复校验失败")
	}
	if count.Int() > 0 {
		return apiErr(consts.CodeConflict, "相同充值账号和关键词的风控规则已存在")
	}
	return nil
}
