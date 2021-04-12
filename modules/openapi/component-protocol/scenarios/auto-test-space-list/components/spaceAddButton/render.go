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

package spaceAddButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type props struct {
	Text       string      `json:"text"`
	Type       string      `json:"type"`
	Operations interface{} `json:"operations"`
}

type AddButtonCandidateOp struct {
	Click struct {
		Reload  bool                   `json:"reload"`
		Key     string                 `json:"key"`
		Command map[string]interface{} `json:"command"`
	} `json:"click"`
}
type AddButtonCandidate struct {
	Disabled    bool                 `json:"disabled"`
	DisabledTip string               `json:"disabledTip"`
	Key         string               `json:"key"`
	Operations  AddButtonCandidateOp `json:"operations"`
	PrefixIcon  string               `json:"prefixIcon"`
	Text        string               `json:"text"`
}

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	prop := props{
		Text: "新建空间",
		Type: "primary",
	}
	c.Props = prop
	c.Operations = map[string]interface{}{
		"click": struct {
			Reload  bool                   `json:"reload"`
			Key     string                 `json:"key"`
			Command map[string]interface{} `json:"command"`
		}{
			Reload: false,
			Key:    "addSpace",
			Command: map[string]interface{}{
				"key":    "set",
				"target": "spaceFormModal",
				"state": map[string]interface{}{
					"visible":  true,
					"formData": nil,
				},
			},
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
