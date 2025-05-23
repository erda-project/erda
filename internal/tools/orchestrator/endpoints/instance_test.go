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

package endpoints

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/instanceinfo"
)

func TestGetContainers(t *testing.T) {
	e := &Endpoints{
		instanceinfoImpl: &instanceinfo.InstanceInfoImpl{},
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(e.instanceinfoImpl), "GetInstanceInfo", func(i *instanceinfo.InstanceInfoImpl, req apistructs.InstanceInfoRequest) (apistructs.InstanceInfoDataList, error) {
		res := apistructs.InstanceInfoDataList{}
		for i := 0; i < 5; i++ {
			infoData := apistructs.InstanceInfoData{
				Meta:        fmt.Sprintf("Meta - %v", i),
				TaskID:      fmt.Sprintf("TaskID - %v", i),
				ContainerID: fmt.Sprintf("ContainerID - %v", i),
				ContainerIP: "127.0.0.1",
				HostIP:      "127.0.0.1",
				Image:       fmt.Sprintf("Image - %v", i),
				CpuRequest:  float64(i),
				MemRequest:  i,
				Phase:       fmt.Sprintf("Phase - %v", i),
				ExitCode:    200,
				Message:     "success",
				StartedAt:   time.Now(),
				ServiceName: fmt.Sprintf("ServiceName - %v", i),
				Cluster:     fmt.Sprintf("Cluster - %v", i),
			}
			if i%2 == 0 {
				finish := time.Now()
				infoData.FinishedAt = &finish
			}
			res = append(res, infoData)
		}
		return res, nil
	})
	_, err := e.getContainers(apistructs.InstanceInfoRequest{})
	if err != nil {
		t.Errorf("getContainers error %v", err.Error())
	}
}

func TestParseMeta(t *testing.T) {
	var (
		metaNamespace     = "namespace"
		metaPodUid        = "5e352011-f819-4dbb-bfea-3060cb866b53"
		metaPodName       = "test-pod"
		metaContainerName = "test-container"
	)
	tests := []struct {
		name  string
		input string
		want  apistructs.K8sInstanceMetaInfo
	}{
		{
			name:  "empty",
			input: "",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
		{
			name:  "no meta",
			input: "hello world",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
		{
			name: "one meta",
			input: strings.Join([]string{
				fmt.Sprintf("%s=%s", apistructs.K8sNamespace, metaNamespace),
				fmt.Sprintf("%s=%s", apistructs.K8sPodUid, metaPodUid),
				fmt.Sprintf("%s=%s", apistructs.K8sPodName, metaPodName),
				fmt.Sprintf("%s=%s", apistructs.K8sContainerName, metaContainerName),
			}, ","),
			want: apistructs.K8sInstanceMetaInfo{
				PodUid:        metaPodUid,
				PodName:       metaPodName,
				PodNamespace:  metaNamespace,
				ContainerName: metaContainerName,
			},
		},
		{
			name:  "invalid meta",
			input: "hello world:meta1=value1:meta2=value2:",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInstanceMeta(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInstanceMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
