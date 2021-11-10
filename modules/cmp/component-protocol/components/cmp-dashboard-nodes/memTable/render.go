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
	"math"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var steveServer cmp.SteveServer
var mServer metrics.Interface

func (mt *MemInfoTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	mserver, ok := ctx.Service("cmp").(metrics.Interface)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a metrics server")
	}
	steveServer = server
	mServer = mserver
	return mt.DefaultProvider.Init(ctx)
}

func (mt *MemInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &mt.State)
	if err != nil {
		return err
	}
	mt.SDK = cputil.SDK(ctx)
	mt.Operations = mt.GetTableOperation()
	mt.Ctx = ctx
	mt.Table.TableComponent = mt
	mt.Server = steveServer
	mt.Metrics = mServer
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
		case common.CMPDashboardUncordonNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardUncordonNode
		case common.CMPDashboardCordonNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardCordonNode
		case common.CMPDashboardDrainNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardDrainNode
		case common.CMPDashboardOfflineNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardOfflineNode
		case common.CMPDashboardOnlineNode:
			(*gs)["SelectedRowKeys"] = mt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardOnlineNode
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	if err = mt.RenderList(c, table.Memory, gs); err != nil {
		return err
	}
	if err = mt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (mt *MemInfoTable) GetRowItems(nodes []data.Object, tableType table.TableType, requests map[string]cmp.AllocatedRes) ([]table.RowItem, error) {
	var (
		err                     error
		status                  *table.SteveStatus
		distribution, dr, usage table.DistributionValue
		clusterName             string
		items                   []table.RowItem
	)
	if mt.SDK.InParams["clusterName"] != nil {
		clusterName = mt.SDK.InParams["clusterName"].(string)
	} else {
		return nil, common.ClusterNotFoundErr
	}
	for _, c := range nodes {
		if status, err = mt.GetItemStatus(c); err != nil {
			return nil, err
		}
		Ip := c.StringSlice("metadata", "fields")[5]
		nodeName := c.StringSlice("metadata", "fields")[0]
		memRequest := requests[nodeName].Mem
		requestQty, _ := resource.ParseQuantity(c.String("status", "allocatable", "memory"))

		distribution = mt.GetDistributionValue(float64(memRequest), float64(requestQty.Value()), table.Memory)
		key := cache.GenerateKey(metrics.Memory, metrics.Node, clusterName, Ip)
		metricsData := metrics.GetCache(key)
		used := 0.0
		if metricsData != nil {
			used = metricsData.Used
		}
		usage = mt.GetUsageValue(used, float64(requestQty.Value()), table.Memory)
		unused := math.Max(float64(memRequest)-used, 0.0)
		dr = mt.GetUnusedRate(unused, float64(memRequest), table.Memory)
		role := c.StringSlice("metadata", "fields")[2]
		ip := c.StringSlice("metadata", "fields")[5]
		if role == "<none>" {
			role = "worker"
		}
		batchOperations := make([]string, 0)
		if !strings.Contains(role, "master") {
			if c.String("spec", "unschedulable") == "true" {
				if !table.IsNodeOffline(c) {
					batchOperations = append(batchOperations, "uncordon")
				}
			} else {
				batchOperations = append(batchOperations, "cordon")
			}
		}
		if role == "worker" && !table.IsNodeLabelInBlacklist(c) {
			//if !table.IsNodeOffline(c) {
			batchOperations = append(batchOperations, "drain")
			//	if c.String("spec", "unschedulable") == "true" && !table.IsNodeOffline(c) {
			//		batchOperations = append(batchOperations, "offline")
			//	}
			//} else {
			//	batchOperations = append(batchOperations, "online")
			//}
		}

		items = append(items, table.RowItem{
			ID:      c.String("metadata", "name"),
			IP:      ip,
			NodeID:  c.String("metadata", "name"),
			Version: c.String("status", "nodeInfo", "kubeletVersion"),
			Role:    role,
			Node: table.Node{
				RenderType: "multiple",
				Renders:    mt.GetRenders(c.String("metadata", "name"), c.Map("metadata", "labels")),
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
		},
		)
	}
	return items, nil
}

func (mt *MemInfoTable) getProps() {
	mt.Props = map[string]interface{}{
		"isLoadMore":     true,
		"rowKey":         "id",
		"sortDirections": []string{"descend", "ascend"},
		"columns": []table.Columns{
			{DataIndex: "Status", Title: mt.SDK.I18n("status"), Sortable: true, Width: 100, Fixed: "left"},
			{DataIndex: "Node", Title: mt.SDK.I18n("node"), Sortable: true, Width: 320},
			{DataIndex: "Distribution", Title: mt.SDK.I18n("distribution"), Sortable: true, Width: 130},
			{DataIndex: "Usage", Title: mt.SDK.I18n("usedRate"), Sortable: true, Width: 130},
			{DataIndex: "UnusedRate", Title: mt.SDK.I18n("unusedRate"), Sortable: true, Width: 140, TitleTip: mt.SDK.I18n("The proportion of allocated resources that are not used")},
			{DataIndex: "IP", Title: mt.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: "Role", Sortable: true, Width: 120},
			{DataIndex: "Version", Title: mt.SDK.I18n("version"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: mt.SDK.I18n("podsList"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"cordon", "uncordon", "drain"},
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
