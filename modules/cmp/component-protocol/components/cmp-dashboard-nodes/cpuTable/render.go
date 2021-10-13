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
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	cputil2 "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var (
	steveServer cmp.SteveServer
	mServer     metrics.Interface
)

func (ct *CpuInfoTable) Init(ctx servicehub.Context) error {
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
	return ct.DefaultProvider.Init(ctx)
}

func (ct *CpuInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &ct.State)
	if err != nil {
		return err
	}
	ct.SDK = cputil.SDK(ctx)
	ct.Operations = ct.GetTableOperation()
	ct.Ctx = ctx
	ct.Table.Server = steveServer
	ct.getProps()
	ct.TableComponent = ct
	ct.Ctx = ctx
	activeKey := (*gs)["activeKey"].(string)
	// Tab name not equal this component name
	if activeKey != tableTabs.CPU_TAB {
		ct.Props["visible"] = false
		return ct.SetComponentValue(c)
	} else {
		ct.Props["visible"] = true
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
			req.ClusterName = ct.SDK.InParams["clusterName"].(string)
			req.OrgID = ct.SDK.Identity.OrgID
			req.UserID = ct.SDK.Identity.UserID
			req.Type = apistructs.K8SNode
			req.Name = nodeId
			err = steveServer.UnlabelNode(ctx, &req, []string{labelKey})
		case common.CMPDashboardUncordonNode:
			(*gs)["SelectedRowKeys"] = ct.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardUncordonNode
		case common.CMPDashboardCordonNode:
			(*gs)["SelectedRowKeys"] = ct.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardCordonNode
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	nodes, err := ct.GetNodes(ctx, gs)
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
		"isLoadMore":     true,
		"rowKey":         "id",
		"sortDirections": []string{"descend", "ascend"},
		"columns": []table.Columns{
			{DataIndex: "Status", Title: ct.SDK.I18n("status"), Sortable: true, Width: 100, Fixed: "left"},
			{DataIndex: "Node", Title: ct.SDK.I18n("node"), Sortable: true, Width: 320},
			{DataIndex: "Distribution", Title: ct.SDK.I18n("distribution"), Sortable: true, Width: 130},
			{DataIndex: "Usage", Title: ct.SDK.I18n("usedRate"), Sortable: true, Width: 130},
			{DataIndex: "UnusedRate", Title: ct.SDK.I18n("unusedRate"), Sortable: true, Width: 140, TitleTip: ct.SDK.I18n("The proportion of allocated resources that are not used")},
			{DataIndex: "IP", Title: ct.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: "Role", Sortable: true, Width: 120},
			{DataIndex: "Version", Title: ct.SDK.I18n("version"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: ct.SDK.I18n("podsList"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"cordon", "uncordon"},
		"scroll":          table.Scroll{X: 1200},
	}
	ct.Props = props
}

func (ct *CpuInfoTable) GetRowItems(nodes []data.Object, tableType table.TableType) ([]table.RowItem, error) {
	var (
		err                     error
		status                  *table.SteveStatus
		distribution, dr, usage table.DistributionValue
		resp                    map[string]*metrics.MetricsData
		nodeLabels              []string
		items                   []table.RowItem
	)
	clusterName := ""
	if ct.SDK.InParams["clusterName"] != nil {
		clusterName = ct.SDK.InParams["clusterName"].(string)
	} else {
		return nil, common.ClusterNotFoundErr
	}
	req := &metrics.MetricsRequest{
		Cluster: ct.SDK.InParams["clusterName"].(string),
		Type:    metrics.Cpu,
		Kind:    metrics.Node,
	}
	for _, node := range nodes {
		req.NodeRequests = append(req.NodeRequests, &metrics.MetricsNodeRequest{
			MetricsRequest: req,
			Ip:             node.StringSlice("metadata", "fields")[5],
		})
	}
	if resp, err = mServer.NodeMetrics(ct.Ctx, req); err != nil || resp == nil {
		logrus.Errorf("metrics error: %v", err)
	}
	for i, c := range nodes {
		nodeLabelsData := c.Map("metadata", "labels")
		for k := range nodeLabelsData {
			nodeLabels = append(nodeLabels, k)
		}
		if status, err = ct.GetItemStatus(c); err != nil {
			return nil, err
		}
		//request := c.Map("status", "allocatable").String("cpu")
		nodeName := c.StringSlice("metadata", "fields")[0]
		cpuRequest, _, _, err := cputil2.GetNodeAllocatedRes(steveServer, clusterName, ct.SDK.Identity.UserID, ct.SDK.Identity.OrgID, nodeName)
		if err != nil {
			return nil, err
		}
		requestQty, _ := resource.ParseQuantity(c.String("status", "allocatable", "cpu"))

		key := req.NodeRequests[i].CacheKey()
		distribution = ct.GetDistributionValue(float64(cpuRequest), float64(requestQty.ScaledValue(resource.Milli)), table.Cpu)
		metricsData, ok := resp[key]
		used := 0.0
		if ok {
			used = metricsData.Used
		}

		usage = ct.GetUsageValue(used, float64(requestQty.Value()), table.Cpu)
		unused := math.Max(float64(cpuRequest)-resp[key].Used*1000, 0.0)
		dr = ct.GetUnusedRate(unused, float64(cpuRequest), table.Cpu)
		role := c.StringSlice("metadata", "fields")[2]
		ip := c.StringSlice("metadata", "fields")[5]
		if role == "<none>" {
			role = "worker"
		}
		batchOperations := make([]string, 0)
		if !strings.Contains(role, "master") {
			if strings.Contains(status.Value, ct.SDK.I18n("SchedulingDisabled")) {
				batchOperations = []string{"uncordon"}
			} else {
				batchOperations = []string{"cordon"}
			}
		}
		items = append(items, table.RowItem{
			ID:      c.String("metadata", "name"),
			IP:      ip,
			Version: c.String("status", "nodeInfo", "kubeletVersion"),
			Role:    role,
			Node: table.Node{
				RenderType: "multiple",
				Renders:    ct.GetRenders(c.String("metadata", "name"), ip, c.Map("metadata", "labels")),
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
			Operate:         ct.GetOperate(c.String("metadata", "name")),
			BatchOperations: batchOperations,
		},
		)
	}
	return items, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "cpuTable", func() servicehub.Provider {
		ci := CpuInfoTable{}
		return &ci
	})
}
