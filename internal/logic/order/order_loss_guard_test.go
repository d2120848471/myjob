package orderlogic

import (
	"testing"

	"myjob/internal/library/channelpricing"

	"github.com/stretchr/testify/require"
)

func TestKakayunMaxMoneyDisallowLossUsesLowerOfSourceTotalAndOrderAmount(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "9.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.NoError(t, err)
	require.Equal(t, "10.0000", maxMoney)
}

func TestKakayunMaxMoneyAllowLossAddsConfiguredTotalLoss(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "20.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "2.5000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.NoError(t, err)
	require.Equal(t, "12.5000", maxMoney)
}

func TestKakayunMaxMoneyCapsAtSourceTotalWhenSourceIsLower(t *testing.T) {
	maxMoney, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "5.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "30.0000"},
		2,
	)

	require.NoError(t, err)
	require.Equal(t, "22.0000", maxMoney)
}

func TestKakayunMaxMoneyRejectsInvalidMoney(t *testing.T) {
	_, err := kakayunMaxMoney(
		orderChannelCandidate{SourceCostPrice: "bad"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		1,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "原始进货价")
}
