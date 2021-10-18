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
	"sync"

	"github.com/go-openapi/strfmt"
	jsi "github.com/json-iterator/go"
	"github.com/pkg/errors"
	types2 "github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-pods/podsTable"
	cmpcputil "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "podsTable", func() servicehub.Provider {
		return &ComponentPodsTable{}
	})
}

var (
	steveServer cmp.SteveServer
	mServer     metrics.Interface
)

func (p *ComponentPodsTable) Init(ctx servicehub.Context) error {
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
	return p.DefaultProvider.Init(ctx)
}

func (p *ComponentPodsTable) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	p.InitComponent(ctx)
	if err := p.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen podsTable component state, %v", err)
	}

	switch event.Operation {
	case cptype.InitializeOperation:
		p.State.PageNo = 1
		p.State.PageSize = 20
	case "changePageSize", "changeSort":
		p.State.PageNo = 1
	}

	if err := p.DecodeURLQuery(); err != nil {
		return fmt.Errorf("failed to decode url query for podsTable component, %v", err)
	}
	if err := p.RenderTable(); err != nil {
		return fmt.Errorf("failed to render podsTable component, %v", err)
	}
	if err := p.EncodeURLQuery(); err != nil {
		return fmt.Errorf("failed to encode url query for podsTable component, %v", err)
	}
	p.SetComponentValue(ctx)
	p.Transfer(component)
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

func (p *ComponentPodsTable) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var tableState State
	jsonData, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &tableState); err != nil {
		return err
	}
	p.State = tableState
	return nil
}

func (p *ComponentPodsTable) DecodeURLQuery() error {
	queryData, ok := p.sdk.InParams["workloadTable__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := json.Unmarshal(decode, &data); err != nil {
		return err
	}
	p.State.PageNo = int(data["pageNo"].(float64))
	p.State.PageSize = int(data["pageSize"].(float64))
	sorterData := data["sorterData"].(map[string]interface{})
	p.State.Sorter.Field = sorterData["field"].(string)
	p.State.Sorter.Order = sorterData["order"].(string)
	return nil
}

func (p *ComponentPodsTable) EncodeURLQuery() error {
	query := make(map[string]interface{})
	query["pageNo"] = p.State.PageNo
	query["pageSize"] = p.State.PageSize
	query["sorterData"] = p.State.Sorter
	jsonData, err := json.Marshal(query)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(jsonData)
	p.State.PodsTableURLQuery = encode
	return nil
}

func (p *ComponentPodsTable) RenderTable() error {
	userID := p.sdk.Identity.UserID
	orgID := p.sdk.Identity.OrgID
	workloadID := p.State.WorkloadID
	splits := strings.Split(workloadID, "_")
	if len(splits) != 3 {
		return fmt.Errorf("invalid workload id %s", workloadID)
	}
	kind, namespace, name := splits[0], splits[1], splits[2]

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SResType(kind),
		ClusterName: p.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}

	resp, err := p.server.GetSteveResource(p.ctx, &req)
	if err != nil {
		return err
	}
	obj := resp.Data()

	labelSelectors := obj.Map("spec", "selector", "matchLabels")
	if kind == string(apistructs.K8SCronJob) {
		labelSelectors = obj.Map("spec", "jobTemplate", "spec", "template", "metadata", "labels")
	}

	podReq := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SPod,
		ClusterName: p.State.ClusterName,
	}

	var list []types2.APIObject
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go func() {
		list, err = p.server.ListSteveResource(p.ctx, &podReq)
		waitGroup.Done()
	}()

	go func() {
		var metricsErr error
		cpuReq := &metrics.MetricsRequest{
			UserId:  userID,
			OrgId:   orgID,
			Cluster: p.State.ClusterName,
			Kind:    metrics.Pod,
			Type:    metrics.Cpu,
		}
		memReq := &metrics.MetricsRequest{
			UserId:  userID,
			OrgId:   orgID,
			Cluster: p.State.ClusterName,
			Kind:    metrics.Pod,
			Type:    metrics.Memory,
		}
		_, metricsErr = mServer.PodMetrics(p.ctx, cpuReq)
		if metricsErr != nil {
			logrus.Errorf("failed to get cpu metrics for pods, %v", metricsErr)
		}
		_, metricsErr = mServer.PodMetrics(p.ctx, memReq)
		if metricsErr != nil {
			logrus.Errorf("failed to get mem metrics for pods, %v", metricsErr)
		}
		waitGroup.Done()
	}()
	waitGroup.Wait()

	if err != nil {
		return err
	}

	var items []Item
	for _, item := range list {
		obj := item.Data()
		labels := obj.Map("metadata", "labels")
		if !matchSelector(labelSelectors, labels) {
			continue
		}

		if kind == string(apistructs.K8SCronJob) {
			ok, err := p.isOwnedByTargetCronJob(obj, name)
			if err != nil {
				logrus.Errorf("failed to check whether pod is owned by target cron job %s, %v", name, err)
				continue
			}
			if !ok {
				continue
			}
		}

		name := obj.String("metadata", "name")
		namespace := obj.String("metadata", "namespace")
		fields := obj.StringSlice("metadata", "fields")
		if len(fields) != 9 {
			logrus.Errorf("length of pod %s:%s fields is invalid", namespace, name)
			continue
		}

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

		cpuStatus, cpuValue, cpuTip := "success", "0", "N/A"
		metricsData := getCache(cache.GenerateKey(p.State.ClusterName, name, namespace, metrics.Cpu, metrics.Pod))
		if metricsData != nil && !cpuLimits.IsZero() {
			usedCPUPercent := metricsData.Used
			cpuStatus, cpuValue, cpuTip = p.parseResPercent(usedCPUPercent, cpuLimits, resource.DecimalSI)
		}

		memStatus, memValue, memTip := "success", "0", "N/A"
		metricsData = getCache(cache.GenerateKey(p.State.ClusterName, name, namespace, metrics.Memory, metrics.Pod))
		if metricsData != nil && !memLimits.IsZero() {
			usedMemPercent := metricsData.Used
			memStatus, memValue, memTip = p.parseResPercent(usedMemPercent, memLimits, resource.BinarySI)
		}

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
			Namespace:      namespace,
			IP:             fields[5],
			Age:            fields[4],
			CPURequests:    cpuRequestStr,
			CPURequestsNum: cpuRequests.MilliValue(),
			CPUPercent: Percent{
				RenderType: "progress",
				Value:      cpuValue,
				Tip:        cpuTip,
				Status:     cpuStatus,
			},
			CPULimits:         cpuLimitsStr,
			CPULimitsNum:      cpuLimits.MilliValue(),
			MemoryRequests:    memRequestsStr,
			MemoryRequestsNum: memRequests.Value(),
			MemoryPercent: Percent{
				RenderType: "progress",
				Value:      memValue,
				Tip:        memTip,
				Status:     memStatus,
			},
			MemoryLimits:    memLimitsStr,
			MemoryLimitsNum: memLimits.Value(),
			Ready:           fields[1],
			NodeName:        fields[6],
		})
	}

	if p.State.Sorter.Field != "" {
		cmpWrapper := func(field, order string) func(int, int) bool {
			ascend := order == "ascend"
			switch field {
			case "status":
				return func(i int, j int) bool {
					less := items[i].Status.Value.Label < items[j].Status.Value.Label
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
			case "age":
				return func(i int, j int) bool {
					ageI, _ := strfmt.ParseDuration(items[i].Age)
					ageJ, _ := strfmt.ParseDuration(items[j].Age)
					less := ageI < ageJ
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
			case "nodeName":
				return func(i int, j int) bool {
					less := items[i].NodeName < items[j].NodeName
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

func (p *ComponentPodsTable) isOwnedByTargetCronJob(pod data.Object, cronJobName string) (bool, error) {
	podOwners := pod.Slice("metadata", "ownerReferences")
	if len(podOwners) == 0 {
		return false, nil
	}
	podOwner := podOwners[0]
	if podOwner.String("kind") != "Job" {
		return false, nil
	}

	req := &apistructs.SteveRequest{
		UserID:      p.sdk.Identity.UserID,
		OrgID:       p.sdk.Identity.OrgID,
		Type:        apistructs.K8SJob,
		ClusterName: p.State.ClusterName,
		Name:        podOwner.String("name"),
		Namespace:   pod.String("metadata", "namespace"),
	}

	job, err := p.server.GetSteveResource(p.ctx, req)
	if err != nil {
		return false, err
	}
	obj := job.Data()

	jobOwners := obj.Slice("metadata", "ownerReferences")
	if len(jobOwners) == 0 {
		return false, nil
	}

	jobOwner := jobOwners[0]
	if jobOwner.String("kind") == "CronJob" && jobOwner.String("name") == cronJobName {
		return true, nil
	}
	return false, nil
}

func (p *ComponentPodsTable) parseResPercent(usedPercent float64, totQty *resource.Quantity, format resource.Format) (string, string, string) {
	var totRes int64
	var usedRes float64
	status, tip, value := "", "", ""
	if format == resource.DecimalSI {
		totRes = totQty.MilliValue()
		usedRes = float64(totRes) * usedPercent / 100
		if usedPercent <= 80 {
			status = "success"
		} else if usedPercent < 100 {
			status = "warning"
		} else {
			status = "error"
		}
		tip = fmt.Sprintf("%s/%s", cmpcputil.ResourceToString(p.sdk, usedRes, format),
			cmpcputil.ResourceToString(p.sdk, float64(totQty.MilliValue()), format))
		value = fmt.Sprintf("%.2f", usedPercent)
	} else {
		totRes = totQty.Value()
		usedRes = float64(totRes) * usedPercent / 100
		if usedPercent <= 80 {
			status = "success"
		} else if usedPercent < 100 {
			status = "warning"
		} else {
			status = "error"
		}
		tip = fmt.Sprintf("%s/%s", cmpcputil.ResourceToString(p.sdk, usedRes, format),
			cmpcputil.ResourceToString(p.sdk, float64(totQty.Value()), format))
		value = fmt.Sprintf("%.2f", usedPercent)
	}
	return status, value, tip
}

func (p *ComponentPodsTable) SetComponentValue(ctx context.Context) {
	p.Props.SortDirections = []string{"descend", "ascend"}
	p.Props.IsLoadMore = true
	p.Props.PageSizeOptions = []string{
		"10", "20", "50", "100",
	}
	p.Props.RowKey = "id"
	p.Props.Columns = []Column{
		{
			DataIndex: "status",
			Title:     cputil.I18n(ctx, "status"),
			Width:     80,
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
		{
			DataIndex: "ready",
			Title:     cputil.I18n(ctx, "ready"),
			Width:     80,
			Sorter:    true,
		},
		{
			DataIndex: "nodeName",
			Title:     cputil.I18n(ctx, "node"),
			Width:     120,
			Sorter:    true,
		},
		{
			DataIndex: "age",
			Title:     cputil.I18n(ctx, "age"),
			Width:     120,
			Sorter:    true,
		},
	}

	p.Operations = map[string]interface{}{
		"changeSort": Operation{
			Key:    "changeSort",
			Reload: true,
		},
	}
}

func (p *ComponentPodsTable) Transfer(component *cptype.Component) {
	component.Props = p.Props
	component.State = map[string]interface{}{
		"clusterName":       p.State.ClusterName,
		"workloadId":        p.State.WorkloadID,
		"pageNo":            p.State.PageNo,
		"pageSize":          p.State.PageSize,
		"sorterData":        p.State.Sorter,
		"total":             p.State.Total,
		"podsTableURLQuery": p.State.PodsTableURLQuery,
	}
	component.Data = map[string]interface{}{
		"list": p.Data.List,
	}
	component.Operations = p.Operations
}

func matchSelector(selector, labels map[string]interface{}) bool {
	for k, v := range selector {
		value, ok := v.(string)
		if !ok {
			return false
		}
		labelV, ok := labels[k]
		if !ok {
			return false
		}
		labelValue, ok := labelV.(string)
		if !ok || labelValue != value {
			return false
		}
	}
	return true
}

func (p *ComponentPodsTable) parsePodStatus(state string) Status {
	color := podsTable.PodStatusToColor[state]
	if color == "" {
		color = "darkslategray"
	}
	return Status{
		RenderType: "tagsRow",
		Size:       "default",
		Value: StatusValue{
			Label: p.sdk.I18n(state),
			Color: color,
		},
	}
}

func parseResource(resStr string, format resource.Format) *resource.Quantity {
	if resStr == "" {
		return resource.NewQuantity(0, format)
	}
	result, _ := resource.ParseQuantity(resStr)
	return &result
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

func getCache(key string) *metrics.MetricsData {

	v, _, err := cache.GetFreeCache().Get(key)
	if err != nil {
		logrus.Errorf("get metrics %v err :%v", key, err)
	}
	d := &metrics.MetricsData{}
	if v != nil {
		err = jsi.Unmarshal(v[0].Value().([]byte), d)
		if err != nil {
			logrus.Errorf("get metrics %v unmarshal to json err :%v", key, err)
		}
	}
	return d
}
