package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalConfigResolvedHost(t *testing.T) {
	require.Equal(t, "https://host.example.com", (&GlobalConfig{Host: "https://host.example.com"}).ResolvedHost())
	require.Equal(t, "https://server.example.com", (&GlobalConfig{Server: "https://server.example.com"}).ResolvedHost())
	require.Empty(t, (&GlobalConfig{}).ResolvedHost())
}

func TestSetAndGetGlobalConfigFromFile(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config")
	expected := &GlobalConfig{
		Version: ConfigVersion,
		Host:    "https://host.example.com",
	}

	err := SetGlobalConfig(configFile, expected)
	require.NoError(t, err)

	actual, err := GetGlobalConfigFrom(configFile)
	require.NoError(t, err)
	require.Equal(t, expected.Host, actual.Host)
	require.Equal(t, expected.Version, actual.Version)
}

func TestGetGlobalConfigFromSupportsLegacyServerField(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config")
	err := os.WriteFile(configFile, []byte("version: v0.0.1\nserver: https://legacy.example.com\n"), 0o644)
	require.NoError(t, err)

	actual, err := GetGlobalConfigFrom(configFile)
	require.NoError(t, err)
	require.Equal(t, "https://legacy.example.com", actual.ResolvedHost())
}
