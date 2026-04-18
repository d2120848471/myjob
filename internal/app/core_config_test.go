package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	modelconfig "myjob/internal/model/config"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoadConfig_UsesFixedBootstrapAdminFromLocalConfig(t *testing.T) {
	t.Setenv("SUPER_ADMIN_PHONE", "")
	t.Setenv("SUPER_ADMIN_PASSWORD", "")

	_, _, cfg, err := loadConfig("../../manifest/config/config.local.yaml")
	require.NoError(t, err)

	require.Equal(t, "15881767197", cfg.Bootstrap.SuperAdminPhone)
	require.Equal(t, "abc123", cfg.Bootstrap.SuperAdminPassword)
}

func TestDefaultConfig_UsesLocalMySQLPort3306(t *testing.T) {
	cfg := modelconfig.Default()
	require.Equal(t, "root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local", cfg.Database.DSN)
}

func TestDefaultConfigPath_FindsRepoConfigFromNestedDir(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	nestedDir := filepath.Join(wd, "..", "..", "test", "integration")
	require.NoError(t, os.Chdir(nestedDir))
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	path, err := defaultConfigPath()
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(path))
	require.True(t, strings.HasSuffix(path, filepath.Join("manifest", "config", "config.local.yaml")))
}

func TestEnsureSuperAdmin_UpdatesExistingAdminCredentials(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	legacyHash, err := bcryptGenerate("legacy123")
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
UPDATE admin_user
SET phone = ?, password_hash = ?, updated_at = ?
WHERE username = ?
`, "13800000000", legacyHash, core.Now(), "admin")
	require.NoError(t, err)

	core.cfg.Bootstrap.SuperAdminPhone = "15881767197"
	core.cfg.Bootstrap.SuperAdminPassword = "abc123"
	require.NoError(t, core.ensureSuperAdmin(ctx))

	user, err := core.GetUserByUsername(ctx, "admin")
	require.NoError(t, err)
	require.Equal(t, "15881767197", user.Phone)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("abc123")))
}
