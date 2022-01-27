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

package issueKanbanV2

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type eventBroadcaster struct{}

func (bc *eventBroadcaster) Product(e websocket.Event) *websocket.Event {
	switch e.Scope.Type {
	case apistructs.ProjectScope:
		switch e.Type {
		case issue.WsTypeIssueCreate:
			return productByIssueCreateEvent(e)
		}
	}
	return nil
}

func productByIssueCreateEvent(e websocket.Event) *websocket.Event {
	ne := &websocket.Event{}
	ne.Scope = e.Scope
	if ne.Scope.Extras == nil {
		ne.Scope.Extras = make(map[string]string)
	}
	ne.Scope.Extras["scenario"] = "issue-kanban"
	ne.Type = e.Type
	ne.Payload = cptype.InvokeRenderWsEventPayload{
		RenderEvent: cptype.ComponentEvent{
			Component:     "", // empty mean init page
			Operation:     cptype.InitializeOperation,
			OperationData: nil,
		},
	}
	return ne
}
