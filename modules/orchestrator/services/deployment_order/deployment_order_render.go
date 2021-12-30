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
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/i18n"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	I18nPermissionDeniedKey   = "DeployPermissionDenied"
	I18nFailedToParseErdaYaml = "FailedToParseErdaYaml"
	I18nEmptyErdaYaml         = "EmptyErdaYaml"
	I18nCustomAddonNotReady   = "CustomAddonNotReady"
	I18nAddonDoesNotExist     = "AddonDoesNotExist"
	I18nApplicationDeploying  = "ApplicationDeploying"
)

func (d *DeploymentOrder) RenderDetail(ctx context.Context, orgId uint64, userId, releaseId, workspace string) (*apistructs.DeploymentOrderDetail, error) {
	releaseResp, err := d.releaseSvc.GetRelease(context.WithValue(ctx, httputil.InternalHeader, "true"), &pb.ReleaseGetRequest{ReleaseID: releaseId})
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s, err: %v", releaseId, err)
	}

	if access, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userId,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(releaseResp.Data.ProjectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		return nil, apierrors.ErrRenderDeploymentOrderDetail.AccessDenied()
	}

	asi, err := d.composeAppsInfoByReleaseResp(releaseResp.Data, workspace)
	if err != nil {
		return nil, err
	}

	err = d.renderAppsPreCheckResult(orgId, releaseResp.Data.ProjectID, userId, workspace, &asi)
	if err != nil {
		return nil, fmt.Errorf("failed to render application precheck result, err: %v", err)
	}

	orderId := uuid.NewString()

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			ID:   orderId,
			Name: utils.ParseOrderName(orderId),
			ReleaseInfo: &apistructs.ReleaseInfo{
				Id:        releaseResp.Data.ReleaseID,
				Version:   releaseResp.Data.Version,
				Type:      convertReleaseType(releaseResp.Data.IsProjectRelease),
				Creator:   releaseResp.Data.UserID,
				CreatedAt: releaseResp.Data.CreatedAt.AsTime(),
				UpdatedAt: releaseResp.Data.UpdatedAt.AsTime(),
			},
			Type:      parseOrderType(releaseResp.Data.IsProjectRelease),
			Workspace: workspace,
		},
		ApplicationsInfo: asi,
	}, nil
}

func (d *DeploymentOrder) renderAppsPreCheckResult(orgId uint64, projectId int64, userId, workspace string, asi *[][]*apistructs.ApplicationInfo) error {
	if asi == nil {
		return nil
	}

	appList := make([]string, 0)
	for _, apps := range *asi {
		for _, info := range apps {
			appList = append(appList, info.Name)
		}
	}

	appStatus, err := d.getDeploymentsStatus(workspace, uint64(projectId), appList)
	if err != nil {
		return err
	}

	for _, apps := range *asi {
		for _, info := range apps {
			failReasons, err := d.staticPreCheck(orgId, userId, workspace, projectId, info.Id, []byte(info.DiceYaml))
			if err != nil {
				return err
			}
			isDeploying, ok := appStatus[info.Id]
			if ok && isDeploying {
				failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nApplicationDeploying, info.Name))
			}

			checkResult := &apistructs.PreCheckResult{
				Success: true,
			}

			if len(failReasons) != 0 {
				checkResult.Success = false
				checkResult.FailReasons = failReasons
			}

			info.PreCheckResult = checkResult
		}
	}

	return nil
}

func (d *DeploymentOrder) getDeploymentsStatus(workspace string, projectId uint64, appList []string) (map[uint64]bool, error) {
	ret := make(map[uint64]bool)

	runtimes, err := d.db.ListRuntimesByAppsName(workspace, projectId, appList)
	if err != nil {
		return nil, err
	}

	for _, r := range *runtimes {
		if runtime.IsDeploying(r.DeploymentStatus) {
			ret[r.ApplicationID] = true
			continue
		}
		ret[r.ApplicationID] = false
	}

	return ret, nil
}

func (d *DeploymentOrder) staticPreCheck(orgId uint64, userId, workspace string, projectId int64, appId uint64, erdaYaml []byte) ([]string, error) {
	failReasons := make([]string, 0)

	// check execute permission
	isOk, err := d.checkExecutePermission(userId, workspace, appId)
	if err != nil {
		return nil, err
	}
	if !isOk {
		failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nPermissionDeniedKey))
	}

	if len(erdaYaml) == 0 {
		failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nEmptyErdaYaml))
		return failReasons, nil
	}

	// parse erda yaml
	dy, err := diceyml.New(erdaYaml, true)
	if err != nil {
		failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nFailedToParseErdaYaml))
		return failReasons, nil
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
			failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nAddonDoesNotExist, instanceName, plan[0]))
			continue
		}

		extension, ok := extensionI.(apistructs.Extension)
		if !ok {
			return nil, fmt.Errorf("failed to assert extension (%s) to Extension", plan[0])
		}

		if extension.Category == "custom" {
			_, ok := customAddonsMap[instanceName]
			if !ok {
				failReasons = append(failReasons, i18n.OrgUintSprintf(orgId, I18nCustomAddonNotReady, instanceName))
				continue
			}
		}
	}

	return failReasons, nil
}

func (d *DeploymentOrder) composeAppsInfoByReleaseResp(releaseResp *pb.ReleaseGetResponseData, workspace string) (
	[][]*apistructs.ApplicationInfo, error) {

	asi := make([][]*apistructs.ApplicationInfo, 0)
	if releaseResp.IsProjectRelease {
		params, err := d.fetchApplicationsParams(releaseResp, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		releasesId := make([]string, 0)
		for i := 0; i < len(releaseResp.ApplicationReleaseList); i++ {
			if releaseResp.ApplicationReleaseList[i] == nil {
				continue
			}
			for _, r := range releaseResp.ApplicationReleaseList[i].List {
				releasesId = append(releasesId, r.ReleaseID)
			}
		}

		releases, err := d.db.ListReleases(releasesId)
		if err != nil {
			return nil, fmt.Errorf("failed to list release, err: %v", err)
		}

		releasesMap := make(map[string]*dbclient.Release)

		for _, r := range releases {
			releasesMap[r.ReleaseId] = r
		}

		for _, batch := range releaseResp.ApplicationReleaseList {
			ai := make([]*apistructs.ApplicationInfo, 0)
			for _, r := range batch.List {
				ret, ok := releasesMap[r.ReleaseID]
				if !ok {
					return nil, fmt.Errorf("failed to get releases %s from dicehub", r.ReleaseID)
				}

				ai = append(ai, &apistructs.ApplicationInfo{
					Id:       uint64(r.ApplicationID),
					Name:     r.ApplicationName,
					Params:   covertParamsType(params[r.ApplicationName]),
					DiceYaml: ret.DiceYaml,
				})
			}
			asi = append(asi, ai)
		}
	} else {
		params, err := d.fetchDeploymentParams(releaseResp.ApplicationID, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		asi = append(asi, []*apistructs.ApplicationInfo{
			{
				Id:       uint64(releaseResp.ApplicationID),
				Name:     releaseResp.ApplicationName,
				Params:   covertParamsType(params),
				DiceYaml: releaseResp.Diceyml,
			},
		})
	}

	return asi, nil
}
