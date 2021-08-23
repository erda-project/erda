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

package issueViewGroup

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct{}

type State struct {
	Value         string            `json:"value,omitempty"`
	ChildrenValue map[string]string `json:"childrenValue,omitempty"`
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	ctxBdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if c.State == nil {
		c.State = map[string]interface{}{}
	}
	state := State{Value: "table", ChildrenValue: map[string]string{"kanban": "deadline"}}

	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		if urlQueryI, ok := ctxBdl.InParams[getStateUrlQueryKey()]; ok {
			if urlQueryStr, ok := urlQueryI.(string); ok && urlQueryStr != "" {
				var urlState State
				b, err := base64.StdEncoding.DecodeString(urlQueryStr)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(b, &urlState); err != nil {
					return err
				}
				state = urlState
			}
		}
	case "changeViewType":
		b, err := json.Marshal(c.State)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &state); err != nil {
			return err
		}
	}

	// props
	props := make(map[string]interface{})
	props["radioType"] = "button"
	props["buttonStyle"] = "solid"
	props["size"] = "small"
	optionTable := map[string]interface{}{
		"text":       "表格",
		"tooltip":    "",
		"prefixIcon": "default-list",
		"key":        "table",
	}
	optionKanban := map[string]interface{}{
		"text":       "看板",
		"tooltip":    "看板视图",
		"prefixIcon": "data-matrix",
		"suffixIcon": "di",
		"key":        "kanban",
	}
	optionGantt := map[string]interface{}{
		"text":       "甘特图",
		"tooltip":    "",
		"prefixIcon": "gantetu",
		"key":        "gantt",
	}
	optionKanbanChildren := []map[string]string{
		{"text": "优先级", "key": "priority"},
		//{"text": "处理人", "key": "assignee"},
		{"text": "截止日期", "key": "deadline"},
		{"text": "自定义", "key": "custom"},
	}
	if ctxBdl.InParams["fixedIssueType"].(string) != "ALL" {
		optionKanbanChildren = append(optionKanbanChildren, map[string]string{"text": "状态", "key": "status"})
	}
	optionKanban["children"] = optionKanbanChildren
	props["options"] = []map[string]interface{}{optionTable, optionKanban, optionGantt}
	c.Props = props

	// set state
	if err := setState(c, state); err != nil {
		return err
	}

	return json.Unmarshal([]byte(`{"onChange":{"key":"changeViewType","reload":true}}`), &c.Operations)
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}

func getStateUrlQueryKey() string {
	return "issueViewGroup__urlQuery"
}

func setState(c *apistructs.Component, state State) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	c.State["value"] = state.Value
	c.State["childrenValue"] = state.ChildrenValue
	c.State[getStateUrlQueryKey()] = base64.StdEncoding.EncodeToString(b)
	return nil
}
