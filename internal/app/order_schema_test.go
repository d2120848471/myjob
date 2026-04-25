package app

import (
	"context"
	"strings"
	"testing"

	modelconfig "myjob/internal/model/config"

	"github.com/stretchr/testify/require"
)

func TestExternalOrderSchemaContainsRequiredTablesAndIndexes(t *testing.T) {
	for _, schema := range []string{sqliteSchema, mysqlSchema} {
		require.Contains(t, schema, "external_order")
		require.Contains(t, schema, "external_order_attempt")
		require.Contains(t, schema, "uk_external_order_order_no")
		require.Contains(t, schema, "idx_external_order_status_poll")
		require.Contains(t, schema, "idx_external_order_attempt_order")
		require.Contains(t, schema, "supplier_us_order_no")
		require.Contains(t, schema, "platform_subject_id")
		require.Contains(t, schema, "platform_subject_name")
		require.Contains(t, schema, "request_snapshot")
		require.Contains(t, schema, "response_snapshot")
	}
}

func TestExternalOrderMySQLCommentsArePresent(t *testing.T) {
	require.Contains(t, mysqlSchema, "COMMENT='外部订单主表'")
	require.Contains(t, mysqlSchema, "COMMENT='外部订单渠道尝试表'")
	for _, column := range []string{"订单号", "充值账号", "订单状态", "上游商家单号", "上游订单号"} {
		require.Contains(t, mysqlSchema, column)
	}
}

func TestEnsureExternalOrderAttemptSchemaAddsSubjectSnapshotColumns(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `DROP TABLE IF EXISTS external_order_attempt`)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
CREATE TABLE external_order_attempt (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    order_no VARCHAR(40) NOT NULL,
    attempt_no INT NOT NULL,
    channel_binding_id BIGINT UNSIGNED NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    platform_account_name VARCHAR(128) NOT NULL DEFAULT '',
    provider_code VARCHAR(32) NOT NULL,
    supplier_goods_no VARCHAR(128) NOT NULL,
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '',
    supplier_us_order_no VARCHAR(64) NOT NULL,
    supplier_order_no VARCHAR(128) NOT NULL DEFAULT '',
    supplier_status VARCHAR(32) NOT NULL DEFAULT '',
    refund_status VARCHAR(32) NOT NULL DEFAULT '',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    receipt VARCHAR(512) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL,
    submitted_at DATETIME NULL,
    last_checked_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)`)
	require.NoError(t, err)

	require.NoError(t, core.ensureExternalOrderAttemptSchema(ctx))
	require.NoError(t, core.ensureExternalOrderAttemptSchema(ctx))

	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	require.NoError(t, core.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM external_order_attempt`))

	columnNames := make([]string, 0, len(rows))
	for _, row := range rows {
		columnNames = append(columnNames, row.Field)
	}
	require.Contains(t, columnNames, "platform_subject_id")
	require.Contains(t, columnNames, "platform_subject_name")
}

func TestOpenOrderDefaultConfigIsUsableForTests(t *testing.T) {
	cfg := modelconfig.Default()
	require.Equal(t, "test-open-order-token", cfg.OpenOrder.Token)
	require.Equal(t, 30, cfg.OpenOrder.PollIntervalSeconds)
	require.Equal(t, 5, cfg.OpenOrder.SubmitScanIntervalSeconds)
	require.False(t, cfg.OpenOrder.WorkerEnabled)
}

func TestOpenOrderConfigFallsBackWhenIntervalsAreNonPositive(t *testing.T) {
	cfg := modelconfig.Default()
	cfg.OpenOrder.PollIntervalSeconds = 0
	cfg.OpenOrder.SubmitScanIntervalSeconds = -1

	modelconfig.Normalize(&cfg)

	require.Equal(t, 30, cfg.OpenOrder.PollIntervalSeconds)
	require.Equal(t, 5, cfg.OpenOrder.SubmitScanIntervalSeconds)
}

func TestExternalOrderSchemaHasNoRequestIPOrUserAgent(t *testing.T) {
	combined := strings.ToLower(sqliteSchema + mysqlSchema)
	require.NotContains(t, combined, "user_agent")
	require.NotContains(t, combined, "request_ip")
}
