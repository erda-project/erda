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
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
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
		{DataIndex: "distributionRate", Title: "cpu分配使用率"},
	},
	"bordered":        true,
	"selectable":      true,
	"pageSizeOptions": []string{"10", "20", "50", "100"},
}

func (mt *MemInfoTable) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var state table.State
	if v := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String());v == nil{
		return common.ResourceEmptyErr
	}else{
		mt.CtxBdl = v.(protocol.ContextBundle)
	}
	common.Transfer(c.State, &state)
	mt.State = state
	if event.Operation != apistructs.InitializeOperation {
		// Tab name not equal this component name
		if c.State["activeKey"].(string) != tab.MEM_TAB {
			return nil
		}
		switch event.Operation {
		case apistructs.CMPDashboardChangePageSizeOperationKey:
			if err := mt.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case apistructs.CMPDashboardChangePageNoOperationKey:
			if err := mt.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case apistructs.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if mt.State.IsFirstFilter {
				mt.State.PageNo = 1
				mt.State.IsFirstFilter = false
			}
		case apistructs.CMPDashboardSortByColumnOperationKey:
			mt.State.PageNo = 1
			mt.State.IsFirstFilter = false
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	} else {
		mt.Props["visible"] = false
	}
	if err := mt.RenderList(c, event); err != nil {
		return err
	}
	if err := mt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (mt *MemInfoTable) RenderList(component *apistructs.Component, event apistructs.ComponentEvent) error {
	var (
		nodeList     []apistructs.SteveResource
		nodes        []apistructs.SteveResource
		err          error
		filter       string
		sortColumn   string
		orgID        int64
		asc          bool
		clusterNames []apistructs.ClusterInfo
	)
	if mt.State.PageNo == 0 {
		mt.State.PageNo = DefaultPageNo
	}
	if mt.State.PageSize == 0 {
		mt.State.PageSize = DefaultPageSize
	}
	pageNo := mt.State.PageNo
	pageSize := mt.State.PageSize
	filter = mt.State.Query["title"].(string)
	sortColumn = mt.State.SortColumnName
	asc = mt.State.Asc

	if mt.State.ClusterName != "" {
		clusterNames = append([]apistructs.ClusterInfo{}, apistructs.ClusterInfo{Name: mt.State.ClusterName})
	} else {
		clusterNames, err = bdl.Bdl.ListClusters("", uint64(orgID))
		if err != nil {
			return err
		}
	}
	// Get all nodes by cluster name
	for _, clusterName := range clusterNames {
		nodeReq := &apistructs.SteveRequest{}
		nodeReq.Name = clusterName.Name
		nodeReq.ClusterName = clusterName.Name
		resp, err := bdl.Bdl.ListSteveResource(nodeReq)
		if err != nil {
			return err
		}
		nodeList = append(nodeList, resp.Data...)
	}
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
	if err = mt.SetData(nodes, v1.ResourceMemory); err != nil {
		return err
	}

	if sortColumn != "" {
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{mt.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Node{}):
			common.SortByNode([]interface{}{mt.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{mt.Data}, sortColumn, asc)
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
func (mt *MemInfoTable) SetData(nodes []apistructs.SteveResource, resName v1.ResourceName) error {
	var (
		lists []table.RowItem
		ri    *table.RowItem
		err   error
	)
	mt.State.Total = len(nodes)
	start := (mt.State.PageNo - 1) * mt.State.PageSize
	end := mathutil.Max(mt.State.PageNo*mt.State.PageSize, mt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = mt.GetRowItem(nodes[i], resName); err != nil {
			return err
		}
		lists = append(lists, *ri)
	}
	return nil
}
func (mt *MemInfoTable) GetRowItem(c apistructs.SteveResource, resName v1.ResourceName) (*table.RowItem, error) {
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
	if status, err = mt.GetItemStatus(&node); err != nil {
		return nil, err
	}
	if distribution, err = mt.GetDistributionValue(&node, resName); err != nil {
		return nil, err
	}
	if usage, err = mt.GetUsageValue(&node, resName); err != nil {
		return nil, err
	}
	if dr, err = mt.GetDistributionRate(&node, resName); err != nil {
		return nil, err
	}
	ri := &table.RowItem{
		ID:      node.Name,
		Version: node.Status.NodeInfo.KubeletVersion,
		Role:    mt.GetRole(nodeLabels),
		Labels: table.Labels{
			RenderType: "tagsColumn",
			Value:      mt.GetPodLabels(node.GetLabels()),
			Operation:  mt.GetLabelOperation(string(node.UID)),
		},
		Node: table.Node{
			RenderType: "linkText",
			Value:      mt.GetNodeAddress(node.Status.Addresses),
			Operation:  mt.GetNodeOperation(),
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

// SetComponentValue transfer MemInfoTable struct to Component
func (mt *MemInfoTable) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(mt.State); err != nil {
		return err
	}
	c.State = state
	c.Operations = mt.Operations
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

func getItemStatus(node *v1.Node) (*common.SteveStatus, error) {
	if node == nil {
		return nil, common.NodeNotFoundErr
	}
	ss := &common.SteveStatus{
		RenderType: "textWithBadge",
	}
	status := common.NodeStatusReady
	if node.Spec.Unschedulable {
		status = common.NodeStatusFreeze
	} else {
		for _, cond := range node.Status.Conditions {
			if cond.Status == v1.ConditionTrue && cond.Type == v1.NodeReady {
				status = common.NodeStatusError
				break
			}
		}
	}
	// 0:English 1:ZH
	ss.Status = common.GetNodeStatus(status)[0]
	ss.Value = common.GetNodeStatus(status)[1]
	return ss, nil
}

func RenderCreator() protocol.CompRender {
	mt := MemInfoTable{}
	mt.Type = "Table"
	mt.Props = getProps()
	mt.Operations = getTableOperation()
	mt.State = table.State{}
	return &mt
}
