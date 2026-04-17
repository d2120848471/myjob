package tradelogic

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// LockSalePrice 基于首个候选绑定锁定对外销售单价。
//
// 规则：
// - 未开启自动改价：sale_price = goodsDefaultSellPrice
// - 固定利润：sale_price = cost_price + default_price
// - 百分比利润：sale_price = cost_price * (1 + default_price/100)
// - 兼容：default_price = -1 视为不改价（沿用 goodsDefaultSellPrice）
func LockSalePrice(goodsDefaultSellPrice decimal.Decimal, binding CandidateBinding) (decimal.Decimal, error) {
	if !binding.IsAutoChange {
		return Round4(goodsDefaultSellPrice), nil
	}
	if binding.DefaultPrice.Equal(decimal.NewFromInt(-1)) {
		return Round4(goodsDefaultSellPrice), nil
	}

	addType := strings.TrimSpace(binding.AddType)
	if addType == "" {
		addType = AddTypeFixed
	}
	switch addType {
	case AddTypeFixed:
		return Round4(binding.CostPrice.Add(binding.DefaultPrice)), nil
	case AddTypePercent:
		factor := decimal.NewFromInt(1).Add(binding.DefaultPrice.Div(decimal.NewFromInt(100)))
		return Round4(binding.CostPrice.Mul(factor)), nil
	default:
		return decimal.Zero, fmt.Errorf("add_type错误")
	}
}
