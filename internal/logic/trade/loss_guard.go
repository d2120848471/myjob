package tradelogic

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// CalcLossAmount 计算本次 attempt 的亏损额（>0 表示亏损）。
func CalcLossAmount(currentAttemptCostPrice decimal.Decimal, lockedSalePrice decimal.Decimal) decimal.Decimal {
	return Round4(currentAttemptCostPrice.Sub(lockedSalePrice))
}

// EnsureLossAllowed 按商品级配置进行亏本保护校验。
//
// - allowLoss=false 且 loss>0：拒绝
// - allowLoss=true 且 maxLoss!=nil 且 loss>maxLoss：拒绝
func EnsureLossAllowed(allowLoss bool, maxLossAmount *decimal.Decimal, currentAttemptCostPrice decimal.Decimal, lockedSalePrice decimal.Decimal) (decimal.Decimal, error) {
	loss := CalcLossAmount(currentAttemptCostPrice, lockedSalePrice)
	if loss.LessThanOrEqual(decimal.Zero) {
		return loss, nil
	}
	if !allowLoss {
		return loss, fmt.Errorf("亏本销售不允许")
	}
	if maxLossAmount != nil && loss.GreaterThan(*maxLossAmount) {
		return loss, fmt.Errorf("超过最大允许亏损额")
	}
	return loss, nil
}
