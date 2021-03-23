package events

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type EventManager struct {
	ch chan Event
}

var mgr EventManager

var defaultEvent DefaultEvent

func Initialize(bdl *bundle.Bundle, wsClient *websocket.Publisher) {
	mgr = EventManager{
		ch: make(chan Event, 100),
	}

	defaultEvent = DefaultEvent{
		bdl:      bdl,
		wsClient: wsClient,
	}

	go func() {
		for {
			e := <-mgr.ch
			go func() {
				logrus.Debugf("received an %s Event: %s (kind: %s, header: %+v, sender: %s, content: %+v)",
					e.Kind(), e, e.Kind(), e.Header(), e.Sender(), e.Content())

				go handle(e, HookTypeWebHook, e.HandleWebhook)
				go handle(e, HookTypeWebSocket, e.HandleWebSocket)
				go handle(e, HookTypeDINGDING, e.HandleDingDing)
				go handle(e, HookTypeHTTP, e.HandleHTTP)
			}()
		}
	}()
}

func handle(e Event, hook HookType, handleFunc func() error) {
	logrus.Debugf("begin handle event, hook: %s, event: %s", hook, e)
	err := handleFunc()
	if err != nil {
		logrus.Errorf("failed to handle event, hook: %s, event: %s, err: %v", hook, e, err)
		return
	}
	logrus.Debugf("success handle event, hook: %s, event: %s", hook, e)
}
