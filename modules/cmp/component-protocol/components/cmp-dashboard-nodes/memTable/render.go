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
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/cmp"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var steveServer cmp.SteveServer

func (mt *MemInfoTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return mt.DefaultProvider.Init(ctx)
}

func (mt *MemInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &mt.State)
	if err != nil {
		return err
	}
	mt.SDK = cputil.SDK(ctx)
	mt.Operations = mt.GetTableOperation()
	mt.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	mt.Table.TableComponent = mt
	mt.Ctx = ctx
	mt.Server = steveServer
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
		//case common.CMPDashboardChangePageSizeOperationKey, common.CMPDashboardChangePageNoOperationKey:
		case common.CMPDashboardSortByColumnOperationKey:
		case common.CMPDashboardRemoveLabel:
			metaName := event.OperationData["fillMeta"].(string)
			label := event.OperationData["meta"].(map[string]interface{})[metaName].(map[string]interface{})["label"].(string)
			labelKey := strings.Split(label, "=")[0]
			nodeId := event.OperationData["meta"].(map[string]interface{})["recordId"].(string)
			req := apistructs.SteveRequest{}
			req.ClusterName = mt.SDK.InParams["clusterName"].(string)
			req.OrgID = mt.SDK.Identity.OrgID
			req.UserID = mt.SDK.Identity.UserID
			req.Type = apistructs.K8SNode
			req.Name = nodeId
			err = mt.Server.UnlabelNode(mt.Ctx, &req, []string{labelKey})
			if err != nil {
				return err
			}
		case common.CMPDashboardUncordonNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardUncordonNode
			return nil
		case common.CMPDashboardCordonNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardCordonNode
			return nil
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
		NodeRequests: []apistructs.MetricsNodeRequest{{IP: c.StringSlice("metadata", "fields")[5]}},
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
	dr = mt.GetUnusedRate(resp[0], table.Memory)
	role := c.StringSlice("metadata", "fields")[2]
	ip := c.StringSlice("metadata", "fields")[5]
	if role == "<none>" {
		role = "worker"
	}
	batchOperations := make([]string, 0)
	if !strings.Contains(role, "master") {
		if strings.Contains(status.Value, mt.SDK.I18n("SchedulingDisabled")) {
			batchOperations = []string{"uncordon"}
		} else {
			batchOperations = []string{"cordon"}
		}
	}
	ri := &table.RowItem{
		ID:      c.String("metadata", "name"),
		IP:      ip,
		Version: c.String("status", "nodeInfo", "kubeletVersion"),
		Role:    role,
		Node: table.Node{
			RenderType: "multiple",
			Renders:    mt.GetRenders(c.String("metadata", "name"), ip, c.Map("metadata", "labels")),
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
		UnusedRate: table.Distribution{
			RenderType: "progress",
			Value:      dr.Percent,
			Status:     table.GetDistributionStatus(dr.Percent),
			Tip:        dr.Text,
		},
		Operate:         mt.GetOperate(c.String("metadata", "name")),
		BatchOperations: batchOperations,
	}
	return ri, nil
}

func (mt *MemInfoTable) getProps() {
	mt.Props = map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "Status", Title: mt.SDK.I18n("status"), Sortable: true, Width: 100, Fixed: "left"},
			{DataIndex: "Node", Title: mt.SDK.I18n("node"), Sortable: true, Width: 320},
			{DataIndex: "Distribution", Title: mt.SDK.I18n("distribution"), Sortable: true, Width: 130},
			{DataIndex: "Usage", Title: mt.SDK.I18n("usedRate"), Sortable: true, Width: 130},
			{DataIndex: "UnusedRate", Title: mt.SDK.I18n("unusedRate"), Sortable: true, Width: 140, TitleTip: mt.SDK.I18n("The proportion of allocated resources that are not used")},
			{DataIndex: "IP", Title: mt.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: "Role", Sortable: true, Width: 120},
			{DataIndex: "Version", Title: mt.SDK.I18n("version"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: mt.SDK.I18n("operate"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"cordon", "uncordon"},
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
