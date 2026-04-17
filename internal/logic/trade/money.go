package tradelogic

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Round4 统一将金额保留 4 位小数。
func Round4(v decimal.Decimal) decimal.Decimal {
	return v.Round(4)
}

// ParseMoney 将字符串解析为 decimal，并要求非空、格式正确。
func ParseMoney(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, fmt.Errorf("money empty")
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("money invalid")
	}
	return amount, nil
}

// MoneyString 返回统一 4 位小数的金额字符串。
func MoneyString(v decimal.Decimal) string {
	return Round4(v).StringFixed(4)
}
