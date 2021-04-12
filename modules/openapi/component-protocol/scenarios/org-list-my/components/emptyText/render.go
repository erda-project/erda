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

package emptyText

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func (i *ComponentEmptyText) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	i.State = state
	propValue, err := json.Marshal(i.Props)
	if err != nil {
		return err
	}
	var prop Props
	err = json.Unmarshal(propValue, &prop)
	if err != nil {
		return err
	}
	i.Props = prop
	return nil
}

func (i *ComponentEmptyText) marshal(c *apistructs.Component) error {
	propValue, err := json.Marshal(i.Props)
	if err != nil {
		return err
	}
	var prop map[string]interface{}
	err = json.Unmarshal(propValue, &prop)
	if err != nil {
		return err
	}
	c.Props = i.Props
	c.Operations = i.Operations
	return nil
}

func (i *ComponentEmptyText) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	err = i.unmarshal(c)
	if err != nil {
		return err
	}

	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	i.initProperty(s)
	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		if i.State.IsEmpty {
			i.Props.Visible = true
		}
	}
	return nil
}

func (i *ComponentEmptyText) initProperty(s apistructs.ComponentProtocolScenario) {
	i.Operations = map[string]interface{}{
		"toPublicOrg": Operation{
			Key:    "toPublicOrg",
			Reload: false,
			Command: Command{
				Key:          "changeScenario",
				ScenarioKey:  "org-list-all",
				ScenarioType: "org-list-all",
			},
		},
	}
	i.Props = Props{
		Visible:    false,
		RenderType: "linkText",
		StyleConfig: Config{
			FontSize:   16,
			LineHeight: 24,
		},
		Value: map[string][]interface{}{
			"text": {
				"您还未加入任何组织，可以选择",
				Redirect{
					Text:         "公开组织",
					OperationKey: "toPublicOrg",
					StyleConfig: RedirectConfig{
						Bold: true,
					},
				},
				"开启您的Erda之旅",
			},
		},
	}
}

func RenderCreator() protocol.CompRender {
	return &ComponentEmptyText{}
}
