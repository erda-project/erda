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

package podsTable

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp"
	cmpcputil "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podsTable", func() servicehub.Provider {
		return &ComponentPodsTable{}
	})
}

var steveServer cmp.SteveServer

func (p *ComponentPodsTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return p.DefaultProvider.Init(ctx)
}

func (p *ComponentPodsTable) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	p.InitComponent(ctx)
	if err := p.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen podsTable component state, %v", err)
	}

	p.SetComponentValue(ctx)
	switch event.Operation {
	case cptype.InitializeOperation:
		p.State.PageNo = 1
		p.State.PageSize = 20
	case cptype.RenderingOperation, "changePageSize", "changeSort":
		if event.Component == "tableTabs" {
			return nil
		} else {
			p.State.PageNo = 1
		}
	}

	if err := p.DecodeURLQuery(); err != nil {
		return fmt.Errorf("failed to decode url query for podsTable component, %v", err)
	}
	if err := p.RenderTable(); err != nil {
		return fmt.Errorf("failed to render podsTable component, %v", err)
	}
	(*gs)["countValues"] = p.State.CountValues
	if err := p.EncodeURLQuery(); err != nil {
		return fmt.Errorf("failed to encode url query for podsTable component, %v", err)
	}
	return nil
}

func (p *ComponentPodsTable) InitComponent(ctx context.Context) {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.bdl = bdl
	sdk := cputil.SDK(ctx)
	p.sdk = sdk
	p.ctx = ctx
	p.server = steveServer
}

func (p *ComponentPodsTable) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(component.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	p.State = state
	return nil
}

func (p *ComponentPodsTable) DecodeURLQuery() error {
	queryData, ok := p.sdk.InParams["podsTable__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}
	query := make(map[string]interface{})
	if err := json.Unmarshal(decode, &query); err != nil {
		return err
	}
	p.State.PageNo = int(query["pageNo"].(float64))
	p.State.PageSize = int(query["pageSize"].(float64))
	sorter := query["sorterData"].(map[string]interface{})
	p.State.Sorter.Field = sorter["field"].(string)
	p.State.Sorter.Order = sorter["order"].(string)
	return nil
}

func (p *ComponentPodsTable) EncodeURLQuery() error {
	query := make(map[string]interface{})
	query["pageNo"] = p.State.PageNo
	query["pageSize"] = p.State.PageSize
	query["sorterData"] = p.State.Sorter
	data, err := json.Marshal(query)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(data)
	p.State.PodsTableURLQuery = encode
	return nil
}

func (p *ComponentPodsTable) RenderTable() error {
	userID := p.sdk.Identity.UserID
	orgID := p.sdk.Identity.OrgID

	podReq := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SPod,
		ClusterName: p.State.ClusterName,
	}

	list, err := p.server.ListSteveResource(p.ctx, &podReq)
	if err != nil {
		return err
	}

	cpuReq := apistructs.MetricsRequest{
		UserID:       userID,
		OrgID:        orgID,
		ClusterName:  p.State.ClusterName,
		ResourceKind: metrics.Pod,
		ResourceType: metrics.Cpu,
	}
	memReq := apistructs.MetricsRequest{
		UserID:       userID,
		OrgID:        orgID,
		ClusterName:  p.State.ClusterName,
		ResourceKind: metrics.Pod,
		ResourceType: metrics.Memory,
	}

	p.State.CountValues = make(map[string]int)
	tempCPULimits := make([]*resource.Quantity, 0)
	tempMemLimits := make([]*resource.Quantity, 0)
	var items []Item
	for _, item := range list {
		obj := item.Data()
		name := obj.String("metadata", "name")
		namespace := obj.String("metadata", "namespace")
		fields := obj.StringSlice("metadata", "fields")
		if len(fields) != 9 {
			logrus.Errorf("length of pod %s:%s fields is invalid", namespace, name)
			continue
		}
		if len(p.State.Values.Namespace) != 0 && !contain(p.State.Values.Namespace, namespace) {
			continue
		}
		if len(p.State.Values.Status) != 0 && !contain(p.State.Values.Status, convertPodStatus(fields[2])) {
			continue
		}
		if len(p.State.Values.Node) != 0 && !contain(p.State.Values.Node, fields[6]) {
			continue
		}
		if p.State.Values.Search != "" && !strings.Contains(name, p.State.Values.Search) &&
			!strings.Contains(fields[5], p.State.Values.Search) {
			continue
		}

		cpuReq.PodRequests = append(cpuReq.PodRequests, apistructs.MetricsPodRequest{
			PodName:   name,
			Namespace: namespace,
		})
		memReq.PodRequests = append(memReq.PodRequests, apistructs.MetricsPodRequest{
			PodName:   name,
			Namespace: namespace,
		})
		p.State.CountValues[fields[2]]++
		status := p.parsePodStatus(fields[2])
		containers := obj.Slice("spec", "containers")
		cpuRequests := resource.NewQuantity(0, resource.DecimalSI)
		cpuLimits := resource.NewQuantity(0, resource.DecimalSI)
		memRequests := resource.NewQuantity(0, resource.BinarySI)
		memLimits := resource.NewQuantity(0, resource.BinarySI)
		for _, container := range containers {
			cpuRequests.Add(*parseResource(container.String("resources", "requests", "cpu"), resource.DecimalSI))
			cpuLimits.Add(*parseResource(container.String("resources", "limits", "cpu"), resource.DecimalSI))
			memRequests.Add(*parseResource(container.String("resources", "requests", "memory"), resource.BinarySI))
			memLimits.Add(*parseResource(container.String("resources", "limits", "memory"), resource.BinarySI))
		}

		cpuRequestStr := cmpcputil.ResourceToString(p.sdk, float64(cpuRequests.MilliValue()), resource.DecimalSI)
		if cpuRequests.MilliValue() == 0 {
			cpuRequestStr = "-"
		}
		cpuLimitsStr := cmpcputil.ResourceToString(p.sdk, float64(cpuLimits.MilliValue()), resource.DecimalSI)
		if cpuLimits.MilliValue() == 0 {
			cpuLimitsStr = "-"
		}
		memRequestsStr := cmpcputil.ResourceToString(p.sdk, float64(memRequests.Value()), resource.BinarySI)
		if memRequests.Value() == 0 {
			memRequestsStr = "-"
		}
		memLimitsStr := cmpcputil.ResourceToString(p.sdk, float64(memLimits.Value()), resource.BinarySI)
		if memLimits.Value() == 0 {
			memLimitsStr = "-"
		}

		tempCPULimits = append(tempCPULimits, cpuLimits)
		tempMemLimits = append(tempMemLimits, memLimits)

		id := fmt.Sprintf("%s_%s", namespace, name)
		items = append(items, Item{
			ID:     id,
			Status: status,
			Name: Link{
				RenderType: "linkText",
				Value:      name,
				Operations: map[string]interface{}{
					"click": LinkOperation{
						Command: Command{
							Key:    "goto",
							Target: "cmpClustersPodDetail",
							State: CommandState{
								Params: map[string]string{
									"podId": id,
								},
								Query: map[string]string{
									"namespace": namespace,
									"podName":   name,
								},
							},
							JumpOut: true,
						},
						Reload: false,
					},
				},
			},
			Namespace:         namespace,
			IP:                fields[5],
			CPURequests:       cpuRequestStr,
			CPURequestsNum:    cpuRequests.MilliValue(),
			CPULimits:         cpuLimitsStr,
			CPULimitsNum:      cpuLimits.MilliValue(),
			MemoryRequests:    memRequestsStr,
			MemoryRequestsNum: memRequests.Value(),
			MemoryLimits:      memLimitsStr,
			MemoryLimitsNum:   memLimits.Value(),
			Ready:             fields[1],
			Node:              fields[6],
		})
	}

	cpuMetrics, err := p.bdl.GetMetrics(cpuReq)
	if err != nil || len(cpuMetrics) == 0 {
		logrus.Errorf("failed to get cpu metrics for pods, %v", err)
		cpuMetrics = make([]apistructs.MetricsData, len(items), len(items))
	}
	memMetrics, err := p.bdl.GetMetrics(memReq)
	if err != nil || len(memMetrics) == 0 {
		logrus.Errorf("failed to get memory metrics for pods, %v", err)
		memMetrics = make([]apistructs.MetricsData, len(items), len(items))
	}

	for i := range items {
		cpuLimits := tempCPULimits[i]
		memLimits := tempMemLimits[i]

		cpuStatus, cpuValue, cpuTip := "success", "0", "N/A"
		usedCPUPercent := cpuMetrics[i].Used
		cpuStatus, cpuValue, cpuTip = p.parseResPercent(usedCPUPercent, cpuLimits, resource.DecimalSI)
		items[i].CPUPercent = Percent{
			RenderType: "progress",
			Value:      cpuValue,
			Tip:        cpuTip,
			Status:     cpuStatus,
		}

		memStatus, memValue, memTip := "success", "0", "N/A"
		usedMemPercent := memMetrics[i].Used
		memStatus, memValue, memTip = p.parseResPercent(usedMemPercent, memLimits, resource.BinarySI)
		items[i].MemoryPercent = Percent{
			RenderType: "progress",
			Value:      memValue,
			Tip:        memTip,
			Status:     memStatus,
		}
	}

	if p.State.Sorter.Field != "" {
		cmpWrapper := func(field, order string) func(int, int) bool {
			ascend := order == "ascend"
			switch field {
			case "status":
				return func(i int, j int) bool {
					less := items[i].Status.Value < items[j].Status.Value
					if ascend {
						return less
					}
					return !less
				}
			case "name":
				return func(i int, j int) bool {
					less := items[i].Name.Value < items[j].Name.Value
					if ascend {
						return less
					}
					return !less
				}
			case "namespace":
				return func(i int, j int) bool {
					less := items[i].Namespace < items[j].Namespace
					if ascend {
						return less
					}
					return !less
				}
			case "ip":
				return func(i int, j int) bool {
					less := items[i].IP < items[j].IP
					if ascend {
						return less
					}
					return !less
				}
			case "cpuRequests":
				return func(i int, j int) bool {
					less := items[i].CPURequestsNum < items[j].CPURequestsNum
					if ascend {
						return less
					}
					return !less
				}
			case "cpuPercent":
				return func(i int, j int) bool {
					vI, _ := strconv.ParseFloat(items[i].CPUPercent.Value, 64)
					vJ, _ := strconv.ParseFloat(items[j].CPUPercent.Value, 64)
					less := vI < vJ
					if ascend {
						return less
					}
					return !less
				}
			case "cpuLimits":
				return func(i int, j int) bool {
					less := items[i].CPULimitsNum < items[j].CPULimitsNum
					if ascend {
						return less
					}
					return !less
				}
			case "memoryRequests":
				return func(i int, j int) bool {
					less := items[i].MemoryRequestsNum < items[j].MemoryRequestsNum
					if ascend {
						return less
					}
					return !less
				}
			case "memoryPercent":
				return func(i int, j int) bool {
					vI, _ := strconv.ParseFloat(items[i].MemoryPercent.Value, 64)
					vJ, _ := strconv.ParseFloat(items[j].MemoryPercent.Value, 64)
					less := vI < vJ
					if ascend {
						return less
					}
					return !less
				}
			case "memoryLimits":
				return func(i int, j int) bool {
					less := items[i].MemoryLimitsNum < items[j].MemoryLimitsNum
					if ascend {
						return less
					}
					return !less
				}
			case "ready":
				return func(i int, j int) bool {
					splits := strings.Split(items[i].Ready, "/")
					readyI := splits[0]
					splits = strings.Split(items[j].Ready, "/")
					readyJ := splits[0]
					less := readyI < readyJ
					if ascend {
						return less
					}
					return !less
				}
			case "node":
				return func(i int, j int) bool {
					less := items[i].Node < items[j].Node
					if ascend {
						return less
					}
					return !less
				}
			default:
				return func(i int, j int) bool {
					return false
				}
			}
		}
		sort.Slice(items, cmpWrapper(p.State.Sorter.Field, p.State.Sorter.Order))
	}

	p.Data.List = items
	p.State.Total = len(items)
	return nil
}

func (p *ComponentPodsTable) parseResPercent(usedPercent float64, totQty *resource.Quantity, format resource.Format) (string, string, string) {
	var totRes int64
	if format == resource.DecimalSI {
		totRes = totQty.MilliValue()
	} else {
		totRes = totQty.Value()
	}
	usedRes := float64(totRes) * usedPercent / 100
	usedQtyString := cmpcputil.ResourceToString(p.sdk, usedRes, format)

	status := ""
	if usedPercent <= 80 {
		status = "success"
	} else if usedPercent < 100 {
		status = "warning"
	} else {
		status = "error"
	}

	tip := ""
	if format == resource.DecimalSI {
		tip = fmt.Sprintf("%s/%s", usedQtyString, cmpcputil.ResourceToString(p.sdk, float64(totQty.MilliValue()), format))
	} else {
		tip = fmt.Sprintf("%s/%s", usedQtyString, cmpcputil.ResourceToString(p.sdk, float64(totQty.Value()), format))
	}
	value := fmt.Sprintf("%.2f", usedPercent)
	if usedRes < 1e-8 {
		tip = "N/A"
		value = "N/A"
	}
	return status, value, tip
}

func (p *ComponentPodsTable) SetComponentValue(ctx context.Context) {
	p.Props.RowKey = "id"
	p.Props.PageSizeOptions = []string{
		"10", "20", "50", "100",
	}
	p.Props.Columns = []Column{
		{
			DataIndex: "status",
			Title:     cputil.I18n(ctx, "status"),
			Width:     120,
			Sorter:    true,
		},
		{
			DataIndex: "name",
			Title:     cputil.I18n(ctx, "name"),
			Width:     180,
			Sorter:    true,
		},
		{
			DataIndex: "namespace",
			Title:     cputil.I18n(ctx, "namespace"),
			Width:     180,
			Sorter:    true,
		},
		{
			DataIndex: "ip",
			Title:     cputil.I18n(ctx, "ip"),
			Width:     120,
			Sorter:    true,
		},
		{
			DataIndex: "ready",
			Title:     cputil.I18n(ctx, "ready"),
			Width:     80,
			Sorter:    true,
		},
		{
			DataIndex: "node",
			Title:     cputil.I18n(ctx, "node"),
			Width:     120,
			Sorter:    true,
		},
	}

	if p.State.ActiveKey == "cpu" {
		p.Props.Columns = append(p.Props.Columns, []Column{
			{
				DataIndex: "cpuRequests",
				Title:     cputil.I18n(ctx, "cpuRequests"),
				Width:     120,
				Sorter:    true,
			},
			{
				DataIndex: "cpuLimits",
				Title:     cputil.I18n(ctx, "cpuLimits"),
				Width:     120,
				Sorter:    true,
			},
			{
				DataIndex: "cpuPercent",
				Title:     cputil.I18n(ctx, "cpuPercent"),
				Width:     120,
				Sorter:    true,
			},
		}...)
	} else {
		p.Props.Columns = append(p.Props.Columns, []Column{
			{
				DataIndex: "memoryRequests",
				Title:     cputil.I18n(ctx, "memoryRequests"),
				Width:     120,
				Sorter:    true,
			},
			{
				DataIndex: "memoryLimits",
				Title:     cputil.I18n(ctx, "memoryLimits"),
				Width:     120,
				Sorter:    true,
			},
			{
				DataIndex: "memoryPercent",
				Title:     cputil.I18n(ctx, "memoryPercent"),
				Width:     120,
				Sorter:    true,
			},
		}...)
	}
	p.Operations = map[string]interface{}{
		"changeSort": Operation{
			Key:    "changeSort",
			Reload: true,
		},
	}
}

var PodStatusToColor = map[string]string{
	"Completed":         "steelblue",
	"ContainerCreating": "orange",
	"CrashLoopBackOff":  "red",
	"Error":             "maroon",
	"Evicted":           "darkgoldenrod",
	"ImagePullBackOff":  "darksalmon",
	"Pending":           "teal",
	"Running":           "green",
	"Terminating":       "brown",
}

func (p *ComponentPodsTable) parsePodStatus(state string) Status {
	color := PodStatusToColor[state]
	if color == "" {
		color = "darkslategray"
	}
	return Status{
		RenderType: "text",
		Value:      p.sdk.I18n(state),
		StyleConfig: StyleConfig{
			Color: color,
		},
	}
}

func contain(arr []string, target string) bool {
	for _, str := range arr {
		if target == str {
			return true
		}
	}
	return false
}

func convertPodStatus(status string) string {
	if _, ok := PodStatusToColor[status]; ok {
		return status
	}
	return "others"
}

func parseResource(str string, format resource.Format) *resource.Quantity {
	if str == "" {
		return resource.NewQuantity(0, format)
	}
	res, _ := resource.ParseQuantity(str)
	return &res
}

func getRange(length, pageNo, pageSize int) (int, int) {
	l := (pageNo - 1) * pageSize
	if l >= length || l < 0 {
		l = 0
	}
	r := l + pageSize
	if r > length || r < 0 {
		r = length
	}
	return l, r
}
