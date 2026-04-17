package adminlogic

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

var allowedGoodsChannelRouteModes = map[string]struct{}{
	"fixed_order":       {},
	"lowest_cost_first": {},
	"weight_percent":    {},
	"time_period":       {},
	"random":            {},
}

func validateGoodsChannelRouteMode(routeMode string) error {
	routeMode = strings.TrimSpace(routeMode)
	if routeMode == "" {
		return fmt.Errorf("选路模式不能为空")
	}
	if _, ok := allowedGoodsChannelRouteModes[routeMode]; !ok {
		return fmt.Errorf("选路模式错误")
	}
	return nil
}

func normalizeGoodsChannelAttemptTimeout(enabled bool, minutes int) (int, error) {
	if !enabled {
		return 0, nil
	}
	if minutes <= 0 {
		return 0, fmt.Errorf("开启超时时必须填写分钟数")
	}
	return minutes, nil
}

func normalizeGoodsChannelMaxLossAmount(allowLoss bool, maxLossAmount string) (*string, error) {
	if !allowLoss {
		return nil, nil
	}
	maxLossAmount = strings.TrimSpace(maxLossAmount)
	if maxLossAmount == "" {
		return nil, nil
	}
	amount, err := decimal.NewFromString(maxLossAmount)
	if err != nil {
		return nil, fmt.Errorf("最大亏损额格式错误")
	}
	if amount.IsNegative() {
		return nil, fmt.Errorf("最大亏损额不能小于0")
	}
	normalized := amount.StringFixed(4)
	return &normalized, nil
}
