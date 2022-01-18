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
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func getFakeErdaYaml() []byte {
	return []byte(`
addons:
  fake-custom:
    plan: custom:basic
  fake-addon:
    plan: hello-world
envs: {}
jobs: {}
services:
  go-demo:
    binds: []
    deployments:
      replicas: 1
    envs: {}
    expose: []
    health_check: {}
    hosts: []
    image: registry.erda.io/erda-demo-erda-demo/go-web:go-demo-1642419767112914097
    ports:
    - 8080
    resources:
      cpu: 1
      mem: 1024
version: "2.0"

`)
}

func TestPreCheck(t *testing.T) {
	order := New()
	bdl := bundle.New()

	defer monkey.UnpatchAll()

	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "ListCustomInstancesByProjectAndEnv",
		func(*dbclient.DBClient, int64, string) ([]dbclient.AddonInstance, error) {
			return []dbclient.AddonInstance{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission",
		func(*bundle.Bundle, *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
			return &apistructs.PermissionCheckResponseData{Access: true}, nil
		},
	)

	got, err := order.preCheck("1", string(apistructs.WorkspaceDev), 1, 1, getFakeErdaYaml())
	assert.NoError(t, err)
	assert.Equal(t, got.Success, false)
	assert.Equal(t, len(got.FailReasons), 2)
}

func TestRenderDetail(t *testing.T) {
	order := New()
	bdl := bundle.New()

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetRelease", func(*bundle.Bundle, string) (*apistructs.ReleaseGetResponseData, error) {
		return &apistructs.ReleaseGetResponseData{
			Diceyml: string(getFakeErdaYaml()),
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(*bundle.Bundle, *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{
			Access: true}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "FetchDeploymentConfigDetail", func(*bundle.Bundle, string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
		return nil, nil, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "ListCustomInstancesByProjectAndEnv",
		func(*dbclient.DBClient, int64, string) ([]dbclient.AddonInstance, error) {
			return []dbclient.AddonInstance{}, nil
		},
	)

	_, err := order.RenderDetail("1", "dd11727fc60945c998c2fcdf6487e9b0", "PROD")
	assert.NoError(t, err)
}
