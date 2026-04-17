package tradelogic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewOrderNo_Format(t *testing.T) {
	now := time.Date(2026, 4, 17, 0, 0, 0, 0, time.Local)
	require.Equal(t, "TO20260417000123", NewOrderNo(now, 123))
}

func TestNewProviderRequestOrderNo_Format(t *testing.T) {
	require.Equal(t, "PRTO20260417000123F001A01", NewProviderRequestOrderNo("TO20260417000123", "F001", 1))
	require.Equal(t, "PRTO20260417000123F002A10", NewProviderRequestOrderNo("TO20260417000123", "F002", 10))
}
