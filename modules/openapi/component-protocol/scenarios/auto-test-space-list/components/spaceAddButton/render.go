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
