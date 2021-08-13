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
		"text":       ctxBdl.I18nPrinter.Sprintf("List"),
		"tooltip":    "",
		"prefixIcon": "default-list",
		"key":        "table",
	}
	optionKanban := map[string]interface{}{
		"text":       ctxBdl.I18nPrinter.Sprintf("Board"),
		"tooltip":    ctxBdl.I18nPrinter.Sprintf("Board View"),
		"prefixIcon": "data-matrix",
		"suffixIcon": "di",
		"key":        "kanban",
	}
	optionGantt := map[string]interface{}{
		"text":       ctxBdl.I18nPrinter.Sprintf("Gantt Chart"),
		"tooltip":    "",
		"prefixIcon": "gantetu",
		"key":        "gantt",
	}
	optionKanbanChildren := []map[string]string{
		{"text": ctxBdl.I18nPrinter.Sprintf("Priority"), "key": "priority"},
		//{"text": "处理人", "key": "assignee"},
		{"text": ctxBdl.I18nPrinter.Sprintf("Deadline"), "key": "deadline"},
		{"text": ctxBdl.I18nPrinter.Sprintf("Custom"), "key": "custom"},
	}
	if ctxBdl.InParams["fixedIssueType"].(string) != "ALL" {
		optionKanbanChildren = append(optionKanbanChildren, map[string]string{"text": ctxBdl.I18nPrinter.Sprintf("State"), "key": "status"})
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
