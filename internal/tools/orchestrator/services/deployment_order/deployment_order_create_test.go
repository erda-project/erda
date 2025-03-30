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
	dicehubRelease "github.com/erda-project/erda/internal/apps/dop/dicehub/release"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/components/runtime"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var ProjectReleaseResp = &pb.ReleaseGetResponseData{
	ReleaseID:        "202cb962ac59075b964b07152d234b70",
	IsProjectRelease: true,
	Labels:           map[string]string{gitBranchLabel: "master"},
	Modes: map[string]*pb.ModeSummary{
		DefaultMode: {
			Expose:   true,
			DependOn: []string{},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{
					List: []*pb.ApplicationReleaseSummary{
						{
							ApplicationID:   1,
							ApplicationName: "app-demo",
							ReleaseID:       "0856df7931494d239abf07a145ade6e9",
							ReleaseName:     "release/1.0.1",
							Version:         "1.0.1+20220210153458",
						},
					},
				},
			},
		},
	},
}

var ProjectReleaseASum = [][]*pb.ApplicationReleaseSummary{
	{
		{
			ApplicationID:   1,
			ApplicationName: "test",
			ReleaseID:       "0856df7931494d239abf07a145ade6e9",
			ReleaseName:     "release/1.0.1",
			Version:         "1.0.1+20220210153458",
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
	monkey.PatchInstanceMethod(reflect.TypeOf(do.db), "ListReleases", func(*dbclient.DBClient, []string) ([]*dbclient.Release, error) {
		return []*dbclient.Release{
			{ReleaseId: "id1"},
			{ReleaseId: "id2"},
		}, nil
	})

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	deployList := [][]string{{"id1"}, {"id2"}}
	data, err := json.Marshal(deployList)
	if err != nil {
		t.Fatal(err)
	}

	_, err = do.composeRtReqWithDeployList(&dbclient.DeploymentOrder{
		BatchSize:    10,
		CurrentBatch: 1,
		Type:         apistructs.TypeProjectRelease,
		Params:       string(paramsJson),
		Workspace:    apistructs.WORKSPACE_PROD,
		DeployList:   string(data),
	}, ProjectReleaseASum, apistructs.SourceDeployCenter)
	assert.NoError(t, err)

	_, err = do.composeRtReqWithRelease(&dbclient.DeploymentOrder{
		BatchSize:    1,
		CurrentBatch: 1,
		Type:         apistructs.TypeApplicationRelease,
		Params:       string(paramsJson),
		Workspace:    apistructs.WORKSPACE_PROD,
		DeployList:   string(data),
	}, AppReleaseResp, apistructs.SourceDeployCenter, false)
	assert.NoError(t, err)
}

func TestFetchDeploymentOrderParam(t *testing.T) {
	bdl := bundle.New()
	order := New(WithBundle(bdl))

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(order), "FetchDeploymentConfigDetail",
		func(*DeploymentOrder, string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
			return []apistructs.EnvConfig{{Key: "key1", Value: "value1", ConfigType: "ENV", Comment: "test1"}},
				[]apistructs.EnvConfig{{Key: "key2", Value: "value2", ConfigType: "FILE", Encrypt: true}}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp", func(*bundle.Bundle, uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			ID: 1,
			Workspaces: []apistructs.ApplicationWorkspace{
				{
					Workspace:       apistructs.WORKSPACE_PROD,
					ConfigNamespace: "1-198-PROD-480",
				},
			},
		}, nil
	})

	got, err := order.fetchDeploymentParams(1, apistructs.WORKSPACE_PROD)
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
	got := New().parseAppsInfoWithDeployList(ProjectReleaseASum)
	assert.Equal(t, got, map[int64]string{1: "test"})
}

func TestParseAppsInfoWithRelease(t *testing.T) {
	order := New()
	got := order.parseAppsInfoWithDeployList([][]*pb.ApplicationReleaseSummary{
		{
			{ApplicationName: "test-1", ApplicationID: 1},
			{ApplicationName: "test-2", ApplicationID: 2},
		},
	})
	assert.Equal(t, got, map[int64]string{1: "test-1", 2: "test-2"})
}

func TestContinueDeployOrder(t *testing.T) {
	order := New()
	bdl := bundle.New()
	rt := runtime.NewRuntimeService()
	releaseSvc := &dicehubRelease.ReleaseService{}
	order.releaseSvc = releaseSvc

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	deployList := [][]string{{"0856df7931494d239abf07a145ade6e9"}}
	data, err := json.Marshal(deployList)
	if err != nil {
		t.Fatal(err)
	}

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "GetDeploymentOrder", func(*dbclient.DBClient, string) (*dbclient.DeploymentOrder, error) {
		return &dbclient.DeploymentOrder{
			ID:           "d9f06aaf-e3b7-4e05-9433-7742162e98f9",
			Modes:        DefaultMode,
			CurrentBatch: 1,
			BatchSize:    3,
			ProjectId:    1,
			Workspace:    apistructs.WORKSPACE_PROD,
			Params:       string(paramsJson),
			Type:         apistructs.TypeProjectRelease,
			DeployList:   string(data),
			ReleaseId:    "202cb962ac59075b964b07152d234b70",
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "UpdateDeploymentOrder", func(*dbclient.DBClient, *dbclient.DeploymentOrder) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(releaseSvc), "GetRelease",
		func(*dicehubRelease.ReleaseService, context.Context, *pb.ReleaseGetRequest) (*pb.ReleaseGetResponse, error) {
			return &pb.ReleaseGetResponse{Data: ProjectReleaseResp}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProjectWithSetter",
		func(*bundle.Bundle, uint64, ...httpclient.RequestSetter) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{ClusterConfig: map[string]string{"PROD": "fake-cluster"}}, nil
		},
	)
	monkey.PatchInstanceMethod(reflect.TypeOf(rt), "Create", func(*runtime.RuntimeService, user.ID, *apistructs.RuntimeCreateRequest) (*apistructs.DeploymentCreateResponseDTO, error) {
		return &apistructs.DeploymentCreateResponseDTO{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetAppIDByNames", func(*bundle.Bundle, uint64, string, []string) (*apistructs.GetAppIDByNamesResponseData, error) {
		return &apistructs.GetAppIDByNamesResponseData{
			AppNameToID: map[string]int64{
				"app-demo": 1,
			},
		}, nil
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

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProjectBranchRules", func(*bundle.Bundle, uint64) ([]*apistructs.BranchRule, error) {
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
	defer monkey.UnpatchAll()

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

func TestRenderDeployList(t *testing.T) {
	modes := map[string]*pb.ModeSummary{
		"A": {
			DependOn: []string{"X"},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "A11"},
					{ReleaseID: "A12"},
					{ReleaseID: "A13"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "A21"},
					{ReleaseID: "A22"},
					{ReleaseID: "A23"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "A31"},
					{ReleaseID: "A32"},
					{ReleaseID: "A33"},
				}},
			},
		},
		"B": {
			DependOn: []string{"X"},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "B11"},
					{ReleaseID: "B12"},
					{ReleaseID: "B13"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "B21"},
					{ReleaseID: "B22"},
					{ReleaseID: "B23"},
				}},
			},
		},
		"C": {
			DependOn: []string{"Y"},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "C11"},
					{ReleaseID: "C12"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "C21"},
					{ReleaseID: "C22"},
					{ReleaseID: "C23"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "C31"},
					{ReleaseID: "C32"},
					{ReleaseID: "C33"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "C41"},
				}},
			},
		},
		"X": {
			DependOn: []string{"Z"},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "X11"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "X21"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "X31"},
					{ReleaseID: "X32"},
				}},
			},
		},
		"Y": {
			DependOn: []string{"Z"},
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Y11"},
					{ReleaseID: "Y12"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Y21"},
				}},
			},
		},
		"Z": {
			ApplicationReleaseList: []*pb.ReleaseSummaryArray{
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Z11"},
					{ReleaseID: "Z12"},
					{ReleaseID: "Z13"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Z21"},
					{ReleaseID: "Z22"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Z31"},
				}},
				{List: []*pb.ApplicationReleaseSummary{
					{ReleaseID: "Z41"},
					{ReleaseID: "Z42"},
				}},
			},
		},
	}

	selectedModes := []string{"A", "B", "C"}

	expected := [][]string{
		{"Z11", "Z12", "Z13"},
		{"Z21", "Z22"},
		{"Z31"},
		{"Z41", "Z42"},
		{"X11", "Y11", "Y12"},
		{"X21", "Y21"},
		{"X31", "X32", "C11", "C12"},
		{"A11", "A12", "A13", "B11", "B12", "B13", "C21", "C22", "C23"},
		{"A21", "A22", "A23", "B21", "B22", "B23", "C31", "C32", "C33"},
		{"A31", "A32", "A33", "C41"},
	}
	deployList := renderDeployList(selectedModes, modes)

	if len(expected) != len(deployList) {
		t.Errorf("deploy list is not expected")
		return
	}
	for i, l := range deployList {
		if len(expected[i]) != len(l) {
			t.Errorf("deploy list is not expected")
			return
		}
		for j, summary := range l {
			if expected[i][j] != summary.ReleaseID {
				t.Errorf("deploy list is not expected")
				return
			}
		}
	}
}

func Test_renderDeployListWithCrossProject(t *testing.T) {
	order := New()

	type args struct {
		selectedModes []string
		releaseResp   *pb.ReleaseGetResponseData
	}

	tests := []struct {
		name    string
		args    args
		want    [][]*pb.ApplicationReleaseSummary
		wantErr bool
	}{
		{
			name: "application repeated",
			args: args{
				selectedModes: []string{"A", "A2"},
				releaseResp: &pb.ReleaseGetResponseData{
					Modes: map[string]*pb.ModeSummary{
						"A": {
							DependOn: []string{},
							ApplicationReleaseList: []*pb.ReleaseSummaryArray{
								{List: []*pb.ApplicationReleaseSummary{
									{ApplicationName: "A11", ReleaseID: "A11"},
									{ApplicationName: "A12", ReleaseID: "A12"},
								}},
								{List: []*pb.ApplicationReleaseSummary{
									{ApplicationName: "A21", ReleaseID: "A21"},
									{ApplicationName: "A22", ReleaseID: "A22"},
								}},
							},
						},
						"A2": {
							DependOn: []string{},
							ApplicationReleaseList: []*pb.ReleaseSummaryArray{
								{List: []*pb.ApplicationReleaseSummary{
									{ApplicationName: "A11", ReleaseID: "A11"},
								}},
								{List: []*pb.ApplicationReleaseSummary{
									{ApplicationName: "A21", ReleaseID: "A21"},
								}},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := order.renderDeployListWithCrossProject(tt.args.selectedModes, 1, "1", tt.args.releaseResp)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderDeployListWithCrossProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renderDeployListWithCrossProject() = %v, want %v", got, tt.want)
			}
		})
	}
}
