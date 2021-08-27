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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/recallsong/go-utils/container/slice"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-workloads/components/filter"
)

func RenderCreator() protocol.CompRender {
	return &ComponentWorkloadTable{}
}

func (w *ComponentWorkloadTable) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	if err := w.SetCtxBundle(ctx); err != nil {
		return fmt.Errorf("failed to set workloadTable component ctx bundle, %v", err)
	}
	if err := w.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadTable component state, %v", err)
	}

	switch event.Operation {
	case apistructs.InitializeOperation:
		w.State.PageNo = 1
		w.State.PageSize = 20
	case apistructs.RenderingOperation, apistructs.OnChangePageSizeOperation, apistructs.OnChangeSortOperation:
		w.State.PageNo = 1
	}

	if err := w.DecodeURLQuery(); err != nil {
		return fmt.Errorf("failed to get url query for workloadTable component, %v", err)
	}
	if err := w.RenderTable(); err != nil {
		return fmt.Errorf("failed to render workloadTable component, %v", err)
	}
	if err := w.EncodeURLQuery(); err != nil {
		return fmt.Errorf("failed to gen url query for workloadTable component, %v", err)
	}
	w.SetComponentValue()
	return nil
}

func (w *ComponentWorkloadTable) DecodeURLQuery() error {
	queryData, ok := w.ctxBdl.InParams["workloadTable__urlQuery"].(string)
	if !ok {
		return nil
	}
	decode, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}
	query := make(map[string]int)
	if err := json.Unmarshal(decode, &query); err != nil {
		return err
	}
	w.State.PageNo = uint64(query["pageNo"])
	w.State.PageSize = uint64(query["pageSize"])
	return nil
}

func (w *ComponentWorkloadTable) EncodeURLQuery() error {
	query := make(map[string]int)
	query["pageNo"] = int(w.State.PageNo)
	query["pageSize"] = int(w.State.PageSize)

	data, err := json.Marshal(query)
	if err != nil {
		return err
	}

	encode := base64.StdEncoding.EncodeToString(data)
	w.State.WorkloadTableURLQuery = encode
	return nil
}

func (w *ComponentWorkloadTable) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("bundle in context can not be empty")
	}
	w.ctxBdl = bdl
	return nil
}

func (w *ComponentWorkloadTable) GenComponentState(component *apistructs.Component) error {
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
	w.State = state
	return nil
}

func (w *ComponentWorkloadTable) RenderTable() error {
	userID := w.ctxBdl.Identity.UserID
	orgID := w.ctxBdl.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		ClusterName: w.State.ClusterName,
	}

	var items []Item
	kinds := getWorkloadKindMap(w.State.Values.Kind)

	// deployment
	if _, ok := kinds[filter.DeploymentType]; ok || len(kinds) == 0 {
		req.Type = apistructs.K8SDeployment
		obj, err := w.ctxBdl.Bdl.ListSteveResource(&req)
		if err != nil {
			return err
		}
		list := obj.Slice("data")

		for _, obj := range list {
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, obj.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(obj.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			fields := obj.StringSlice("metadata", "fields")
			if len(fields) != 8 {
				logrus.Errorf("deployment %s has invalid fields length", obj.String("metadata", "name"))
				continue
			}

			status := Status{RenderType: "text"}
			if fields[2] == fields[3] {
				status.Value = "Active"
				status.StyleConfig.Color = "green"
			} else {
				status.Value = "Error"
				status.StyleConfig.Color = "red"
			}

			name := fields[0]
			namespace := obj.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", apistructs.K8SDaemonSet, namespace, name)
			ageNum, _ := time.ParseDuration(fields[4])
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
								Target: "cmpClustersWorkload",
								State: CommandState{
									Params: map[string]string{
										"workloadId": id,
									},
								},
								JumpOut: true,
							},
							Reload: false,
						},
					},
				},
				Namespace: namespace,
				Kind:      obj.String("kind"),
				Age:       fields[4],
				AgeNum:    ageNum.Nanoseconds(),
				Ready:     fields[1],
				UpToDate:  fields[2],
				Available: fields[3],
			})
		}
	}

	// daemonSet
	if _, ok := kinds[filter.DaemonSetType]; ok || len(kinds) != 0 {
		req.Type = apistructs.K8SDaemonSet
		obj, err := w.ctxBdl.Bdl.ListSteveResource(&req)
		if err != nil {
			return err
		}
		list := obj.Slice("data")

		for _, obj := range list {
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, obj.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(obj.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			fields := obj.StringSlice("metadata", "fields")
			if len(fields) != 11 {
				logrus.Errorf("daemonset %s has invalid fields length", obj.String("metadata", "name"))
				continue
			}

			status := Status{RenderType: "text"}
			if fields[3] == fields[1] {
				status.Value = "Active"
				status.StyleConfig.Color = "green"
			} else {
				status.Value = "Error"
				status.StyleConfig.Color = "red"
			}

			name := fields[0]
			namespace := obj.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", apistructs.K8SDaemonSet, namespace, name)
			ageNum, _ := time.ParseDuration(fields[7])
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
								Target: "cmpClustersWorkload",
								State: CommandState{
									Params: map[string]string{
										"workloadId": id,
									},
								},
								JumpOut: true,
							},
							Reload: false,
						},
					},
				},
				Namespace: namespace,
				Kind:      obj.String("kind"),
				Age:       fields[7],
				AgeNum:    ageNum.Nanoseconds(),
				Ready:     fields[3],
				UpToDate:  fields[4],
				Available: fields[5],
				Desired:   fields[1],
				Current:   fields[2],
			})
		}
	}

	// statefulSet
	if _, ok := kinds[filter.StatefulSetType]; ok || len(kinds) == 0 {
		req.Type = apistructs.K8SStatefulSet
		obj, err := w.ctxBdl.Bdl.ListSteveResource(&req)
		if err != nil {
			return err
		}
		list := obj.Slice("data")

		for _, obj := range list {
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, obj.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(obj.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			fields := obj.StringSlice("metadata", "fields")
			if len(fields) != 5 {
				logrus.Errorf("statefulSet %s has invalid fields length", obj.String("metadata", "name"))
				continue
			}

			status := Status{RenderType: "text"}
			readyPods := strings.Split(fields[1], "/")
			if readyPods[0] == readyPods[1] {
				status.Value = "Active"
				status.StyleConfig.Color = "green"
			} else {
				status.Value = "Error"
				status.StyleConfig.Color = "red"
			}

			name := fields[0]
			namespace := obj.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", apistructs.K8SStatefulSet, namespace, name)
			ageNum, _ := time.ParseDuration(fields[4])
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
								Target: "cmpClustersWorkload",
								State: CommandState{
									Params: map[string]string{
										"workloadId": id,
									},
								},
								JumpOut: true,
							},
							Reload: false,
						},
					},
				},
				Namespace: namespace,
				Kind:      obj.String("kind"),
				Age:       fields[2],
				AgeNum:    ageNum.Nanoseconds(),
				Ready:     fields[1],
			})
		}
	}

	// job
	if _, ok := kinds[filter.JobType]; ok || len(kinds) == 0 {
		req.Type = apistructs.K8SJob
		obj, err := w.ctxBdl.Bdl.ListSteveResource(&req)
		if err != nil {
			return err
		}
		list := obj.Slice("data")

		for _, obj := range list {
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, obj.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(obj.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			fields := obj.StringSlice("metadata", "fields")
			if len(fields) != 7 {
				logrus.Errorf("job %s has invalid fields length", obj.String("metadata", "name"))
				continue
			}

			status := Status{RenderType: "text"}
			succeed := obj.String("status", "succeeded")
			active := obj.String("status", "active")
			failed := obj.String("status", "failed")
			if failed != "" {
				status.Value = "Failed"
				status.StyleConfig.Color = "red"
			} else if active != "" {
				status.Value = "Active"
				status.StyleConfig.Color = "green"
			} else if succeed != "" {
				status.Value = "Succeeded"
				status.StyleConfig.Color = "steelBlue"
			}

			name := fields[0]
			namespace := obj.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", apistructs.K8SJob, namespace, name)
			ageNum, _ := time.ParseDuration(fields[3])
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
								Target: "cmpClustersWorkload",
								State: CommandState{
									Params: map[string]string{
										"workloadId": id,
									},
								},
								JumpOut: true,
							},
							Reload: false,
						},
					},
				},
				Namespace:   namespace,
				Kind:        obj.String("kind"),
				Age:         fields[3],
				AgeNum:      ageNum.Nanoseconds(),
				Completions: fields[1],
				Duration:    fields[2],
			})
		}
	}

	// cronjob
	if _, ok := kinds[filter.CronJobType]; ok || len(kinds) == 0 {
		req.Type = apistructs.K8SCronJob
		obj, err := w.ctxBdl.Bdl.ListSteveResource(&req)
		if err != nil {
			return err
		}
		list := obj.Slice("data")

		for _, obj := range list {
			if w.State.Values.Namespace != nil && !contain(w.State.Values.Namespace, obj.String("metadata", "namespace")) {
				continue
			}
			if w.State.Values.Search != "" && !strings.Contains(obj.String("metadata", "name"), w.State.Values.Search) {
				continue
			}

			fields := obj.StringSlice("metadata", "fields")
			if len(fields) != 9 {
				logrus.Errorf("cronjob %s has invalid fields length", obj.String("metadata", "name"))
				continue
			}

			status := Status{
				RenderType: "text",
				Value:      "Active",
				StyleConfig: StyleConfig{
					Color: "green",
				},
			}
			name := fields[0]
			namespace := obj.String("metadata", "namespace")
			id := fmt.Sprintf("%s_%s_%s", apistructs.K8SCronJob, namespace, name)
			ageNum, _ := time.ParseDuration(fields[5])
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
								Target: "cmpClustersWorkload",
								State: CommandState{
									Params: map[string]string{
										"workloadId": id,
									},
								},
								JumpOut: true,
							},
							Reload: false,
						},
					},
				},
				Namespace:    namespace,
				Kind:         obj.String("kind"),
				Age:          fields[5],
				AgeNum:       ageNum.Nanoseconds(),
				Schedule:     fields[1],
				LastSchedule: fields[4],
			})
		}
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
					less := items[i].AgeNum < items[j].AgeNum
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
					nI, _ := time.ParseDuration(items[i].Duration)
					nJ, _ := time.ParseDuration(items[j].Duration)
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
					nI, _ := time.ParseDuration(items[i].LastSchedule)
					nJ, _ := time.ParseDuration(items[j].LastSchedule)
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

	l, r := getRange(len(items), int(w.State.PageNo), int(w.State.PageSize))
	w.Data.List = items[l:r]
	w.State.Total = uint64(len(items))
	return nil
}

func (w *ComponentWorkloadTable) SetComponentValue() {
	w.Props.PageSizeOptions = []string{"10", "20", "50", "100"}

	statusColumn := Column{
		DataIndex: "status",
		Title:     "Status",
		Width:     120,
		Sorter:    true,
	}
	kindColumn := Column{
		DataIndex: "kind",
		Title:     "Kind",
		Width:     120,
		Sorter:    true,
	}
	nameColumn := Column{
		DataIndex: "name",
		Title:     "Name",
		Width:     180,
		Sorter:    true,
	}
	namespaceColumn := Column{
		DataIndex: "namespace",
		Title:     "Namespace",
		Width:     180,
		Sorter:    true,
	}
	ageColumn := Column{
		DataIndex: "age",
		Title:     "Age",
		Width:     100,
		Sorter:    true,
	}
	readyColumn := Column{
		DataIndex: "ready",
		Title:     "Ready",
		Width:     100,
		Sorter:    true,
	}
	upToDateColumn := Column{
		DataIndex: "upToDate",
		Title:     "Up-to-Date",
		Width:     100,
		Sorter:    true,
	}
	availableColumn := Column{
		DataIndex: "available",
		Title:     "Available",
		Width:     100,
		Sorter:    true,
	}
	desiredColumn := Column{
		DataIndex: "desired",
		Title:     "Desired",
		Width:     100,
		Sorter:    true,
	}
	currentColumn := Column{
		DataIndex: "current",
		Title:     "Current",
		Width:     100,
		Sorter:    true,
	}
	completionsColumn := Column{
		DataIndex: "completions",
		Title:     "Completions",
		Width:     100,
		Sorter:    true,
	}
	durationColumn := Column{
		DataIndex: "duration",
		Title:     "Duration",
		Width:     120,
		Sorter:    true,
	}
	scheduleColumn := Column{
		DataIndex: "schedule",
		Title:     "Schedule",
		Width:     100,
		Sorter:    true,
	}
	lastScheduleColumn := Column{
		DataIndex: "lastSchedule",
		Title:     "LastSchedule",
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
		apistructs.OnChangePageNoOperation.String(): Operation{
			Key:    apistructs.ChangePageNoOperation.String(),
			Reload: true,
		},
		apistructs.OnChangePageSizeOperation.String(): Operation{
			Key:    apistructs.OnChangePageSizeOperation.String(),
			Reload: true,
		},
		apistructs.OnChangeSortOperation.String(): Operation{
			Key:    apistructs.OnChangeSortOperation.String(),
			Reload: true,
		},
	}
}

func getWorkloadKindMap(kinds []string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, kind := range kinds {
		res[kind] = struct{}{}
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
	if l >= length {
		l = 0
	}
	r := l + pageSize
	if r > length {
		r = length
	}
	return l, r
}
