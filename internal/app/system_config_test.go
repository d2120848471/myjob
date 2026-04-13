package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFinanceTaxConfig_ParsesScaledRates(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	_, err = core.DB().Exec(ctx, `
INSERT INTO system_config (config_key, config_value, description, created_at, updated_at)
VALUES
    (?, ?, ?, ?, ?),
    (?, ?, ?, ?, ?)
`,
		"finance_tax_exclusive_rate", "4.5", "未税->含税税率", core.Now(), core.Now(),
		"finance_tax_inclusive_rate", "3.8", "含税->未税税率", core.Now(), core.Now(),
	)
	require.NoError(t, err)

	cfg, err := core.LoadFinanceTaxConfig(ctx)
	require.NoError(t, err)
	require.Equal(t, "4.5", cfg.TaxExclusiveRate)
	require.EqualValues(t, 45000, cfg.TaxExclusiveRateScaled)
	require.Equal(t, "3.8", cfg.TaxInclusiveRate)
	require.EqualValues(t, 38000, cfg.TaxInclusiveRateScaled)
}

func TestLoadSystemConfigGroup_ReturnsUnconfiguredIntegrationItem(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	state, err := core.LoadSystemConfigGroup(context.Background(), "integration")
	require.NoError(t, err)
	require.Equal(t, "integration", state.Group)
	require.Len(t, state.Items, 1)
	require.Equal(t, "robot_webhook_url", state.Items[0].Key)
	require.Equal(t, "url", state.Items[0].ValueType)
	require.False(t, state.Items[0].Configured)
	require.Empty(t, state.Items[0].Value)
}
