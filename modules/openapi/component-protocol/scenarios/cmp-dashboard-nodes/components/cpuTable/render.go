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

package cpuTable

import (
	"context"
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/components/tab"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var tableProperties = map[string]interface{}{
	"rowKey": "id",
	"columns": []table.Columns{
		{DataIndex: "status", Title: "状态"},
		{DataIndex: "node", Title: "节点"},
		{DataIndex: "role", Title: "角色"},
		{DataIndex: "version", Title: "版本"},
		{DataIndex: "distribution", Title: "cpu分配率"},
		{DataIndex: "use", Title: "cpu使用率"},
		{DataIndex: "distribution", Title: "cpu分配使用率"},
	},
	"bordered":        true,
	"selectable":      true,
	"pageSizeOptions": []string{"10", "20", "50", "100"},
}

func (ct *CpuInfoTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		err   error
		state table.State
	)
	ct.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	ct.Ctx = ctx
	err = common.Transfer(c.State, &state)
	if err != nil {
		return err
	}
	ct.State = state
	if event.Operation != apistructs.InitializeOperation {
		// Tab name not equal this component name
		if c.State["activeKey"].(string) != tab.CPU_TAB {
			return nil
		}
		switch event.Operation {
		case apistructs.CMPDashboardChangePageSizeOperationKey:
			if err = ct.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case apistructs.CMPDashboardChangePageNoOperationKey:
			if err = ct.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case apistructs.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if ct.State.IsFirstFilter {
				ct.State.PageNo = 1
				ct.State.IsFirstFilter = false
			}
		case apistructs.CMPDashboardSortByColumnOperationKey:
			ct.State.PageNo = 1
			ct.State.IsFirstFilter = false
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	} else {
		ct.Props["visible"] = true
		return nil
	}
	if err = ct.RenderList(c, event, v1.ResourceCPU); err != nil {
		return err
	}
	if err = ct.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (ct *CpuInfoTable) RenderList(component *apistructs.Component, event apistructs.ComponentEvent, resName v1.ResourceName) error {
	var (
		nodeList   []apistructs.SteveResource
		nodes      []apistructs.SteveResource
		resp       *apistructs.SteveCollection
		err        error
		filter     string
		sortColumn string
		orgID      string
		userID     string
		asc        bool
	)
	if ct.State.PageNo == 0 {
		ct.State.PageNo = DefaultPageNo
	}
	if ct.State.PageSize == 0 {
		ct.State.PageSize = DefaultPageSize
	}
	pageNo := ct.State.PageNo
	pageSize := ct.State.PageSize
	filter = ct.State.Query["title"].(string)
	sortColumn = ct.State.SortColumnName
	asc = ct.State.Asc
	clusterName := ct.CtxBdl.InParams["clusterName"].(string)
	orgID = ct.CtxBdl.Identity.OrgID
	userID = ct.CtxBdl.Identity.UserID
	nodeReq := &apistructs.SteveRequest{}
	nodeReq.ClusterName = clusterName
	nodeReq.Type = apistructs.K8SNode
	nodeReq.OrgID = orgID
	nodeReq.UserID = userID
	resp, err = ct.CtxBdl.Bdl.ListSteveResource(nodeReq)
	if err != nil {
		return err
	}
	nodeList = append(nodeList, resp.Data...)
	if filter == "" {
		nodes = nodeList
	} else {
		// Filter by node name or node uid
		for _, node := range nodeList {
			if strings.Contains(node.Metadata.Name, filter) || strings.Contains(node.ID, filter) {
				nodes = append(nodes, node)
			}
		}
	}
	// transfer and set data into table
	if err = ct.SetData(nodes, resName); err != nil {
		return err
	}
	if sortColumn != "" {
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{ct.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Node{}):
			common.SortByNode([]interface{}{ct.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{ct.Data}, sortColumn, asc)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}

	nodes = nodes[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(nodes))]
	component.Data["list"] = nodes
	return nil
}

// SetData assemble rowItem of table
func (ct *CpuInfoTable) SetData(nodes []apistructs.SteveResource, resName v1.ResourceName) error {
	var (
		lists []table.RowItem
		ri    *table.RowItem
		err   error
	)
	ct.State.Total = len(nodes)
	start := (ct.State.PageNo - 1) * ct.State.PageSize
	end := mathutil.Max(ct.State.PageNo*ct.State.PageSize, ct.State.Total)

	for i := start; i < end; i++ {
		if ri, err = ct.GetRowItem(nodes[i], resName); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	return nil
}
func getProps() map[string]interface{} {
	return tableProperties
}
func getTableOperation() map[string]interface{} {
	ops := map[string]table.Operation{
		"changePageNo": {
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": {
			Key:    "changePageSize",
			Reload: true,
		},
	}
	res := map[string]interface{}{}
	for key, op := range ops {
		res[key] = interface{}(op)
	}
	return res
}

func (ct *CpuInfoTable) GetRowItem(c apistructs.SteveResource, resName v1.ResourceName) (*table.RowItem, error) {
	var (
		err                     error
		status                  *common.SteveStatus
		distribution, dr, usage *table.DistributionValue
		node                    v1.Node
	)
	if err = common.Transfer(c, &node); err != nil {
		return nil, err
	}
	nodeLabels := c.Metadata.Labels
	if status, err = ct.GetItemStatus(&node); err != nil {
		return nil, err
	}
	if distribution, err = ct.GetDistributionValue(&node, resName); err != nil {
		return nil, err
	}
	if usage, err = ct.GetUsageValue(&node, resName); err != nil {
		return nil, err
	}
	if dr, err = ct.GetDistributionRate(&node, resName); err != nil {
		return nil, err
	}
	ri := &table.RowItem{
		ID:      node.Name,
		Version: node.Status.NodeInfo.KubeletVersion,
		Role:    ct.GetRole(nodeLabels),
		Labels: table.Labels{
			RenderType: "tagsColumn",
			Value:      ct.GetPodLabels(node.GetLabels()),
			Operation:  ct.GetLabelOperation(string(node.UID)),
		},
		Node: table.Node{
			RenderType: "linkText",
			Value:      ct.GetNodeAddress(node.Status.Addresses),
			Operation:  ct.GetNodeOperation(),
			Reload:     false,
		},
		Status: *status,
		Distribution: table.Distribution{
			RenderType: "bgProgress",
			Value:      *distribution,
		},
		Usage: table.Distribution{
			RenderType: "bgProgress",
			Value:      *usage,
		},
		DistributionRate: table.Distribution{
			RenderType: "bgProgress",
			Value:      *dr,
		},
	}
	return ri, nil
}

func RenderCreator() protocol.CompRender {
	ci := CpuInfoTable{}
	ci.Type = "Table"
	ci.Props = getProps()
	ci.Operations = getTableOperation()
	ci.State = table.State{}
	return &ci
}
