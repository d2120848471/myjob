package orderlogic

import (
	"fmt"
	"strings"

	"myjob/internal/library/channelpricing"
	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/shopspring/decimal"
)

type segmentSafetyPrice struct {
	Value          string
	SendToSupplier bool
}

func computeSegmentSafetyPrice(candidate orderChannelCandidate, config reorderConfig, snapshot channelpricing.OrderPriceSnapshot, capabilities supplierprovider.OrderProviderCapabilities, orderQuantity, segmentQuantity int) (segmentSafetyPrice, error) {
	if orderQuantity <= 0 || segmentQuantity <= 0 {
		return segmentSafetyPrice{}, fmt.Errorf("购买数量必须大于0")
	}

	sourceUnit, err := decimal.NewFromString(strings.TrimSpace(candidate.SourceCostPrice))
	if err != nil || sourceUnit.IsNegative() {
		return segmentSafetyPrice{}, fmt.Errorf("原始进货价格式错误")
	}
	orderAmount, err := decimal.NewFromString(strings.TrimSpace(snapshot.OrderAmount))
	if err != nil || orderAmount.IsNegative() {
		return segmentSafetyPrice{}, fmt.Errorf("订单金额格式错误")
	}

	allowedLoss := decimal.Zero
	if config.AllowLossSaleEnabled == 1 {
		allowedLoss, err = decimal.NewFromString(strings.TrimSpace(config.MaxLossAmount))
		if err != nil || allowedLoss.IsNegative() {
			return segmentSafetyPrice{}, fmt.Errorf("允许亏本金额格式错误")
		}
	}

	orderQty := decimal.NewFromInt(int64(orderQuantity))
	segmentQty := decimal.NewFromInt(int64(segmentQuantity))
	segmentOrderAmount := orderAmount.Div(orderQty).Mul(segmentQty).Round(4)
	segmentAllowedLoss := allowedLoss.Div(orderQty).Mul(segmentQty).Round(4)
	segmentSourceTotal := sourceUnit.Mul(segmentQty).Round(4)
	ceiling := segmentOrderAmount.Add(segmentAllowedLoss).Round(4)
	if capabilities.SafetyPrice.Mode == supplierprovider.SafetyPriceModeUnsupported {
		// 没有上游防亏损字段的平台只能依赖本地快照，发现本地进价已超过允许上限时直接拒绝提交。
		if segmentSourceTotal.GreaterThan(ceiling) {
			return segmentSafetyPrice{}, fmt.Errorf("本地防亏损校验失败：进货总额%s超过允许上限%s", segmentSourceTotal.StringFixed(4), ceiling.StringFixed(4))
		}
		return segmentSafetyPrice{}, nil
	}

	value := segmentSourceTotal
	if ceiling.LessThan(value) {
		value = ceiling
	}
	if capabilities.SafetyPrice.Mode == supplierprovider.SafetyPriceModeUnit {
		value = value.Div(segmentQty).Round(4)
	}
	return segmentSafetyPrice{Value: value.StringFixed(4), SendToSupplier: true}, nil
}
