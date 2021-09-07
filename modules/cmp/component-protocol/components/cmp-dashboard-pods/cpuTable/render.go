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
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	common "github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/common"
	table "github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/common/table"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/tableTabs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
	"reflect"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
)

const (
	DefaultPageSize = 10
	DefaultPageNo   = 1
)

var tableProperties = map[string]interface{}{

}

func (ct *CpuTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	var (
		err   error
		state table.State
	)
	isFirstFilter := (*gs)["isFirstFilter"].(bool)
	ct.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	ct.Ctx = ctx
	err = common.Transfer(c.State, &state)
	if err != nil {
		return err
	}
	ct.State = state
	if event.Operation != cptype.InitializeOperation {
		// Tab name not equal this component name
		if c.State["activeKey"].(string) != tableTabs.CPU_TAB {
			return nil
		}
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey:

		case common.CMPDashboardChangePageNoOperationKey:

		case cptype.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if isFirstFilter {
				ct.State.PageNo = 1
				(*gs)["isFirstFilter"] = false
			}
		case common.CMPDashboardSortByColumnOperationKey:
			ct.State.PageNo = 1
			(*gs)["isFirstFilter"] = false
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	} else {
		ct.Props["visible"] = true
		return nil
	}
	if err = ct.RenderList(c, event, v1.ResourceCPU, gs); err != nil {
		return err
	}
	if err = ct.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (ct *CpuTable) RenderList(component *cptype.Component, event cptype.ComponentEvent, resName v1.ResourceName, gs *cptype.GlobalStateData) error {
	var (
		podList    []data.Object
		pods       []data.Object
		rowItems   []table.RowItem
		err        error
		sortColumn string
	)
	if ct.State.PageNo == 0 {
		ct.State.PageNo = DefaultPageNo
	}
	if ct.State.PageSize == 0 {
		ct.State.PageSize = DefaultPageSize
	}
	pageNo := ct.State.PageNo
	pageSize := ct.State.PageSize
	podList = (*gs)["nodes"].([]data.Object)

	namespace := (*gs)["namespace"].(string)
	status := (*gs)["status"].(string)
	nodeName := (*gs)["node"].(string)
	q := (*gs)["Q"].(string)
	if q == "" && namespace == "" && status == "" && nodeName == "" {
		pods = podList
	} else {
		// Filter by pod name or pod uid , namespace status and node
		for _, pod := range podList {
			exist := false
			if q != "" && (strings.Contains(pod.String("metadata", "name"), q) || strings.Contains(pod.String("id"), q)) {
				exist = true
			}
			if namespace != "" && pod.String("metadata", "namespace") == namespace {
				exist = true
			}
			if status != "" && strings.ToLower(pod.StringSlice("metadata", "status", "fields")[2]) == strings.ToLower(status) {
				exist = true
			}
			if nodeName != "" && strings.ToLower(pod.StringSlice("spec", "status", "fields")[2]) == strings.ToLower(status) {
				exist = true
			}
			if exist {
				pods = append(pods, pod)
			}
		}
	}

	if ct.State.SorterData.Field != "" {
		sorter := string(common.Asc) == strings.ToLower(ct.State.SorterData.Order)
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{ct.Data}, sortColumn, sorter)
		case reflect.TypeOf(table.Name{}):
			common.SortByName([]interface{}{ct.Data}, sortColumn, sorter)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{ct.Data}, sortColumn, sorter)
		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	} // transfer and set data into table

	pods = pods[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(pods))]
	if rowItems, err = ct.SetData(pods, resName); err != nil {
		return err
	}
	component.Data["list"] = rowItems
	return nil
}

func (ct *CpuTable) getProps() {
	ct.Props = map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "status", Title: ct.SDK.I18n("status"), Width: 120},
			{DataIndex: "name", Title: ct.SDK.I18n("name"), Width: 180},
			{DataIndex: "namespace", Title: ct.SDK.I18n("namespace")},
			{DataIndex: "ip", Title: ct.SDK.I18n("ip"), Width: 120},
			{DataIndex: "request", Title:"cpu"+ ct.SDK.I18n("request"), Width: 120},
			{DataIndex: "limit", Title: "cpu"+ ct.SDK.I18n("limit"), Width: 120},
			{DataIndex: "usedPercent", Title: "cpu"+ct.SDK.I18n("usedPercent"), Width: 120},
			{DataIndex: "ready", Title: "ready", Width: 80},
		},
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
	}
}
// SetData assemble rowItem of table
func (ct *CpuTable) SetData(pods []data.Object, resName v1.ResourceName) ([]table.RowItem, error) {
	var (
		lists []table.RowItem
		ri    *table.RowItem
		err   error
	)
	ct.State.Total = len(pods)
	start := (ct.State.PageNo - 1) * ct.State.PageSize
	end := mathutil.Max(ct.State.PageNo*ct.State.PageSize, ct.State.Total)

	for i := start; i < end; i++ {
		if ri, err = ct.GetRowItem(pods[i], resName); err != nil {
			return nil, err
		}
		lists = append(lists, *ri)
	}
	return lists, nil
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

func (ct *CpuTable) GetRowItem(c data.Object, resName v1.ResourceName) (*table.RowItem, error) {
	var (
		err    error
		status *table.Status
		dr     *table.DistributionValue
		resp   []apistructs.MetricsData
	)

	req := apistructs.MetricsRequest{
		ClusterName:  ct.SDK.InParams["clusterName"].(string),
		Names:        []string{c.String("id")},
		ResourceType: resName,
		ResourceKind: "pod",
		OrgID:        ct.SDK.Identity.OrgID,
		UserID:       ct.SDK.Identity.UserID,
	}

	if resp, err = ct.CtxBdl.GetMetrics(req); err != nil {
		return nil, err
	}
	dr = ct.GetDistributionRate(resp[0])
	status = ct.GetItemStatus(dr.Percent)
	ri := &table.RowItem{
		ID:     c.String("id"),
		Status: *status,
		IP:     c.StringSlice("metadata", "fields")[5],
		Limit:  ct.GetResourceReq(c, "limits", v1.ResourceCPU),
		Request:  ct.GetResourceReq(c, "requests", v1.ResourceCPU),
		UsedPercent: table.Distribution{
			RenderType: "bgProgress",
			Value:      *dr,
		},
	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "cpuTable", func() servicehub.Provider {
		ci := CpuTable{}
		ci.Type = "Table"
		ci.Props = getProps()
		ci.Operations = getTableOperation()
		ci.State = table.State{}
		return &ci
	})
}
