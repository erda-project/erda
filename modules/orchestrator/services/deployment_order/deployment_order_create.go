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
	"github.com/erda-project/erda/modules/pkg/user"
)

func (d *DeploymentOrder) Create(req *apistructs.DeploymentOrderCreateRequest) (*apistructs.DeploymentOrderCreateResponse, error) {
	// get release info
	releaseResp, err := d.bdl.GetRelease(req.ReleaseId)
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", req.ReleaseId, err)
		return nil, err
	}

	// permission check
	if err := d.checkExecutePermission(req.Operator, req.Workspace, releaseResp); err != nil {
		return nil, apierrors.ErrCreateDeploymentOrder.InternalError(err)
	}

	// parse the type of deployment order
	orderType := parseOrderType(req.Type, releaseResp.IsProjectRelease)

	// compose deployment order
	order, err := d.composeDeploymentOrder(orderType, releaseResp, req.Workspace, user.ID(req.Operator))
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
		Name:            order.Name,
		Type:            order.Type,
		ReleaseId:       order.ReleaseId,
		ProjectId:       order.ProjectId,
		ProjectName:     order.ProjectName,
		ApplicationId:   order.ApplicationId,
		ApplicationName: order.ApplicationName,
		Status:          parseDeploymentOrderStatus(nil),
	}

	if req.AutoRun {
		executeDeployResp, err := d.executeDeploy(order, releaseResp)
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

	// permission check
	if err := d.checkExecutePermission(req.Operator, order.Workspace, nil, order.ReleaseId); err != nil {
		logrus.Errorf("failed to check execute permission, err: %v", err)
		return nil, apierrors.ErrCreateDeploymentOrder.InternalError(err)
	}

	order.Operator = user.ID(req.Operator)

	releaseResp, err := d.bdl.GetRelease(order.ReleaseId)
	if err != nil {
		logrus.Errorf("failed to get release, err: %v", err)
		return nil, err
	}

	if _, err := d.executeDeploy(order, releaseResp); err != nil {
		logrus.Errorf("failed to execute deploy, order id: %s, err: %v", req.DeploymentOrderId, err)
		return nil, err
	}

	return order, nil
}

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
		return nil, apierrors.ErrListDeploymentOrder.AccessDenied()
	}

	orderName, err := d.renderDeploymentOrderName(uint64(releaseResp.ProjectID), releaseId, releaseResp.IsProjectRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to render deployment order name, project: %d, release: %s, err: %v",
			releaseResp.ProjectID, releaseId, err)
	}

	asi := make([]*apistructs.ApplicationInfo, 0)

	if releaseResp.IsProjectRelease {
		params, err := d.fetchApplicationsParams(apistructs.TypeProjectRelease, releaseResp, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		for _, r := range releaseResp.ApplicationReleaseList {
			asi = append(asi, &apistructs.ApplicationInfo{
				Id:     uint64(r.ApplicationID),
				Name:   r.ApplicationName,
				Params: params[r.ApplicationName],
			})
		}

	} else {
		params, err := d.fetchDeploymentParams(releaseResp.ApplicationID, workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
		}

		asi = append(asi, &apistructs.ApplicationInfo{
			Id:     uint64(releaseResp.ApplicationID),
			Name:   releaseResp.ApplicationName,
			Params: params,
		})
	}

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			Name: orderName,
		},
		ApplicationsInfo: asi,
	}, nil
}

func (d *DeploymentOrder) executeDeploy(order *dbclient.DeploymentOrder, releaseResp *apistructs.ReleaseGetResponseData) (map[uint64]*apistructs.DeploymentCreateResponseDTO, error) {
	// compose runtime create requests
	rtCreateReqs, err := d.composeRuntimeCreateRequests(order, releaseResp)
	if err != nil {
		return nil, fmt.Errorf("failed to compose runtime create requests, err: %v", err)
	}

	deployResponse := make(map[uint64]*apistructs.DeploymentCreateResponseDTO)
	applicationsStatus := make(apistructs.DeploymentOrderStatusMap)

	// create runtimes
	for _, rtCreateReq := range rtCreateReqs {
		runtimeCreateResp, err := d.rt.Create(order.Operator, rtCreateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create runtime %s, cluster: %s, release id: %s, err: %v",
				rtCreateReq.Name, rtCreateReq.ClusterName, rtCreateReq.ReleaseID, err)
		}
		deployResponse[runtimeCreateResp.ApplicationID] = runtimeCreateResp
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

func (d *DeploymentOrder) composeDeploymentOrder(t string, r *apistructs.ReleaseGetResponseData, workspace string, operator user.ID) (*dbclient.DeploymentOrder, error) {
	order := &dbclient.DeploymentOrder{
		ID:          uuid.New().String(),
		Type:        t,
		ReleaseId:   r.ReleaseID,
		ProjectId:   uint64(r.ProjectID),
		ProjectName: r.ProjectName,
		Workspace:   workspace,
		Operator:    operator,
	}

	params, err := d.fetchApplicationsParams(t, r, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment params, err: %v", err)
	}

	paramsJson, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params, err: %v", err)
	}

	order.Params = string(paramsJson)

	switch t {
	case apistructs.TypePipeline:
		branch, ok := r.Labels[gitBranchLabel]
		if !ok {
			return nil, fmt.Errorf("failed to get release branch in release %s", r.ReleaseID)
		}

		order.Name = branch
		order.ApplicationId, order.ApplicationName = r.ApplicationID, r.ApplicationName
		return order, nil
	case apistructs.TypeApplicationRelease:
		order.ApplicationId, order.ApplicationName = r.ApplicationID, r.ApplicationName
	}

	orderName, err := d.renderDeploymentOrderName(uint64(r.ProjectID), r.ReleaseID, r.IsProjectRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to render deployment order name, err: %v", err)
	}

	order.Name = orderName

	return order, nil
}

func (d *DeploymentOrder) renderDeploymentOrderName(projectId uint64, releaseId string, isProjectRange bool) (string, error) {
	var (
		orderName  string
		namePrefix = appOrderPrefix
		orderType  = apistructs.TypeApplicationRelease
	)

	if isProjectRange {
		namePrefix = projectOrderPrefix
		orderType = apistructs.TypeProjectRelease
	}

	c, err := d.db.GetOrderCountByProject(projectId, orderType)
	if err != nil {
		return orderName, fmt.Errorf("count order in project %d error: %v", projectId, err)
	}

	return namePrefix + fmt.Sprintf(orderNameTmpl, releaseId, c), nil
}

func (d *DeploymentOrder) fetchApplicationsParams(t string, r *apistructs.ReleaseGetResponseData, workspace string) (map[string]*apistructs.DeploymentOrderParam, error) {
	ret := make(map[string]*apistructs.DeploymentOrderParam, 0)

	switch t {
	case apistructs.TypePipeline, apistructs.TypeApplicationRelease:
		params, err := d.fetchDeploymentParams(r.ApplicationID, workspace)
		if err != nil {
			return nil, err
		}
		ret[r.ApplicationName] = params
	case apistructs.TypeProjectRelease:
		for _, ar := range r.ApplicationReleaseList {
			params, err := d.fetchDeploymentParams(ar.ApplicationID, workspace)
			if err != nil {
				return nil, err
			}
			ret[ar.ApplicationName] = params
		}
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

func (d *DeploymentOrder) composeRuntimeCreateRequests(order *dbclient.DeploymentOrder, r *apistructs.ReleaseGetResponseData) ([]*apistructs.RuntimeCreateRequest, error) {
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

	t := order.Type

	switch t {
	case apistructs.TypePipeline, apistructs.TypeApplicationRelease:
		branch, ok := r.Labels[gitBranchLabel]
		if !ok {
			return nil, fmt.Errorf("failed to get release branch in release %s", r.ReleaseID)
		}

		rtCreateReq := &apistructs.RuntimeCreateRequest{
			Name:                branch,
			DeploymentOrderId:   deploymentOrderId,
			DeploymentOrderName: parseDeploymentOrderShowName(order.Name),
			ReleaseVersion:      r.Version,
			ReleaseID:           r.ReleaseID,
			Source:              apistructs.TypePipeline,
			Operator:            operator,
			ClusterName:         clusterName,
			Extra: apistructs.RuntimeCreateRequestExtra{
				OrgID:           orgId,
				ProjectID:       projectId,
				ApplicationID:   uint64(r.ApplicationID),
				ApplicationName: r.ApplicationName,
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

		if t == apistructs.TypeApplicationRelease {
			rtCreateReq.Name = order.ApplicationName
			rtCreateReq.Source = release
			rtCreateReq.Extra.DeployType = release
		}

		ret = append(ret, rtCreateReq)
	case apistructs.TypeProjectRelease:
		for _, ar := range r.ApplicationReleaseList {
			rtCreateReq := &apistructs.RuntimeCreateRequest{
				Name:                ar.ApplicationName,
				DeploymentOrderId:   deploymentOrderId,
				DeploymentOrderName: parseDeploymentOrderShowName(order.Name),
				ReleaseVersion:      r.Version,
				ReleaseID:           ar.ReleaseID,
				Source:              release,
				Operator:            operator,
				ClusterName:         clusterName,
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
			ret = append(ret, rtCreateReq)

			paramJson, err := json.Marshal(orderParams[ar.ApplicationName])
			if err != nil {
				return nil, err
			}

			rtCreateReq.Param = string(paramJson)
		}
	}

	return ret, nil
}

func parseOrderType(t string, isProjectRelease bool) string {
	var orderType string
	if t == apistructs.TypePipeline {
		orderType = apistructs.TypePipeline
	} else if isProjectRelease {
		orderType = apistructs.TypeProjectRelease
	} else {
		orderType = apistructs.TypeApplicationRelease
	}

	return orderType
}

func parseDeploymentOrderShowName(orderName string) string {
	if strings.HasPrefix(orderName, appOrderPrefix) || strings.HasPrefix(orderName, projectOrderPrefix) {
		nameSlice := strings.Split(orderName, "_")
		if len(nameSlice) != 3 {
			return orderName
		}
		if len(nameSlice[1]) < 6 {
			return orderName
		}
		nameSlice[1] = nameSlice[1][:6]
		return strings.Join(nameSlice, "_")
	} else {
		return orderName
	}
}
