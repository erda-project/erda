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
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type EventManager struct {
	ch chan Event
}

var mgr EventManager

var defaultEvent DefaultEvent

func Initialize(bdl *bundle.Bundle, wsClient *websocket.Publisher, dbClient *dbclient.Client) {
	mgr = EventManager{
		ch: make(chan Event, 100),
	}

	defaultEvent = DefaultEvent{
		bdl:      bdl,
		wsClient: wsClient,
		dbClient: dbClient,
	}

	go func() {
		for {
			e := <-mgr.ch
			go func(ev Event) {
				//logrus.Debugf("received an %s Event: %s (kind: %s, header: %+v, sender: %s, content: %+v)",
				//	e.Kind(), e, e.Kind(), e.Header(), e.Sender(), e.Content())

				go handle(ev, HookTypeWebHook, ev.HandleWebhook)
				go handle(ev, HookTypeWebSocket, ev.HandleWebSocket)
				go handle(ev, HookTypeDINGDING, ev.HandleDingDing)
				go handle(ev, HookTypeHTTP, ev.HandleHTTP)
				go handle(ev, HookTypeDB, ev.HandleDB)
			}(e)
		}
	}()
}

func handle(e Event, hook HookType, handleFunc func() error) {
	//logrus.Debugf("begin handle event, hook: %s, event: %s", hook, e)
	err := handleFunc()
	if err != nil {
		logrus.Errorf("failed to handle event, hook: %s, event: %s, err: %v", hook, e, err)
		return
	}
	//logrus.Debugf("success handle event, hook: %s, event: %s", hook, e)
}
