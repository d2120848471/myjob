package channelpricing

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	// AddTypeFixed 表示渠道自动改价按固定利润加价。
	AddTypeFixed = "fixed"
	// AddTypePercent 表示渠道自动改价按成本百分比加价。
	AddTypePercent = "percent"
)

var percentBase = decimal.NewFromInt(100)

// Rule 表示单条商品渠道绑定的售价规则快照。
type Rule struct {
	DefaultSellPrice string
	CostPrice        string
	IsAutoChange     int
	AddType          string
	ProfitValue      string
}

// OrderPriceSnapshot 表示订单按选中渠道计算后的金额快照。
type OrderPriceSnapshot struct {
	UnitPrice    string
	OrderAmount  string
	CostAmount   string
	ProfitAmount string
}

// EffectiveSellPrice 按渠道绑定自动改价规则计算单件可售价格。
func EffectiveSellPrice(rule Rule) (string, error) {
	if rule.IsAutoChange == 0 {
		if strings.TrimSpace(rule.DefaultSellPrice) == "" {
			return "", nil
		}
		basePrice, err := decimal.NewFromString(strings.TrimSpace(rule.DefaultSellPrice))
		if err != nil {
			return "", fmt.Errorf("商品默认售价格式错误: %w", err)
		}
		return basePrice.Round(4).StringFixed(4), nil
	}

	costAmount, err := decimal.NewFromString(strings.TrimSpace(rule.CostPrice))
	if err != nil {
		return "", fmt.Errorf("比较成本价格式错误: %w", err)
	}
	profitAmount, err := decimal.NewFromString(strings.TrimSpace(rule.ProfitValue))
	if err != nil {
		return "", fmt.Errorf("利润值格式错误: %w", err)
	}

	switch strings.TrimSpace(rule.AddType) {
	case AddTypeFixed:
		return costAmount.Add(profitAmount).Round(4).StringFixed(4), nil
	case AddTypePercent:
		multiplier := decimal.NewFromInt(1).Add(profitAmount.Div(percentBase))
		return costAmount.Mul(multiplier).Round(4).StringFixed(4), nil
	default:
		return "", fmt.Errorf("自动改价类型错误")
	}
}

// OrderSnapshot 按选中渠道售价和数量计算订单金额、成本和利润快照。
func OrderSnapshot(rule Rule, quantity int) (OrderPriceSnapshot, error) {
	if quantity <= 0 {
		return OrderPriceSnapshot{}, fmt.Errorf("购买数量必须大于0")
	}
	unitPrice, err := EffectiveSellPrice(rule)
	if err != nil {
		return OrderPriceSnapshot{}, err
	}
	if strings.TrimSpace(unitPrice) == "" {
		unitPrice = "0.0000"
	}
	unitAmount, err := decimal.NewFromString(unitPrice)
	if err != nil {
		return OrderPriceSnapshot{}, fmt.Errorf("订单单价格式错误: %w", err)
	}
	costAmount, err := decimal.NewFromString(strings.TrimSpace(rule.CostPrice))
	if err != nil {
		return OrderPriceSnapshot{}, fmt.Errorf("比较成本价格式错误: %w", err)
	}

	quantityAmount := decimal.NewFromInt(int64(quantity))
	totalCost := costAmount.Mul(quantityAmount).Round(4)
	totalOrder := unitAmount.Mul(quantityAmount).Round(4)
	return OrderPriceSnapshot{
		UnitPrice:    unitAmount.Round(4).StringFixed(4),
		OrderAmount:  totalOrder.StringFixed(4),
		CostAmount:   totalCost.StringFixed(4),
		ProfitAmount: totalOrder.Sub(totalCost).Round(4).StringFixed(4),
	}, nil
}
