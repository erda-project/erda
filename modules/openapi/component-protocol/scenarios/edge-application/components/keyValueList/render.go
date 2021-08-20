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

package keyvaluelist

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
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

func (c *ComponentKVList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	identity := c.ctxBundle.Identity

	if event.Operation == apistructs.RenderingOperation {
		if err := c.OperationRendering(identity); err != nil {
			return err
		}
	}

	return nil
}

func getProps(visible bool) apistructs.EdgeKVListProps {
	return apistructs.EdgeKVListProps{
		Visible:    visible,
		Pagination: false,
		RowKey:     "keyName",
		Columns: []apistructs.EdgeKVListColumns{
			{
				EdgeColumns: apistructs.EdgeColumns{
					DataIndex: "keyName",
				},
				ColSpan: 0,
			},
			{
				EdgeColumns: apistructs.EdgeColumns{
					DataIndex: "valueName",
				},
				ColSpan: 0,
			},
		},
	}
}
