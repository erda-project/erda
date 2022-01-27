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

package websocket

import (
	"github.com/erda-project/erda/apistructs"
)

var eventBroadcasterFactory = map[apistructs.ScopeType]map[string][]EventProductor{}

func RegisterEventProductor(scopeType apistructs.ScopeType, eventType string, productor EventProductor) {
	if eventBroadcasterFactory[scopeType] == nil {
		eventBroadcasterFactory[scopeType] = make(map[string][]EventProductor)
	}
	eventBroadcasterFactory[scopeType][eventType] = append(eventBroadcasterFactory[scopeType][eventType], productor)
}

type EventProductor interface {
	Product(e Event) *Event
}

// broadcastEvent receive one event and output batch events
func broadcastEvent(e Event) (events []Event) {
	// init
	events = []Event{e}

	// find broadcaster
	scopedMap, ok := eventBroadcasterFactory[e.Scope.Type]
	if !ok {
		return
	}
	productors, ok := scopedMap[e.Type]
	if !ok {
		return
	}

	// product events
	for _, productor := range productors {
		ne := productor.Product(e)
		if ne != nil {
			events = append(events, *ne)
		}
	}

	return
}
