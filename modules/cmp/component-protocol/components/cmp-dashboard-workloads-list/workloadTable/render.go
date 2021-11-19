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

package workloadTable

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/recallsong/go-utils/container/slice"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-workloads-list/filter"
	cmpcputil "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workloads-list", "workloadTable", func() servicehub.Provider {
		return &ComponentWorkloadTable{}
	})
}

var steveServer cmp.SteveServer

func (w *ComponentWorkloadTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return w.DefaultProvider.Init(ctx)
}

func (w *ComponentWorkloadTable) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	w.InitComponent(ctx)
	if err := w.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadTable component state, %v", err)
	}

	switch event.Operation {
	case cptype.InitializeOperation:
		w.State.PageNo = 1
		w.State.PageSize = 20
		if err := w.DecodeURLQuery(); err != nil {
			return fmt.Errorf("failed to get url query for workloadTable component, %v", err)
		}
	case cptype.RenderingOperation, "changePageSize", "changeSort":
		w.State.PageNo = 1
	}

	if err := w.RenderTable(); err != nil {
		return fmt.Errorf("failed to render workloadTable component, %v", err)
	}
	if err := w.EncodeURLQuery(); err != nil {
		return fmt.Errorf("failed to gen url query for workloadTable component, %v", err)
	}
	w.SetComponentValue(ctx)
	w.Transfer(component)
	return nil
}

func (w *ComponentWorkloadTable) DecodeURLQuery() error {
	urlQuery, ok := w.sdk.InParams["workloadTable__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(urlQuery)
	if err != nil {
		return err
	}
	queryData := make(map[string]interface{})
	if err := json.Unmarshal(decode, &queryData); err != nil {
		return err
	}
	w.State.PageNo = uint64(queryData["pageNo"].(float64))
	w.State.PageSize = uint64(queryData["pageSize"].(float64))
	sorterData := queryData["sorterData"].(map[string]interface{})
	w.State.Sorter.Field, _ = sorterData["field"].(string)
	w.State.Sorter.Order, _ = sorterData["order"].(string)
	return nil
}

func (w *ComponentWorkloadTable) EncodeURLQuery() error {
	query := make(map[string]interface{})
	query["pageNo"] = w.State.PageNo
	query["pageSize"] = w.State.PageSize
	query["sorterData"] = w.State.Sorter
	jsonData, err := json.Marshal(query)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(jsonData)
	w.State.WorkloadTableURLQuery = encode
	return nil
}

func (w *ComponentWorkloadTable) InitComponent(ctx context.Context) {
	sdk := cputil.SDK(ctx)
	w.sdk = sdk
	w.ctx = ctx
	w.server = steveServer
}

func (w *ComponentWorkloadTable) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	w.State = state
	return nil
}

func (w *ComponentWorkloadTable) RenderTable() error {
	userID := w.sdk.Identity.UserID
	orgID := w.sdk.Identity.OrgID

	steveRequest := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		ClusterName: w.State.ClusterName,
	}

	var items []Item
	kinds := getWorkloadKindMap(w.State.Values.Kind)

	activeCount := map[apistructs.K8SResType]int{}
	abnormalCount := map[apistructs.K8SResType]int{}
	succeededCount := map[apistructs.K8SResType]int{}
	failedCount := map[apistructs.K8SResType]int{}
	updateCount := map[apistructs.K8SResType]int{}

	for _, kind := range []apistructs.K8SResType{apistructs.K8SDeployment, apistructs.K8SStatefulSet,
		apistructs.K8SDaemonSet, apistructs.K8SJob, apistructs.K8SCronJob} {
		if _, ok := kinds[string(kind)]; !ok && len(kinds) != 0 {
			continue
		}
		steveRequest.Type = kind
		var (
			list []types.APIObject
			err  error
		)
		list, err = cmpcputil.ListSteveResourceByNamespaces(w.ctx, w.server, &steveRequest, w.State.Values.Namespace)
		if err != nil {
			return err
		}

		for _, obj := range list {
			workload := obj.Data()
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, workload.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(workload.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			statusValue, statusColor, err := cmpcputil.ParseWorkloadStatus(workload)
			if err != nil {
				logrus.Error(err)
				continue
			}
			if w.State.Values.Status != nil && !contain(w.State.Values.Status, statusValue) {
				continue
			}
			status := Status{
				RenderType: "text",
				Value:      w.sdk.I18n(statusValue),
				StyleConfig: StyleConfig{
					Color: statusColor,
				},
			}

			name := workload.String("metadata", "name")
			namespace := workload.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", kind, namespace, name)
			link := Link{
				RenderType: "linkText",
				Value:      name,
				Operations: map[string]interface{}{
					"click": LinkOperation{
						Reload: false,
						Key:    "openWorkloadDetail",
					},
				},
			}

			fields := workload.StringSlice("metadata", "fields")
			item := Item{
				ID:        id,
				Status:    status,
				Name:      link,
				Namespace: namespace,
				Kind:      workload.String("kind"),
			}

			switch kind {
			case apistructs.K8SDeployment:
				if len(fields) != 8 {
					logrus.Errorf("deployment %s:%s has invalid fields length %d", namespace, name, len(fields))
					continue
				}
				if statusValue == "Active" {
					activeCount[kind]++
				} else if statusValue == "Updating" {
					updateCount[kind]++
				} else {
					abnormalCount[kind]++
				}
				item.Age = fields[4]
				item.Ready = fields[1]
				item.UpToDate = fields[2]
				item.Available = fields[3]
			case apistructs.K8SDaemonSet:
				if len(fields) != 11 {
					logrus.Errorf("daemonset %s:%s has invalid fields length %d", namespace, name, len(fields))
					continue
				}
				if statusValue == "Active" {
					activeCount[kind]++
				} else {
					abnormalCount[kind]++
				}
				item.Age = fields[7]
				item.Ready = fields[3]
				item.UpToDate = fields[4]
				item.Available = fields[5]
				item.Desired = fields[1]
				item.Current = fields[2]
			case apistructs.K8SStatefulSet:
				if len(fields) != 5 {
					logrus.Errorf("statefulSet %s:%s has invalid fields length %d", namespace, name, len(fields))
					continue
				}
				if statusValue == "Active" {
					activeCount[kind]++
				} else {
					abnormalCount[kind]++
				}
				item.Age = fields[2]
				item.Ready = fields[1]
			case apistructs.K8SJob:
				if len(fields) != 7 {
					logrus.Errorf("job %s:%s has invalid fields length %d", namespace, name, len(fields))
					continue
				}
				switch statusValue {
				case "Active":
					activeCount[kind]++
				case "Succeeded":
					succeededCount[kind]++
				case "Failed":
					failedCount[kind]++
				}
				item.Age = fields[3]
				item.Completions = fields[1]
				item.Duration = fields[2]
			case apistructs.K8SCronJob:
				if len(fields) != 9 {
					logrus.Errorf("cronJob %s:%s has invalid fields length %d", namespace, name, len(fields))
					continue
				}
				activeCount[kind]++
				item.Age = fields[5]
				item.Schedule = fields[1]
				item.LastSchedule = fields[4]
			}
			items = append(items, item)
		}
	}

	w.State.CountValues = CountValues{
		DeploymentsCount: Count{
			Active:   activeCount[apistructs.K8SDeployment],
			Abnormal: abnormalCount[apistructs.K8SDeployment],
			Updating: updateCount[apistructs.K8SDeployment],
		},
		DaemonSetCount: Count{
			Active:   activeCount[apistructs.K8SDaemonSet],
			Abnormal: abnormalCount[apistructs.K8SDaemonSet],
		},
		StatefulSetCount: Count{
			Active:   activeCount[apistructs.K8SStatefulSet],
			Abnormal: abnormalCount[apistructs.K8SStatefulSet],
		},
		JobCount: Count{
			Active:    activeCount[apistructs.K8SJob],
			Succeeded: succeededCount[apistructs.K8SJob],
			Failed:    failedCount[apistructs.K8SJob],
		},
		CronJobCount: Count{
			Active: activeCount[apistructs.K8SCronJob],
		},
	}

	if w.State.Sorter.Field != "" {
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
			case "kind":
				return func(i int, j int) bool {
					less := items[i].Kind < items[j].Kind
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
			case "ready":
				return func(i int, j int) bool {
					less := false
					if strings.Contains(items[i].Ready, "/") {
						readyI := strings.Split(items[i].Ready, "/")[0]
						readyJ := strings.Split(items[j].Ready, "/")[0]
						less = readyI < readyJ
					} else {
						readyI, _ := strconv.ParseInt(items[i].Ready, 10, 64)
						readyJ, _ := strconv.ParseInt(items[j].Ready, 10, 64)
						less = readyI < readyJ
					}
					if ascend {
						return less
					}
					return !less
				}
			case "upToDate":
				return func(i int, j int) bool {
					nI, _ := strconv.ParseInt(items[i].UpToDate, 10, 64)
					nJ, _ := strconv.ParseInt(items[j].UpToDate, 10, 64)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "available":
				return func(i int, j int) bool {
					nI, _ := strconv.ParseInt(items[i].Available, 10, 64)
					nJ, _ := strconv.ParseInt(items[j].Available, 10, 64)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "desired":
				return func(i int, j int) bool {
					nI, _ := strconv.ParseInt(items[i].Desired, 10, 64)
					nJ, _ := strconv.ParseInt(items[j].Desired, 10, 64)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "current":
				return func(i int, j int) bool {
					nI, _ := strconv.ParseInt(items[i].Current, 10, 64)
					nJ, _ := strconv.ParseInt(items[j].Current, 10, 64)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "completions":
				return func(i int, j int) bool {
					nI, _ := strconv.ParseInt(items[i].Completions, 10, 64)
					nJ, _ := strconv.ParseInt(items[j].Completions, 10, 64)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "duration":
				return func(i int, j int) bool {
					nI, _ := strfmt.ParseDuration(items[i].Duration)
					nJ, _ := strfmt.ParseDuration(items[j].Duration)
					less := nI < nJ
					if ascend {
						return less
					}
					return !less
				}
			case "schedule":
				return func(i int, j int) bool {
					less := items[i].Schedule < items[j].Schedule
					if ascend {
						return less
					}
					return !less
				}
			case "lastSchedule":
				return func(i int, j int) bool {
					nI, _ := strfmt.ParseDuration(items[i].LastSchedule)
					nJ, _ := strfmt.ParseDuration(items[j].LastSchedule)
					less := nI < nJ
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
		slice.Sort(items, cmpWrapper(w.State.Sorter.Field, w.State.Sorter.Order))
	}

	w.Data.List = items
	w.State.Total = uint64(len(items))
	return nil
}

func (w *ComponentWorkloadTable) SetComponentValue(ctx context.Context) {
	w.Props.SortDirections = []string{"descend", "ascend"}
	w.Props.RowKey = "workloadId"
	w.Props.PageSizeOptions = []string{"10", "20", "50", "100"}
	w.Props.RequestIgnore = []string{"data"}

	statusColumn := Column{
		DataIndex: "status",
		Title:     cputil.I18n(ctx, "status"),
		Width:     120,
		Sorter:    true,
	}
	kindColumn := Column{
		DataIndex: "kind",
		Title:     cputil.I18n(ctx, "workloadKind"),
		Width:     120,
		Sorter:    true,
	}
	nameColumn := Column{
		DataIndex: "name",
		Title:     cputil.I18n(ctx, "name"),
		Width:     180,
		Sorter:    true,
	}
	namespaceColumn := Column{
		DataIndex: "namespace",
		Title:     cputil.I18n(ctx, "namespace"),
		Width:     180,
		Sorter:    true,
	}
	ageColumn := Column{
		DataIndex: "age",
		Title:     cputil.I18n(ctx, "age"),
		Width:     100,
		Sorter:    true,
	}
	readyColumn := Column{
		DataIndex: "ready",
		Title:     cputil.I18n(ctx, "ready"),
		Width:     100,
		Sorter:    true,
	}
	upToDateColumn := Column{
		DataIndex: "upToDate",
		Title:     cputil.I18n(ctx, "upToDate"),
		Width:     100,
		Sorter:    true,
	}
	availableColumn := Column{
		DataIndex: "available",
		Title:     cputil.I18n(ctx, "available"),
		Width:     100,
		Sorter:    true,
	}
	desiredColumn := Column{
		DataIndex: "desired",
		Title:     cputil.I18n(ctx, "desired"),
		Width:     100,
		Sorter:    true,
	}
	currentColumn := Column{
		DataIndex: "current",
		Title:     cputil.I18n(ctx, "current"),
		Width:     100,
		Sorter:    true,
	}
	completionsColumn := Column{
		DataIndex: "completions",
		Title:     cputil.I18n(ctx, "completions"),
		Width:     100,
		Sorter:    true,
	}
	durationColumn := Column{
		DataIndex: "duration",
		Title:     cputil.I18n(ctx, "jobDuration"),
		Width:     120,
		Sorter:    true,
	}
	scheduleColumn := Column{
		DataIndex: "schedule",
		Title:     cputil.I18n(ctx, "schedule"),
		Width:     100,
		Sorter:    true,
	}
	lastScheduleColumn := Column{
		DataIndex: "lastSchedule",
		Title:     cputil.I18n(ctx, "lastSchedule"),
		Width:     120,
		Sorter:    true,
	}

	if len(w.State.Values.Kind) != 1 {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, ageColumn,
		}
	} else if w.State.Values.Kind[0] == filter.DeploymentType {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, readyColumn, upToDateColumn, availableColumn, ageColumn,
		}
	} else if w.State.Values.Kind[0] == filter.DaemonSetType {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, desiredColumn, currentColumn, readyColumn, upToDateColumn, availableColumn, ageColumn,
		}
	} else if w.State.Values.Kind[0] == filter.StatefulSetType {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, readyColumn, ageColumn,
		}
	} else if w.State.Values.Kind[0] == filter.JobType {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, completionsColumn, durationColumn, ageColumn,
		}
	} else if w.State.Values.Kind[0] == filter.CronJobType {
		w.Props.Columns = []Column{
			statusColumn, kindColumn, nameColumn, namespaceColumn, scheduleColumn, lastScheduleColumn, ageColumn,
		}
	}

	w.Operations = make(map[string]interface{})
	w.Operations = map[string]interface{}{
		apistructs.OnChangeSortOperation.String(): Operation{
			Key:    apistructs.OnChangeSortOperation.String(),
			Reload: true,
		},
	}
}

func (w *ComponentWorkloadTable) Transfer(component *cptype.Component) {
	component.Props = w.Props
	component.Data = map[string]interface{}{
		"list": w.Data.List,
	}
	component.State = map[string]interface{}{
		"clusterName":             w.State.ClusterName,
		"countValues":             w.State.CountValues,
		"pageNo":                  w.State.PageNo,
		"pageSize":                w.State.PageSize,
		"sorterData":              w.State.Sorter,
		"total":                   w.State.Total,
		"values":                  w.State.Values,
		"workloadTable__urlQuery": w.State.WorkloadTableURLQuery,
	}
	component.Operations = w.Operations
}

func getWorkloadKindMap(kinds []string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, kind := range kinds {
		switch kind {
		case filter.DeploymentType:
			res[string(apistructs.K8SDeployment)] = struct{}{}
		case filter.StatefulSetType:
			res[string(apistructs.K8SStatefulSet)] = struct{}{}
		case filter.DaemonSetType:
			res[string(apistructs.K8SDaemonSet)] = struct{}{}
		case filter.JobType:
			res[string(apistructs.K8SJob)] = struct{}{}
		case filter.CronJobType:
			res[string(apistructs.K8SCronJob)] = struct{}{}
		default:
			res[kind] = struct{}{}
		}
	}
	return res
}

func contain(arr []string, target string) bool {
	for _, str := range arr {
		if target == str {
			return true
		}
	}
	return false
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
