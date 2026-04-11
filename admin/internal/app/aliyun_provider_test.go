package app

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func TestNewApplicationFromConfig_AliyunProviderDoesNotUseMockSender(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cfg := defaultConfig()
	cfg.Database.Driver = "sqlite"
	cfg.Database.DSN = ":memory:"
	cfg.Redis.Addr = mr.Addr()
	cfg.Bootstrap.SuperAdminPhone = "13800000000"
	cfg.Bootstrap.SuperAdminPassword = "Admin_123"
	cfg.SMS.Provider = "aliyun"

	app, err := NewApplicationFromConfig(cfg)
	require.NoError(t, err)
	defer func() { require.NoError(t, app.Close()) }()

	_, ok := app.sender.(*MockSMSSender)
	require.False(t, ok, "aliyun provider should use real sender instead of mock sender")
}
