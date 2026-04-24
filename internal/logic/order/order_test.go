package orderlogic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCanReorderRequiresSmartAndTimeoutWindow(t *testing.T) {
	created := time.Date(2026, 4, 24, 15, 0, 0, 0, time.Local)
	require.False(t, canReorder(reorderConfig{SmartEnabled: 0, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 0, TimeoutMinutes: 10}, created, created.Add(time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 0}, created, created.Add(time.Minute)))
	require.True(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(9*time.Minute)))
	require.False(t, canReorder(reorderConfig{SmartEnabled: 1, TimeoutEnabled: 1, TimeoutMinutes: 10}, created, created.Add(11*time.Minute)))
}

func TestSelectCandidateFixedAndLowestCost(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 20, CostPrice: "9.0000"},
		{BindingID: 11, Sort: 10, CostPrice: "12.0000"},
	}
	fixed := selectCandidate(candidates, map[int64]struct{}{}, "fixed_order", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 11, fixed.BindingID)
	lowest := selectCandidate(candidates, map[int64]struct{}{}, "lowest_cost", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 10, lowest.BindingID)
}

func TestSelectCandidateSkipsAttempted(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 10, CostPrice: "9.0000"},
		{BindingID: 11, Sort: 20, CostPrice: "12.0000"},
	}
	selected := selectCandidate(candidates, map[int64]struct{}{10: {}}, "fixed_order", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
	require.EqualValues(t, 11, selected.BindingID)
}

func TestSelectCandidateWeightedPercentSkipsZeroWeight(t *testing.T) {
	candidates := []orderChannelCandidate{
		{BindingID: 10, Sort: 10, CostPrice: "9.0000", OrderWeight: "0.0000"},
		{BindingID: 11, Sort: 20, CostPrice: "12.0000", OrderWeight: "100.0000"},
	}

	for range 100 {
		selected := selectCandidate(candidates, map[int64]struct{}{}, "weighted_percent", time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local))
		require.EqualValues(t, 11, selected.BindingID)
	}
}

func TestPollIntervalDurationUsesConfiguredSeconds(t *testing.T) {
	require.Equal(t, 7*time.Second, pollIntervalDuration(7))
	require.Equal(t, 30*time.Second, pollIntervalDuration(0))
}
