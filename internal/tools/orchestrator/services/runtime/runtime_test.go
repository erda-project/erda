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

// Package runtime 应用实例相关操作
package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestGetRollbackConfig(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAllProjects",
		func(*bundle.Bundle) ([]apistructs.ProjectDTO, error) {
			return []apistructs.ProjectDTO{
				{ID: 1, RollbackConfig: map[string]int{"DEV": 3, "TEST": 5, "STAGING": 4, "PROD": 6}},
				{ID: 2, RollbackConfig: map[string]int{"DEV": 4, "TEST": 6, "STAGING": 5, "PROD": 7}},
				{ID: 3, RollbackConfig: map[string]int{"DEV": 5, "TEST": 7, "STAGING": 6, "PROD": 8}},
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

var diceYml = `version: 2.0
services:
  web:
    ports:
      - 8080
      - port: 20880
      - port: 1234
        protocol: "UDP"
      - port: 4321
        protocol: "HTTP"
      - port: 53
        protocol: "DNS"
        l4_protocol: "UDP"
        default: true
    deployments:
      replicas: 1
    resources:
      cpu: 0.1
      mem: 512
    k8s_snippet:
      container:
        name: abc
        stdin: true
        workingDir: aaa
        imagePullPolicy: Always
        securityContext:
          privileged: true
`

func TestGetServicesNames(t *testing.T) {
	name, err := getServicesNames(diceYml)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, []string{"web"}, name)
}

func TestConvertRuntimeDeployDto(t *testing.T) {
	app := &apistructs.ApplicationDTO{
		ID:          1,
		Name:        "foo",
		OrgID:       2,
		OrgName:     "erda",
		ProjectID:   3,
		ProjectName: "bar",
	}

	release := &pb.ReleaseGetResponseData{
		Diceyml: diceYml,
	}

	dto := &basepb.PipelineDTO{ID: 4}

	want := apistructs.RuntimeDeployDTO{
		PipelineID:      4,
		ApplicationID:   1,
		ApplicationName: "foo",
		ProjectID:       3,
		ProjectName:     "bar",
		OrgID:           2,
		OrgName:         "erda",
		ServicesNames:   []string{"web"},
	}
	deployDto, err := convertRuntimeDeployDto(app, release, dto)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, want, *deployDto)
}

func Test_setClusterName(t *testing.T) {
	var bdl *bundle.Bundle
	var clusterinfoImpl *clusterinfo.ClusterInfoImpl
	m1 := monkey.PatchInstanceMethod(reflect.TypeOf(clusterinfoImpl), "Info", func(_ *clusterinfo.ClusterInfoImpl, clusterName string) (apistructs.ClusterInfoData, error) {
		if clusterName == "erda-edge" {
			return apistructs.ClusterInfoData{apistructs.JOB_CLUSTER: "erda-center", apistructs.DICE_IS_EDGE: "true"}, nil
		}
		return apistructs.ClusterInfoData{apistructs.DICE_IS_EDGE: "false"}, nil
	})
	defer m1.Unpatch()
	runtimeSvc := New(WithBundle(bdl), WithClusterInfo(clusterinfoImpl))
	rt := &dbclient.Runtime{
		ClusterName: "erda-edge",
	}
	runtimeSvc.setClusterName(rt)
	assert.Equal(t, "erda-center", rt.ClusterName)
}

func generateMultiAddons(t *testing.T, randCount int) string {
	var addonsCount int
	if randCount != 0 {
		addonsCount = rand.Intn(randCount)
	}

	addons := make(diceyml.AddOns)
	for i := 0; i < addonsCount; i++ {
		name := fmt.Sprintf("existAddon%d", i)
		plan := fmt.Sprintf("existAddon%d:%s", i, apistructs.AddonUltimate)
		addons[name] = &diceyml.AddOn{
			Plan: plan,
			Options: map[string]string{
				"version": "1.0.0",
			},
		}
	}

	if addonsCount != 0 {
		addons["nonExistAddon1"] = &diceyml.AddOn{
			Plan: fmt.Sprintf("nonExistAddon1:%s", apistructs.AddonBasic),
			Options: map[string]string{
				"version": "1.0.0",
			},
		}
	}

	diceObj := diceyml.Object{
		Version: "2.0",
		AddOns:  addons,
	}

	diceYaml, err := yaml.Marshal(diceObj)
	assert.NoError(t, err)

	return string(diceYaml)
}

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "" {
		return nil, fmt.Errorf("the IdOrName is empty")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{}}, nil
}

func TestRuntime_GetOrg(t *testing.T) {
	type fields struct {
		org org.ClientInterface
	}
	type args struct {
		orgID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				org: orgMock{},
			},
			args:    args{orgID: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				org: orgMock{},
			},
			args:    args{orgID: 1},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runtime{
				org: tt.fields.org,
			}
			got, err := r.GetOrg(tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}
