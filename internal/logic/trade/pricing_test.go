package tradelogic

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestLockSalePrice_NotAutoChange_UsesGoodsDefault(t *testing.T) {
	price, err := LockSalePrice(decimal.RequireFromString("99.0000"), CandidateBinding{
		CostPrice:    decimal.RequireFromString("10.0000"),
		IsAutoChange: false,
		AddType:      AddTypeFixed,
		DefaultPrice: decimal.RequireFromString("5.0000"),
	})
	require.NoError(t, err)
	require.Equal(t, "99.0000", MoneyString(price))
}

func TestLockSalePrice_FixedProfit(t *testing.T) {
	price, err := LockSalePrice(decimal.RequireFromString("99.0000"), CandidateBinding{
		CostPrice:    decimal.RequireFromString("10.0000"),
		IsAutoChange: true,
		AddType:      AddTypeFixed,
		DefaultPrice: decimal.RequireFromString("5.5000"),
	})
	require.NoError(t, err)
	require.Equal(t, "15.5000", MoneyString(price))
}

func TestLockSalePrice_PercentProfit(t *testing.T) {
	price, err := LockSalePrice(decimal.RequireFromString("99.0000"), CandidateBinding{
		CostPrice:    decimal.RequireFromString("100.0000"),
		IsAutoChange: true,
		AddType:      AddTypePercent,
		DefaultPrice: decimal.RequireFromString("10.0000"),
	})
	require.NoError(t, err)
	require.Equal(t, "110.0000", MoneyString(price))
}

func TestLockSalePrice_DefaultPriceMinusOne_TreatAsNoChange(t *testing.T) {
	price, err := LockSalePrice(decimal.RequireFromString("99.0000"), CandidateBinding{
		CostPrice:    decimal.RequireFromString("100.0000"),
		IsAutoChange: true,
		AddType:      AddTypeFixed,
		DefaultPrice: decimal.NewFromInt(-1),
	})
	require.NoError(t, err)
	require.Equal(t, "99.0000", MoneyString(price))
}

func TestLockSalePrice_InvalidAddType(t *testing.T) {
	_, err := LockSalePrice(decimal.RequireFromString("99.0000"), CandidateBinding{
		CostPrice:    decimal.RequireFromString("100.0000"),
		IsAutoChange: true,
		AddType:      "oops",
		DefaultPrice: decimal.RequireFromString("1.0000"),
	})
	require.Error(t, err)
}
