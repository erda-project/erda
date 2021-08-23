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

package memTable

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common/table"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/components/tab"
)

var tableProperties = map[string]interface{}{
	"rowKey": "name",
	// todo update dataindex
	"columns": []table.Columns{
		{DataIndex: "id"},
		{DataIndex: "status", Title: "状态", Sortable: true},
		{DataIndex: "pod", Title: "名称", Sortable: true},
		{DataIndex: "namespace", Title: "命名空间"},
		{DataIndex: "cpuPercent", Title: "cpu分配量"},
		{DataIndex: "cpuUsed", Title: "cpu水位"},
		{DataIndex: "ready", Title: "Ready"},
	},
	"bordered":        true,
	"selectable":      true,
	"pageSizeOptions": []string{"10", "20", "50", "100"},
}

func (pt *PodCpuTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	pt.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	err := common.Transfer(c.State, &pt.State)
	if err != nil {
		return err
	}
	if c.State["activeKey"] != tab.CPU_TAB {
		return nil
	}
	if event.Operation != apistructs.InitializeOperation {

		switch event.Operation {
		//case apistructs.CMPDashboardChangePageSizeOperationKey:
		//	if err := pt.RenderChangePageSize(event.OperationData); err != nil {
		//		return err
		//	}
		case apistructs.CMPDashboardChangePageNoOperationKey:
			if err := pt.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case apistructs.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if pt.State.IsFirstFilter {
				pt.State.PageNo = 1
				pt.State.IsFirstFilter = false
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	if err := pt.RenderList(c, event); err != nil {
		return err
	}
	if err := pt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func getProps() map[string]interface{} {
	return tableProperties
}

func RenderCreator() protocol.CompRender {
	pi := PodCpuTable{}
	pi.Type = "Table"
	pi.Props = getProps()
	pi.Operations = table.GetTableOperation()
	pi.State = table.State{}
	return &pi
}
