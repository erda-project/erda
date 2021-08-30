// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package runtime 应用实例相关操作
package runtime

import (
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestModifyStatusIfNotForDisplay(t *testing.T) {
	runtime := apistructs.RuntimeInspectDTO{
		Status: "Unknown",
		Services: map[string]*apistructs.RuntimeInspectServiceDTO{
			"test": {
				Status: "Stopped",
			},
		},
	}
	updateStatusToDisplay(&runtime)
	assert.Equal(t, "Unknown", runtime.Status)
	for _, s := range runtime.Services {
		assert.Equal(t, "Stopped", s.Status)
	}
}

func TestFillRuntimeDataWithServiceGroup(t *testing.T) {
	var (
		data          apistructs.RuntimeInspectDTO
		targetService diceyml.Services
		sg            apistructs.ServiceGroup
		domainMap     = make(map[string][]string, 0)
		status        string
	)

	fakeData := `{"id":1,"name":"develop","serviceGroupName":"1f1a1k1e11","serviceGroupNamespace":"services","source":"PIPELINE","status":"Healthy","deployStatus":"CANCELED","deleteStatus":"","releaseId":"11f1a1k1e11111111111111111111111","clusterId":1,"clusterName":"fake-cluster","clusterType":"k8s","resources":{"cpu":0.6,"mem":3072,"disk":0},"extra":{"applicationId":1,"buildId":0,"workspace":"DEV"},"projectID":1,"services":{"fake-service":{"status":"Healthy","deployments":{"replicas":1},"resources":{"cpu":0.3,"mem":1536,"disk":0},"envs":{"fakeEnv":"1"},"addrs":["fake-service.services--1f1a1k1e11.svc.cluster.local:8060"],"expose":["http://fake-service-dev-1-app.dev.fake.io"],"errors":null}},"lastMessage":{},"timeCreated":"2021-06-11T16:40:49+08:00","createdAt":"2021-06-11T16:40:49+08:00","updatedAt":"2021-06-17T13:53:49+08:00","errors":null}`
	err := json.Unmarshal([]byte(fakeData), &data)
	assert.NoError(t, err)

	fakeTargetService := `{"fake-service":{"image":"registry.fake.com/dice/fake-service:fake","image_username":"","image_password":"","cmd":"","ports":[{"port":8060,"protocol":"TCP","l4_protocol":"TCP","expose":true,"default":false}],"envs":{"fakeEnv":"1"},"resources":{"cpu":0.3,"mem":1536,"max_cpu":1,"max_mem":1536,"disk":0,"network":{"mode":"container"}},"deployments":{"replicas":2,"policies":""},"health_check":{"http":{"port":8060,"path":"/fake/health","duration":200},"exec":{}},"traffic_security":{}}}`
	err = json.Unmarshal([]byte(fakeTargetService), &targetService)
	assert.NoError(t, err)

	fakeSG := `{"created_time":1623400858,"last_modified_time":1623908973,"executor":"fake","clusterName":"fake-cluster","force":true,"name":"1f1a1k1e11","namespace":"services","services":[{"name":"fake-service","namespace":"services--1f1a1k1e11","image":"registry.fake.com/dice/fake-service:fake","image_username":"","image_password":"","Ports":[{"port":8060,"protocol":"TCP","l4_protocol":"TCP","expose":true,"default":false}],"proxyPorts":[8060],"vip":"fake-service.services--1f1a1k1e11.svc.cluster.local","shortVIP":"192.168.1.1","proxyIp":"192.168.1.1","scale":2,"resources":{"cpu":0.8,"mem":1537},"health_check":{"http":{"port":8060,"path":"/fake","duration":200}},"traffic_security":{},"status":"Healthy","reason":"","unScheduledReasons":{}}],"serviceDiscoveryKind":"","serviceDiscoveryMode":"DEPEND","projectNamespace":"","status":"Healthy","reason":"","unScheduledReasons":{}}`
	err = json.Unmarshal([]byte(fakeSG), &sg)
	assert.NoError(t, err)

	domainMap["fake-service"] = []string{"http://fake-services-dev-1-app.fake.io"}
	status = "CANCELED"

	fillRuntimeDataWithServiceGroup(&data, targetService, &sg, domainMap, status)
	assert.Equal(t, "", data.ModuleErrMsg["fake-service"]["Msg"])
	assert.Equal(t, "", data.ModuleErrMsg["fake-service"]["Reason"])
	assert.Equal(t, 1.6, data.Resources.CPU)
	assert.Equal(t, 3074, data.Resources.Mem)
	assert.Equal(t, 0, data.Resources.Disk)

	assert.Equal(t, 0.8, data.Services["fake-service"].Resources.CPU)
	assert.Equal(t, 1537, data.Services["fake-service"].Resources.Mem)
	assert.Equal(t, 0, data.Services["fake-service"].Resources.Disk)
	assert.Equal(t, "Healthy", data.Services["fake-service"].Status)
	assert.Equal(t, 2, data.Services["fake-service"].Deployments.Replicas)
}

func TestGetRollbackConfig(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAllProjects",
		func(*bundle.Bundle) ([]apistructs.ProjectDTO, error) {
			return []apistructs.ProjectDTO{
				apistructs.ProjectDTO{ID: 1, RollbackConfig: map[string]int{"DEV": 3, "TEST": 5, "STAGING": 4, "PROD": 6}},
				apistructs.ProjectDTO{ID: 2, RollbackConfig: map[string]int{"DEV": 4, "TEST": 6, "STAGING": 5, "PROD": 7}},
				apistructs.ProjectDTO{ID: 3, RollbackConfig: map[string]int{"DEV": 5, "TEST": 7, "STAGING": 6, "PROD": 8}},
			}, nil
		},
	)
	defer monkey.UnpatchAll()

	r := New(WithBundle(bdl))
	cfg, err := r.getRollbackConfig()
	assert.NoError(t, err)
	assert.Equal(t, 3, cfg[1]["DEV"])
	assert.Equal(t, 6, cfg[2]["TEST"])
	assert.Equal(t, 6, cfg[3]["STAGING"])
	assert.Equal(t, 8, cfg[3]["PROD"])
}

func Test_getRedeployPipelineYmlName(t *testing.T) {
	type args struct {
		runtime dbclient.Runtime
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases
		{
			name: "Filled in the space and scene set",
			args: args{
				runtime: dbclient.Runtime{
					ApplicationID: 1,
					Workspace:     "PORD",
					Name:          "master",
				},
			},
			want: "1/PORD/master/pipeline.yml",
		},
		{
			name: "Filled in the space and scene set",
			args: args{
				runtime: dbclient.Runtime{
					ApplicationID: 4,
					Workspace:     "TEST",
					Name:          "master",
				},
			},
			want: "4/TEST/master/pipeline.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRedeployPipelineYmlName(tt.args.runtime); got != tt.want {
				t.Errorf("getRedeployPipelineYmlName() = %v, want %v", got, tt.want)
			}
		})
	}
}
