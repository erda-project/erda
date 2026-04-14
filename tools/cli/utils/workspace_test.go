package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGitRemoteURLWithHTTPS(t *testing.T) {
	u, err := parseGitRemoteURL("https://erda.cloud/erda/dop/erda-project/erda")
	require.NoError(t, err)
	require.Equal(t, "https", u.Scheme)
	require.Equal(t, "erda.cloud", u.Host)
	require.Equal(t, "/erda/dop/erda-project/erda", u.Path)
}

func TestParseGitRemoteURLWithSSHScpStyle(t *testing.T) {
	u, err := parseGitRemoteURL("git@github.com:iutx/erda.git")
	require.NoError(t, err)
	require.Equal(t, "ssh", u.Scheme)
	require.Equal(t, "github.com", u.Host)
	require.Equal(t, "/iutx/erda.git", u.Path)
}

func TestParseGitRemoteURLRejectsInvalidRemote(t *testing.T) {
	_, err := parseGitRemoteURL(":::")
	require.Error(t, err)
}
