package events

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

type EventboxPublisher struct {
	manager *EventManager
}

func NewEventboxPublisher(manager *EventManager) *EventListener {
	var l EventListener = &EventboxPublisher{manager: manager}
	return &l
}

func (p *EventboxPublisher) OnEvent(event *RuntimeEvent) {
	if err := p.manager.publishWebhook(event); err != nil {
		logrus.Errorf("[alert] failed to publish webhook, event: %v, err is %v", event, err.Error())
	}
}

func (m *EventManager) publishWebhook(event *RuntimeEvent) error {
	w := apistructs.EventHeader{}
	switch event.EventName {
	case RuntimeCreated:
		w.Event = "runtime"
		w.Action = "create"
		w.OrgID = strconv.FormatUint(event.Runtime.OrgID, 10)
		w.ProjectID = strconv.FormatUint(event.Runtime.ProjectID, 10)
		w.ApplicationID = strconv.FormatUint(event.Runtime.ApplicationID, 10)
		w.Env = event.Runtime.Workspace
	case RuntimeDeleting:
		w.Event = "runtime"
		w.Action = "delete"
		w.OrgID = strconv.FormatUint(event.Runtime.OrgID, 10)
		w.ProjectID = strconv.FormatUint(event.Runtime.ProjectID, 10)
		w.ApplicationID = strconv.FormatUint(event.Runtime.ApplicationID, 10)
		w.Env = event.Runtime.Workspace
	default:
		// TODO: support more webhooks
		return nil
	}
	ev := apistructs.EventCreateRequest{
		Sender:      "orchestrator",
		EventHeader: w,
		Content:     event,
	}
	return m.bdl.CreateEvent(&ev)
}
