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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/modules/pkg/user"
)

func (d *DeploymentOrder) Create(req *apistructs.DeploymentOrderCreateRequest) (*apistructs.DeploymentOrderCreateResponse, error) {
	// generate order id
	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	// parse release id
	releaseId, err := d.getReleaseIdFromReq(req)
	if err != nil {
		logrus.Errorf("failed to get release id, err: %v", err)
		return nil, err
	}

	// get release info
	releaseResp, err := d.bdl.GetRelease(releaseId)
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", releaseId, err)
		return nil, err
	}

	// permission check
	if err := d.batchCheckExecutePermission(req.Operator, req.Workspace, d.parseAppsInfoWithRelease(releaseResp)); err != nil {
		return nil, apierrors.ErrCreateDeploymentOrder.InternalError(err)
	}

	// compose deployment order
	order, err := d.composeDeploymentOrder(releaseResp, req)
	if err != nil {
		logrus.Errorf("failed to compose deployment order, error: %v", err)
		return nil, err
	}

	// save order to db
	if err := d.db.UpdateDeploymentOrder(order); err != nil {
		logrus.Errorf("failed to update deployment order, err: %v", err)
		return nil, err
	}

	createResp := &apistructs.DeploymentOrderCreateResponse{
		Id:              order.ID,
		Name:            utils.ParseOrderName(order.ID),
		Type:            order.Type,
		ReleaseId:       order.ReleaseId,
		ProjectId:       order.ProjectId,
		ProjectName:     order.ProjectName,
		ApplicationId:   order.ApplicationId,
		ApplicationName: order.ApplicationName,
		Status:          parseDeploymentOrderStatus(nil),
	}

	if req.AutoRun {
		executeDeployResp, err := d.executeDeploy(order, releaseResp, req.Source, parseRuntimeNameFromBranch(req))
		if err != nil {
			logrus.Errorf("failed to executeDeploy, err: %v", err)
			return nil, err
		}
		createResp.Deployments = executeDeployResp
	}

	return createResp, nil
}

func (d *DeploymentOrder) Deploy(req *apistructs.DeploymentOrderDeployRequest) (*dbclient.DeploymentOrder, error) {
	order, err := d.db.GetDeploymentOrder(req.DeploymentOrderId)
	if err != nil {
		logrus.Errorf("failed to get deployment order, err: %v", err)
		return nil, err
	}

	appsInfo, err := d.parseAppsInfoWithOrder(order)
	if err != nil {
		logrus.Errorf("failed to parse application info with order, err: %v", err)
		return nil, err
	}

	// permission check
	if err := d.batchCheckExecutePermission(req.Operator, order.Workspace, appsInfo); err != nil {
		logrus.Errorf("failed to check execute permission, err: %v", err)
		return nil, apierrors.ErrDeployDeploymentOrder.InternalError(err)
	}

	order.Operator = user.ID(req.Operator)

	releaseResp, err := d.bdl.GetRelease(order.ReleaseId)
	if err != nil {
		logrus.Errorf("failed to get release, err: %v", err)
		return nil, err
	}

	if _, err := d.executeDeploy(order, releaseResp, apistructs.SourceDeployCenter, false); err != nil {
		logrus.Errorf("failed to execute deploy, order id: %s, err: %v", req.DeploymentOrderId, err)
		return nil, err
	}

	return order, nil
}

func (d *DeploymentOrder) executeDeploy(order *dbclient.DeploymentOrder, releaseResp *apistructs.ReleaseGetResponseData,
	source string, isRuntimeNameFromBranch bool) (map[string]*apistructs.DeploymentCreateResponseDTO, error) {
	// compose runtime create requests
	rtCreateReqs, err := d.composeRuntimeCreateRequests(order, releaseResp, source, isRuntimeNameFromBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to compose runtime create requests, err: %v", err)
	}

	deployResponse := make(map[string]*apistructs.DeploymentCreateResponseDTO)
	applicationsStatus := make(apistructs.DeploymentOrderStatusMap)

	// create runtimes
	for _, rtCreateReq := range rtCreateReqs {
		runtimeCreateResp, err := d.rt.Create(order.Operator, rtCreateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create runtime %s, cluster: %s, release id: %s, err: %v",
				rtCreateReq.Name, rtCreateReq.ClusterName, rtCreateReq.ReleaseID, err)
		}
		deployResponse[rtCreateReq.Extra.ApplicationName] = runtimeCreateResp
		applicationsStatus[rtCreateReq.Extra.ApplicationName] = apistructs.DeploymentOrderStatusItem{
			DeploymentID:     runtimeCreateResp.DeploymentID,
			AppID:            runtimeCreateResp.ApplicationID,
			DeploymentStatus: apistructs.DeploymentStatusInit,
			RuntimeID:        runtimeCreateResp.RuntimeID,
		}
	}

	// marshal applications status
	jsonAppStatus, err := json.Marshal(applicationsStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal applications status, err: %v", err)
	}

	// update deployment order status
	order.StartedAt = time.Now()
	order.Status = string(jsonAppStatus)
	if err := d.db.UpdateDeploymentOrder(order); err != nil {
		return nil, err
	}

	return deployResponse, nil
}

func (d *DeploymentOrder) composeDeploymentOrder(release *apistructs.ReleaseGetResponseData,
	req *apistructs.DeploymentOrderCreateRequest) (*dbclient.DeploymentOrder, error) {
	var (
		orderId   = req.Id
		orderType = parseOrderType(release.IsProjectRelease)
		workspace = req.Workspace
	)

	order := &dbclient.DeploymentOrder{
		ID:          orderId,
		Type:        orderType,
		Workspace:   workspace,
		Operator:    user.ID(req.Operator),
		ProjectId:   uint64(release.ProjectID),
		ReleaseId:   release.ReleaseID,
		ProjectName: release.ProjectName,
	}

	params, err := d.fetchApplicationsParams(release, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
	}

	paramsJson, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params, err: %v", err)
	}

	order.Params = string(paramsJson)

	if orderType == apistructs.TypeApplicationRelease {
		order.ApplicationId, order.ApplicationName = release.ApplicationID, release.ApplicationName
	}

	return order, nil
}

func (d *DeploymentOrder) fetchApplicationsParams(r *apistructs.ReleaseGetResponseData, workspace string) (map[string]*apistructs.DeploymentOrderParam, error) {
	ret := make(map[string]*apistructs.DeploymentOrderParam, 0)

	if r.IsProjectRelease {
		for i := 0; i < len(r.ApplicationReleaseList); i++ {
			for _, ar := range r.ApplicationReleaseList[i] {
				params, err := d.fetchDeploymentParams(ar.ApplicationID, workspace)
				if err != nil {
					return nil, err
				}
				ret[ar.ApplicationName] = params
			}
		}
	} else {
		params, err := d.fetchDeploymentParams(r.ApplicationID, workspace)
		if err != nil {
			return nil, err
		}
		ret[r.ApplicationName] = params
	}

	return ret, nil
}

func (d *DeploymentOrder) fetchDeploymentParams(applicationId int64, workspace string) (*apistructs.DeploymentOrderParam, error) {
	configNsTmpl := "app-%d-%s"

	deployConfig, fileConfig, err := d.bdl.FetchDeploymentConfigDetail(fmt.Sprintf(configNsTmpl, applicationId, strings.ToUpper(workspace)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment config, err: %v", err)
	}

	params := make(apistructs.DeploymentOrderParam, 0)

	for _, c := range deployConfig {
		params = append(params, &apistructs.DeploymentOrderParamData{
			Key:     c.Key,
			Value:   c.Value,
			Type:    "ENV",
			Encrypt: c.Encrypt,
			Comment: c.Comment,
		})
	}

	for _, c := range fileConfig {
		params = append(params, &apistructs.DeploymentOrderParamData{
			Key:     c.Key,
			Value:   c.Value,
			Type:    "FILE",
			Encrypt: c.Encrypt,
			Comment: c.Comment,
		})
	}

	return &params, nil
}

func (d *DeploymentOrder) composeRuntimeCreateRequests(order *dbclient.DeploymentOrder, r *apistructs.ReleaseGetResponseData,
	source string, isRuntimeNameFromBranch bool) ([]*apistructs.RuntimeCreateRequest, error) {

	if order == nil || r == nil {
		return nil, fmt.Errorf("deployment order or release response data is nil")
	}

	var (
		ret       = make([]*apistructs.RuntimeCreateRequest, 0)
		projectId = uint64(r.ProjectID)
		orgId     = uint64(r.OrgID)
		workspace = order.Workspace
	)

	projectInfo, err := d.bdl.GetProject(projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project info, id: %d, err: %v", projectId, err)
	}

	// get cluster name with workspace
	clusterName, ok := projectInfo.ClusterConfig[workspace]
	if !ok {
		return nil, fmt.Errorf("cluster not found at workspace: %s", workspace)
	}

	// parse operator
	operator := order.Operator.String()

	// deployment order id
	deploymentOrderId := order.ID

	// parse params
	var orderParams map[string]*apistructs.DeploymentOrderParam
	if err := json.Unmarshal([]byte(order.Params), &orderParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params, err: %v", err)
	}

	if source != apistructs.SourceDeployPipeline {
		source = release
	}

	runtimeSource := apistructs.RuntimeSource(source)

	if r.IsProjectRelease {
		for i := 0; i < len(r.ApplicationReleaseList); i++ {
			for _, ar := range r.ApplicationReleaseList[i] {
				rtCreateReq := &apistructs.RuntimeCreateRequest{
					Name:              ar.ApplicationName,
					DeploymentOrderId: deploymentOrderId,
					ReleaseVersion:    r.Version,
					ReleaseID:         ar.ReleaseID,
					Source:            runtimeSource,
					Operator:          operator,
					ClusterName:       clusterName,
					Extra: apistructs.RuntimeCreateRequestExtra{
						OrgID:           orgId,
						ProjectID:       projectId,
						ApplicationName: ar.ApplicationName,
						ApplicationID:   uint64(ar.ApplicationID),
						DeployType:      release,
						Workspace:       workspace,
						BuildID:         0, // Deprecated
					},
					SkipPushByOrch: false,
				}

				paramJson, err := json.Marshal(orderParams[ar.ApplicationName])
				if err != nil {
					return nil, err
				}
				rtCreateReq.Param = string(paramJson)

				ret = append(ret, rtCreateReq)
			}
		}
	} else {
		rtCreateReq := &apistructs.RuntimeCreateRequest{
			Name:              order.ApplicationName,
			DeploymentOrderId: deploymentOrderId,
			ReleaseVersion:    r.Version,
			ReleaseID:         r.ReleaseID,
			Source:            runtimeSource,
			Operator:          operator,
			ClusterName:       clusterName,
			Extra: apistructs.RuntimeCreateRequestExtra{
				OrgID:           orgId,
				ProjectID:       projectId,
				ApplicationID:   uint64(r.ApplicationID),
				ApplicationName: r.ApplicationName,
				DeployType:      release,
				Workspace:       workspace,
				BuildID:         0, // Deprecated
			},
			SkipPushByOrch: false,
		}

		paramJson, err := json.Marshal(orderParams[r.ApplicationName])
		if err != nil {
			return nil, err
		}
		rtCreateReq.Param = string(paramJson)

		if isRuntimeNameFromBranch {
			branch, ok := r.Labels[gitBranchLabel]
			if !ok {
				return nil, fmt.Errorf("failed to get release branch in release %s", r.ReleaseID)
			}
			rtCreateReq.Name = branch
			rtCreateReq.Extra.DeployType = ""
		}

		ret = append(ret, rtCreateReq)
	}

	return ret, nil
}

func (d *DeploymentOrder) parseAppsInfoWithOrder(order *dbclient.DeploymentOrder) (map[int64]string, error) {
	ret := make(map[int64]string)
	switch order.Type {
	case apistructs.TypeProjectRelease:
		releaseResp, err := d.bdl.GetRelease(order.ReleaseId)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(releaseResp.ApplicationReleaseList); i++ {
			for _, r := range releaseResp.ApplicationReleaseList[i] {
				ret[r.ApplicationID] = r.ApplicationName
			}
		}
	default:
		ret[order.ApplicationId] = order.ApplicationName
	}
	return ret, nil
}

func (d *DeploymentOrder) parseAppsInfoWithRelease(releaseResp *apistructs.ReleaseGetResponseData) map[int64]string {
	ret := make(map[int64]string)
	if releaseResp.IsProjectRelease {
		for i := 0; i < len(releaseResp.ApplicationReleaseList); i++ {
			for _, r := range releaseResp.ApplicationReleaseList[i] {
				ret[r.ApplicationID] = r.ApplicationName
			}
		}
	} else {
		ret[releaseResp.ApplicationID] = releaseResp.ApplicationName
	}

	return ret
}

func (d *DeploymentOrder) getReleaseIdFromReq(req *apistructs.DeploymentOrderCreateRequest) (string, error) {
	if req.ReleaseId != "" {
		return req.ReleaseId, nil
	}

	if req.ReleaseName == "" {
		return "", fmt.Errorf("please specified release name or release id")
	}

	switch req.Type {
	case apistructs.TypeProjectRelease:
		if req.ProjectId == 0 {
			return "", fmt.Errorf("deploy project release must provide effective project id")
		}
		r, err := d.db.GetProjectReleaseByVersion(req.ReleaseName, req.ProjectId)
		if err != nil {
			return "", err
		}
		return r.ReleaseId, nil
	case apistructs.TypeApplicationRelease:
		if req.ReleaseName == "" {
			return "", fmt.Errorf("deploy application relealse must provide effective application name")
		}
		r, err := d.db.GetApplicationReleaseByVersion(req.ReleaseName, req.ApplicationName)
		if err != nil {
			return "", err
		}
		return r.ReleaseId, nil
	default:
		return "", fmt.Errorf("deploy type %s is not support", req.Type)
	}
}

func parseRuntimeNameFromBranch(r *apistructs.DeploymentOrderCreateRequest) bool {
	if r.Source != apistructs.SourceDeployPipeline || r.ReleaseId == "" {
		return false
	}
	// TODO: from pipeline build deploy but want runtime name is application name
	return true
}

func parseOrderType(isProjectRelease bool) string {
	if isProjectRelease {
		return apistructs.TypeProjectRelease
	}
	return apistructs.TypeApplicationRelease
}

func covertParamsType(param *apistructs.DeploymentOrderParam) *apistructs.DeploymentOrderParam {
	if param == nil {
		return param
	}
	for _, data := range *param {
		data.Type = convertConfigType(data.Type)
	}
	return param
}
