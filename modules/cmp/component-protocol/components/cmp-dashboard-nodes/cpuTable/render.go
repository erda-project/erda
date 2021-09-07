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

package cpuTable

import (
	"context"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (ct *CpuInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &ct.State)
	if err != nil {
		return err
	}
	ct.SDK = cputil.SDK(ctx)
	ct.Operations = ct.GetTableOperation()
	ct.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	ct.getProps()
	ct.GetBatchOperation()
	ct.TableComponent = ct
	activeKey := (*gs)["activeKey"].(string)
	if activeKey != tableTabs.CPU_TAB {
		ct.Props["visible"] = false
		return ct.SetComponentValue(c)
	} else {
		ct.Props["visible"] = true
	}
	if event.Operation != cptype.InitializeOperation {
		// Tab name not equal this component name
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey:
			if err = ct.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardChangePageNoOperationKey:
			if err = ct.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardSortByColumnOperationKey:
			ct.State.PageNo = 1
		case common.CMPDashboardDeleteNode:
			err := ct.DeleteNode(ct.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		case common.CMPDashboardUnfreezeNode:
			err := ct.UnFreezeNode(ct.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		case common.CMPDashboardFreezeNode:
			err := ct.FreezeNode(ct.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	nodes, err := ct.GetNodes(gs)
	if err != nil {
		return err
	}
	if err = ct.RenderList(c, table.Cpu, nodes); err != nil {
		return err
	}
	if err = ct.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (ct *CpuInfoTable) getProps() {
	props := map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "Status", Title: ct.SDK.I18n("status"), Sortable: true, Width: 80, Fixed: "left"},
			{DataIndex: "Node", Title: ct.SDK.I18n("node"), Sortable: true},
			{DataIndex: "IP", Title: ct.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: ct.SDK.I18n("role"), Sortable: true, Width: 120},
			{DataIndex: "Version", Title: ct.SDK.I18n("version"), Width: 120},
			{DataIndex: "Distribution", Title: "cpu" + ct.SDK.I18n("distribution"), Sortable: true, Width: 120},
			{DataIndex: "Usage", Title: "cpu" + ct.SDK.I18n("use"), Sortable: true, Width: 120},
			{DataIndex: "UsageRate", Title: "cpu" + ct.SDK.I18n("distributionRate"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: ct.SDK.I18n("operate"), Sortable: true, Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},

		"scroll": table.Scroll{X: 1200},
	}
	ct.Props = props
}

func (ct *CpuInfoTable) GetRowItem(c data.Object, tableType table.TableType) (*table.RowItem, error) {
	var (
		err                     error
		status                  *table.SteveStatus
		distribution, dr, usage table.DistributionValue
		resp                    []apistructs.MetricsData
		nodeLabels              []string
	)
	nodeLabelsData := c.Map("metadata", "labels")
	for k := range nodeLabelsData {
		nodeLabels = append(nodeLabels, k)
	}
	if status, err = ct.GetItemStatus(c); err != nil {
		return nil, err
	}
	req := apistructs.MetricsRequest{
		ClusterName:  ct.SDK.InParams["clusterName"].(string),
		Names:        []string{c.String("id")},
		ResourceType: metrics.Cpu,
		ResourceKind: metrics.Node,
		OrgID:        ct.SDK.Identity.OrgID,
		UserID:       ct.SDK.Identity.UserID,
	}

	if resp, err = ct.CtxBdl.GetMetrics(req); err != nil || resp == nil {
		logrus.Errorf("metrics error: %v", err)
		resp = []apistructs.MetricsData{{Used: 0}}
	}
	//request := c.Map("status", "allocatable").String("cpu")
	limit := c.Map("status", "capacity").String("cpu")
	resp[0].Total = cast.ToFloat64(limit)
	distribution = ct.GetDistributionValue(resp[0])
	usage = ct.GetUsageValue(resp[0])
	dr = ct.GetDistributionRate(resp[0])
	role := c.StringSlice("metadata", "fields")[2]
	if role == "<none>" {
		role = "worker"
	}
	ri := &table.RowItem{
		ID:      c.String("id"),
		IP:      c.StringSlice("metadata", "fields")[5],
		Version: c.String("status", "nodeInfo", "kubeletVersion"),
		Role:    role,
		Node: table.Node{
			RenderType: "multiple",
			Renders:    ct.GetRenders(c.String("id"), c.Map("metadata", "labels")),
		},
		Status: *status,
		Distribution: table.Distribution{
			RenderType: "progress",
			Value:      distribution.Percent,
			Status:     table.GetDistributionStatus(distribution.Percent),
			Tip:        distribution.Text,
		},
		Usage: table.Distribution{
			RenderType: "progress",
			Value:      usage.Percent,
			Status:     table.GetDistributionStatus(usage.Percent),
			Tip:        usage.Text,
		},
		UsageRate: table.Distribution{
			RenderType: "progress",
			Value:      dr.Percent,
			Status:     table.GetDistributionStatus(dr.Percent),
			Tip:        dr.Text,
		},
		Operate:      ct.GetOperate(c.String("id")),
		BatchOptions: []string{"delete", "freeze", "unfreeze"},
	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "cpuTable", func() servicehub.Provider {
		ci := CpuInfoTable{}
		return &ci
	})
}
