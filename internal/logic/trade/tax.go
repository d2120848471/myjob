package tradelogic

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// TaxAdjustResult 表示税态换算结果。
type TaxAdjustResult struct {
	CostPrice          decimal.Decimal
	TaxAdjustDirection string
	TaxAdjustRate      decimal.Decimal
	TaxAdjustAmount    decimal.Decimal
}

// TaxAdjust 根据商品税态与渠道税态计算比较成本价。
//
// 规则对齐：
// - 商品未税、渠道未税：cost_price = source_cost_price
// - 商品有税、渠道无税：cost_price = source_cost_price * (1 + untaxedToTaxedRate/100)
// - 商品有税、渠道有税：cost_price = source_cost_price
// - 商品无税、渠道有税：cost_price = source_cost_price * (1 - taxedToUntaxedRate/100)
func TaxAdjust(goodsHasTax int, channelHasTax int, sourceCostPrice decimal.Decimal, untaxedToTaxedRate decimal.Decimal, taxedToUntaxedRate decimal.Decimal) (TaxAdjustResult, error) {
	if goodsHasTax == channelHasTax {
		return TaxAdjustResult{
			CostPrice:          Round4(sourceCostPrice),
			TaxAdjustDirection: "none",
			TaxAdjustRate:      decimal.Zero,
			TaxAdjustAmount:    decimal.Zero,
		}, nil
	}

	if goodsHasTax == 1 && channelHasTax == 0 {
		adjust := sourceCostPrice.Mul(untaxedToTaxedRate).Div(decimal.NewFromInt(100))
		return TaxAdjustResult{
			CostPrice:          Round4(sourceCostPrice.Add(adjust)),
			TaxAdjustDirection: "untaxed_to_taxed",
			TaxAdjustRate:      Round4(untaxedToTaxedRate),
			TaxAdjustAmount:    Round4(adjust),
		}, nil
	}

	if goodsHasTax == 0 && channelHasTax == 1 {
		adjust := sourceCostPrice.Mul(taxedToUntaxedRate).Div(decimal.NewFromInt(100))
		return TaxAdjustResult{
			CostPrice:          Round4(sourceCostPrice.Sub(adjust)),
			TaxAdjustDirection: "taxed_to_untaxed",
			TaxAdjustRate:      Round4(taxedToUntaxedRate),
			TaxAdjustAmount:    Round4(adjust),
		}, nil
	}

	return TaxAdjustResult{}, fmt.Errorf("tax flags invalid")
}
