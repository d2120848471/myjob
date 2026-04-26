package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureSupplierProductPushSchemaCreatesTables(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS supplier_product_subscription`)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS product_goods_channel_price_change_log`)
	require.NoError(t, err)

	err = core.ensureSupplierProductPushSchema(ctx)
	require.NoError(t, err)

	subscriptionColumns := loadColumnNames(t, core, "supplier_product_subscription")
	require.Contains(t, subscriptionColumns, "provider_code")
	require.Contains(t, subscriptionColumns, "platform_account_id")
	require.Contains(t, subscriptionColumns, "supplier_goods_no")
	require.Contains(t, subscriptionColumns, "callback_url")
	require.Contains(t, subscriptionColumns, "status")
	require.Contains(t, subscriptionColumns, "last_action")
	require.Contains(t, subscriptionColumns, "last_error")
	require.Contains(t, subscriptionColumns, "request_snapshot")
	require.Contains(t, subscriptionColumns, "response_snapshot")
	require.Contains(t, subscriptionColumns, "subscribed_at")
	require.Contains(t, subscriptionColumns, "canceled_at")

	priceLogColumns := loadColumnNames(t, core, "product_goods_channel_price_change_log")
	require.Contains(t, priceLogColumns, "source")
	require.Contains(t, priceLogColumns, "provider_code")
	require.Contains(t, priceLogColumns, "platform_account_id")
	require.Contains(t, priceLogColumns, "binding_id")
	require.Contains(t, priceLogColumns, "goods_id")
	require.Contains(t, priceLogColumns, "old_source_cost_price")
	require.Contains(t, priceLogColumns, "new_source_cost_price")
	require.Contains(t, priceLogColumns, "old_effective_sell_price")
	require.Contains(t, priceLogColumns, "new_effective_sell_price")
	require.Contains(t, priceLogColumns, "change_amount")
	require.Contains(t, priceLogColumns, "description")
	require.Contains(t, priceLogColumns, "raw_payload")
}

func loadColumnNames(t *testing.T, core *Core, table string) []string {
	t.Helper()
	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	err := core.DB().GetCore().GetScan(context.Background(), &rows, `SHOW COLUMNS FROM `+table)
	require.NoError(t, err)

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Field)
	}
	return names
}
