package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureProductGoodsSchema_AddsMissingSubjectIDColumn(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE product_goods`)
	require.NoError(t, err)

	_, err = core.DB().Exec(ctx, `
CREATE TABLE product_goods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    goods_code VARCHAR(32) NOT NULL,
    brand_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    goods_type VARCHAR(32) NOT NULL,
    supply_type VARCHAR(32) NOT NULL DEFAULT 'channel',
    is_export TINYINT NOT NULL DEFAULT 1,
    is_douyin TINYINT NOT NULL DEFAULT 0,
    has_tax TINYINT NOT NULL DEFAULT 0,
    exception_notify TINYINT NOT NULL DEFAULT 1,
    product_template_id BIGINT UNSIGNED NULL,
    purchase_limit_strategy_id BIGINT UNSIGNED NULL,
    purchase_notice TEXT NULL,
    terminal_price_limit DECIMAL(10, 4) NULL,
    balance_limit DECIMAL(10, 4) NOT NULL DEFAULT 0.0000,
    default_sell_price DECIMAL(10, 4) NULL,
    min_purchase_qty INT NOT NULL DEFAULT 1,
    max_purchase_qty INT NOT NULL DEFAULT 1,
    status TINYINT NOT NULL DEFAULT 1,
    is_deleted TINYINT NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_product_goods_code (goods_code)
)
`)
	require.NoError(t, err)

	err = core.ensureProductGoodsSchema(ctx)
	require.NoError(t, err)

	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	err = core.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM product_goods`)
	require.NoError(t, err)

	columnNames := make([]string, 0, len(rows))
	for _, row := range rows {
		columnNames = append(columnNames, row.Field)
	}
	require.Contains(t, columnNames, "subject_id")
}
