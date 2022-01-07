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
		Type:   apistructs.TypePipeline,
		Params: string(paramsJson),
	}, releaseResp, "PROD")
	assert.NoError(t, err)

	releaseResp.ApplicationReleaseList = []*apistructs.ApplicationReleaseSummary{
		{ReleaseID: "8781f475e5617a04"},
	}

	_, err = do.composeRuntimeCreateRequests(&dbclient.DeploymentOrder{
		Type:   apistructs.TypeProjectRelease,
		Params: string(paramsJson),
	}, releaseResp, "PROD")
	assert.NoError(t, err)
}

func TestFetchDeploymentOrderParam(t *testing.T) {
	order := New()
	bdl := bundle.New()

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "FetchDeploymentConfigDetail",
		func(*bundle.Bundle, string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
			return []apistructs.EnvConfig{{Key: "key1", Value: "value1", ConfigType: "ENV"}},
				[]apistructs.EnvConfig{{Key: "key2", Value: "value2", ConfigType: "FILE", Encrypt: true}}, nil
		},
	)

	got, err := order.fetchDeploymentOrderParam(1, "STAGING")
	assert.NoError(t, err)
	assert.Equal(t, got.Env[0].Key, "key1")
	assert.Equal(t, got.File[0].IsEncrypt, true)
}
