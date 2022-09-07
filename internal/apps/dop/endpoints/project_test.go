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
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/bdl"
	"github.com/erda-project/erda/internal/apps/dop/services/project"
	"github.com/erda-project/erda/internal/pkg/user"
)

func TestAddOnsFilterIn(t *testing.T) {
	addOns := []apistructs.AddonFetchResponseData{
		{
			ID:                  "1",
			PlatformServiceType: 1,
		},
		{
			ID:                  "2",
			PlatformServiceType: 0,
		},
		{
			ID:                  "3",
			PlatformServiceType: 1,
		},
	}
	newAddOns := addOnsFilterIn(addOns, func(addOn *apistructs.AddonFetchResponseData) bool {
		return addOn.PlatformServiceType == 0
	})
	assert.Equal(t, 1, len(newAddOns))
}

func TestExportProjectTemplate(t *testing.T) {
	pm1 := monkey.Patch(user.GetIdentityInfo, func(r *http.Request) (apistructs.IdentityInfo, error) {
		return apistructs.IdentityInfo{UserID: "1"}, nil
	})
	defer pm1.Unpatch()

	proSvc := &project.Project{}
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
	ep := Endpoints{
		bdl:                bdl.Bdl,
		project:            proSvc,
		queryStringDecoder: queryStringDecoder,
	}
	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(ep.bdl), "CheckPermission", func(b *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return nil, errors.Errorf("invalid permission")
	})
	defer pm3.Unpatch()

	r := http.Request{URL: &url.URL{}}
	_, err := ep.ExportProjectTemplate(context.Background(), &r, map[string]string{})
	assert.NoError(t, err)
}

func TestImportProjectTemplate(t *testing.T) {
	pm1 := monkey.Patch(user.GetIdentityInfo, func(r *http.Request) (apistructs.IdentityInfo, error) {
		return apistructs.IdentityInfo{UserID: "1"}, nil
	})
	defer pm1.Unpatch()

	proSvc := &project.Project{}
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
	ep := Endpoints{
		bdl:                bdl.Bdl,
		project:            proSvc,
		queryStringDecoder: queryStringDecoder,
	}
	pm3 := monkey.PatchInstanceMethod(reflect.TypeOf(ep.bdl), "CheckPermission", func(b *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return nil, errors.Errorf("invalid permission")
	})
	defer pm3.Unpatch()

	r := http.Request{URL: &url.URL{}}
	_, err := ep.ImportProjectTemplate(context.Background(), &r, map[string]string{})
	assert.NoError(t, err)
}

func Test_getProjectID(t *testing.T) {
	vars := map[string]string{
		"projectID": "1",
	}
	ep := &Endpoints{}
	projectID, err := ep.getProjectID(vars)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), projectID)
}

func Test_getOrgID(t *testing.T) {
	vars := map[string]string{
		"orgID": "1",
	}
	ep := &Endpoints{}
	orgID, err := ep.getOrgID(vars)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), orgID)
}

func Test_checkOrgIDAndProjectID(t *testing.T) {
	type args struct {
		orgIDStr     string
		projectIDStr string
	}
	tests := []struct {
		name          string
		args          args
		wantOrgID     uint64
		wantProjectID uint64
		wantHaveError bool
	}{
		{
			name:          "empty org id",
			args:          args{orgIDStr: ""},
			wantOrgID:     0,
			wantProjectID: 0,
			wantHaveError: true,
		},
		{
			name:          "empty project id",
			args:          args{orgIDStr: "1", projectIDStr: ""},
			wantOrgID:     1,
			wantProjectID: 0,
			wantHaveError: true,
		},
		{
			name:          "invalid org id",
			args:          args{orgIDStr: "s", projectIDStr: "1"},
			wantOrgID:     0,
			wantProjectID: 0,
			wantHaveError: true,
		},
		{
			name:          "invalid project id",
			args:          args{orgIDStr: "1", projectIDStr: "s"},
			wantOrgID:     1,
			wantProjectID: 0,
			wantHaveError: true,
		},
		{
			name:          "all right",
			args:          args{orgIDStr: "1", projectIDStr: "2"},
			wantOrgID:     1,
			wantProjectID: 2,
			wantHaveError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrgID, gotProjectID, err := checkOrgIDAndProjectID(tt.args.orgIDStr, tt.args.projectIDStr)
			if (err != nil) != tt.wantHaveError {
				t.Errorf("checkOrgIDAndProjectID() have error = %v, wantHaveError %v", err != nil, tt.wantHaveError)
				return
			}
			assert.Equalf(t, tt.wantOrgID, gotOrgID, "checkOrgIDAndProjectID(%v, %v)", tt.args.orgIDStr, tt.args.projectIDStr)
			assert.Equalf(t, tt.wantProjectID, gotProjectID, "checkOrgIDAndProjectID(%v, %v)", tt.args.orgIDStr, tt.args.projectIDStr)
		})
	}
}
