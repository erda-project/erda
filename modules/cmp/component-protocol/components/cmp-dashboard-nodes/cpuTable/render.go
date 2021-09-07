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

func (ct *CpuInfoTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	var (
		err   error
		state table.State
	)
	ct.Ctx = ctx
	ct.SDK = cputil.SDK(ctx)
	ct.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	ct.getProps()
	ct.Operations = getTableOperation()
	ct.State = table.State{}
	activeKey:= (*gs)["activeKey"].(string)
	err = common.Transfer(c.State, &state)
	if err != nil {
		return err
	}
	ct.State = state
	if event.Operation != cptype.InitializeOperation {
		// Tab name not equal this component name
		if activeKey != tableTabs.CPU_TAB {
			return ct.SetComponentValue(c)
		}
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
	} else {
		ct.Props["visible"] = true
	}
	nodes := (*gs)["nodes"].([]data.Object)
	if err = ct.RenderList(c, event, v1.ResourceCPU, nodes); err != nil {
		return err
	}
	if err = ct.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (ct *CpuInfoTable) RenderList(component *cptype.Component, event cptype.ComponentEvent, resName v1.ResourceName, nodes []data.Object) error {
	var (
		err        error
		sortColumn string
		asc        bool
		items      []table.RowItem
	)
	if ct.State.PageNo == 0 {
		ct.State.PageNo = DefaultPageNo
	}
	if ct.State.PageSize == 0 {
		ct.State.PageSize = DefaultPageSize
	}
	if ct.State.Sorter.Field != ""{
		sortColumn = ct.State.Sorter.Field
		asc = strings.ToLower(ct.State.Sorter.Order) == "ascend"
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

	if items, err = ct.SetData(nodes, v1.ResourceMemory); err != nil {
		return err
	}
	component.Data = map[string]interface{}{"list":items}
	return nil
}

// SetData assemble rowItem of table
func (ct *CpuInfoTable) SetData(nodes []data.Object, resName v1.ResourceName) ([]table.RowItem, error) {
	var (
		list []table.RowItem
		ri   *table.RowItem
		err  error
	)
	ct.State.Total = len(nodes)
	start := (ct.State.PageNo - 1) * ct.State.PageSize
	end := mathutil.Min(ct.State.PageNo*ct.State.PageSize, ct.State.Total)

	for i := start; i < end; i++ {
		if ri, err = ct.GetRowItem(nodes[i], resName); err != nil {
			return nil, err
		}
		list = append(list, *ri)
	}
	return list, nil
}

func (ct *CpuInfoTable) getProps() {
	props := map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "status", Title: ct.SDK.I18n("status"), Sortable: true, Width: 80, Fixed: "left"},
			{DataIndex: "node", Title: ct.SDK.I18n("node"), Sortable: true, Width: 120},
			{DataIndex: "role", Title: ct.SDK.I18n("role"), Sortable: true, Width: 120},
			{DataIndex: "version", Title: ct.SDK.I18n("version"), Width: 120},
			{DataIndex: "distribution", Title: "cpu" + ct.SDK.I18n("request"), Width: 120},
			{DataIndex: "usage", Title: "cpu" + ct.SDK.I18n("usage"), Width: 120},
			{DataIndex: "usageRate", Title: "cpu" + ct.SDK.I18n("distributionRate"), Width: 120},
			{DataIndex: "operate", Title: "操作", Width: 120, Fixed: "right"},
		},
		"bordered":        true,
		"selectable":      true,
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
		"batchOperations": map[string]table.BatchOperation{
			"delete":   {Key: "delete", Text: ct.SDK.I18n("delete"), Reload: true, ShowIndex: nil},
			"freeze":   {Key: "freeze", Text: ct.SDK.I18n("freeze"), Reload: true, ShowIndex: nil},
			"unfreeze": {Key: "unfreeze", Text: ct.SDK.I18n("unfreeze"), Reload: true, ShowIndex: nil},
		},
		"scroll": table.Scroll{X: 1200},
	}
	ct.Props = props
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

func (ct *CpuInfoTable) GetRowItem(c data.Object, resName v1.ResourceName) (*table.RowItem, error) {
	var (
		err                     error
		status                  *common.SteveStatus
		distribution, dr, usage table.DistributionValue
		//resp                    []apistructs.MetricsData
		nodeLabels []string
	)
	nodeLabelsData := c.Map("metadata", "labels")
	for k,_:= range nodeLabelsData{
		nodeLabels = append(nodeLabels, k)
	}
	if status, err = ct.GetItemStatus(c); err != nil {
		return nil, err
	}
	//req := apistructs.MetricsRequest{
	//	ClusterName:  ct.SDK.InParams["clusterName"].(string),
	//	Names:        []string{c.String("id")},
	//	ResourceType: resName,
	//	ResourceKind: "node",
	//	OrgID:        ct.SDK.Identity.OrgID,
	//	UserID:       ct.SDK.Identity.UserID,
	//}

	//if resp, err = ct.CtxBdl.GetMetrics(req); err != nil {
	//	return nil, err
	//}
	//distribution = ct.GetDistributionValue(resp[0])
	//usage = ct.GetUsageValue(resp[0])
	//dr = ct.GetDistributionRate(resp[0])
	ri := &table.RowItem{
		ID:      c.String("id"),
		Version: c.String("status", "nodeInfo", "kubeletVersion"),
		Role:    c.StringSlice("metadata","fields")[2],
		Node: table.Node{
			RenderType: "multiple",
			Renders:    ct.GetRenders(c.String("id"), ct.GetIp(c), c.Map("metadata", "labels")),
		},
		Status: *status,
		Distribution: table.Distribution{
			RenderType: "bgProgress",
			Value:      distribution,
		},
		Usage: table.Distribution{
			RenderType: "bgProgress",
			Value:      usage,
		},
		UsageRate: table.Distribution{
			RenderType: "bgProgress",
			Value:      dr,
		},
		Operate: ct.GetOperate(c.String("id")),
		BatchOptions: []string{"delete","freeze","unfreeze"},

	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "cpuTable", func() servicehub.Provider {
		ci := CpuInfoTable{}
		return &ci
	})
}
