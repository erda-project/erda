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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/utils"
)

func (d *DeploymentOrder) Get(userId string, orderId string) (*apistructs.DeploymentOrderDetail, error) {
	order, err := d.db.GetDeploymentOrder(orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment order, err: %v", err)
	}

	if access, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userId,
		Scope:    apistructs.ProjectScope,
		ScopeID:  order.ProjectId,
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		return nil, apierrors.ErrListDeploymentOrder.AccessDenied()
	}

	// parse params
	var params map[string]apistructs.DeploymentOrderParam

	if err := json.Unmarshal([]byte(order.Params), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params, err: %v", err)
	}

	// parse status
	appsStatus := make(apistructs.DeploymentOrderStatusMap)
	if order.Status != "" {
		if err := json.Unmarshal([]byte(order.Status), &appsStatus); err != nil {
			return nil, fmt.Errorf("failed to unmarshal applications status, err: %v", err)
		}
	}

	curRelease, err := d.db.GetReleases(order.ReleaseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get release, err: %v", err)
	}

	releases := make([]*dbclient.Release, 0)

	switch order.Type {
	case apistructs.TypePipeline, apistructs.TypeApplicationRelease:
		releases = append(releases, curRelease)
	case apistructs.TypeProjectRelease:
		subReleasesId := make([]string, 0)
		if err := json.Unmarshal([]byte(curRelease.ApplicationReleaseList), &subReleasesId); err != nil {
			return nil, fmt.Errorf("failed to get sub release, err: %v", err)
		}

		subReleases, err := d.db.ListReleases(subReleasesId)
		if err != nil {
			return nil, fmt.Errorf("failed to list sub release, err: %v", err)
		}
		for _, subRelease := range subReleases {
			releases = append(releases, subRelease)
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
			ID:   order.ID,
			Name: utils.ParseOrderName(order.ID),
			ReleaseInfo: &apistructs.ReleaseInfo{
				Id:        order.ReleaseId,
				Version:   curRelease.Version,
				Type:      convertReleaseType(curRelease.IsProjectRelease),
				Creator:   curRelease.UserId,
				CreatedAt: curRelease.CreatedAt,
				UpdatedAt: curRelease.UpdatedAt,
			},
			Type:      order.Type,
			Workspace: order.Workspace,
			Status:    parseDeploymentOrderStatus(appsStatus),
			Operator:  order.Operator.String(),
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
			StartedAt: parseStartedTime(order.StartedAt),
		},
		ApplicationsInfo: asi,
	}, nil
}

func composeApplicationsInfo(releases []*dbclient.Release, params map[string]apistructs.DeploymentOrderParam,
	appsStatus apistructs.DeploymentOrderStatusMap) ([]*apistructs.ApplicationInfo, error) {

	asi := make([]*apistructs.ApplicationInfo, 0)

	for _, subRelease := range releases {
		applicationName := subRelease.ApplicationName

		// parse deployment order
		orderParamsData := make(apistructs.DeploymentOrderParam, 0)

		param, ok := params[applicationName]
		if ok {
			for _, data := range param {
				if data.Encrypt {
					data.Value = ""
				}
				orderParamsData = append(orderParamsData, &apistructs.DeploymentOrderParamData{
					Key:     data.Key,
					Value:   data.Value,
					Encrypt: data.Encrypt,
					Type:    convertConfigType(data.Type),
					Comment: data.Comment,
				})
			}
		}

		var status apistructs.DeploymentStatus
		app, ok := appsStatus[subRelease.ApplicationName]
		if ok {
			status = app.DeploymentStatus
		}

		labels := make(map[string]string)
		if err := json.Unmarshal([]byte(subRelease.Labels), &labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal release labels, err: %v", err)
		}

		asi = append(asi, &apistructs.ApplicationInfo{
			Id:             subRelease.ApplicationId,
			Name:           applicationName,
			DeploymentId:   app.DeploymentID,
			Params:         &orderParamsData,
			ReleaseId:      subRelease.ReleaseId,
			ReleaseVersion: subRelease.Version,
			Branch:         labels["gitBranch"],
			DiceYaml:       subRelease.DiceYaml,
			CommitId:       labels["gitCommitId"],
			Status:         status,
		})
	}

	return asi, nil
}

func parseDeploymentOrderStatus(appStatus apistructs.DeploymentOrderStatusMap) apistructs.DeploymentOrderStatus {
	if appStatus == nil || len(appStatus) == 0 {
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

func parseStartedTime(t time.Time) *time.Time {
	// TODO: equal started default time with unix zero
	if t.Year() < 2000 {
		return nil
	}
	return &t
}

func convertConfigType(configType string) string {
	if configType == "dice-file" || configType == "kv" {
		return configType
	}
	if configType == "FILE" {
		return "dice-file"
	}
	return "kv"
}

func convertReleaseType(isProjectRelease bool) string {
	if isProjectRelease {
		return apistructs.ReleaseTypeProject
	}
	return apistructs.ReleaseTypeApplication
}
