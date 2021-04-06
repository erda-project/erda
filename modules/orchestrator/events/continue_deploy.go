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
