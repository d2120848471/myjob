package adminlogic

import (
	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

func rechargeRiskRuleItemFromEntity(row entity.RechargeRiskRule) adminapi.RechargeRiskRuleItem {
	return adminapi.RechargeRiskRuleItem{
		ID:            row.ID,
		Account:       row.Account,
		GoodsKeyword:  row.GoodsKeyword,
		Reason:        row.Reason,
		Status:        row.Status,
		StatusText:    rechargeRiskStatusText(row.Status),
		HitCount:      row.HitCount,
		CreatedByName: row.CreatedByName,
		UpdatedByName: row.UpdatedByName,
		CreatedAt:     formatAppTime(row.CreatedAt),
		UpdatedAt:     formatAppTime(row.UpdatedAt),
	}
}

func rechargeRiskRecordItemFromEntity(row entity.RechargeRiskRecord) adminapi.RechargeRiskRecordItem {
	return adminapi.RechargeRiskRecordItem{
		ID:             row.ID,
		RuleID:         row.RuleID,
		OrderNo:        row.OrderNo,
		Account:        row.Account,
		MatchedKeyword: row.MatchedKeyword,
		GoodsCode:      row.GoodsCode,
		GoodsName:      row.GoodsName,
		Reason:         row.Reason,
		InterceptedAt:  formatAppTime(row.InterceptedAt),
	}
}
