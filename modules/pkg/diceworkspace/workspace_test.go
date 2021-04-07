// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package diceworkspace

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
)

func TestGetWorkspaceByBranch(t *testing.T) {
	rules := []*apistructs.BranchRule{
		{
			Rule:              "feature/*",
			IsProtect:         false,
			Workspace:         "DEV",
			ArtifactWorkspace: "DEV",
		},
		{
			Rule:              "develop",
			IsProtect:         false,
			Workspace:         "TEST",
			ArtifactWorkspace: "TEST",
		},
		{
			Rule:              "release/*",
			IsProtect:         true,
			Workspace:         "STAGING",
			ArtifactWorkspace: "STAGING",
		},
		{
			Rule:              "hotfix/*",
			IsProtect:         true,
			Workspace:         "STAGING",
			ArtifactWorkspace: "STAGING",
		},
		{
			Rule:              "support/*",
			IsProtect:         true,
			Workspace:         "PROD",
			ArtifactWorkspace: "PROD",
		},
		{
			Rule:              "master",
			IsProtect:         true,
			Workspace:         "PROD",
			ArtifactWorkspace: "PROD",
		},
	}
	ws, err := GetByGitReference("bugfix/bugfix", rules)
	require.Error(t, err)

	// PROD
	ws, err = GetByGitReference("master", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.ProdWorkspace, ws)

	ws, err = GetByGitReference("support/2.13.1", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.ProdWorkspace, ws)

	// STAGING
	ws, err = GetByGitReference("release/2.13.1", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.StagingWorkspace, ws)

	ws, err = GetByGitReference("hotfix/2.13.1-hotfix", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.StagingWorkspace, ws)

	// TEST
	ws, err = GetByGitReference("develop", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.TestWorkspace, ws)

	ws, err = GetByGitReference("develop-sonar", rules)
	require.Error(t, err)

	// DEV
	ws, err = GetByGitReference("feature/gitflow", rules)
	require.NoError(t, err)
	require.Equal(t, apistructs.DevWorkspace, ws)

	ws, err = GetByGitReference("upgrade/1", rules)
	require.Error(t, err)
}
