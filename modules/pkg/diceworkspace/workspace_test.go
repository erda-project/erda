package diceworkspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWorkspaceByBranch(t *testing.T) {
	ws, err := GetByGitReference("bugfix/bugfix")
	require.Error(t, err)

	// PROD
	ws, err = GetByGitReference("master")
	require.NoError(t, err)
	require.Equal(t, PROD, ws)

	ws, err = GetByGitReference("support/2.13.1")
	require.NoError(t, err)
	require.Equal(t, PROD, ws)

	// STAGING
	ws, err = GetByGitReference("release/2.13.1")
	require.NoError(t, err)
	require.Equal(t, STAGING, ws)

	ws, err = GetByGitReference("hotfix/2.13.1-hotfix")
	require.NoError(t, err)
	require.Equal(t, STAGING, ws)

	// TEST
	ws, err = GetByGitReference("develop")
	require.NoError(t, err)
	require.Equal(t, TEST, ws)

	ws, err = GetByGitReference("develop-sonar")
	require.Error(t, err)

	// DEV
	ws, err = GetByGitReference("feature/gitflow")
	require.NoError(t, err)
	require.Equal(t, DEV, ws)

	ws, err = GetByGitReference("upgrade/1")
	require.Error(t, err)
}
