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
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/queue"
)

type EventManager struct {
	c         chan *RuntimeEvent
	listeners []*EventListener
	db        *dbclient.DBClient
	bdl       *bundle.Bundle
}

func NewEventManager(cap int, queue *queue.PusherQueue, db *dbclient.DBClient, bdl *bundle.Bundle) *EventManager {
	m := &EventManager{
		c:   make(chan *RuntimeEvent, cap),
		db:  db,
		bdl: bdl,
	}
	m.listeners = []*EventListener{
		NewActivityPublisher(m),
		NewEventboxPublisher(m),
		NewWsPublisher(m),
		NewDeployPusher(m, queue),
		NewDeployErrorCollector(m, db, bdl),
		NewDeployTimeCollector(m, db),
	}
	return m
}

type EventListener interface {
	OnEvent(e *RuntimeEvent)
}

func (m *EventManager) EmitEvent(e *RuntimeEvent) {
	go func(c chan *RuntimeEvent) {
		c <- e
	}(m.c)
}

func (m *EventManager) Start() {
	go func() {
		for {
			e := <-m.c
			logrus.Debugf("received RuntimeEvent %v", e)
			for i := range m.listeners {
				go onEvent(m.listeners[i], e)
			}
		}
	}()
}

func onEvent(l *EventListener, e *RuntimeEvent) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logrus.Errorf("[alert] failed to emit event to listener: %v", err)
		}
	}()
	// do send event to listener's OnEvent
	(*l).OnEvent(e)
}
