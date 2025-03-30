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
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	infrai18n "github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/components/runtime"
	"github.com/erda-project/erda/internal/tools/orchestrator/i18n"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	I18nPermissionDeniedKey   = "DeployPermissionDenied"
	I18nFailedToParseErdaYaml = "FailedToParseErdaYaml"
	I18nEmptyErdaYaml         = "EmptyErdaYaml"
	I18nCustomAddonNotReady   = "CustomAddonNotReady"
	I18nAddonDoesNotExist     = "AddonDoesNotExist"
	I18nApplicationDeploying  = "ApplicationDeploying"
	I18nAddonPlanIllegal      = "AddonPlanIllegal"
	I18nAddonFormatIllegal    = "AddonFormatIllegal"
)

var (
	lang struct{ Lang string }
)

func (d *DeploymentOrder) RenderDetail(ctx context.Context, id, userId, releaseId, workspace string,
	projectId uint64, modes []string) (*apistructs.DeploymentOrderDetail, error) {
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	langCodes, _ := ctx.Value(lang).(infrai18n.LanguageCodes)

	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: releaseId})
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s, err: %v", releaseId, err)
	}

	if access, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userId,
		Scope:    apistructs.ProjectScope,
		ScopeID:  projectId,
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		return nil, apierrors.ErrRenderDeploymentOrderDetail.AccessDenied()
	}

	asi, err := d.composeAppsInfoByReleaseResp(projectId, userId, releaseResp.Data, workspace, modes)
	if err != nil {
		return nil, err
	}

	err = d.renderAppsPreCheckResult(langCodes, projectId, userId, workspace, &asi)
	if err != nil {
		return nil, fmt.Errorf("failed to render application precheck result, err: %v", err)
	}

	orderId := id
	if orderId == "" {
		orderId = uuid.NewString()
	}

	releaseData := releaseResp.GetData()

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			ID:   orderId,
			Name: utils.ParseOrderName(orderId),
			ReleaseInfo: &apistructs.ReleaseInfo{
				Id:        releaseData.GetReleaseID(),
				Version:   releaseData.GetVersion(),
				Type:      convertReleaseType(releaseData.GetIsProjectRelease()),
				Creator:   releaseData.GetUserID(),
				CreatedAt: releaseData.GetCreatedAt().AsTime(),
				UpdatedAt: releaseData.GetUpdatedAt().AsTime(),
			},
			Type:      parseOrderType(releaseData.GetIsProjectRelease()),
			Workspace: workspace,
		},
		ApplicationsInfo: asi,
	}, nil
}

func (d *DeploymentOrder) renderAppsPreCheckResult(langCodes infrai18n.LanguageCodes, projectId uint64, userId, workspace string, asi *[][]*apistructs.ApplicationInfo) error {
	if asi == nil {
		return nil
	}

	appList := make([]string, 0)
	for _, apps := range *asi {
		for _, info := range apps {
			appList = append(appList, info.Name)
		}
	}

	appStatus, err := d.getDeploymentsStatus(workspace, projectId, appList)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	for i := range *asi {
		for j := range (*asi)[i] {
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()
				app := (*asi)[i][j]
				failReasons, err := d.staticPreCheck(langCodes, userId, workspace, projectId, app.Id, []byte(app.DiceYaml))
				if err != nil {
					logrus.Errorf("failed to static pre check app %s, %v", app.Name, err)
					failReasons = append(failReasons, err.Error())
				}
				mux.Lock()
				isDeploying, ok := appStatus[app.Id]
				mux.Unlock()
				if ok && isDeploying {
					failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nApplicationDeploying, app.Name))
				}

				checkResult := &apistructs.PreCheckResult{
					Success: true,
				}

				if len(failReasons) != 0 {
					checkResult.Success = false
					checkResult.FailReasons = failReasons
				}

				app.PreCheckResult = checkResult
			}(i, j)
		}
	}
	wg.Wait()
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

func (d *DeploymentOrder) staticPreCheck(langCodes infrai18n.LanguageCodes, userId, workspace string,
	projectId, appId uint64, diceYml []byte) ([]string, error) {
	failReasons := make([]string, 0)

	// check execute permission
	isOk, err := d.checkExecutePermission(userId, workspace, appId)
	if err != nil {
		return nil, err
	}
	if !isOk {
		failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nPermissionDeniedKey))
	}

	if len(diceYml) == 0 {
		failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nEmptyErdaYaml))
		return failReasons, nil
	}

	// parse erda yaml
	dy, err := diceyml.New(diceYml, true)
	if err != nil {
		failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nFailedToParseErdaYaml))
		return failReasons, nil
	}

	customAddonsMap := make(map[string]byte)
	customAddons, err := d.db.ListCustomInstancesByProjectAndEnv(projectId, strings.ToUpper(workspace))
	if err != nil {
		return nil, err
	}

	for _, v := range customAddons {
		customAddonsMap[strings.ToUpper(v.Name)] = 0
	}

	// check addon
	for instanceName, addOn := range dy.Obj().AddOns {
		addonName, addonPlan, err := d.addon.ParseAddonFullPlan(addOn.Plan)
		if err != nil {
			failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nAddonFormatIllegal, addonName))
			continue
		}

		ok, err := d.addon.CheckDeployCondition(addonName, addonPlan, workspace)
		if err != nil {
			failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nAddonFormatIllegal, addonName))
			continue
		}

		if !ok {
			failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nAddonPlanIllegal, instanceName))
			continue
		}

		extensionI, ok := addon.AddonInfos.Load(addonName)
		if !ok {
			// addon doesn't support
			failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nAddonDoesNotExist, instanceName, addonName))
			continue
		}

		extension, ok := extensionI.(apistructs.Extension)
		if !ok {
			return nil, fmt.Errorf("failed to assert extension (%s) to Extension", addonName)
		}

		if extension.Category == apistructs.AddonCustomCategory {
			_, ok := customAddonsMap[strings.ToUpper(instanceName)]
			if !ok {
				failReasons = append(failReasons, i18n.LangCodesSprintf(langCodes, I18nCustomAddonNotReady, instanceName))
				continue
			}
		}
	}

	return failReasons, nil
}

func (d *DeploymentOrder) composeAppsInfoByReleaseResp(projectId uint64, userId string, releaseResp *pb.ReleaseGetResponseData,
	workspace string, modes []string) ([][]*apistructs.ApplicationInfo, error) {

	asi := make([][]*apistructs.ApplicationInfo, 0)

	switch parseOrderType(releaseResp.IsProjectRelease) {
	case apistructs.TypeProjectRelease:
		// mode exist check
		for _, mode := range modes {
			if _, ok := releaseResp.Modes[mode]; !ok {
				return nil, errors.Errorf("mode %s does not exist in release modes list", mode)
			}
		}
		// render deploy list with cross project
		deployList, err := d.renderDeployListWithCrossProject(modes, projectId, userId, releaseResp)
		if err != nil {
			return nil, err
		}

		params, err := d.fetchApplicationsParams(releaseResp, deployList, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		for _, batch := range deployList {
			ai := make([]*apistructs.ApplicationInfo, 0, len(batch))
			for _, r := range batch {
				appName := r.GetApplicationName()
				ai = append(ai, &apistructs.ApplicationInfo{
					Id:       uint64(r.GetApplicationID()),
					Name:     appName,
					Params:   covertParamsType(params[appName]),
					DiceYaml: r.GetDiceYml(),
				})
			}
			asi = append(asi, ai)
		}
	case apistructs.TypeApplicationRelease:
		// APPLICATION_RELEASE, application info from release
		params, err := d.fetchDeploymentParams(releaseResp.GetApplicationID(), workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		asi = append(asi, []*apistructs.ApplicationInfo{
			{
				Id:       uint64(releaseResp.GetApplicationID()),
				Name:     releaseResp.GetApplicationName(),
				Params:   covertParamsType(params),
				DiceYaml: releaseResp.GetDiceyml(),
			},
		})
	default:
	}

	return asi, nil
}
