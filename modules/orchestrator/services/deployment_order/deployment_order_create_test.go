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
	"github.com/erda-project/erda/pkg/http/httpclient"
)

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
			return &apistructs.ReleaseGetResponseData{
				Labels: map[string]string{gitBranchLabel: "master"},
			}, nil
		},
	)

	params := map[string]*apistructs.DeploymentOrderParam{}

	paramsJson, err := json.Marshal(params)
	assert.NoError(t, err)

	releaseResp := &apistructs.ReleaseGetResponseData{
		Labels: map[string]string{gitBranchLabel: "master"},
	}

	_, err = do.composeRuntimeCreateRequests(&dbclient.DeploymentOrder{
		Type:      apistructs.TypeApplicationRelease,
		Params:    string(paramsJson),
		Workspace: apistructs.WORKSPACE_PROD,
	}, releaseResp, apistructs.SourceDeployCenter, false)
	assert.NoError(t, err)

	releaseResp.ApplicationReleaseList = [][]*apistructs.ApplicationReleaseSummary{
		{
			{ReleaseID: "8781f475e5617a04"},
		},
	}

	_, err = do.composeRuntimeCreateRequests(&dbclient.DeploymentOrder{
		Type:      apistructs.TypeProjectRelease,
		Params:    string(paramsJson),
		Workspace: apistructs.WORKSPACE_PROD,
	}, releaseResp, apistructs.SourceDeployCenter, false)
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
