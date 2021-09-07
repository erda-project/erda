// Copyright (c) 2021 Terminus, Innode.
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
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
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
	pt.Props = pt.getProps()
	pt.Operations = getTableOperation()
	activeKey:= (*gs)["activeKey"].(string)
	if event.Operation != cptype.InitializeOperation {
		if activeKey != tableTabs.POD_TAB {
			return pt.SetComponentValue(c)
		}
		if c.State["activeKey"] != tableTabs.POD_TAB {
			return nil
		}
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
	} else {
		pt.Props["visible"] = false
	}
	nodes := (*gs)["nodes"].([]data.Object)
	if err := pt.RenderList(c, event, nodes); err != nil {
		return err
	}
	if err := pt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}
func (pt *PodInfoTable) RenderList(component *cptype.Component, event cptype.ComponentEvent, nodes []data.Object) error {
	var (
		err        error
		sortColumn string
		asc        bool
		items      []table.RowItem
	)
	if pt.State.PageNo == 0 {
		pt.State.PageNo = DefaultPageNo
	}
	if pt.State.PageSize == 0 {
		pt.State.PageSize = DefaultPageSize
	}
	if pt.State.Sorter.Field != ""{
		sortColumn = pt.State.Sorter.Field
		asc = strings.ToLower(pt.State.Sorter.Order) == "ascend"
	}

	if sortColumn != "" {
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{pt.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Node{}):
			common.SortByNode([]interface{}{pt.Data}, sortColumn, asc)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{pt.Data}, sortColumn, asc)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}

	nodes = nodes[(pt.State.PageNo-1)*pt.State.PageSize : mathutil.Min(pt.State.PageNo*pt.State.PageSize, len(nodes))]
	if items, err = pt.SetData(nodes); err != nil {
		return err
	}
	component.Data = map[string]interface{}{"list":items}
	return nil
}

// SetComponentValue transfer CpuInfoTable struct to Component
func (pt *PodInfoTable) SetComponentValue(c *cptype.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(pt.State); err != nil {
		return err
	}
	c.State = state
	c.Operations = pt.Operations
	return nil
}

func (pt *PodInfoTable) getProps() map[string]interface{} {
	return map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "status", Title: pt.SDK.I18n("status")},
			{DataIndex: "node", Title: pt.SDK.I18n("node")},
			{DataIndex: "role", Title: pt.SDK.I18n("role")},
			{DataIndex: "version", Title: pt.SDK.I18n("version")},
			{DataIndex: "usageRate", Title: "pod" + pt.SDK.I18n("usage")},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
		"batchOperations": map[string]table.BatchOperation{
			"delete":   {Key: "delete", Text: pt.SDK.I18n("delete"), Reload: true, ShowIndex: nil},
			"freeze":   {Key: "freeze", Text: pt.SDK.I18n("freeze"), Reload: true, ShowIndex: nil},
			"unfreeze": {Key: "unfreeze", Text: pt.SDK.I18n("unfreeze"), Reload: true, ShowIndex: nil},
		},
		"scroll": table.Scroll{X: 1200},
	}
}

// SetData assemble rowItem of table
func (pt *PodInfoTable) SetData(nodes []data.Object) ([]table.RowItem, error) {
	var (
		list []table.RowItem
		ri   *table.RowItem
		err  error
	)
	pt.State.Total = len(nodes)
	start := (pt.State.PageNo - 1) * pt.State.PageSize
	end := mathutil.Min(pt.State.PageNo*pt.State.PageSize, pt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = pt.GetRowItem(nodes[i]); err != nil {
			return nil, err
		}
		list = append(list, *ri)
	}
	return list, nil
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

func (pt *PodInfoTable) GetRowItem(node data.Object) (*table.RowItem, error) {
	var (
		err    error
		status *common.SteveStatus
	)
	status, err = pt.GetItemStatus(node)
	if err != nil {
		return nil, err
	}
	if status, err = pt.GetItemStatus(node); err != nil{
		return nil, err
	}
	allocatable := cast.ToFloat64(node.String("status", "allocatable","pods"))
	capacity := cast.ToFloat64(node.String("status", "capacity","pods"))
	ri := &table.RowItem{
		ID:      node.String("id"),
		Version: node.String("status", "nodeInfo", "kubeletVersion"),
		Role:    node.StringSlice("metadata","fields")[2],
		Node: table.Node{
			RenderType: "multiple",
			Renders:    pt.GetRenders(node.String("id"), pt.GetIp(node), node.Map("metadata", "labels")),
		},
		Status: *status,
		UsageRate: table.Distribution{
			RenderType: "bgProgress",
			Value:      table.DistributionValue{Text: "pod" + pt.SDK.I18n("usage"), Percent: common.GetPercent(capacity-allocatable, capacity)},
		},
		Operate: pt.GetOperate(node.String("id")),
		BatchOptions: []string{"delete","freeze","unfreeze"},
	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "podTable", func() servicehub.Provider {
		pi := PodInfoTable{}
		pi.Type = "Table"
		pi.Operations = getTableOperation()
		pi.State = table.State{}
		return &pi
	})
}
