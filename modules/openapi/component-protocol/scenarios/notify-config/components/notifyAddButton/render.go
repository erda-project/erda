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

package notifyAddButton

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentButton struct {
	ctxBdl     protocol.ContextBundle
	Type       string             `json:"type"`
	Operations AddButtonOperation `json:"operations"`
	Props      Props              `json:"props"`
	State      State              `json:"state"`
}

type State struct {
	Operation string `json:"operation"`
	Visible   bool   `json:"visible"`
	EditId    uint64 `json:"editId"`
}

type Props struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type AddButtonOperation struct {
	Click Click `json:"click"`
}

type Click struct {
	key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type CommandState struct {
	Visible  bool        `json:"visible"`
	FormData interface{} `json:"formData"`
}

func (cb *ComponentButton) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := cb.Import(c); err != nil {
		logrus.Errorf("import button component is failed err is %v", err)
		return err
	}
	cb.ctxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err := cb.handlerAddOperation(); err != nil {
		return err
	}
	if err := cb.Export(c); err != nil {
		logrus.Errorf("export button component is failed err is %v", err)
		return err
	}
	return nil
}

func (cb *ComponentButton) handlerAddOperation() error {
	buttonResp := ComponentButton{
		Type: "Button",
		Operations: AddButtonOperation{
			Click: Click{
				key:    "addNotify",
				Reload: true,
			},
		},
		Props: Props{
			Text: "新建通知",
			Type: "primary",
		},
	}
	cb.Props = buttonResp.Props
	cb.Type = buttonResp.Type
	//使弹窗可见
	cb.State.Visible = true
	data, err := json.Marshal(buttonResp.Operations)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &cb.Operations)
	if err != nil {
		return err
	}
	return nil
}

func (cb *ComponentButton) Import(c *apistructs.Component) error {
	com, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(com, cb); err != nil {
		return err
	}
	return nil
}

func (cb *ComponentButton) Export(c *apistructs.Component) error {
	b, err := json.Marshal(cb)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentButton{}
}
