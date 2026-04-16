package adminlogic

import (
	"fmt"
	"strings"
)

type normalizedPurchaseLimitInput struct {
	Name       string
	LimitType  int
	PeriodType int
	Period     int
	LimitNums  int
	LimitTimes int
}

func normalizePurchaseLimitInput(name string, limitType, periodType, period, limitNums, limitTimes int) (normalizedPurchaseLimitInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("策略名称不能为空")
	}
	if _, ok := purchaseLimitTypeTitles[limitType]; !ok {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制类型错误")
	}
	if _, ok := purchaseLimitPeriodTypeTitles[periodType]; !ok {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("周期类型错误")
	}
	if period <= 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制周期必须大于0")
	}
	// 数量和笔数允许填 0，语义是“不限制”，但不允许负数落库。
	if limitNums < 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制数量不能小于0")
	}
	if limitTimes < 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制笔数不能小于0")
	}
	return normalizedPurchaseLimitInput{
		Name:       name,
		LimitType:  limitType,
		PeriodType: periodType,
		Period:     period,
		LimitNums:  limitNums,
		LimitTimes: limitTimes,
	}, nil
}
