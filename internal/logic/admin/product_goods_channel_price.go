package adminlogic

import (
	"fmt"
	"strings"

	modelruntime "myjob/internal/model/runtime"

	"github.com/shopspring/decimal"
)

const (
	taxAdjustDirectionNone           = "none"
	taxAdjustDirectionUntaxedToTaxed = "untaxed_to_taxed"
	taxAdjustDirectionTaxedToUntaxed = "taxed_to_untaxed"

	autoPriceAddTypeFixed   = "fixed"
	autoPriceAddTypePercent = "percent"
)

type productGoodsChannelCostSnapshot struct {
	SourceCostPrice    string
	CostPrice          string
	TaxAdjustDirection string
	TaxAdjustRate      string
	TaxAdjustAmount    string
}

var channelPricePercentBase = decimal.NewFromInt(100)

// computeChannelCostSnapshot 根据商品税态、渠道税态和系统税率配置，计算绑定比较成本价。
func computeChannelCostSnapshot(sourceCostPrice string, goodsHasTax, channelHasTax int, cfg modelruntime.FinanceTaxConfig) (productGoodsChannelCostSnapshot, error) {
	sourceAmount, err := decimal.NewFromString(strings.TrimSpace(sourceCostPrice))
	if err != nil {
		return productGoodsChannelCostSnapshot{}, fmt.Errorf("原始进货价格式错误: %w", err)
	}

	snapshot := productGoodsChannelCostSnapshot{
		SourceCostPrice:    sourceAmount.StringFixed(4),
		CostPrice:          sourceAmount.StringFixed(4),
		TaxAdjustDirection: taxAdjustDirectionNone,
		TaxAdjustRate:      "0.0000",
		TaxAdjustAmount:    "0.0000",
	}
	if goodsHasTax == channelHasTax {
		return snapshot, nil
	}

	rateText := cfg.TaxExclusiveRate
	direction := taxAdjustDirectionUntaxedToTaxed
	if goodsHasTax == 0 && channelHasTax == 1 {
		rateText = cfg.TaxInclusiveRate
		direction = taxAdjustDirectionTaxedToUntaxed
	}
	rate, err := decimal.NewFromString(strings.TrimSpace(rateText))
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return productGoodsChannelCostSnapshot{}, fmt.Errorf("税率未配置")
	}

	adjustAmount := sourceAmount.Mul(rate).Div(channelPricePercentBase).Round(4)
	costAmount := sourceAmount.Add(adjustAmount)
	if direction == taxAdjustDirectionTaxedToUntaxed {
		costAmount = sourceAmount.Sub(adjustAmount)
	}

	snapshot.CostPrice = costAmount.Round(4).StringFixed(4)
	snapshot.TaxAdjustDirection = direction
	snapshot.TaxAdjustRate = rate.Round(4).StringFixed(4)
	snapshot.TaxAdjustAmount = adjustAmount.StringFixed(4)
	return snapshot, nil
}

// computeChannelEffectiveSellPrice 按绑定自动改价规则，计算当前绑定的可售价格。
func computeChannelEffectiveSellPrice(defaultSellPrice, costPrice string, isAutoChange int, addType, defaultPrice string) (string, error) {
	if isAutoChange == 0 {
		if strings.TrimSpace(defaultSellPrice) == "" {
			return "", nil
		}
		basePrice, err := decimal.NewFromString(strings.TrimSpace(defaultSellPrice))
		if err != nil {
			return "", fmt.Errorf("商品默认售价格式错误: %w", err)
		}
		return basePrice.Round(4).StringFixed(4), nil
	}

	costAmount, err := decimal.NewFromString(strings.TrimSpace(costPrice))
	if err != nil {
		return "", fmt.Errorf("比较成本价格式错误: %w", err)
	}
	profitAmount, err := decimal.NewFromString(strings.TrimSpace(defaultPrice))
	if err != nil {
		return "", fmt.Errorf("利润值格式错误: %w", err)
	}

	switch strings.TrimSpace(addType) {
	case autoPriceAddTypeFixed:
		return costAmount.Add(profitAmount).Round(4).StringFixed(4), nil
	case autoPriceAddTypePercent:
		multiplier := decimal.NewFromInt(1).Add(profitAmount.Div(channelPricePercentBase))
		return costAmount.Mul(multiplier).Round(4).StringFixed(4), nil
	default:
		return "", fmt.Errorf("自动改价类型错误")
	}
}
