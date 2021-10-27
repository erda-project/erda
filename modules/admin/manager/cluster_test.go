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

package manager

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestRemoveSensitiveInfo(t *testing.T) {
	cluster := apistructs.ClusterInfo{
		Name: "fake-cluster",
		SchedConfig: &apistructs.ClusterSchedConfig{
			MasterURL:    "FakeMasterURL",
			AuthType:     "FakeAuthType",
			AuthUsername: "FakeAuthUsername",
			AuthPassword: "FakeAuthPassword",
			CACrt:        "FakeCACrt",
			ClientKey:    "FakeClientKey",
			ClientCrt:    "FakeClientCrt",
			AccessKey:    "FakeAccessKey",
			AccessSecret: "FakeAccessSecret",
		},
		OpsConfig: &apistructs.OpsConfig{
			AccessKey: "Fake AccessKey",
		},
		System: &apistructs.Sysconf{
			SSH: apistructs.SSH{
				User:     "FakeUser",
				Password: "FakePassword",
			},
			Storage: apistructs.Storage{
				MountPoint: "FakeMountPoint",
			},
		},
		ManageConfig: &apistructs.ManageConfig{
			CaData:           "FakeCaData",
			CredentialSource: "FakeCredentialSource",
		},
	}
	removeSensitiveInfo(&cluster)
	// remove assert
	assert.Equal(t, "", cluster.SchedConfig.AuthPassword)
	assert.Equal(t, (*apistructs.OpsConfig)(nil), cluster.OpsConfig)
	assert.Equal(t, "", cluster.System.SSH.Password)
	assert.Equal(t, "", cluster.ManageConfig.CaData)
	// keep assert
	assert.Equal(t, "FakeMasterURL", cluster.SchedConfig.MasterURL)
	assert.Equal(t, "FakeMountPoint", cluster.System.Storage.MountPoint)
	assert.Equal(t, "FakeCredentialSource", cluster.ManageConfig.CredentialSource)

}

func TestIsManager(t *testing.T) {
	var bdl *bundle.Bundle
	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ScopeRoleAccess", func(_ *bundle.Bundle, _ string, _ *apistructs.ScopeRoleAccessRequest) (*apistructs.ScopeRole, error) {
		return &apistructs.ScopeRole{Access: true,
			Roles: []string{"fake-role"},
		}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckIfRoleIsManager", func(_ *bundle.Bundle, _ string) bool {
		return true
	})

	err := IsManager(bdl, "", apistructs.OrgScope, "")
	assert.NoError(t, err)
}

func TestNonManager(t *testing.T) {
	var bdl *bundle.Bundle
	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ScopeRoleAccess", func(_ *bundle.Bundle, _ string, _ *apistructs.ScopeRoleAccessRequest) (*apistructs.ScopeRole, error) {
		return &apistructs.ScopeRole{Access: false,
			Roles: []string{"fake-role"},
		}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckIfRoleIsManager", func(_ *bundle.Bundle, _ string) bool {
		return false
	})

	err := IsManager(bdl, "", apistructs.OrgScope, "")
	assert.Error(t, err)
}

func TestOrgPermission(t *testing.T) {
	var bdl *bundle.Bundle
	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(_ *bundle.Bundle, _ *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{
			Access: true,
		}, nil
	})

	err := OrgPermCheck(bdl, "1", "2", "GET")
	assert.NoError(t, err)
}

func TestOrgPermissionFailed(t *testing.T) {
	var bdl *bundle.Bundle
	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(_ *bundle.Bundle, _ *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{
			Access: false,
		}, nil
	})

	err := OrgPermCheck(bdl, "1", "2", "GET")
	assert.Error(t, err)
}

func TestOrgPermissionCheckFailed(t *testing.T) {
	var bdl *bundle.Bundle
	// monkey patch Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(_ *bundle.Bundle, _ *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return nil, fmt.Errorf("socle role access failed")
	})

	err := OrgPermCheck(bdl, "1", "2", "GET")
	assert.Error(t, err)
}
