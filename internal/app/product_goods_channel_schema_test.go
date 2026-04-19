package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureProductGoodsChannelSchema_AddsInventoryColumnsAndConfigTable(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS product_goods_channel_config`)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS product_goods_channel_binding`)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
CREATE TABLE product_goods_channel_binding (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    goods_id BIGINT UNSIGNED NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    supplier_goods_no VARCHAR(128) NOT NULL,
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '',
    source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_direction VARCHAR(32) NOT NULL DEFAULT 'none',
    tax_adjust_rate DECIMAL(10,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    dock_status TINYINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    validate_template_id BIGINT UNSIGNED NULL,
    is_auto_change TINYINT NOT NULL DEFAULT 0,
    add_type VARCHAR(16) NOT NULL DEFAULT '',
    default_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    is_deleted TINYINT NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_product_goods_channel_binding_active (goods_id, platform_account_id, supplier_goods_no, is_deleted)
)
`)
	require.NoError(t, err)

	err = core.ensureProductGoodsChannelSchema(ctx)
	require.NoError(t, err)

	bindingColumns := make([]struct {
		Field string `db:"Field"`
	}, 0)
	err = core.DB().GetCore().GetScan(ctx, &bindingColumns, `SHOW COLUMNS FROM product_goods_channel_binding`)
	require.NoError(t, err)

	columnNames := make([]string, 0, len(bindingColumns))
	for _, row := range bindingColumns {
		columnNames = append(columnNames, row.Field)
	}
	require.Contains(t, columnNames, "order_weight")
	require.Contains(t, columnNames, "order_time_start")
	require.Contains(t, columnNames, "order_time_end")

	configColumns := make([]struct {
		Field string `db:"Field"`
	}, 0)
	err = core.DB().GetCore().GetScan(ctx, &configColumns, `SHOW COLUMNS FROM product_goods_channel_config`)
	require.NoError(t, err)

	configColumnNames := make([]string, 0, len(configColumns))
	for _, row := range configColumns {
		configColumnNames = append(configColumnNames, row.Field)
	}
	require.Contains(t, configColumnNames, "goods_id")
	require.Contains(t, configColumnNames, "smart_reorder_enabled")
	require.Contains(t, configColumnNames, "order_strategy")
	require.Contains(t, configColumnNames, "max_loss_amount")
}
