package orderlogic

import (
	"testing"

	"myjob/internal/library/channelpricing"
	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/stretchr/testify/require"
)

func TestSegmentSafetyPriceTotalModeDisallowLoss(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "9.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "20.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeTotal, FieldName: "maxmoney"}},
		2,
		2,
	)

	require.NoError(t, err)
	require.Equal(t, "20.0000", result.Value)
	require.True(t, result.SendToSupplier)
}

func TestSegmentSafetyPriceUnitModeUsesUnitCeiling(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "20.0000"},
		reorderConfig{AllowLossSaleEnabled: 1, MaxLossAmount: "4.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "40.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnit, FieldName: "safe_cost"}},
		2,
		2,
	)

	require.NoError(t, err)
	require.Equal(t, "20.0000", result.Value)
	require.True(t, result.SendToSupplier)
}

func TestSegmentSafetyPriceUnsupportedRunsLocalGuard(t *testing.T) {
	result, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "9.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnsupported}},
		1,
		1,
	)

	require.NoError(t, err)
	require.False(t, result.SendToSupplier)
	require.Empty(t, result.Value)
}

func TestSegmentSafetyPriceUnsupportedRejectsLocalLoss(t *testing.T) {
	_, err := computeSegmentSafetyPrice(
		orderChannelCandidate{SourceCostPrice: "11.0000"},
		reorderConfig{AllowLossSaleEnabled: 0, MaxLossAmount: "0.0000"},
		channelpricing.OrderPriceSnapshot{OrderAmount: "10.0000"},
		supplierprovider.OrderProviderCapabilities{SafetyPrice: supplierprovider.SafetyPriceCapability{Mode: supplierprovider.SafetyPriceModeUnsupported}},
		1,
		1,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "防亏损")
}
