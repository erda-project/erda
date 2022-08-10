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
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/queue"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httputil"
)

var (
	defaultGroup = 20
)

func (d *DeploymentOrder) PushOnDeploymentOrderPolling() (abort bool, err0 error) {
	// polling deployment order which is deploying and project release
	// application release status update with deployment fsm
	deploymentOrders, err := d.db.FindUnfinishedDeploymentOrders()
	if err != nil {
		logrus.Errorf("failed to find unfinished deployment order to continue, (%v)", err)
		return
	}
	if len(deploymentOrders) == 0 {
		logrus.Debugf("find empty unfinished deployment orders to continue")
		return
	}

	fn := func(order dbclient.DeploymentOrder, wg *sync.WaitGroup) {
		if wg == nil {
			logrus.Errorf("wait group is nil")
			return
		}

		defer wg.Done()
		// current batch == 0 -> deployment order not stared
		if order.CurrentBatch == 0 {
			logrus.Debugf("deployment order %s is not stared, current batch: %d, batch size: %d", order.ID,
				order.CurrentBatch, order.BatchSize)
			return
		}

		// compose current order status map
		statusMap, err := d.composeDeploymentOrderStatusMap(order)
		if err != nil {
			logrus.Errorf("failed to compose deployment order status map, (%v)", err)
			return
		}

		// status update, only update status of current batch
		if err := inspectDeploymentStatusDetail(&order, statusMap); err != nil {
			logrus.Errorf("failed to inspect deployment status detail, (%v)", err)
			return
		}

		// if current batch is success and is not last batch, deploy next batch
		if utils.ParseDeploymentOrderStatus(statusMap) == apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusOK) {
			if order.CurrentBatch == order.BatchSize {
				logrus.Debugf("deployment order %s is finished", order.ID)
				order.Status = string(apistructs.DeploymentStatusOK)
				if err := d.db.UpdateDeploymentOrder(&order); err != nil {
					logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				}
				// finished or try update status again.
				return
			}

			order.CurrentBatch++
			order.Status = string(apistructs.DeploymentStatusDeploying)
			if err := d.db.UpdateDeploymentOrder(&order); err != nil {
				logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				// try advance batch again
				return
			}

			// if event lost, reprocess will be triggered
			logrus.Infof("start to push deployment order %s to queue, batch size: %d, current batch: %d", order.ID,
				order.BatchSize, order.CurrentBatch)
			if err := d.queue.Push(queue.DEPLOYMENT_ORDER_BATCHES, order.ID); err != nil {
				logrus.Errorf("failed to push on DEPLOYMENT_ORDER_BATCHES, deploymentOrderID: %s, (%v)", order.ID, err)
				return
			}
		} else {
			order.Status = string(utils.ParseDeploymentOrderStatus(statusMap))
			if err := d.db.UpdateDeploymentOrder(&order); err != nil {
				logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				return
			}
		}
	}

	for i := 0; i < len(deploymentOrders); i += defaultGroup {
		group := make([]dbclient.DeploymentOrder, 0, defaultGroup)
		if len(deploymentOrders)-i < defaultGroup {
			group = deploymentOrders[i:]
		} else {
			group = deploymentOrders[i : i+defaultGroup]
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(group))

		for _, order := range group {
			go fn(order, wg)
		}
		wg.Wait()
	}

	return
}

func (d *DeploymentOrder) composeDeploymentOrderStatusMap(order dbclient.DeploymentOrder) (apistructs.DeploymentOrderStatusMap, error) {
	statusMap := make(apistructs.DeploymentOrderStatusMap)

	ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: order.ReleaseId})
	if err != nil {
		logrus.Errorf("failed to get release, err: %v", err)
		return nil, err
	}

	// use for project release
	deployList, err := d.renderDeployListWithCrossProject(strings.Split(order.Modes, ","), order.ProjectId,
		order.Operator.String(), releaseResp.GetData())
	if err != nil {
		return nil, errors.Errorf("failed to render deploy list with cross project, err: %v", err)
	}

	if len(deployList) == 0 {
		return statusMap, errors.Errorf("deployment order %s has invalid deploy list", order.ID)
	}

	// get runtime status, and update
	for _, r := range deployList[order.CurrentBatch-1] {
		if r == nil {
			// reset deployment order status is failed
			order.Status = string(apistructs.DeploymentStatusFailed)
			if err := d.db.UpdateDeploymentOrder(&order); err != nil {
				logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				return statusMap, err
			}
			return statusMap, errors.New("release is nil")
		}

		rt, errGetRuntime := d.db.GetRuntimeByAppName(order.Workspace, order.ProjectId, r.GetApplicationName())
		if errGetRuntime != nil {
			if !errors.Is(errGetRuntime, gorm.ErrRecordNotFound) {
				return statusMap, errors.Errorf("failed to get runtime by app name %s, (%v)",
					r.GetApplicationName(), errGetRuntime)
			}

			statusMap[r.ApplicationName] = apistructs.DeploymentOrderStatusItem{
				AppID:            uint64(r.GetApplicationID()),
				DeploymentStatus: apistructs.DeploymentStatusFailed,
			}
			continue
		}

		lastDeployment, err := d.db.FindLastDeployment(rt.ID)
		if err != nil {
			return statusMap, errors.Errorf("failed to find last deployment for runtime %d, (%v)", rt.ID, err)
		}
		if lastDeployment == nil {
			return statusMap, errors.Errorf("failed to find last deployment for runtime, last deployment is nil, runtime id: %d", rt.ID)
		}

		statusMap[r.GetApplicationName()] = apistructs.DeploymentOrderStatusItem{
			RuntimeID:        rt.ID,
			AppID:            rt.ApplicationID,
			DeploymentID:     lastDeployment.ID,
			DeploymentStatus: lastDeployment.Status,
		}
	}
	return statusMap, nil
}

func inspectDeploymentStatusDetail(order *dbclient.DeploymentOrder, newOrderStatusMap apistructs.DeploymentOrderStatusMap) error {
	curOrderStatusMap := make(apistructs.DeploymentOrderStatusMap, 0)

	if order.StatusDetail != "" {
		if err := json.Unmarshal([]byte(order.StatusDetail), &curOrderStatusMap); err != nil {
			return errors.Wrapf(err, "failed to unmarshal to deployment order status (%s)",
				order.ID)
		}
	}

	for appName, newStatus := range newOrderStatusMap {
		curStatus, ok := curOrderStatusMap[appName]
		if !ok {
			curStatus = apistructs.DeploymentOrderStatusItem{}
		}
		// doesn't need to update the status (ok) even if the Runtime is deleted
		if newStatus.DeploymentStatus == "" || curStatus.DeploymentStatus == apistructs.DeploymentStatusOK {
			continue
		}
		// if status deployment id is nil, append status
		if newStatus.DeploymentID == 0 {
			curStatus.DeploymentStatus = newStatus.DeploymentStatus
		} else {
			curStatus = newStatus
		}
		curOrderStatusMap[appName] = curStatus
	}

	orderStatusMapJson, err := json.Marshal(curOrderStatusMap)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal to deployment order status (%s)",
			order.ID)
	}

	order.StatusDetail = string(orderStatusMapJson)
	return nil
}
