package dbclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetBuildCache(t *testing.T) {
	cache, err := client.GetBuildCache("terminus-dev", "registry.marathon.l4lb.thisdcos.directory:5000/bc4f384766d395fe11bb97d5f2c9c72b/cidepcache:latest")
	require.NoError(t, err)
	require.True(t, cache.ID == 1)
}

func TestClient_DeleteBuildCache(t *testing.T) {
	require.NoError(t, client.DeleteBuildCache(2))
}
