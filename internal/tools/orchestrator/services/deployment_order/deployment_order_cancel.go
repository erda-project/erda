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

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (d *DeploymentOrder) Cancel(ctx context.Context, req *apistructs.DeploymentOrderCancelRequest) (*dbclient.DeploymentOrder, error) {
	order, err := d.db.GetDeploymentOrder(req.DeploymentOrderId)
	if err != nil {
		logrus.Errorf("failed to get order, id: %s, err: %v", req.DeploymentOrderId, err)
		return nil, err
	}

	// get release info
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: order.ReleaseId})
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", order.ReleaseId, err)
		return nil, err
	}

	releaseData := releaseResp.GetData()

	appsInfo := make(map[int64]string)
	switch order.Type {
	case apistructs.TypeProjectRelease:
		deployList, err := d.renderDeployListWithCrossProject(strings.Split(order.Modes, ","), order.ProjectId,
			req.Operator, releaseData)
		if err != nil {
			return nil, fmt.Errorf("failed to render deploy list with cross project, err: %v", err)
		}
		appsInfo = d.parseAppsInfoWithDeployList(deployList)
	case apistructs.TypeApplicationRelease:
		appsInfo[releaseData.ApplicationID] = releaseData.ApplicationName
	}

	if err := d.batchCheckExecutePermission(req.Operator, order.Workspace, appsInfo); err != nil {
		return nil, apierrors.ErrCancelDeploymentOrder.InternalError(err)
	}

	runtimes, err := d.db.GetRuntimeByDeployOrderId(order.ProjectId, order.ID)
	if err != nil {
		logrus.Errorf("failed to get runtime by deployment order id: %s, project: %s, err: %v", order.ID,
			order.ProjectName, err)
		return nil, err
	}

	if len(*runtimes) == 0 {
		logrus.Warnf("none runtimes need cancel deploying")
		// project release order deal with sync
		if order.Status == string(apistructs.DeploymentStatusCanceled) || order.Type == apistructs.TypeProjectRelease {
			return nil, nil
		}
		defaultStatusItem := apistructs.DeploymentOrderStatusItem{
			AppID:            uint64(order.ApplicationId),
			DeploymentStatus: apistructs.DeploymentStatusCanceled,
		}

		// deal with application release order
		order.Status = string(apistructs.DeploymentStatusCanceled)
		statusMap := make(apistructs.DeploymentOrderStatusMap)
		if err := json.Unmarshal([]byte(order.StatusDetail), &statusMap); err != nil {
			logrus.Errorf("failed to unmarshal status detail, err: %v", err)
			statusMap[order.ApplicationName] = defaultStatusItem
		} else {
			if app, ok := statusMap[order.ApplicationName]; ok {
				app.DeploymentStatus = apistructs.DeploymentStatusFailed
				statusMap[order.ApplicationName] = app
			} else {
				statusMap[order.ApplicationName] = defaultStatusItem
			}
		}

		newStatusDetail, err := json.Marshal(statusMap)
		if err == nil {
			order.StatusDetail = string(newStatusDetail)
		} else {
			logrus.Errorf("failed to marshal status detail, err: %v", err)
		}

		if err := d.db.UpdateDeploymentOrder(order); err != nil {
			logrus.Errorf("failed to update deployment order, id: %s, err: %v", order.ID, err)
			return nil, err
		}
		return nil, nil
	}

	for _, runtime := range *runtimes {
		if err := d.deploy.CancelLastDeploy(runtime.ID, req.Operator, req.Force); err != nil {
			logrus.Errorf("failed to cancel deploy, runtime: %d, err: %v", runtime.ID, err)
			return nil, err
		}
	}

	return order, nil
}
