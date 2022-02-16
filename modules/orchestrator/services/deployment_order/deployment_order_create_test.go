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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var ProjectReleaseResp = &apistructs.ReleaseGetResponseData{
	IsProjectRelease: true,
	Labels:           map[string]string{gitBranchLabel: "master"},
	ApplicationReleaseList: [][]*apistructs.ApplicationReleaseSummary{
		{
			{
				ReleaseID:   "0856df7931494d239abf07a145ade6e9",
				ReleaseName: "release/1.0.1",
				Version:     "1.0.1+20220210153458",
			},
		},
	},
}

var AppReleaseResp = &apistructs.ReleaseGetResponseData{
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

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetRelease",
		func(*bundle.Bundle, string) (*apistructs.ReleaseGetResponseData, error) {
			return ProjectReleaseResp, nil
		},
	)

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	ProjectReleaseResp.ApplicationReleaseList = [][]*apistructs.ApplicationReleaseSummary{
		{
			{ReleaseID: "8781f475e5617a04"},
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
	order := New()
	bdl := bundle.New()

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "FetchDeploymentConfigDetail",
		func(*bundle.Bundle, string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
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
	got, err := order.parseAppsInfoWithOrder(&dbclient.DeploymentOrder{
		ApplicationName: "test",
		ApplicationId:   1,
		Type:            apistructs.TypeApplicationRelease,
	})
	assert.NoError(t, err)
	assert.Equal(t, got, map[int64]string{1: "test"})
}

func TestParseAppsInfoWithRelease(t *testing.T) {
	order := New()
	got := order.parseAppsInfoWithRelease(&apistructs.ReleaseGetResponseData{
		IsProjectRelease: true,
		ApplicationReleaseList: [][]*apistructs.ApplicationReleaseSummary{
			{
				{ApplicationName: "test-1", ApplicationID: 1},
				{ApplicationName: "test-2", ApplicationID: 2},
			},
		},
	})
	assert.Equal(t, got, map[int64]string{1: "test-1", 2: "test-2"})
}

func TestContinueDeployOrder(t *testing.T) {
	order := New()
	bdl := bundle.New()
	rt := runtime.New()

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
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetRelease",
		func(*bundle.Bundle, string) (*apistructs.ReleaseGetResponseData, error) {
			return ProjectReleaseResp, nil
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
