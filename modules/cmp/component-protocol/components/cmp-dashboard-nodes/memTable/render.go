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

package memTable

import (
	"context"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

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

func (mt *MemInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &mt.State)
	if err != nil {
		return err
	}
	mt.SDK = cputil.SDK(ctx)
	mt.Operations = mt.GetTableOperation()
	mt.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	mt.Table.TableComponent = mt
	mt.getProps()
	activeKey := (*gs)["activeKey"].(string)
	// Tab name not equal this component name
	if activeKey != tableTabs.MEM_TAB {
		mt.Props["visible"] = false
		return mt.SetComponentValue(c)
	} else {
		mt.Props["visible"] = true
	}
	if event.Operation != cptype.InitializeOperation {
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey, common.CMPDashboardChangePageNoOperationKey:
		case common.CMPDashboardSortByColumnOperationKey:
		case common.CMPDashboardRemoveLabel:
			if event.Operation == common.CMPDashboardRemoveLabel {
				metaName := event.OperationData["fillMeta"].(string)
				label := event.OperationData["meta"].(map[string]interface{})[metaName].(map[string]interface{})["label"].(string)
				nodeId := event.OperationData["meta"].(map[string]interface{})[metaName].(map[string]interface{})["recordId"].(string)
				req := apistructs.SteveRequest{}
				req.ClusterName = mt.SDK.InParams["clusterName"].(string)
				req.OrgID = mt.SDK.Identity.OrgID
				req.UserID = mt.SDK.Identity.UserID
				req.Type = apistructs.K8SNode
				req.Name = nodeId
				err = mt.CtxBdl.UnlabelNode(&req, []string{label})
				if err != nil {
					return err
				}
			}
		case common.CMPDashboardUncordonNode:
			err := mt.UncordonNode(mt.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		case common.CMPDashboardCordonNode:
			err := mt.CordonNode(mt.State.SelectedRowKeys)
			if err != nil {
				return err
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	nodes, err := mt.GetNodes(gs)
	if err != nil {
		return err
	}
	if err = mt.RenderList(c, table.Memory, nodes); err != nil {
		return err
	}
	if err = mt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (mt *MemInfoTable) GetRowItem(c data.Object, tableType table.TableType) (*table.RowItem, error) {
	var (
		err                     error
		status                  *table.SteveStatus
		distribution, dr, usage table.DistributionValue
		resp                    []apistructs.MetricsData
	)
	if status, err = mt.GetItemStatus(c); err != nil {
		return nil, err
	}
	req := apistructs.MetricsRequest{
		ClusterName:  mt.SDK.InParams["clusterName"].(string),
		IP:           []string{c.StringSlice("metadata", "fields")[5]},
		ResourceType: metrics.Memory,
		ResourceKind: metrics.Node,
		OrgID:        mt.SDK.Identity.OrgID,
		UserID:       mt.SDK.Identity.UserID,
	}
	if resp, err = mt.CtxBdl.GetMetrics(req); err != nil || resp == nil {
		logrus.Errorf("metrics error: %v", err)
		resp = []apistructs.MetricsData{{Used: 0}}
	}
	limitStr := c.Map("extra", "parsedResource", "capacity").String("Memory")
	limitQuantity, _ := resource.ParseQuantity(limitStr)
	requestStr := c.Map("extra", "parsedResource", "allocated").String("Memory")
	requestQuantity, _ := resource.ParseQuantity(requestStr)
	resp[0].Total = float64(limitQuantity.Value())
	resp[0].Request = float64(requestQuantity.Value())
	distribution = mt.GetDistributionValue(resp[0], table.Memory)
	usage = mt.GetUsageValue(resp[0], table.Memory)
	dr = mt.GetDistributionRate(resp[0], table.Memory)
	role := c.StringSlice("metadata", "fields")[2]
	ip := c.StringSlice("metadata", "fields")[5]
	if role == "<none>" {
		role = "worker"
	}
	ri := &table.RowItem{
		ID:      c.String("id"),
		IP:      ip,
		Version: c.String("status", "nodeInfo", "kubeletVersion"),
		Role:    role,
		Node: table.Node{
			RenderType: "multiple",
			Renders:    mt.GetRenders(c.String("id"), ip, c.Map("metadata", "labels")),
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
		Operate:         mt.GetOperate(c.String("id")),
		BatchOperations: []string{"cordon", "uncordon"},
	}
	return ri, nil
}

func (mt *MemInfoTable) getProps() {
	mt.Props = map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "Status", Title: mt.SDK.I18n("status"), Sortable: true, Width: 100, Fixed: "left"},
			{DataIndex: "Node", Title: mt.SDK.I18n("node"), Sortable: true, Width: 340},
			{DataIndex: "IP", Title: mt.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: mt.SDK.I18n("role"), Sortable: true, Width: 120},
			{DataIndex: "Version", Title: mt.SDK.I18n("version"), Width: 120},
			{DataIndex: "Distribution", Title: mt.SDK.I18n("memDistribution"), Sortable: true, Width: 120},
			{DataIndex: "Usage", Title: mt.SDK.I18n("memUsed"), Sortable: true, Width: 120},
			{DataIndex: "UsageRate", Title: mt.SDK.I18n("memDistributionRate"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: mt.SDK.I18n("operate"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"cordon", "cordon"},
		"scroll":          table.Scroll{X: 1200},
	}

}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "memTable", func() servicehub.Provider {
		mt := MemInfoTable{}
		mt.Type = "Table"
		return &mt
	})
}
