package contract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIProtocolLayout_IsFlatUnderAPIDirectory(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	require.DirExists(t, filepath.Join(root, "api"))
	require.NoDirExists(t, filepath.Join(root, "api", "admin"))
	for _, name := range []string{"auth.go", "brand.go", "common.go", "group.go", "industry.go", "log.go", "product_template.go", "purchase_limit.go", "settings.go", "settings_sms.go", "settings_system.go", "subject.go", "supplier_platform.go", "user.go"} {
		require.FileExists(t, filepath.Join(root, "api", name))
	}
	require.NoDirExists(t, filepath.Join(root, "internal", "kernel"))
}

func TestAPIProtocolLayout_HasNoLegacyPackagePathReferences(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	for _, relativePath := range []string{
		"README.md",
		filepath.Join("docs", "architecture.md"),
		filepath.Join("docs", "module-map.md"),
		filepath.Join("internal", "service", "interfaces.go"),
		filepath.Join("internal", "controller", "admin", "auth.go"),
		filepath.Join("internal", "logic", "admin", "auth.go"),
	} {
		content, err := os.ReadFile(filepath.Join(root, relativePath))
		require.NoError(t, err)
		require.NotContains(t, string(content), "api/admin/v1")
		require.NotContains(t, string(content), "myjob/api/admin/v1")
	}

	var goFiles []string
	err := filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, walkErr error) error {
		require.NoError(t, walkErr)
		if info != nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, goFiles)
	for _, path := range goFiles {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.False(t, strings.Contains(string(content), "myjob/api/admin/v1"), path)
	}
}
