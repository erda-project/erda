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
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var steveServer cmp.SteveServer

func (pt *PodInfoTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return pt.DefaultProvider.Init(ctx)
}

func (pt *PodInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &pt.State)
	if err != nil {
		return err
	}
	pt.SDK = cputil.SDK(ctx)
	pt.Operations = pt.GetTableOperation()
	pt.Ctx = ctx
	pt.Table.TableComponent = pt
	pt.Ctx = ctx
	pt.Server = steveServer
	pt.getProps()
	activeKey := (*gs)["activeKey"].(string)
	// Tab name not equal this component name
	if activeKey != tableTabs.POD_TAB {
		pt.Props["visible"] = false
		return pt.SetComponentValue(c)
	} else {
		pt.Props["visible"] = true
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
			req.ClusterName = pt.SDK.InParams["clusterName"].(string)
			req.OrgID = pt.SDK.Identity.OrgID
			req.UserID = pt.SDK.Identity.UserID
			req.Type = apistructs.K8SNode
			req.Name = nodeId
			err = pt.Server.UnlabelNode(pt.Ctx, &req, []string{labelKey})
			if err != nil {
				return err
			}
		case common.CMPDashboardUncordonNode:
			(*gs)["SelectedRowKeys"] = pt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardUncordonNode
		case common.CMPDashboardCordonNode:
			(*gs)["SelectedRowKeys"] = pt.State.SelectedRowKeys
			(*gs)["OperationKey"] = common.CMPDashboardCordonNode
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	}
	nodes, err := pt.GetNodes(ctx, gs)
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
		"isLoadMore":     true,
		"rowKey":         "id",
		"sortDirections": []string{"descend", "ascend"},
		"columns": []table.Columns{
			{DataIndex: "Status", Title: pt.SDK.I18n("status"), Sortable: true, Width: 100, Fixed: "left"},
			{DataIndex: "Node", Title: pt.SDK.I18n("node"), Sortable: true, Width: 320},
			{DataIndex: "Usage", Title: pt.SDK.I18n("usedRate"), Sortable: true},
			{DataIndex: "IP", Title: pt.SDK.I18n("ip"), Sortable: true, Width: 100},
			{DataIndex: "Role", Title: "Role", Sortable: true, Width: 120},
			{DataIndex: "Version", Title: pt.SDK.I18n("version"), Sortable: true, Width: 120},
			{DataIndex: "Operate", Title: pt.SDK.I18n("operate"), Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"cordon", "uncordon"},
		"scroll":          table.Scroll{X: 1200},
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
	ip := node.StringSlice("metadata", "fields")[5]
	if role == "<none>" {
		role = "worker"
	}
	batchOperations := make([]string, 0)
	if !strings.Contains(role, "master") {
		if strings.Contains(status.Value, pt.SDK.I18n("SchedulingDisabled")) {
			batchOperations = []string{"uncordon"}
		} else {
			batchOperations = []string{"cordon"}
		}
	}
	ri := &table.RowItem{
		ID:      node.String("metadata", "name"),
		IP:      ip,
		Version: node.String("status", "nodeInfo", "kubeletVersion"),
		Role:    role,
		Node: table.Node{
			RenderType: "multiple",
			Renders:    pt.GetRenders(node.String("metadata", "name"), ip, node.Map("metadata", "labels")),
		},
		Status: *status,
		Usage: table.Distribution{
			RenderType: "progress",
			Value:      ur.Percent,
			Status:     table.GetDistributionStatus(ur.Percent),
			Tip:        pt.GetScaleValue(allocatable, capacity, table.Pod),
		},
		Operate:         pt.GetOperate(node.String("metadata", "name")),
		BatchOperations: batchOperations,
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
