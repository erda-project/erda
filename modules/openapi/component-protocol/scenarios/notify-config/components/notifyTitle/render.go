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

package notifyTitle

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentAction struct {
	CtxBdl protocol.ContextBundle
}

type NotifyTitle struct {
	Type  string `json:"type"`
	Props Props  `json:"props"`
}

type Props struct {
	Title string `json:"title"`
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := ca.Import(c); err != nil {
		logrus.Errorf("import component is failed err is %v", err)
		return err
	}
	notifyTitle := NotifyTitle{
		Type: "Title",
		Props: Props{
			Title: "帮助您更好地组织通知项",
		},
	}
	c.Props = notifyTitle.Props
	c.Type = notifyTitle.Type
	return nil
}

func (ca *ComponentAction) Import(c *apistructs.Component) error {
	com, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(com, ca); err != nil {
		return err
	}
	return nil
}
