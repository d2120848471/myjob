package app

import (
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

func TestOpenOrderDefaultConfigIsUsableForTests(t *testing.T) {
	cfg := modelconfig.Default()
	require.Equal(t, "test-open-order-token", cfg.OpenOrder.Token)
	require.Equal(t, 30, cfg.OpenOrder.PollIntervalSeconds)
	require.Equal(t, 5, cfg.OpenOrder.SubmitScanIntervalSeconds)
	require.False(t, cfg.OpenOrder.WorkerEnabled)
}

func TestExternalOrderSchemaHasNoRequestIPOrUserAgent(t *testing.T) {
	combined := strings.ToLower(sqliteSchema + mysqlSchema)
	require.NotContains(t, combined, "user_agent")
	require.NotContains(t, combined, "request_ip")
}
