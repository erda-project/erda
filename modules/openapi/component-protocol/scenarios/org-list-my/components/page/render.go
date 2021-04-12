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

package page

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func (i *ComponentPage) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	c.State = state
	c.Props = i.Props
	return nil
}

func (i *ComponentPage) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	i.initProperty(s)

	return nil
}

func (i *ComponentPage) initProperty(s apistructs.ComponentProtocolScenario) {
	myOrgs := Option{
		Key:  "my",
		Name: "我的组织",
		Operations: map[string]interface{}{
			"click": ClickOperation{
				Reload: false,
				Key:    "myOrg",
				Command: Command{
					Key:          "changeScenario",
					ScenarioKey:  s.ScenarioKey,
					ScenarioType: s.ScenarioType,
				},
			},
		},
	}
	publicOrgs := Option{
		Key:  "public",
		Name: "公开组织",
		Operations: map[string]interface{}{
			"click": ClickOperation{
				Reload: false,
				Key:    "publicOrg",
				Command: Command{
					Key:          "changeScenario",
					ScenarioKey:  "org-list-all",
					ScenarioType: "org-list-all",
				},
			},
		},
	}

	i.Props = Props{
		TabMenu: []Option{myOrgs, publicOrgs},
	}
	i.State.ActiveKey = myOrgs.Key
}

func RenderCreator() protocol.CompRender {
	return &ComponentPage{}
}
