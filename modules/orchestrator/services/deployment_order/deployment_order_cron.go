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

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/queue"
	"github.com/erda-project/erda/modules/orchestrator/utils"
)

func (d *DeploymentOrder) PushOnDeploymentOrderPolling() (abort bool, err0 error) {
	// polling deployment order which is deploying and project release
	// application release status update with deployment fsm
	deploymentOrders, err := d.db.FindUnfinishedDeploymentOrders()
	if err != nil {
		logrus.Warnf("failed to find unfinished deployment order to continue, (%v)", err)
	}
	if len(deploymentOrders) == 0 {
		logrus.Debugf("find empty unfinished deployment orders to continue")
		return
	}

	for _, order := range deploymentOrders {
		// current batch == 0 -> deployment order not stared
		if order.CurrentBatch == 0 {
			logrus.Debugf("deployment order %s is not stared, current batch: %d, batch size: %d", order.ID,
				order.CurrentBatch, order.BatchSize)
			continue
		}

		var (
			deployList [][]string
			releaseIds []string
			releases   []*dbclient.Release
			id2Release = make(map[string]*dbclient.Release)
		)
		deployList, err = unmarshalDeployList(order.DeployList)
		if err != nil {
			logrus.Warnf("failed to unmarshal deploy list for order %s, %v", order.ID, err)
			continue
		}
		for _, l := range deployList {
			releaseIds = append(releaseIds, l...)
		}
		releases, err = d.db.ListReleases(releaseIds)
		if err != nil {
			return
		}
		for _, release := range releases {
			id2Release[release.ReleaseId] = release
		}

		// get runtime status, and update
		statusMap := apistructs.DeploymentOrderStatusMap{}
		for _, id := range deployList[order.CurrentBatch-1] {
			rt, errGetRuntime := d.db.GetRuntimeByAppName(order.Workspace, order.ProjectId, id2Release[id].ApplicationName)
			if errGetRuntime != nil {
				if !errors.Is(errGetRuntime, gorm.ErrRecordNotFound) {
					logrus.Errorf("failed to get runtime by app name %s, (%v)", id2Release[id].ApplicationName, errGetRuntime)
					return
				}

				logrus.Debugf("runtime not found, app name: %s", id2Release[id].ApplicationName)
				// if current batch is 0, polling first batch and production event
				if order.CurrentBatch == 0 {
					order.CurrentBatch++
					if err := d.db.UpdateDeploymentOrder(&order); err != nil {
						return
					}
				}
				// if current batch is not 0, last event may be lost, reproduction event.
				if err := d.queue.Push(queue.DEPLOYMENT_ORDER_BATCHES, order.ID); err != nil {
					logrus.Errorf("failed to push on DEPLOYMENT_ORDER_BATCHES, deploymentOrderID: %s, (%v)", order.ID, err)
					return
				}
				return
			}

			lastDeployment, err := d.db.FindLastDeployment(rt.ID)
			if err != nil {
				logrus.Errorf("failed to find last deployment for runtime %d, (%v)", rt.ID, err)
				return
			}
			if lastDeployment == nil {
				logrus.Errorf("failed to find last deployment for runtime, last deployment is nil, runtime id: %d", rt.ID)
				return
			}

			statusMap[id2Release[id].ApplicationName] = apistructs.DeploymentOrderStatusItem{
				RuntimeID:        rt.ID,
				AppID:            rt.ApplicationID,
				DeploymentID:     lastDeployment.ID,
				DeploymentStatus: lastDeployment.Status,
			}
		}

		// status update, only update status of current batch
		if err := inspectDeploymentStatusDetail(&order, statusMap); err != nil {
			logrus.Errorf("failed to update deployment order %s status, (%v)", order.ID, err)
			continue
		}

		// if current batch is success and is not last batch, deploy next batch
		if utils.ParseDeploymentOrderStatus(statusMap) == apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusOK) {
			if order.CurrentBatch == order.BatchSize {
				logrus.Debugf("deployment order %s is finished", order.ID)
				order.Status = string(apistructs.DeploymentStatusOK)
				if err := d.db.UpdateDeploymentOrder(&order); err != nil {
					logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
					return
				}
				continue
			}

			order.CurrentBatch++
			order.Status = string(apistructs.DeploymentStatusDeploying)
			if err := d.db.UpdateDeploymentOrder(&order); err != nil {
				logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				return
			}

			// if event lost, reprocess will be triggered
			logrus.Infof("start to push deployment order %s to queue, batch size: %d, current batch: %d", order.ID,
				order.BatchSize, order.CurrentBatch)
			if err := d.queue.Push(queue.DEPLOYMENT_ORDER_BATCHES, order.ID); err != nil {
				logrus.Errorf("failed to push on DEPLOYMENT_ORDER_BATCHES, deploymentOrderID: %s, (%v)", order.ID, err)
				continue
			}
		} else {
			order.Status = string(utils.ParseDeploymentOrderStatus(statusMap))
			if err := d.db.UpdateDeploymentOrder(&order); err != nil {
				logrus.Errorf("failed to update deployment order %s, (%v)", order.ID, err)
				return
			}
		}
	}

	return
}

func inspectDeploymentStatusDetail(order *dbclient.DeploymentOrder, newOrderStatusMap apistructs.DeploymentOrderStatusMap) error {
	curOrderStatusMap := make(apistructs.DeploymentOrderStatusMap, 0)

	if order.StatusDetail != "" {
		if err := json.Unmarshal([]byte(order.StatusDetail), &curOrderStatusMap); err != nil {
			return errors.Wrapf(err, "failed to unmarshal to deployment order status (%s)",
				order.ID)
		}
	}

	for appName, status := range newOrderStatusMap {
		if status.DeploymentID == 0 || status.DeploymentStatus == "" {
			continue
		}
		curOrderStatusMap[appName] = status
	}

	orderStatusMapJson, err := json.Marshal(curOrderStatusMap)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal to deployment order status (%s)",
			order.ID)
	}

	order.StatusDetail = string(orderStatusMapJson)
	return nil
}
