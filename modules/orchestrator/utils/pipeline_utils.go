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

package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

// GenRedeployPipelineYaml gen pipeline.yml for redeploy
func GenRedeployPipelineYaml(runtimeID uint64) apistructs.PipelineYml {
	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{
			{{
				Type:    "dice-deploy-redeploy",
				Alias:   "dice-deploy-redeploy",
				Version: "1.0",
				Params: map[string]interface{}{
					"runtime_id": strconv.FormatUint(runtimeID, 10),
				},
			}},
			{{
				Type:    "dice-deploy-addon",
				Version: "1.0",
				Params: map[string]interface{}{
					"deployment_id": "${dice-deploy-redeploy:OUTPUT:deployment_id}",
				},
			}},
			{{
				Type:    "dice-deploy-service",
				Version: "1.0",
				Params: map[string]interface{}{
					"deployment_id": "${dice-deploy-redeploy:OUTPUT:deployment_id}",
				},
			}},
			{{
				Type:    "dice-deploy-domain",
				Version: "1.0",
				Params: map[string]interface{}{
					"deployment_id": "${dice-deploy-redeploy:OUTPUT:deployment_id}",
				},
			}},
		},
	}

	return yml
}

// GenCreateByReleasePipelineYaml gen pipeline.yml for create runtime by releaseID
func GenCreateByReleasePipelineYaml(releaseID string, workspaces []string) apistructs.PipelineYml {
	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{
			{},
			{},
			{},
			{},
		},
	}
	for _, workspace := range workspaces {
		yml.Stages[0] = append(yml.Stages[0], &apistructs.PipelineYmlAction{
			Type:    "dice-deploy-release",
			Alias:   fmt.Sprintf("dice-deploy-release-%s", workspace),
			Version: "1.0",
			Params: map[string]interface{}{
				"release_id": releaseID,
				"workspace":  workspace,
			},
		})
		yml.Stages[1] = append(yml.Stages[1], &apistructs.PipelineYmlAction{
			Type:    "dice-deploy-addon",
			Alias:   fmt.Sprintf("dice-deploy-addon-%s", workspace),
			Version: "1.0",
			Params: map[string]interface{}{
				"deployment_id": fmt.Sprintf("${dice-deploy-release-%s:OUTPUT:deployment_id}", workspace),
			},
		})
		yml.Stages[2] = append(yml.Stages[2], &apistructs.PipelineYmlAction{
			Type:    "dice-deploy-service",
			Alias:   fmt.Sprintf("dice-deploy-service-%s", workspace),
			Version: "1.0",
			Params: map[string]interface{}{
				"deployment_id": fmt.Sprintf("${dice-deploy-release-%s:OUTPUT:deployment_id}", workspace),
			},
		})
		yml.Stages[3] = append(yml.Stages[3], &apistructs.PipelineYmlAction{
			Type:    "dice-deploy-domain",
			Alias:   fmt.Sprintf("dice-deploy-domain-%s", workspace),
			Version: "1.0",
			Params: map[string]interface{}{
				"deployment_id": fmt.Sprintf("${dice-deploy-release-%s:OUTPUT:deployment_id}", workspace),
			},
		})
	}

	return yml
}

// FindCreatingRuntimesByRelease find those runtimes created through the release
func FindCreatingRuntimesByRelease(appID uint64, env string, ymlName string, bdl *bundle.Bundle) ([]apistructs.RuntimeSummaryDTO, error) {
	var result []apistructs.RuntimeSummaryDTO
	pipelinePageReq := apistructs.PipelinePageListRequest{
		Sources:  []apistructs.PipelineSource{apistructs.PipelineSourceDice},
		Statuses: []string{"Running"},
	}
	if appID != 0 {
		pipelinePageReq.MustMatchLabelsQueryParams = []string{"appID=" + strconv.FormatUint(appID, 10)}
	}
	if ymlName != "" {
		pipelinePageReq.YmlNames = []string{ymlName}
	}

	resp, err := bdl.PageListPipeline(pipelinePageReq)
	if err != nil {
		return nil, err
	}

	for _, v := range resp.Pipelines {
		if strings.Contains(v.YmlName, "dice-deploy-release") &&
			strings.ToUpper(v.Extra.DiceWorkspace) == strings.ToUpper(env) {
			result = append(result, apistructs.RuntimeSummaryDTO{
				RuntimeInspectDTO: apistructs.RuntimeInspectDTO{
					Name:         v.FilterLabels["branch"],
					Source:       apistructs.RELEASE,
					Status:       "Init",
					DeployStatus: apistructs.DeploymentStatusDeploying,
					ClusterName:  v.ClusterName,
					Extra: map[string]interface{}{"applicationId": v.FilterLabels["appID"], "buildId": v.ID,
						"workspace": v.Extra.DiceWorkspace, "commitId": v.Commit, "fakeRuntime": true},
					TimeCreated: *v.TimeBegin,
					CreatedAt:   *v.TimeBegin,
					UpdatedAt:   *v.TimeBegin,
				},
				LastOperateTime: *v.TimeBegin,
				LastOperator:    fmt.Sprintf("%v", v.Extra.RunUser.ID),
			})
		}
	}

	return result, nil
}
