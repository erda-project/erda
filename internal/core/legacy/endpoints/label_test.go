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
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/legacy/services/permission"
)

func TestEndpoints_checkProjectPermission(t *testing.T) {
	e := Endpoints{}
	s := &permission.Permission{}
	e.permission = s
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "CheckPermission",
		func(s *permission.Permission, req *apistructs.PermissionCheckRequest) (bool, error) {
			// suppose we have project permission for: 1
			if req.ScopeID == 1 {
				return true, nil
			}
			return false, fmt.Errorf("no permission")
		},
	)

	// no projectID
	err := e.checkProjectPermission(apistructs.IdentityInfo{}, 0, "")
	assert.Error(t, err)

	// internal-client
	err = e.checkProjectPermission(apistructs.IdentityInfo{InternalClient: "test"}, 1, "")
	assert.NoError(t, err)

	// project: 1
	err = e.checkProjectPermission(apistructs.IdentityInfo{UserID: "1"}, 1, "mock")
	assert.NoError(t, err)

	// project: 2
	err = e.checkProjectPermission(apistructs.IdentityInfo{UserID: "1"}, 2, "mock")
	assert.Error(t, err)
}
