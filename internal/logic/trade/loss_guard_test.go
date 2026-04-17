package tradelogic

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestEnsureLossAllowed_DisallowLoss(t *testing.T) {
	_, err := EnsureLossAllowed(false, nil, decimal.RequireFromString("110.0000"), decimal.RequireFromString("100.0000"))
	require.Error(t, err)
}

func TestEnsureLossAllowed_AllowLossWithinMax(t *testing.T) {
	maxLoss := decimal.RequireFromString("20.0000")
	loss, err := EnsureLossAllowed(true, &maxLoss, decimal.RequireFromString("110.0000"), decimal.RequireFromString("100.0000"))
	require.NoError(t, err)
	require.Equal(t, "10.0000", MoneyString(loss))
}

func TestEnsureLossAllowed_AllowLossExceedsMax(t *testing.T) {
	maxLoss := decimal.RequireFromString("5.0000")
	_, err := EnsureLossAllowed(true, &maxLoss, decimal.RequireFromString("110.0000"), decimal.RequireFromString("100.0000"))
	require.Error(t, err)
}

func TestEnsureLossAllowed_NotLoss(t *testing.T) {
	loss, err := EnsureLossAllowed(false, nil, decimal.RequireFromString("90.0000"), decimal.RequireFromString("100.0000"))
	require.NoError(t, err)
	require.Equal(t, "-10.0000", MoneyString(loss))
}
