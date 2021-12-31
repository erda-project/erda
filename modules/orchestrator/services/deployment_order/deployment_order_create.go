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
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
)

func (d *DeploymentOrder) Create(req *apistructs.DeploymentOrderCreateRequest) error {
	// get release info
	releaseResp, err := d.bdl.GetRelease(req.ReleaseId)
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", req.ReleaseId, err)
		return err
	}

	// parse the type of deployment order
	orderType := parseOrderType(req.Type, releaseResp.IsProjectRelease)

	// compose deployment order
	order, err := d.composeDeploymentOrder(orderType, releaseResp, req.Workspace, user.ID(req.Operator))
	if err != nil {
		logrus.Errorf("failed to compose deployment order, error: %v", err)
		return err
	}

	// compose runtime create requests
	rtCreateReqs, err := d.composeRuntimeCreateRequests(order, releaseResp, req.Workspace)
	if err != nil {
		logrus.Errorf("failed to compose runtime create requests, err: %v", err)
		return err
	}

	// save order to db
	if err := d.db.GetOrCreateDeploymentOrder(order); err != nil {
		return err
	}

	// create runtimes
	for _, rtCreateReq := range rtCreateReqs {
		_, err := d.rt.Create(user.ID(req.Operator), rtCreateReq)
		if err != nil {
			logrus.Errorf("failed to create runtime %s, cluster: %s, release id: %s, err: %v",
				rtCreateReq.Name, rtCreateReq.ClusterName, rtCreateReq.ReleaseID, err)
			return err
		}
	}

	return nil
}

func (d *DeploymentOrder) composeDeploymentOrder(t string, r *apistructs.ReleaseGetResponseData, workspace string, operator user.ID) (*dbclient.DeploymentOrder, error) {
	order := &dbclient.DeploymentOrder{
		ID:          uuid.New().String(),
		Type:        t,
		ReleaseId:   r.ReleaseID,
		ProjectId:   uint64(r.ProjectID),
		ProjectName: r.ProjectName,
		Operator:    operator,
	}

	params, err := d.fetchDeploymentParams(t, r, workspace)
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
	case apistructs.TypeProjectRelease:
		order.Name = projectOrderPrefix
	case apistructs.TypeApplicationRelease:
		order.ApplicationId, order.ApplicationName = r.ApplicationID, r.ApplicationName
		order.Name = appOrderPrefix
	}

	c, err := d.db.GetOrderCountByProject(uint64(r.ProjectID), order.Type)
	if err != nil {
		return nil, fmt.Errorf("count order in project %s error: %v", r.ProjectName, err)
	}

	order.Name = order.Name + fmt.Sprintf(orderNameTmpl, r.ReleaseID, c)

	return order, nil
}

func (d *DeploymentOrder) fetchDeploymentParams(t string, r *apistructs.ReleaseGetResponseData, workspace string) (map[string]*apistructs.DeploymentOrderParam, error) {
	ret := make(map[string]*apistructs.DeploymentOrderParam, 0)

	switch t {
	case apistructs.TypePipeline, apistructs.TypeApplicationRelease:
		params, err := d.fetchDeploymentOrderParam(r.ApplicationID, workspace)
		if err != nil {
			return nil, err
		}
		ret[r.ApplicationName] = params
	case apistructs.TypeProjectRelease:
		for _, ar := range r.ApplicationReleaseList {
			params, err := d.fetchDeploymentOrderParam(ar.ApplicationID, workspace)
			if err != nil {
				return nil, err
			}
			ret[ar.ApplicationName] = params
		}
	}

	return ret, nil
}

func (d *DeploymentOrder) fetchDeploymentOrderParam(applicationId int64, workspace string) (*apistructs.DeploymentOrderParam, error) {
	configNsTmpl := "app-%d-%s"

	deployConfig, fileConfig, err := d.bdl.FetchDeploymentConfigDetail(fmt.Sprintf(configNsTmpl, applicationId, workspace))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment config, err: %v", err)
	}
	envParams := make([]apistructs.DeploymentOrderParamItem, 0)
	fileParams := make([]apistructs.DeploymentOrderParamItem, 0)

	for _, c := range deployConfig {
		envParams = append(envParams, apistructs.DeploymentOrderParamItem{
			Key:       c.Key,
			Value:     c.Value,
			IsEncrypt: c.Encrypt,
		})
	}

	for _, c := range fileConfig {
		fileParams = append(fileParams, apistructs.DeploymentOrderParamItem{
			Key:       c.Key,
			Value:     c.Value,
			IsEncrypt: c.Encrypt,
		})
	}

	return &apistructs.DeploymentOrderParam{
		Env:  envParams,
		File: fileParams,
	}, nil
}

func (d *DeploymentOrder) composeRuntimeCreateRequests(order *dbclient.DeploymentOrder, r *apistructs.ReleaseGetResponseData,
	workspace string) ([]*apistructs.RuntimeCreateRequest, error) {

	ret := make([]*apistructs.RuntimeCreateRequest, 0)

	projectId := uint64(r.ProjectID)
	orgId := uint64(r.OrgID)

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
			Name:              branch,
			DeploymentOrderId: deploymentOrderId,
			ReleaseID:         r.ReleaseID,
			Source:            apistructs.TypePipeline,
			Operator:          operator,
			ClusterName:       clusterName,
			Extra: apistructs.RuntimeCreateRequestExtra{
				OrgID:           orgId,
				ProjectID:       projectId,
				ApplicationID:   uint64(r.ApplicationID),
				ApplicationName: r.ApplicationName,
				Workspace:       workspace,
				BuildID:         0, // Deprecated
				DeployType:      release,
			},
			SkipPushByOrch: false,
		}

		paramJson, err := json.Marshal(orderParams[r.ApplicationName])
		if err != nil {
			return nil, err
		}

		rtCreateReq.Param = string(paramJson)

		if t == apistructs.TypeApplicationRelease {
			rtCreateReq.Source = release
		}

		ret = append(ret, rtCreateReq)
	case apistructs.TypeProjectRelease:
		for _, ar := range r.ApplicationReleaseList {
			rl, err := d.bdl.GetRelease(ar.ReleaseID)
			if err != nil {
				return nil, err
			}

			branch, ok := rl.Labels[gitBranchLabel]
			if !ok {
				return nil, fmt.Errorf("failed to get release branch in release %s", rl.ReleaseID)
			}

			rtCreateReq := &apistructs.RuntimeCreateRequest{
				Name:              branch,
				DeploymentOrderId: deploymentOrderId,
				ReleaseID:         ar.ReleaseID,
				Source:            release,
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
