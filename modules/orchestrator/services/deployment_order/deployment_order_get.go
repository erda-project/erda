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

	"github.com/erda-project/erda/apistructs"
)

func (d *DeploymentOrder) Get(orderId string) (*apistructs.DeploymentOrderDetail, error) {
	order, err := d.db.GetDeploymentOrder(orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment order, err: %v", err)
	}

	// get release info
	releaseResp, err := d.bdl.GetRelease(order.ReleaseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get release, err: %v", err)
	}

	// parse params
	var params map[string]apistructs.DeploymentOrderParam

	if err := json.Unmarshal([]byte(order.Params), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params, err: %v", err)
	}

	// parse status
	appsStatus := make(map[string]apistructs.DeploymentOrderStatusItem, 0)
	if order.Status != "" {
		if err := json.Unmarshal([]byte(order.Status), &appsStatus); err != nil {
			return nil, fmt.Errorf("failed to unmarshal applications status, err: %v", err)
		}
	}

	releases := make([]*apistructs.ReleaseGetResponseData, 0)

	switch order.Type {
	case apistructs.TypePipeline, apistructs.TypeApplicationRelease:
		releases = append(releases, releaseResp)
	case apistructs.TypeProjectRelease:
		for _, r := range releaseResp.ApplicationReleaseList {
			ret, err := d.bdl.GetRelease(r.ReleaseID)
			if err != nil {
				return nil, fmt.Errorf("failed to get release repsonse, err: %v", err)
			}
			releases = append(releases, ret)
		}
	default:
		return nil, fmt.Errorf("deployment order type %s is illegal", order.Type)
	}

	// compose applications info
	asi, err := composeApplicationsInfo(releases, params, appsStatus)
	if err != nil {
		return nil, err
	}

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			ID:        order.ID,
			Name:      order.Name,
			ReleaseID: order.ReleaseId,
			Type:      order.Type,
			Workspace: order.Workspace,
			Status:    parseDeploymentOrderStatus(appsStatus),
			Operator:  order.Operator.String(),
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		ApplicationsInfo: asi,
		ReleaseVersion:   releaseResp.Version,
	}, nil
}

func composeApplicationsInfo(releases []*apistructs.ReleaseGetResponseData, params map[string]apistructs.DeploymentOrderParam,
	appsStatus apistructs.DeploymentOrderStatusMap) ([]*apistructs.ApplicationInfo, error) {

	asi := make([]*apistructs.ApplicationInfo, 0)

	for _, subRelease := range releases {
		applicationName := subRelease.ApplicationName

		// parse deployment order
		orderParamsData := make([]*apistructs.DeploymentOrderParamData, 0)

		param, ok := params[applicationName]
		if ok {
			for _, env := range param.Env {
				if env.IsEncrypt {
					env.Value = ""
				}
				orderParamsData = append(orderParamsData, &apistructs.DeploymentOrderParamData{
					Key:        env.Key,
					Value:      env.Value,
					ConfigType: "ENV",
				})
			}

			for _, env := range param.File {
				if env.IsEncrypt {
					env.Value = ""
				}
				orderParamsData = append(orderParamsData, &apistructs.DeploymentOrderParamData{
					Key:        env.Key,
					Value:      env.Value,
					ConfigType: "FILE",
				})
			}
		}

		paramJson, err := json.Marshal(orderParamsData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal order param, err: %v", err)
		}

		var status apistructs.DeploymentStatus
		app, ok := appsStatus[subRelease.ApplicationName]
		if ok {
			status = app.DeploymentStatus
		}

		asi = append(asi, &apistructs.ApplicationInfo{
			Name:           applicationName,
			DeploymentId:   app.DeploymentID,
			Param:          string(paramJson),
			ReleaseVersion: subRelease.Version,
			Branch:         subRelease.Labels["gitBranch"],
			CommitId:       subRelease.Labels["gitCommitId"],
			DiceYaml:       subRelease.Diceyml,
			Status:         status,
		})
	}

	return asi, nil
}

func parseDeploymentOrderStatus(appStatus apistructs.DeploymentOrderStatusMap) apistructs.DeploymentOrderStatus {
	if appStatus == nil {
		return orderStatusWaitDeploy
	}

	status := make([]apistructs.DeploymentStatus, 0)
	for _, a := range appStatus {
		if a.DeploymentStatus == apistructs.DeploymentStatusWaitApprove ||
			a.DeploymentStatus == apistructs.DeploymentStatusInit ||
			a.DeploymentStatus == apistructs.DeploymentStatusWaiting ||
			a.DeploymentStatus == apistructs.DeploymentStatusDeploying {
			return apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusDeploying)
		}
		status = append(status, a.DeploymentStatus)
	}

	var isFailed bool

	for _, s := range status {
		if s == apistructs.DeploymentStatusCanceling ||
			s == apistructs.DeploymentStatusCanceled {
			return apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusCanceled)
		}
		if s == apistructs.DeploymentStatusFailed {
			isFailed = true
		}
	}

	if isFailed {
		return apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusFailed)
	}

	return apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusOK)
}
