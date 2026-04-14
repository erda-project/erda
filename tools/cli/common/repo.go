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

// ResolveWorkspaceApplication resolves DOP project and application IDs from names under orgID.
// Use this for workspace commands instead of GetRepoStats when the git remote URL encodes the
// target org: GetRepoStats is proxied to gittar /wb/:project/:app, which resolves org from the
// OpenAPI session Org-ID (or host), not from the org derived from the remote, so the wrong app
// can be returned when project/app names exist under the session org.
func ResolveWorkspaceApplication(ctx *command.Context, orgID uint64, projectName, applicationName string) (projectID uint64, applicationID int64, err error) {
	projectID, err = GetProjectIDByName(ctx, orgID, projectName)
	if err != nil {
		return 0, 0, err
	}
	appID, err := GetApplicationIdByName(ctx, orgID, projectID, applicationName)
	if err != nil {
		return 0, 0, err
	}
	return projectID, int64(appID), nil
}

func GetRepoStats(ctx *command.Context, orgID uint64, project, application string) (apistructs.GittarStatsData, error) {
	var gitResp apistructs.GittarStatsResponse
	resp, err := ctx.Get().Path(fmt.Sprintf("/api/repo/%s/%s/stats/", project, application)).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().JSON(&gitResp)
	if err != nil {
		return apistructs.GittarStatsData{}, err
	}
	if !resp.IsOK() {
		return apistructs.GittarStatsData{}, fmt.Errorf("faild to find application stats, status code: %d", resp.StatusCode())
	}
	if !gitResp.Success {
		return apistructs.GittarStatsData{}, fmt.Errorf("failed to find application stats, %+v", gitResp.Error)
	}

	return gitResp.Data, nil
}
