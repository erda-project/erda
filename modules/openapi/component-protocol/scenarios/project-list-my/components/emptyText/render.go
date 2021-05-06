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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// GenComponentState 获取state
func (i *ComponentText) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentText) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := i.GenComponentState(c); err != nil {
		return err
	}

	// 如果list组件的数据不为空，则直接返回
	if !i.State.IsEmpty {
		c.Props = map[string]interface{}{"visible": false}
		return nil
	}

	i.Props = Props{
		Visible:    true,
		RenderType: "linkText",
		StyleConfig: StyleConfig{
			FontSize:   16,
			LineHeight: 24,
		},
		Value: map[string]interface{}{
			"text": []interface{}{
				"您还未加入任何项目，可以选择",
				map[string]interface{}{
					"text":         "公开项目",
					"operationKey": "toPublicProject",
					"styleConfig":  map[string]interface{}{"bold": true},
				},
				"开启您的Erda项目之旅",
			},
		},
	}

	i.Operations = map[string]interface{}{
		"toPublicProject": Operation{
			Key:    "toPublicProject",
			Reload: false,
			Command: Command{
				Key:          "changeScenario",
				ScenarioType: "project-list-all",
				ScenarioKey:  "project-list-all",
			},
		},
	}

	c.Operations = i.Operations
	c.Props = i.Props
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentText{}
}
