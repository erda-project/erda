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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (d *DeploymentOrder) Get(ctx context.Context, userId string, orderId string) (*apistructs.DeploymentOrderDetail, error) {
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
	if order.StatusDetail != "" {
		if err := json.Unmarshal([]byte(order.StatusDetail), &appsStatus); err != nil {
			return nil, fmt.Errorf("failed to unmarshal applications status, err: %v", err)
		}
	}

	// get release info
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: order.ReleaseId})
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", order.ReleaseId, err)
		return nil, err
	}

	curRelease := releaseResp.GetData()

	var appsInfo [][]*apistructs.ApplicationInfo

	switch order.Type {
	case apistructs.TypeProjectRelease:
		appReleases, err := d.renderDeployListWithCrossProject(strings.Split(order.Modes, ","), order.ProjectId, userId, curRelease)
		if err != nil {
			return nil, err
		}
		appsInfo, err = composeApplicationsInfo(appReleases, params, appsStatus)
		if err != nil {
			return nil, err
		}
	case apistructs.TypeApplicationRelease:
		appsInfo, err = composeApplicationsInfo([][]*pb.ApplicationReleaseSummary{
			{
				{
					ReleaseID:       curRelease.GetReleaseID(),
					Version:         curRelease.GetVersion(),
					ApplicationID:   curRelease.GetApplicationID(),
					ApplicationName: curRelease.GetApplicationName(),
					DiceYml:         curRelease.GetDiceyml(),
				},
			},
		}, params, appsStatus)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid deployment order type: %s", order.Type)
	}

	return &apistructs.DeploymentOrderDetail{
		DeploymentOrderItem: apistructs.DeploymentOrderItem{
			ID:   order.ID,
			Name: utils.ParseOrderName(order.ID),
			ReleaseInfo: &apistructs.ReleaseInfo{
				Id:        order.ReleaseId,
				Type:      convertReleaseType(curRelease.GetIsProjectRelease()),
				Version:   curRelease.GetVersion(),
				Creator:   curRelease.GetUserID(),
				CreatedAt: curRelease.GetCreatedAt().AsTime(),
				UpdatedAt: curRelease.GetUpdatedAt().AsTime(),
			},
			Type:         order.Type,
			Workspace:    order.Workspace,
			BatchSize:    order.BatchSize,
			CurrentBatch: order.CurrentBatch,
			Status:       apistructs.DeploymentOrderStatus(order.Status),
			Operator:     order.Operator.String(),
			DeployList:   order.DeployList,
			CreatedAt:    order.CreatedAt,
			UpdatedAt:    order.UpdatedAt,
			StartedAt:    parseStartedTime(order.StartedAt),
		},
		ApplicationsInfo: appsInfo,
	}, nil
}

func composeApplicationsInfo(appReleases [][]*pb.ApplicationReleaseSummary, params map[string]apistructs.DeploymentOrderParam,
	appsStatus apistructs.DeploymentOrderStatusMap) ([][]*apistructs.ApplicationInfo, error) {
	asi := make([][]*apistructs.ApplicationInfo, 0)

	for _, releases := range appReleases {
		ai := make([]*apistructs.ApplicationInfo, 0, len(releases))
		for _, rs := range releases {
			// parse deployment order
			orderParamsData := make(apistructs.DeploymentOrderParam, 0)
			applicationName := rs.GetApplicationName()
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

			var status apistructs.DeploymentStatus = apistructs.DeployStatusWaitDeploy
			app, ok := appsStatus[applicationName]
			if ok {
				status = app.DeploymentStatus
			}

			ai = append(ai, &apistructs.ApplicationInfo{
				Id:             uint64(rs.GetApplicationID()),
				Name:           applicationName,
				DeploymentId:   app.DeploymentID,
				Params:         &orderParamsData,
				ReleaseId:      rs.GetReleaseID(),
				ReleaseVersion: rs.GetVersion(),
				DiceYaml:       rs.GetDiceYml(),
				Status:         utils.ParseDeploymentStatus(status),
			})
		}
		asi = append(asi, ai)
	}

	return asi, nil
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
