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
	"fmt"

	"github.com/erda-project/erda/apistructs"
)

var broadcasterFactory = map[apistructs.ScopeType]map[string]EventBroadcaster{}

func RegisterEventBroadcaster(scopeType apistructs.ScopeType, eventType string, bc EventBroadcaster) {
	if broadcasterFactory[scopeType] == nil {
		broadcasterFactory[scopeType] = make(map[string]EventBroadcaster)
	}
	bc, exist := broadcasterFactory[scopeType][eventType]
	if exist {
		panic(fmt.Errorf("broadcaster already exists, scopeType: %s, eventType: %s", scopeType, eventType))
	}
	broadcasterFactory[scopeType][eventType] = bc
}

type EventBroadcaster interface {
	Product(e Event) (events []Event)
}

// broadcastEvent receive one event and output batch events
func broadcastEvent(e Event) (events []Event) {
	// init
	events = []Event{e}

	// find broadcaster
	scopedMap, ok := broadcasterFactory[e.Scope.Type]
	if !ok {
		return
	}
	broadcaster, ok := scopedMap[e.Type]
	if !ok {
		return
	}

	// product events
	events = append(events, broadcaster.Product(e)...)

	return
}
