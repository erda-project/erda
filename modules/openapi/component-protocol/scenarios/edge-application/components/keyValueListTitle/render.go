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

package keyvaluelisttitle

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

type EdgeKVListDataItem struct {
	KeyName   string    `json:"keyName"`
	ValueName ItemValue `json:"valueName"`
}

type ItemValue struct {
	RenderType string           `json:"renderType"`
	Value      ItemValueContent `json:"value"`
}

type ItemValueContent struct {
	Text     string `json:"text"`
	CopyText string `json:"copyText"`
}

func (c *ComponentKVListTitle) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	if err := c.SetComponent(component); err != nil {
		return err
	}

	if event.Operation == apistructs.RenderingOperation {
		if err := c.OperationRendering(); err != nil {
			return err
		}
	}

	return nil
}

func getProps(visible bool) apistructs.EdgeKVListTitleProps {
	return apistructs.EdgeKVListTitleProps{
		Visible: visible,
		Title:   "链接信息",
		Level:   2,
	}
}
