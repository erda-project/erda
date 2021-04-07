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
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/orchestrator/queue"
)

type DeployPusher struct {
	manager *EventManager
	queue   *queue.PusherQueue
}

func NewDeployPusher(manager *EventManager, queue *queue.PusherQueue) *EventListener {
	var l EventListener = &DeployPusher{manager: manager, queue: queue}
	return &l
}

func (p *DeployPusher) OnEvent(event *RuntimeEvent) {
	switch event.EventName {
	case RuntimeDeployStatusChanged:
		if err := p.queue.Push(queue.DEPLOY_CONTINUING, strconv.Itoa(int(event.Deployment.ID))); err != nil {
			logrus.Errorf("[alert] failed to continue deploy, event: %v, err is %v", event, err.Error())
		}
	}
}
