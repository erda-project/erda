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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rancher/apiserver/pkg/types"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-workloads-list/filter"
)

func TestComponentWorkloadTable_GenComponentState(t *testing.T) {
	component := &cptype.Component{
		State: map[string]interface{}{
			"clusterName": "test",
			"pageNo":      1,
			"pageSize":    20,
			"total":       100,
			"sorterData": Sorter{
				Field: "test",
				Order: "test",
			},
			"values": Values{
				Namespace: []string{"test"},
				Kind:      []string{"test"},
				Status:    []string{"test"},
				Search:    "test",
			},
			"countValues": CountValues{
				DeploymentsCount: Count{
					Active: 1,
					Error:  1,
				},
				DaemonSetCount: Count{
					Active: 1,
					Error:  1,
				},
				StatefulSetCount: Count{
					Active: 1,
					Error:  1,
				},
				JobCount: Count{
					Active:    1,
					Succeeded: 1,
					Failed:    1,
				},
				CronJobCount: Count{
					Active: 1,
				},
			},
		},
	}
	src, err := json.Marshal(component.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	f := &ComponentWorkloadTable{}
	if err := f.GenComponentState(component); err != nil {
		t.Errorf("test failed, %v", err)
	}

	dst, err := json.Marshal(f.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	fmt.Println(string(src))
	fmt.Println(string(dst))
	if string(src) != string(dst) {
		t.Error("test failed, generate result is unexpected")
	}
}

type MockSteveServer struct {
	cmp.SteveServer
}

func (m *MockSteveServer) ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error) {
	switch req.Type {
	case apistructs.K8SDeployment:
		return []types.APIObject{
			{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"name":      "deploy-test",
						"namespace": "default",
						"fields": []interface{}{
							"deploy-test",
							"1/1",
							1,
							1,
							"1d",
							"deploy-test",
							"deploy-image",
							"",
						},
					},
				},
			},
		}, nil
	case apistructs.K8SDaemonSet:
		return []types.APIObject{
			{
				Object: map[string]interface{}{
					"kind": "DaemonSet",
					"metadata": map[string]interface{}{
						"name":      "daemonSet-test",
						"namespace": "default",
						"fields": []interface{}{
							"daemonSet-test",
							1,
							1,
							1,
							1,
							1,
							"<none>",
							"1d",
							"daemonSet-test",
							"daemonSet-image",
							"",
						},
					},
				},
			},
		}, nil
	case apistructs.K8SStatefulSet:
		return []types.APIObject{
			{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
					"metadata": map[string]interface{}{
						"name":      "statefulSet-test",
						"namespace": "default",
						"fields": []interface{}{
							"daemonSet-test",
							"1/1",
							"1d",
							"daemonSet-test",
							"daemonSet-image",
						},
					},
				},
			},
		}, nil
	case apistructs.K8SJob:
		return []types.APIObject{
			{
				Object: map[string]interface{}{
					"kind": "Job",
					"metadata": map[string]interface{}{
						"name":      "job-test",
						"namespace": "default",
						"fields": []interface{}{
							"job-test",
							"1/1",
							"10s",
							"1d",
							"job-test",
							"job-image",
							"",
						},
					},
				},
			},
		}, nil
	case apistructs.K8SCronJob:
		return []types.APIObject{
			{
				Object: map[string]interface{}{
					"kind": "CronJob",
					"metadata": map[string]interface{}{
						"name":      "cronJob-test",
						"namespace": "default",
						"fields": []interface{}{
							"cronJob-test",
							"0 * * * *",
							"False",
							0,
							"1m",
							"1d",
							"k8s",
							"cronJob-image",
							"<none>",
						},
					},
				},
			},
		}, nil
	}
	return []types.APIObject{}, nil
}

type MockTran struct {
	i18n.Translator
}

func (m *MockTran) Text(lang i18n.LanguageCodes, key string) string {
	return ""
}

func (m *MockTran) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return ""
}

func TestComponentWorkloadTable_RenderTable(t *testing.T) {
	w := ComponentWorkloadTable{
		sdk: &cptype.SDK{
			Tran: &MockTran{},
			Identity: &pb.IdentityInfo{
				UserID: "1",
				OrgID:  "1",
			},
		},
		server: &MockSteveServer{},
	}
	if err := w.RenderTable(); err != nil {
		t.Errorf("test failed, %v", err)
	}
}

func TestComponentWorkloadTable_SetComponentValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), cptype.GlobalInnerKeyCtxSDK, &cptype.SDK{Tran: &MockTran{}})
	w := ComponentWorkloadTable{}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 5 {
		t.Errorf("test failed, expected length of columns in props is 5, actual %d", len(w.Props.Columns))
	}

	w.State.Values.Kind = []string{filter.DeploymentType}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 8 {
		t.Errorf("test failed, expected length of columns in props is 8, actual %d", len(w.Props.Columns))
	}

	w.State.Values.Kind = []string{filter.DaemonSetType}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 10 {
		t.Errorf("test failed, expected length of columns in props is 10, actual %d", len(w.Props.Columns))
	}

	w.State.Values.Kind = []string{filter.StatefulSetType}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 6 {
		t.Errorf("test failed, expected length of columns in props is 6, actual %d", len(w.Props.Columns))
	}

	w.State.Values.Kind = []string{filter.JobType}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 7 {
		t.Errorf("test failed, expected length of columns in props is 7, actual %d", len(w.Props.Columns))
	}

	w.State.Values.Kind = []string{filter.CronJobType}
	w.SetComponentValue(ctx)
	if len(w.Props.Columns) != 7 {
		t.Errorf("test failed, expected length of columns in props is 7, actual %d", len(w.Props.Columns))
	}
}

func TestGetWorkloadKindMap(t *testing.T) {
	kinds := []string{"test1", "test2"}
	mp := getWorkloadKindMap(kinds)
	if _, ok := mp["test1"]; !ok {
		t.Errorf("test failed, expect key is not exist in res")
	}
	if _, ok := mp["test2"]; !ok {
		t.Errorf("test failed, expect key is not exist in res")
	}
}

func TestContain(t *testing.T) {
	arr := []string{
		"a", "b", "c", "d",
	}
	if contain(arr, "e") {
		t.Errorf("test failed, expected not contain \"e\", actual do")
	}
	if !contain(arr, "a") || !contain(arr, "b") || !contain(arr, "c") || !contain(arr, "d") {
		t.Errorf("test failed, expected contain \"a\" , \"b\", \"c\" and \"d\", actual not")
	}
}

func TestGetRange(t *testing.T) {
	length := 0
	pageNo := 1
	pageSize := 20
	l, r := getRange(length, pageNo, pageSize)
	if l != 0 {
		t.Errorf("test failed, l is unexpected, expected 0, actual %d", l)
	}
	if r != 0 {
		t.Errorf("test failed, r is unexpected, expected 0, actual %d", r)
	}

	length = 21
	pageNo = 2
	pageSize = 20
	l, r = getRange(length, pageNo, pageSize)
	if l != 20 {
		t.Errorf("test failed, l is unexpected, expected 20, actual %d", l)
	}
	if r != 21 {
		t.Errorf("test failed, r is unexpected, expected 21, actual %d", r)
	}

	length = 20
	pageNo = 2
	pageSize = 50
	l, r = getRange(length, pageNo, pageSize)
	if l != 0 {
		t.Errorf("test failed, l is unexpected, expected 0, actual %d", l)
	}
	if r != 20 {
		t.Errorf("test failed, r is unexpected, expected 20, actual %d", r)
	}
}
