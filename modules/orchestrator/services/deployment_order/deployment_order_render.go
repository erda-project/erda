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

package deployment_order

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	preCheckPermissionDenied = "permission denied"
	parseErdaYamlError       = "parse erda.yml error"
	erdaYamlIsEmpty          = "erda.yml is empty"
	addonDoesNotSupportTmpl  = "addon %s (plan: %s) doesn't support"
	customAddonNotReadyTmpl  = "custom addon %s is not ready"
)

func (d *DeploymentOrder) RenderDetail(userId, releaseId, workspace string) (*apistructs.DeploymentOrderDetail, error) {
	releaseResp, err := d.bdl.GetRelease(releaseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s, err: %v", releaseId, err)
	}

	if access, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userId,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(releaseResp.ProjectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		return nil, apierrors.ErrRenderDeploymentOrderDetail.AccessDenied()
	}

	asi := make([]*apistructs.ApplicationInfo, 0)

	orderType := apistructs.TypeApplicationRelease

	if releaseResp.IsProjectRelease {
		params, err := d.fetchApplicationsParams(releaseResp, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		releasesId := make([]string, 0)
		for _, r := range releaseResp.ApplicationReleaseList {
			releasesId = append(releasesId, r.ReleaseID)
		}

		releases, err := d.db.ListReleases(releasesId)
		if err != nil {
			return nil, fmt.Errorf("failed to list release, err: %v", err)
		}

		releasesMap := make(map[string]*dbclient.Release)

		for _, r := range releases {
			releasesMap[r.ReleaseId] = r
		}

		for _, r := range releaseResp.ApplicationReleaseList {
			ret, ok := releasesMap[r.ReleaseID]
			if !ok {
				return nil, fmt.Errorf("failed to get releases %s from dicehub", r.ReleaseID)
			}

			asi = append(asi, &apistructs.ApplicationInfo{
				Id:       uint64(r.ApplicationID),
				Name:     r.ApplicationName,
				Params:   covertParamsType(params[r.ApplicationName]),
				DiceYaml: ret.DiceYaml,
			})
		}

		orderType = apistructs.TypeProjectRelease
	} else {
		params, err := d.fetchDeploymentParams(releaseResp.ApplicationID, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		asi = append(asi, &apistructs.ApplicationInfo{
			Id:       uint64(releaseResp.ApplicationID),
			Name:     releaseResp.ApplicationName,
			Params:   covertParamsType(params),
			DiceYaml: releaseResp.Diceyml,
		})
	}

	err = d.renderAppsPreCheckResult(releaseResp.ProjectID, userId, workspace, &asi)
	if err != nil {
		return nil, fmt.Errorf("failed to render application precheck result, err: %v", err)
	}

	orderId := uuid.NewString()

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			ID:   orderId,
			Name: utils.ParseOrderName(orderId),
			ReleaseInfo: &apistructs.ReleaseInfo{
				Id:        releaseResp.ReleaseID,
				Version:   releaseResp.Version,
				Type:      convertReleaseType(releaseResp.IsProjectRelease),
				Creator:   releaseResp.UserID,
				CreatedAt: releaseResp.CreatedAt,
				UpdatedAt: releaseResp.UpdatedAt,
			},
			Type:      orderType,
			Workspace: workspace,
		},
		ApplicationsInfo: asi,
	}, nil
}

func (d *DeploymentOrder) renderAppsPreCheckResult(projectId int64, userId, workspace string, asi *[]*apistructs.ApplicationInfo) error {
	if asi == nil {
		return nil
	}

	for _, info := range *asi {
		ret, err := d.preCheck(userId, workspace, projectId, info.Id, []byte(info.DiceYaml))
		if err != nil {
			return err
		}
		info.PreCheckResult = ret
	}

	return nil
}

func (d *DeploymentOrder) preCheck(userId, workspace string, projectId int64, appId uint64, erdaYaml []byte) (*apistructs.PreCheckResult, error) {
	failReasons := make([]string, 0)
	ret := &apistructs.PreCheckResult{
		Success: true,
	}

	// check execute permission
	isOk, err := d.checkExecutePermission(userId, workspace, appId)
	if err != nil {
		return nil, err
	}
	if !isOk {
		failReasons = append(failReasons, preCheckPermissionDenied)
	}

	if len(erdaYaml) == 0 {
		failReasons = append(failReasons, erdaYamlIsEmpty)
		return &apistructs.PreCheckResult{
			FailReasons: failReasons,
		}, nil
	}

	// parse erda yaml
	dy, err := diceyml.New(erdaYaml, true)
	if err != nil {
		failReasons = append(failReasons, parseErdaYamlError)
		return &apistructs.PreCheckResult{
			FailReasons: failReasons,
		}, nil
	}

	customAddonsMap := make(map[string]byte)
	customAddons, err := d.db.ListCustomInstancesByProjectAndEnv(projectId, strings.ToUpper(workspace))
	if err != nil {
		return nil, err
	}

	for _, v := range customAddons {
		customAddonsMap[v.Name] = 0
	}

	// check addon
	for instanceName, addOn := range dy.Obj().AddOns {
		plan := strutil.Split(addOn.Plan, ":", true)
		if len(plan) < 2 {
			plan = append(plan, apistructs.AddonDefaultPlan)
		}

		extensionI, ok := addon.AddonInfos.Load(plan[0])
		if !ok {
			// addon doesn't support
			failReasons = append(failReasons, fmt.Sprintf(addonDoesNotSupportTmpl, plan[0], plan[1]))
			continue
		}

		extension, ok := extensionI.(apistructs.Extension)
		if !ok {
			return nil, fmt.Errorf("failed to assert extension (%s) to Extension", plan[0])
		}

		if extension.Category == "custom" {
			_, ok := customAddonsMap[instanceName]
			if !ok {
				failReasons = append(failReasons, fmt.Sprintf(customAddonNotReadyTmpl, instanceName))
				continue
			}
		}
	}

	if len(failReasons) != 0 {
		ret.FailReasons = failReasons
		ret.Success = false
	}

	return ret, nil
}
