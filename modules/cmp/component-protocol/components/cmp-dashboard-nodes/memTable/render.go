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
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/tableTabs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

func (mt *MemInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	err := common.Transfer(c.State, &mt.State)
	if err != nil {
		return err
	}
	mt.SDK = cputil.SDK(ctx)
	mt.Operations = getTableOperation()
	mt.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	mt.Props = mt.getProps()
	activeKey := (*gs)["activeKey"].(string)
	if event.Operation != cptype.InitializeOperation {
		// Tab name not equal this component name
		if activeKey != tableTabs.MEM_TAB {
			return mt.SetComponentValue(c)
		}
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey:
			if err := mt.RenderChangePageSize(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardChangePageNoOperationKey:
			if err := mt.RenderChangePageNo(event.OperationData); err != nil {
				return err
			}
		case common.CMPDashboardSortByColumnOperationKey:
			mt.State.PageNo = 1
		case common.CMPDashboardDeleteNode:
			if err := mt.DeleteNode(mt.State.SelectedRowKeys); err != nil {
				return err
			}
		case common.CMPDashboardUnfreezeNode:
			if err := mt.UnFreezeNode(mt.State.SelectedRowKeys); err != nil {
				return err
			}
		case common.CMPDashboardFreezeNode:
			if err := mt.FreezeNode(mt.State.SelectedRowKeys); err != nil {
				return err
			}
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	} else {
		mt.Props["visible"] = false
	}
	nodesList := (*gs)["nodes"].([]data.Object)
	if err := mt.RenderList(c, event, nodesList); err != nil {
		return err
	}
	if err := mt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (mt *MemInfoTable) RenderList(component *cptype.Component, event cptype.ComponentEvent, nodeList []data.Object) error {
	var (
		items      []table.RowItem
		err        error
		sortColumn string
		asc        bool
	)
	if mt.State.PageNo == 0 {
		mt.State.PageNo = DefaultPageNo
	}
	if mt.State.PageSize == 0 {
		mt.State.PageSize = DefaultPageSize
	}
	if mt.State.Sorter.Field != ""{
		sortColumn = mt.State.Sorter.Field
		asc = strings.ToLower(mt.State.Sorter.Order) == "ascend"
	}
	if sortColumn != "" {
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{nodeList}, sortColumn, asc)
		case reflect.TypeOf(table.Node{}):
			common.SortByNode([]interface{}{nodeList}, sortColumn, asc)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{nodeList}, sortColumn, asc)

		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}
	if items, err = mt.SetData(nodeList, v1.ResourceMemory); err != nil {
		return err
	}
	component.Data = map[string]interface{}{"list":items}
	return nil
}

// SetData assemble rowItem of table
func (mt *MemInfoTable) SetData(nodes []data.Object, resName v1.ResourceName) ([]table.RowItem, error) {
	var (
		lists []table.RowItem
		ri    *table.RowItem
		err   error
	)
	mt.State.Total = len(nodes)
	start := (mt.State.PageNo - 1) * mt.State.PageSize
	end := mathutil.Min(mt.State.PageNo*mt.State.PageSize, mt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = mt.GetRowItem(nodes[i], resName); err != nil {
			return nil, err
		}
		lists = append(lists, *ri)
	}
	return lists, err
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

func (mt *MemInfoTable) GetRowItem(c data.Object, resName v1.ResourceName) (*table.RowItem, error) {
	var (
		err                     error
		status                  *common.SteveStatus
		distribution, dr, usage table.DistributionValue
		//resp                    []apistructs.MetricsData
	)
	if status, err = mt.GetItemStatus(c); err != nil {
		return nil, err
	}
	//req := apistructs.MetricsRequest{
	//	ClusterName:  mt.SDK.InParams["clusterName"].(string),
	//	Names:        []string{c.String("id")},
	//	ResourceType: resName,
	//	OrgID:        mt.SDK.Identity.OrgID,
	//	UserID:       mt.SDK.Identity.UserID,
	//}
	//if resp, err = mt.CtxBdl.GetMetrics(req); err != nil {
	//	return nil, err
	//}
	//distribution = mt.GetDistributionValue(resp[0])
	//usage = mt.GetUsageValue(resp[0])
	//dr = mt.GetDistributionRate(resp[0])
	ri := &table.RowItem{
		ID:      c.String("id"),
		Version: c.String("status", "nodeInfo", "kubeletVersion"),
		Role:    c.StringSlice("metadata","fields")[2],
		Node: table.Node{
			RenderType: "multiple",
			Renders:    mt.GetRenders(c.String("id"), mt.GetIp(c), c.Map("metadata", "labels")),
		},
		Status: *status,
		Distribution: table.Distribution{
			RenderType: "bgProgress",
			Value:      distribution,
			Status:     c.StringSlice("metadata", "fields")[1],
		},
		Usage: table.Distribution{
			RenderType: "bgProgress",
			Value:      usage,
			Status:     c.StringSlice("metadata", "fields")[1],
		},
		UsageRate: table.Distribution{
			RenderType: "bgProgress",
			Value:      dr,
			Status:     c.StringSlice("metadata", "fields")[1],
		},
		BatchOptions: []string{"delete","freeze","unfreeze"},
	}
	return ri, nil
}

// SetComponentValue transfer MemInfoTable struct to Component
func (mt *MemInfoTable) SetComponentValue(c *cptype.Component) error {
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

func (mt *MemInfoTable) getProps() map[string]interface{} {
	return map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "status", Title: mt.SDK.I18n("status"), Sortable: true, Width: 100},
			{DataIndex: "node", Title: mt.SDK.I18n("node"), Sortable: true, Width: 240},
			{DataIndex: "role", Title: mt.SDK.I18n("role"), Width: 120},
			{DataIndex: "version", Title: mt.SDK.I18n("version"), Width: 120},
			{DataIndex: "distribution", Title: "mem" + mt.SDK.I18n("distribution"), Width: 120},
			{DataIndex: "use", Title: "mem" + mt.SDK.I18n("use"), Width: 120},
			{DataIndex: "distributionRate", Title: "mem" + mt.SDK.I18n("distributionRate"), Width: 120},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
		"batchOperations": map[string]table.BatchOperation{
			"delete":   {Key: "delete", Text: mt.SDK.I18n("delete"), Reload: true, ShowIndex: nil},
			"freeze":   {Key: "freeze", Text: mt.SDK.I18n("freeze"), Reload: true, ShowIndex: nil},
			"unfreeze": {Key: "unfreeze", Text: mt.SDK.I18n("unfreeze"), Reload: true, ShowIndex: nil},
		},
		"scroll": table.Scroll{X: 1200},
	}

}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "memTable", func() servicehub.Provider {
		mt := MemInfoTable{}
		mt.Type = "Table"
		return &mt
	})
}
