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

package deployment_order

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	release2 "github.com/erda-project/erda/modules/dicehub/release"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var ProjectReleaseResp = &pb.ReleaseGetResponseData{
	IsProjectRelease: true,
	Labels:           map[string]string{gitBranchLabel: "master"},
	ApplicationReleaseList: []*pb.ReleaseSummaryArray{
		{
			List: []*pb.ApplicationReleaseSummary{
				{
					ReleaseID:   "0856df7931494d239abf07a145ade6e9",
					ReleaseName: "release/1.0.1",
					Version:     "1.0.1+20220210153458",
				},
			},
		},
	},
}

var AppReleaseResp = &pb.ReleaseGetResponseData{
	Labels:          map[string]string{gitBranchLabel: "master"},
	ApplicationID:   1,
	ApplicationName: "test",
}

func TestComposeRuntimeCreateRequests(t *testing.T) {
	do := New()
	bdl := bundle.New()

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProjectWithSetter",
		func(*bundle.Bundle, uint64, ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{ClusterConfig: map[string]string{"PROD": "fake-cluster"}}, nil
		},
	)

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	ProjectReleaseResp.ApplicationReleaseList = []*pb.ReleaseSummaryArray{
		{
			List: []*pb.ApplicationReleaseSummary{
				{
					ReleaseID: "8781f475e5617a04",
				},
			},
		},
	}

	_, err = do.composeRuntimeCreateRequests(&dbclient.DeploymentOrder{
		BatchSize:    10,
		CurrentBatch: 1,
		Type:         apistructs.TypeProjectRelease,
		Params:       string(paramsJson),
		Workspace:    apistructs.WORKSPACE_PROD,
	}, ProjectReleaseResp, apistructs.SourceDeployCenter, false)
	assert.NoError(t, err)

	_, err = do.composeRuntimeCreateRequests(&dbclient.DeploymentOrder{
		BatchSize:    1,
		CurrentBatch: 1,
		Type:         apistructs.TypeApplicationRelease,
		Params:       string(paramsJson),
		Workspace:    apistructs.WORKSPACE_PROD,
	}, AppReleaseResp, apistructs.SourceDeployCenter, false)
	assert.NoError(t, err)
}

func TestFetchDeploymentOrderParam(t *testing.T) {
	var order *DeploymentOrder

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(order), "FetchDeploymentConfigDetail",
		func(*DeploymentOrder, string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
			return []apistructs.EnvConfig{{Key: "key1", Value: "value1", ConfigType: "ENV", Comment: "test1"}},
				[]apistructs.EnvConfig{{Key: "key2", Value: "value2", ConfigType: "FILE", Encrypt: true}}, nil
		},
	)

	got, err := order.fetchDeploymentParams(1, "STAGING")
	assert.NoError(t, err)
	assert.Equal(t, got, &apistructs.DeploymentOrderParam{
		{Key: "key1", Value: "value1", Type: "ENV", Comment: "test1"},
		{Key: "key2", Value: "value2", Type: "FILE", Encrypt: true},
	})
}

func TestParseShowParams(t *testing.T) {
	data := apistructs.DeploymentOrderParam{
		{Key: "key1", Value: "value1", Type: "FILE"},
		{Key: "key2", Value: "value2", Type: "ENV"},
	}
	got := covertParamsType(&data)
	for _, p := range *got {
		if !(p.Type == "dice-file" || p.Type == "kv") {
			panic(fmt.Errorf("params type error: %v", p.Type))
		}
	}
}

func TestParseAppsInfoWithOrder(t *testing.T) {
	order := New()
	got, err := order.parseAppsInfoWithOrder(context.Background(), &dbclient.DeploymentOrder{
		ApplicationName: "test",
		ApplicationId:   1,
		Type:            apistructs.TypeApplicationRelease,
	})
	assert.NoError(t, err)
	assert.Equal(t, got, map[int64]string{1: "test"})
}

func TestParseAppsInfoWithRelease(t *testing.T) {
	order := New()
	got := order.parseAppsInfoWithRelease(&pb.ReleaseGetResponseData{
		IsProjectRelease: true,
		ApplicationReleaseList: []*pb.ReleaseSummaryArray{
			{
				List: []*pb.ApplicationReleaseSummary{
					{ApplicationName: "test-1", ApplicationID: 1},
					{ApplicationName: "test-2", ApplicationID: 2},
				},
			},
		},
	})
	assert.Equal(t, got, map[int64]string{1: "test-1", 2: "test-2"})
}

func TestContinueDeployOrder(t *testing.T) {
	order := New()
	bdl := bundle.New()
	rt := runtime.New()
	releaseSvc := &release2.ReleaseService{}
	order.releaseSvc = releaseSvc

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "GetDeploymentOrder", func(*dbclient.DBClient, string) (*dbclient.DeploymentOrder, error) {
		return &dbclient.DeploymentOrder{
			CurrentBatch: 1,
			BatchSize:    3,
			Workspace:    apistructs.WORKSPACE_PROD,
			Params:       string(paramsJson),
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "UpdateDeploymentOrder", func(*dbclient.DBClient, *dbclient.DeploymentOrder) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(releaseSvc), "GetRelease",
		func(*release2.ReleaseService, context.Context, *pb.ReleaseGetRequest) (*pb.ReleaseGetResponse, error) {
			return &pb.ReleaseGetResponse{Data: ProjectReleaseResp}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProjectWithSetter",
		func(*bundle.Bundle, uint64, ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{ClusterConfig: map[string]string{"PROD": "fake-cluster"}}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(rt), "Create", func(*runtime.Runtime, user.ID, *apistructs.RuntimeCreateRequest) (*apistructs.DeploymentCreateResponseDTO, error) {
		return &apistructs.DeploymentCreateResponseDTO{}, nil
	})

	err = order.ContinueDeployOrder("d9f06aaf-e3b7-4e05-9433-7742162e98f9")
	assert.NoError(t, err)
}

func TestParseRuntimeNameFromBranch(t *testing.T) {
	// from deploy center
	assert.Equal(t, parseRuntimeNameFromBranch(&apistructs.DeploymentOrderCreateRequest{
		Source: apistructs.SourceDeployCenter,
	}), false)
	// from pipeline, if not specified deployWithoutBranch, then use branch name
	assert.Equal(t, parseRuntimeNameFromBranch(&apistructs.DeploymentOrderCreateRequest{
		Source:    apistructs.SourceDeployPipeline,
		ReleaseId: "fake-release-id",
	}), true)
	// from pipeline, specified deployWithoutBranch
	assert.Equal(t, parseRuntimeNameFromBranch(&apistructs.DeploymentOrderCreateRequest{
		Source:              apistructs.SourceDeployPipeline,
		ReleaseId:           "fake-release-id",
		DeployWithoutBranch: true,
	}), false)
	// deploy application release
	assert.Equal(t, parseRuntimeNameFromBranch(&apistructs.DeploymentOrderCreateRequest{
		Type:            apistructs.TypeProjectRelease,
		ReleaseName:     "1.0.0",
		ApplicationName: "app-1",
		Source:          apistructs.SourceDeployPipeline,
	}), false)
}

func TestGetWorkspaceFromBranch(t *testing.T) {
	order := New()
	bdl := bundle.New()

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetBranchRules", func(*bundle.Bundle, apistructs.ScopeType, uint64) ([]*apistructs.BranchRule, error) {
		return []*apistructs.BranchRule{
			{
				Rule:      "feature/*",
				Workspace: apistructs.WORKSPACE_DEV,
			},
			{
				Rule:      "develop",
				Workspace: apistructs.WORKSPACE_TEST,
			},
			{
				Rule:      "release/*,hotfix/*",
				Workspace: apistructs.WORKSPACE_STAGING,
			},
			{
				Rule:      "master,support/*",
				Workspace: apistructs.WORKSPACE_PROD,
			},
		}, nil
	})

	type args struct {
		Branch string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "apps-1",
			args: args{
				Branch: "feature/demo",
			},
			want: apistructs.WORKSPACE_DEV,
		},
		{
			name: "apps-2",
			args: args{
				Branch: "develop",
			},
			want: apistructs.WORKSPACE_TEST,
		},
		{
			name: "apps-3",
			args: args{
				Branch: "release/1.0",
			},
			want: apistructs.WORKSPACE_STAGING,
		},
		{
			name: "apps-4",
			args: args{
				Branch: "master",
			},
			want: apistructs.WORKSPACE_PROD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := order.getWorkspaceFromBranch(1, tt.args.Branch)
			assert.NoError(t, err)
			if tt.want != got {
				t.Errorf("getWorkspaceFromBranch got = %v, want %v", got, tt.want)
			}
		})
	}
}
