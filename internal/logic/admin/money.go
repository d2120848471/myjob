package adminlogic

import (
	"strings"

	"github.com/shopspring/decimal"
)

func formatMoney(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return value
	}
	return amount.StringFixed(4)
}
