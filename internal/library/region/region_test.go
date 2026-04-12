package region

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRegionResolver_LoadsXDBAndResolvesPublicIP(t *testing.T) {
	t.Parallel()

	resolver := NewRegionResolver(DefaultDBPaths()...)
	require.NotEmpty(t, resolver.Resolve("8.8.8.8"))
}

func TestRegionResolver_ReturnsEmptyForPrivateOrMissingDatabase(t *testing.T) {
	t.Parallel()

	empty := NewRegionResolver(filepath.Join(t.TempDir(), "missing.xdb"))
	require.Equal(t, "", empty.Resolve("8.8.8.8"))

	resolver := NewRegionResolver(DefaultDBPaths()...)
	require.Equal(t, "", resolver.Resolve("127.0.0.1"))
}
