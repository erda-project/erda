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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// GenComponentState 获取state
func (i *ComponentPage) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	return nil
}

func (i *ComponentPage) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := i.GenComponentState(c); err != nil {
		return err
	}
	i.Props = Props{
		TabMenu: []TabMenu{
			{
				Key:  "my",
				Name: "我的项目",
				Operations: map[string]interface{}{
					apistructs.ClickOperation.String(): Operation{
						Reload: false,
						Key:    "myProject",
						Command: Command{
							Key:          "changeScenario",
							ScenarioType: "project-list-my",
							ScenarioKey:  "project-list-my",
						},
					},
				},
			},
			{
				Key:  "all",
				Name: "公开项目",
				Operations: map[string]interface{}{
					apistructs.ClickOperation.String(): Operation{
						Reload: false,
						Key:    "allProject",
						Command: Command{
							Key:          "changeScenario",
							ScenarioType: "project-list-all",
							ScenarioKey:  "project-list-all",
						},
					},
				},
			},
		},
	}

	if event.Operation == apistructs.InitializeOperation {
		i.State.ActiveKey = "my"
	}

	c.Props = i.Props
	c.State = map[string]interface{}{
		"activeKey": i.State.ActiveKey,
	}
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentPage{}
}
