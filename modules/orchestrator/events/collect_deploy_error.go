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

package events

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/log"
)

type DeployErrorCollector struct {
	manager *EventManager
	db      *dbclient.DBClient
	bdl     *bundle.Bundle
}

func NewDeployErrorCollector(manager *EventManager, db *dbclient.DBClient, bdl *bundle.Bundle) *EventListener {
	var l EventListener = &DeployErrorCollector{manager: manager, db: db, bdl: bdl}
	return &l
}

func (c *DeployErrorCollector) OnEvent(event *RuntimeEvent) {
	c.reportRuntimeServiceErrors(event)
	c.collectDeployErrors(event)
}

func (c *DeployErrorCollector) collectDeployErrors(event *RuntimeEvent) {
	// only care InstanceChanged events
	if event.EventName != RuntimeServiceInstancesChanged {
		return
	}

	deployment, err := c.db.FindLastDeployment(event.Runtime.ID)
	if err != nil {
		logrus.Warnf("failed to find last deployment of runtime: %v for error collector, err: %v",
			event.Runtime.ID, err.Error())
		return
	}
	if deployment == nil {
		// not found
		return
	}

	if deployment.Status != apistructs.DeploymentStatusDeploying {
		// last deployment not in Deploying
		logrus.Debugf("deployment not in Deploying, thus cannot collect error into deployment: %v log",
			deployment.ID)
		return
	}

	d := &log.DeployLogHelper{DeploymentID: deployment.ID, Bdl: c.bdl}
	for _, i := range event.Instances {
		toLink := fmt.Sprintf("##to_link:projectId:%d,applicationId:%d,runtimeId:%d,serviceName:%s,containerId:%s",
			event.Runtime.ProjectID, event.Runtime.ApplicationID, event.Runtime.ID, event.Service.ServiceName, i.InstanceID)
		d.Log(fmt.Sprintf("%s/%s -- service(%s) instances changed %s",
			i.Status, i.Stage, event.Service.ServiceName, toLink))
	}
}

func (c *DeployErrorCollector) reportRuntimeServiceErrors(event *RuntimeEvent) {
	// only care InstanceChanged events
	if event.EventName != RuntimeServiceInstancesChanged {
		return
	}

	// clear the error if Healthy
	if event.Service.Status == apistructs.ServiceStatusHealthy {
		if err := c.db.ClearRuntimeServiceErrors(event.Service.ID); err != nil {
			logrus.Errorf("[alert] failed to clear RuntimeService(%v/%v) errors, err: %v",
				event.Service.RuntimeID, event.Service.ServiceName, err)
		} else {
			// TODO: we modified the pointer, it may cause unknown result
			event.Service.Errors = nil
			newEvent := RuntimeEvent{
				EventName: RuntimeServiceStatusChanged,
				Runtime:   event.Runtime,
				Service:   event.Service,
			}
			c.manager.EmitEvent(&newEvent)
		}
		return
	}

	// set errors if not Healthy
	var errs []apistructs.ErrorResponse
	for _, i := range event.Instances {
		switch i.Status {
		case apistructs.InstanceStatusKilled, apistructs.InstanceStatusFailed, apistructs.InstanceStatusFinished:
			// TODO: this error msg should provided by scheduler
			errs = append(errs, apistructs.ErrorResponse{
				Msg: fmt.Sprintf("实例启动失败"),
				Ctx: map[string]interface{}{
					"refLog":     true,
					"instanceId": i.InstanceID,
					"status":     i.Status,
				},
			})
			// TODO: to prevent errors explore, currently only one error to put in
			break
		default:
			// don't care other statuses
		}
	}
	if len(errs) != 0 {
		if err := c.db.SetRuntimeServiceErrors(event.Service.ID, errs); err != nil {
			logrus.Errorf("[alert] failed to set RuntimeService(%v/%v) errors, err: %v",
				event.Service.RuntimeID, event.Service.ServiceName, err)
		} else {
			// TODO: we modified the pointer, it may cause unknown result
			event.Service.Errors = errs
			newEvent := RuntimeEvent{
				EventName: RuntimeServiceStatusChanged,
				Runtime:   event.Runtime,
				Service:   event.Service,
			}
			c.manager.EmitEvent(&newEvent)

			// TODO: should we emit RuntimeStatusChanged here?
			event.Runtime.Errors = errs
			newRuntimeEvent := RuntimeEvent{
				EventName: RuntimeStatusChanged,
				Runtime:   event.Runtime,
			}
			c.manager.EmitEvent(&newRuntimeEvent)
		}
	}
}
