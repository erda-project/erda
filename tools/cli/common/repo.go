// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

var (
	resolveWorkspaceRepoStats = GetWorkspaceRepoStats
)

func GetWorkspaceRepoStats(ctx *command.Context, orgID uint64, projectName, applicationName string) (apistructs.GittarStatsData, error) {
	var resp apistructs.GittarStatsResponse

	response, err := ctx.Get().
		Path(fmt.Sprintf("/api/repo/%s/%s/stats", projectName, applicationName)).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().JSON(&resp)
	if err != nil {
		return apistructs.GittarStatsData{}, err
	}
	if !response.IsOK() {
		return apistructs.GittarStatsData{}, fmt.Errorf("failed to find application stats, status code: %d", response.StatusCode())
	}
	if !resp.Success {
		return apistructs.GittarStatsData{}, fmt.Errorf("failed to find application stats, %+v", resp.Error)
	}

	return resp.Data, nil
}

// ResolveWorkspaceApplication resolves DOP project and application IDs from workspace context.
// Resolution order:
//  1. current project config/context IDs when names match
//  2. repo stats by org/repo
//  3. return an explicit error without falling back to high-permission list APIs
func ResolveWorkspaceApplication(ctx *command.Context, orgID uint64, projectName, applicationName string) (projectID uint64, applicationID int64, err error) {
	if ctx != nil &&
		ctx.CurrentProject.ProjectID > 0 &&
		ctx.CurrentApplication.ApplicationID > 0 &&
		ctx.CurrentProject.Project == projectName &&
		ctx.CurrentApplication.Application == applicationName {
		return ctx.CurrentProject.ProjectID, int64(ctx.CurrentApplication.ApplicationID), nil
	}

	stats, err := resolveWorkspaceRepoStats(ctx, orgID, projectName, applicationName)
	if err == nil && stats.ProjectID > 0 && stats.ApplicationID > 0 {
		return stats.ProjectID, stats.ApplicationID, nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to resolve workspace application from local config or repo stats for %s/%s: %w", projectName, applicationName, err)
	}
	return 0, 0, fmt.Errorf("failed to resolve workspace application from local config or repo stats for %s/%s", projectName, applicationName)
}
