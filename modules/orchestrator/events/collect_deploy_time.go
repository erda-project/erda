// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package events

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

type DeployTimeCollector struct {
	manager *EventManager
	db      *dbclient.DBClient
}

func NewDeployTimeCollector(manager *EventManager, db *dbclient.DBClient) *EventListener {
	var l EventListener = &DeployTimeCollector{manager: manager, db: db}
	return &l
}

func (c *DeployTimeCollector) OnEvent(event *RuntimeEvent) {
	c.collectDeployTimes(event)
}

func (c *DeployTimeCollector) collectDeployTimes(event *RuntimeEvent) {
	// only care DeployStatusChanged events
	if event.EventName != RuntimeDeployStatusChanged {
		return
	}

	deployment, err := c.db.GetDeployment(event.Deployment.ID)
	if err != nil {
		logrus.Warnf("failed to get deployment: %v for time collector, err: %v",
			event.Deployment.ID, err.Error())
		return
	}
	if deployment == nil {
		// not found
		return
	}

	isChanged := true
	now := time.Now()
	// TODO: need refactor, should introduce deployment_phases table to tracking times
	// TODO: Script is a virtual Phase now, it indicate the `pre-service` Phase
	switch deployment.Status {
	case apistructs.DeploymentStatusDeploying:
		// Status is Deploying, entering the Phase, so set the StartAt
		switch deployment.Phase {
		case apistructs.DeploymentPhaseAddon:
			deployment.Extra.AddonPhaseStartAt = &now
		case apistructs.DeploymentPhaseScript:
			// the end of addon, service start at same time
			deployment.Extra.AddonPhaseEndAt = &now
			deployment.Extra.ServicePhaseStartAt = &now
		case apistructs.DeploymentPhaseCompleted:
			deployment.Extra.ServicePhaseEndAt = &now
		default:
			isChanged = false
		}
	case apistructs.DeploymentStatusFailed, apistructs.DeploymentStatusCanceled:
		// Status is Failed or Canceled, Phase broken (end) at current time
		switch deployment.Phase {
		case apistructs.DeploymentPhaseAddon:
			deployment.Extra.AddonPhaseEndAt = &now
		case apistructs.DeploymentPhaseScript, apistructs.DeploymentPhaseService:
			deployment.Extra.ServicePhaseEndAt = &now
		default:
			isChanged = false
		}
	default:
		isChanged = false
	}

	if isChanged {
		if err := c.db.UpdateDeployment(deployment); err != nil {
			logrus.Errorf("[alert] failed to update Phase Times into deployment: %d, %v",
				deployment.ID, err)
		}
	}
}
