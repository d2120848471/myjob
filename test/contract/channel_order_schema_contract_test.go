package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelOrderSchemaTablesExistInTestSQLite(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	for _, table := range []string{
		"product_goods_channel_config",
		"product_goods_channel_binding",
		"trade_order",
		"trade_order_attempt",
		"provider_callback_log",
		"provider_price_notify_log",
		"open_caller",
	} {
		count, err := h.app.Core().DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table)
		require.NoError(t, err)
		require.Equal(t, 1, count.Int(), table)
	}
}

func TestChannelOrderSchemaSeedsTradeTaxConfigKeys(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := context.Background()

	for _, key := range []string{
		"trade.tax.untaxed_to_taxed_rate",
		"trade.tax.taxed_to_untaxed_rate",
	} {
		count, err := h.app.Core().DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key)
		require.NoError(t, err)
		require.Equal(t, 1, count.Int(), key)
	}
}
