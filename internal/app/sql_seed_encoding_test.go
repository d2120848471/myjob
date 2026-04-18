package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUTF8SeedSQLFilesDeclareUTF8Connection(t *testing.T) {
	files := []string{
		"002_seed_menu.sql",
		"004_seed_config.sql",
		"005_supplier_platform.sql",
	}

	for _, name := range files {
		path := filepath.Join("..", "..", "manifest", "sql", name)
		content, err := os.ReadFile(path)
		require.NoError(t, err)

		trimmed := bytes.TrimSpace(content)
		require.Truef(
			t,
			bytes.HasPrefix(trimmed, []byte("SET NAMES utf8mb4;")),
			"%s must declare utf8mb4 before inserting Chinese seed text",
			name,
		)
	}
}
