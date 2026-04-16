package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

const (
	purchaseLimitTypeMember  = 1
	purchaseLimitTypeAccount = 2

	purchaseLimitPeriodDay      = 1
	purchaseLimitPeriodInterval = 2
)

var purchaseLimitTypeItems = []entity.PurchaseLimitEnumItem{
	{ID: purchaseLimitTypeMember, Title: "同一会员"},
	{ID: purchaseLimitTypeAccount, Title: "同一充值账号"},
}

var purchaseLimitPeriodTypeItems = []entity.PurchaseLimitEnumItem{
	{ID: purchaseLimitPeriodDay, Title: "按天"},
	{ID: purchaseLimitPeriodInterval, Title: "按区间(分钟)"},
}

var purchaseLimitTypeTitles = enumTitleMap(purchaseLimitTypeItems)
var purchaseLimitPeriodTypeTitles = enumTitleMap(purchaseLimitPeriodTypeItems)

// Enums 返回购买数量限制策略相关枚举（限制类型、周期类型）。
func (l *PurchaseLimitLogic) Enums(ctx context.Context, _ *adminapi.PurchaseLimitStrategyEnumsReq) (*adminapi.PurchaseLimitStrategyEnumsRes, error) {
	limitTypes := append([]entity.PurchaseLimitEnumItem(nil), purchaseLimitTypeItems...)
	periodTypes := append([]entity.PurchaseLimitEnumItem(nil), purchaseLimitPeriodTypeItems...)
	return &adminapi.PurchaseLimitStrategyEnumsRes{
		LimitTypes:  limitTypes,
		PeriodTypes: periodTypes,
	}, nil
}

func enumTitleMap(items []entity.PurchaseLimitEnumItem) map[int]string {
	result := make(map[int]string, len(items))
	for _, item := range items {
		result[item.ID] = item.Title
	}
	return result
}
