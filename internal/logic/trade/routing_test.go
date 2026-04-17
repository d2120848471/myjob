package tradelogic

import (
	"math/rand"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPickFirstBinding_FixedOrder(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 2, Sort: 20, CostPrice: decimal.NewFromInt(10)},
		{ID: 3, Sort: 10, CostPrice: decimal.NewFromInt(10)},
		{ID: 1, Sort: 10, CostPrice: decimal.NewFromInt(10)},
	}
	first, err := PickFirstBinding(RouteModeFixedOrder, bindings, now, nil)
	require.NoError(t, err)
	require.Equal(t, int64(1), first.ID)
}

func TestPickFirstBinding_LowestCostFirst(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 1, Sort: 10, CostPrice: decimal.RequireFromString("10.0000")},
		{ID: 2, Sort: 5, CostPrice: decimal.RequireFromString("9.0000")},
		{ID: 3, Sort: 1, CostPrice: decimal.RequireFromString("9.0000")},
	}
	first, err := PickFirstBinding(RouteModeLowestCostFirst, bindings, now, nil)
	require.NoError(t, err)
	// cost_price 相同 -> sort asc, id asc
	require.Equal(t, int64(3), first.ID)
}

func TestPickNextBinding_TimePeriod_FilterAndSort(t *testing.T) {
	now := time.Date(2026, 4, 17, 1, 30, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 1, Sort: 20, StartTime: "23:00", EndTime: "02:00"},
		{ID: 2, Sort: 10, StartTime: "10:00", EndTime: "12:00"},
		{ID: 3, Sort: 5, StartTime: "23:00", EndTime: "02:00"},
	}
	first, ok, err := PickNextBinding(RouteModeTimePeriod, bindings, now, nil, nil)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(3), first.ID)
}

func TestPickNextBinding_TimePeriod_NoMatch(t *testing.T) {
	now := time.Date(2026, 4, 17, 9, 0, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 1, Sort: 1, StartTime: "10:00", EndTime: "12:00"},
		{ID: 2, Sort: 2, StartTime: "23:00", EndTime: "02:00"},
	}
	_, ok, err := PickNextBinding(RouteModeTimePeriod, bindings, now, nil, nil)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestPickFirstBinding_WeightPercent_FiltersWeightZero(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 1, Weight: 0},
		{ID: 2, Weight: 5},
	}
	rng := rand.New(rand.NewSource(1))
	first, err := PickFirstBinding(RouteModeWeightPercent, bindings, now, rng)
	require.NoError(t, err)
	require.Equal(t, int64(2), first.ID)
}

func TestPickNextBinding_Random_ExcludeAttempted(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.Local)
	bindings := []CandidateBinding{
		{ID: 1},
		{ID: 2},
	}
	attempted := map[int64]struct{}{2: {}}
	rng := rand.New(rand.NewSource(2))
	next, ok, err := PickNextBinding(RouteModeRandom, bindings, now, attempted, rng)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(1), next.ID)
}
