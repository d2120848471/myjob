package tradelogic

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTaxAdjust_SameTax(t *testing.T) {
	res, err := TaxAdjust(1, 1, decimal.RequireFromString("100.0000"), decimal.RequireFromString("13.0000"), decimal.RequireFromString("9.0000"))
	require.NoError(t, err)
	require.Equal(t, "100.0000", MoneyString(res.CostPrice))
	require.Equal(t, "none", res.TaxAdjustDirection)
}

func TestTaxAdjust_UntaxedToTaxed(t *testing.T) {
	res, err := TaxAdjust(1, 0, decimal.RequireFromString("100.0000"), decimal.RequireFromString("13.0000"), decimal.RequireFromString("9.0000"))
	require.NoError(t, err)
	require.Equal(t, "113.0000", MoneyString(res.CostPrice))
	require.Equal(t, "untaxed_to_taxed", res.TaxAdjustDirection)
	require.Equal(t, "13.0000", MoneyString(res.TaxAdjustRate))
	require.Equal(t, "13.0000", MoneyString(res.TaxAdjustAmount))
}

func TestTaxAdjust_TaxedToUntaxed(t *testing.T) {
	res, err := TaxAdjust(0, 1, decimal.RequireFromString("100.0000"), decimal.RequireFromString("13.0000"), decimal.RequireFromString("9.0000"))
	require.NoError(t, err)
	require.Equal(t, "91.0000", MoneyString(res.CostPrice))
	require.Equal(t, "taxed_to_untaxed", res.TaxAdjustDirection)
	require.Equal(t, "9.0000", MoneyString(res.TaxAdjustRate))
	require.Equal(t, "9.0000", MoneyString(res.TaxAdjustAmount))
}
