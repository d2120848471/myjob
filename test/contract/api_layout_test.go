package contract_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIProtocolLayout_IsFlatUnderAPIDirectory(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	require.NoDirExists(t, filepath.Join(root, "api", "admin"))
	for _, dir := range []string{"auth", "config", "group", "log", "subject", "user"} {
		require.DirExists(t, filepath.Join(root, "api", dir))
	}
}
