package tradelogic

import (
	"fmt"
)

// FulfillmentNo 生成履约分片编号：F001、F002...
func FulfillmentNo(index int) string {
	if index <= 0 {
		index = 1
	}
	return fmt.Sprintf("F%03d", index)
}

// PlanFulfillments 根据 provider 是否支持原生 quantity，输出履约分片规划。
func PlanFulfillments(quantity int, providerSupportsNativeQuantity bool) ([]FulfillmentPlanItem, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity错误")
	}
	if providerSupportsNativeQuantity {
		return []FulfillmentPlanItem{{FulfillmentNo: FulfillmentNo(1), AttemptQuantity: quantity}}, nil
	}
	items := make([]FulfillmentPlanItem, 0, quantity)
	for i := 1; i <= quantity; i++ {
		items = append(items, FulfillmentPlanItem{FulfillmentNo: FulfillmentNo(i), AttemptQuantity: 1})
	}
	return items, nil
}
