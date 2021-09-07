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

package podTable

import (
	"context"
	"fmt"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

func (pt *PodInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &pt.State)
	if err != nil {
		return err
	}
	pt.SDK = cputil.SDK(ctx)
	pt.Operations = pt.GetTableOperation()
	pt.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	pt.Table.TableComponent = pt
	pt.getProps()
	pt.GetBatchOperation()
	activeKey := (*gs)["activeKey"].(string)
	if activeKey != tableTabs.POD_TAB {
		pt.Props["visible"] = false
		return pt.SetComponentValue(c)
	} else {
		pt.Props["visible"] = true
	}
	if event.Operation != cptype.InitializeOperation {
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey:
			if err := pt.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardChangePageNoOperationKey:
			if err := pt.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardDeleteNode:
			err := pt.DeleteNode(pt.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		case common.CMPDashboardUnfreezeNode:
			err := pt.UnFreezeNode(pt.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		case common.CMPDashboardFreezeNode:
			err := pt.FreezeNode(pt.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	nodes, err := pt.GetNodes(gs)
	if err != nil {
		return err
	}
	if err = pt.RenderList(c, table.Pod, nodes); err != nil {
		return err
	}
	if err = pt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (pt *PodInfoTable) getProps() {
	p := map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "Status", Title: pt.SDK.I18n("status"), Sortable: true, Width: 80, Fixed: "left"},
			{DataIndex: "Node", Title: pt.SDK.I18n("node"), Sortable: true},
			{DataIndex: "IP", Title: pt.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: pt.SDK.I18n("role"), Sortable: true},
			{DataIndex: "Version", Title: pt.SDK.I18n("version")},
			{DataIndex: "UsageRate", Title: "pod" + pt.SDK.I18n("usage"), Sortable: true},
			{DataIndex: "Operate", Title: pt.SDK.I18n("operate"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
		"scroll": table.Scroll{X: 1200},
	}
	pt.Props = p
}

func (pt *PodInfoTable) GetRowItem(node data.Object, tableType table.TableType) (*table.RowItem, error) {
	var (
		err    error
		status *table.SteveStatus
	)
	status, err = pt.GetItemStatus(node)
	if err != nil {
		return nil, err
	}
	if status, err = pt.GetItemStatus(node); err != nil {
		return nil, err
	}
	allocatable := cast.ToFloat64(node.String("extra", "parsedResource", "allocated", "Pods"))
	capacity := cast.ToFloat64(node.String("extra", "parsedResource", "capacity", "Pods"))
	ur := table.DistributionValue{Percent: common.GetPercent(allocatable, capacity)}
	role := node.StringSlice("metadata", "fields")[2]
	if role == "<none>" {
		role = "worker"
	}
	ri := &table.RowItem{
		ID:      node.String("id"),
		IP:      node.StringSlice("metadata", "fields")[5],
		Version: node.String("status", "nodeInfo", "kubeletVersion"),
		Role:    role,
		Node: table.Node{
			RenderType: "multiple",
			Renders:    pt.GetRenders(node.String("id"), node.Map("metadata", "labels")),
		},
		Status: *status,
		UsageRate: table.Distribution{
			RenderType: "progress",
			Value:      ur.Percent,
			Status:     table.GetDistributionStatus(ur.Percent),
			Tip:        fmt.Sprintf("%d/%d", int64(allocatable), int64(capacity)),
		},
		Operate:      pt.GetOperate(node.String("id")),
		BatchOptions: []string{"delete", "freeze", "unfreeze"},
	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "podTable", func() servicehub.Provider {
		pi := PodInfoTable{}
		pi.Type = "Table"
		pi.State = table.State{}
		return &pi
	})
}
