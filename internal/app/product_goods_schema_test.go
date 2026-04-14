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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    goods_code TEXT NOT NULL UNIQUE,
    brand_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    goods_type TEXT NOT NULL,
    supply_type TEXT NOT NULL DEFAULT 'channel',
    is_export INTEGER NOT NULL DEFAULT 1,
    is_douyin INTEGER NOT NULL DEFAULT 0,
    has_tax INTEGER NOT NULL DEFAULT 0,
    exception_notify INTEGER NOT NULL DEFAULT 1,
    product_template_id INTEGER NULL,
    purchase_limit_strategy_id INTEGER NULL,
    purchase_notice TEXT NULL,
    terminal_price_limit TEXT NULL,
    balance_limit TEXT NOT NULL DEFAULT '0.0000',
    default_sell_price TEXT NULL,
    min_purchase_qty INTEGER NOT NULL DEFAULT 1,
    max_purchase_qty INTEGER NOT NULL DEFAULT 1,
    status INTEGER NOT NULL DEFAULT 1,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)
`)
	require.NoError(t, err)

	err = core.ensureProductGoodsSchema(ctx)
	require.NoError(t, err)

	rows := make([]struct {
		Name string `db:"name"`
	}, 0)
	err = core.DB().GetCore().GetScan(ctx, &rows, `PRAGMA table_info(product_goods)`)
	require.NoError(t, err)

	columnNames := make([]string, 0, len(rows))
	for _, row := range rows {
		columnNames = append(columnNames, row.Name)
	}
	require.Contains(t, columnNames, "subject_id")
}
