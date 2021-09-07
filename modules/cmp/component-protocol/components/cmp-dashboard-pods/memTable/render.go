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
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
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

func (mt *MemTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	var (
		err   error
		state table.State
	)
	isFirstFilter := (*gs)["isFirstFilter"].(bool)
	mt.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	mt.Ctx = ctx
	mt.SDK = cputil.SDK(ctx)
	mt.getProps()
	mt.Operations = getTableOperation()
	mt.State = table.State{}
	err = common.Transfer(c.State, &state)
	if err != nil {
		return err
	}
	mt.State = state
	if event.Operation != cptype.InitializeOperation {
		// Tab name not equal this component name
		if c.State["activeKey"].(string) != tableTabs.MEM_TAB {
			return nil
		}
		switch event.Operation {
		case common.CMPDashboardChangePageSizeOperationKey:
		case common.CMPDashboardChangePageNoOperationKey:
		case cptype.RenderingOperation:
			// IsFirstFilter delivered from filer component
			if isFirstFilter {
				mt.State.PageNo = 1
				(*gs)["isFirstFilter"] = false
			}
		case common.CMPDashboardSortByColumnOperationKey:
			mt.State.PageNo = 1
			(*gs)["isFirstFilter"] = false
		default:
			logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
		}
	} else {
		mt.Props["visible"] = true
		return nil
	}
	if err = mt.RenderList(c, event, v1.ResourceMemory, gs); err != nil {
		return err
	}
	if err = mt.SetComponentValue(c); err != nil {
		return err
	}
	return nil
}

func (mt *MemTable) getProps() {
	mt.Props = map[string]interface{}{
		"rowKey": "id",
		"columns": []table.Columns{
			{DataIndex: "status", Title: mt.SDK.I18n("status"), Width: 120},
			{DataIndex: "name", Title: mt.SDK.I18n("name"), Width: 180},
			{DataIndex: "namespace", Title: mt.SDK.I18n("namespace")},
			{DataIndex: "ip", Title: mt.SDK.I18n("ip"), Width: 120},
			{DataIndex: "request", Title: mt.SDK.I18n("memUsed"), Width: 120},
			{DataIndex: "limit", Title: mt.SDK.I18n("memUsage"), Width: 120},
			{DataIndex: "usedPercent", Title: "mem"+mt.SDK.I18n("usedPercent"), Width: 120},
			{DataIndex: "ready", Title: mt.SDK.I18n("namespace"), Width: 80},
		},
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"operations": map[string]table.Operation{
			"changePageNo": {Key: "changePageNo", Reload: true},
			"changeSort":   {Key: "changeSort", Reload: true},
		},
	}
}

func (mt *MemTable) RenderList(component *cptype.Component, event cptype.ComponentEvent, resName v1.ResourceName, gs *cptype.GlobalStateData) error {
	var (
		items      []table.RowItem
		podList    []data.Object
		pods       []data.Object
		err        error
		sortColumn string
		asc        bool
	)
	sortColumn = event.OperationData["field"].(string)
	if event.OperationData["order"].(string) == "ascend" {
		asc = true
	}
	if mt.State.PageNo == 0 {
		mt.State.PageNo = DefaultPageNo
	}
	if mt.State.PageSize == 0 {
		mt.State.PageSize = DefaultPageSize
	}
	pageNo := mt.State.PageNo
	pageSize := mt.State.PageSize
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
	if sortColumn != "" {
		refCol := reflect.ValueOf(table.RowItem{}).FieldByName(sortColumn)
		switch refCol.Type() {
		case reflect.TypeOf(""):
			common.SortByString([]interface{}{pods}, sortColumn, asc)
		case reflect.TypeOf(table.Distribution{}):
			common.SortByDistribution([]interface{}{pods}, sortColumn, asc)

		default:
			logrus.Errorf("sort by column %s not found", sortColumn)
			return common.TypeNotAvailableErr
		}
	}
	pods = pods[(pageNo-1)*pageSize : mathutil.Max((pageNo-1)*pageSize, len(pods))]
	if items, err = mt.SetData(pods, v1.ResourceMemory); err != nil {
		return err
	}
	component.Data["list"] = items
	return nil
}

// SetData assemble rowItem of table
func (mt *MemTable) SetData(pods []data.Object, resName v1.ResourceName) ([]table.RowItem, error) {
	var (
		lists []table.RowItem
		ri    *table.RowItem
		err   error
	)
	mt.State.Total = len(pods)
	start := (mt.State.PageNo - 1) * mt.State.PageSize
	end := mathutil.Max(mt.State.PageNo*mt.State.PageSize, mt.State.Total)

	for i := start; i < end; i++ {
		if ri, err = mt.GetRowItem(pods[i], resName); err != nil {
			return nil, err
		}
		lists = append(lists, *ri)
	}
	return lists, nil
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

func (mt *MemTable) GetRowItem(c data.Object, resName v1.ResourceName) (*table.RowItem, error) {
	var (
		err    error
		status *table.Status
		dr     *table.DistributionValue
		resp   []apistructs.MetricsData
	)

	req := apistructs.MetricsRequest{
		ClusterName:  mt.SDK.InParams["clusterName"].(string),
		Names:        []string{c.String("id")},
		ResourceType: resName,
		ResourceKind: "pod",
		OrgID:        mt.SDK.Identity.OrgID,
		UserID:       mt.SDK.Identity.UserID,
	}

	if resp, err = mt.CtxBdl.GetMetrics(req); err != nil {
		return nil, err
	}
	dr = mt.GetDistributionRate(resp[0])
	status = mt.GetItemStatus(dr.Percent)
	ri := &table.RowItem{
		ID:     c.String("id"),
		Status: *status,
		IP:     c.StringSlice("metadata", "fields")[5],
		Limit:  mt.GetResourceReq(c, "limits", v1.ResourceMemory),
		Request:  mt.GetResourceReq(c, "requests", v1.ResourceMemory),
		UsedPercent: table.Distribution{
			RenderType: "bgProgress",
			Value:      *dr,
		},
	}
	return ri, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "memTable", func() servicehub.Provider {
		ci := MemTable{}
		ci.Type = "Table"
		return &ci
	})
}
