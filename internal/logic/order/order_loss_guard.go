package orderlogic

import (
	"fmt"
	"strings"

	"myjob/internal/library/channelpricing"

	"github.com/shopspring/decimal"
)

// kakayunMaxMoney 按卡卡云“最大进货总金额”口径计算防亏本阈值。
func kakayunMaxMoney(candidate orderChannelCandidate, config reorderConfig, snapshot channelpricing.OrderPriceSnapshot, quantity int) (string, error) {
	if quantity <= 0 {
		return "", fmt.Errorf("购买数量必须大于0")
	}

	sourceUnit, err := decimal.NewFromString(strings.TrimSpace(candidate.SourceCostPrice))
	if err != nil || sourceUnit.IsNegative() {
		return "", fmt.Errorf("原始进货价格式错误")
	}
	orderAmount, err := decimal.NewFromString(strings.TrimSpace(snapshot.OrderAmount))
	if err != nil || orderAmount.IsNegative() {
		return "", fmt.Errorf("订单金额格式错误")
	}

	allowedLoss := decimal.Zero
	if config.AllowLossSaleEnabled == 1 {
		allowedLoss, err = decimal.NewFromString(strings.TrimSpace(config.MaxLossAmount))
		if err != nil || allowedLoss.IsNegative() {
			return "", fmt.Errorf("允许亏本金额格式错误")
		}
	}

	sourceTotal := sourceUnit.Mul(decimal.NewFromInt(int64(quantity))).Round(4)
	salesCeiling := orderAmount.Add(allowedLoss).Round(4)
	maxMoney := sourceTotal
	if salesCeiling.LessThan(maxMoney) {
		maxMoney = salesCeiling
	}
	if maxMoney.IsNegative() {
		return "", fmt.Errorf("最大进货金额格式错误")
	}
	return maxMoney.Round(4).StringFixed(4), nil
}
