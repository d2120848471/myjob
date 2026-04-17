package tradelogic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlanFulfillments_SupportsNativeQuantity(t *testing.T) {
	items, err := PlanFulfillments(3, true)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "F001", items[0].FulfillmentNo)
	require.Equal(t, 3, items[0].AttemptQuantity)
}

func TestPlanFulfillments_SplitByOne(t *testing.T) {
	items, err := PlanFulfillments(3, false)
	require.NoError(t, err)
	require.Len(t, items, 3)
	require.Equal(t, "F001", items[0].FulfillmentNo)
	require.Equal(t, 1, items[0].AttemptQuantity)
	require.Equal(t, "F002", items[1].FulfillmentNo)
	require.Equal(t, "F003", items[2].FulfillmentNo)
}
