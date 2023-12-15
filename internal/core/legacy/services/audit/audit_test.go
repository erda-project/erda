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

package audit

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/legacy/model"
)

type mockTran struct {
	i18n.Translator
}

func (t mockTran) Get(lang i18n.LanguageCodes, key, def string) string {
	return key
}

func (t mockTran) Text(lang i18n.LanguageCodes, key string) string {
	return errMap[key]
}

func (t mockTran) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return fmt.Sprintf(key, args...)
}

var errMap = map[string]string{
	ErrInvalidOrg:          "ErrInvalidOrg",
	ErrInvalidAppInOrg:     "ErrInvalidAppInOrg",
	ErrInvalidProjectInOrg: "ErrInvalidProjectInOrg",
	ErrInvalidAppInProject: "ErrInvalidAppInProject",
}

func TestGetAllProjectIdInOrg(t *testing.T) {
	// Create a new instance of Audit
	audit := &Audit{}

	mockProjectMap := make(map[uint64][]model.Project)

	mockProjectMap[1] = []model.Project{
		{BaseModel: model.BaseModel{ID: 1}},
		{BaseModel: model.BaseModel{ID: 2}},
		{BaseModel: model.BaseModel{ID: 3}},
		{BaseModel: model.BaseModel{ID: 4}},
	}

	mockProjectMap[2] = []model.Project{
		{BaseModel: model.BaseModel{ID: 5}},
		{BaseModel: model.BaseModel{ID: 6}},
		{BaseModel: model.BaseModel{ID: 7}},
	}

	// Replace the ListProjectByOrgID method with a mock implementation
	monkey.PatchInstanceMethod(reflect.TypeOf(audit.db), "ListProjectByOrgID", func(_ *dao.DBClient, orgId uint64) ([]model.Project, error) {
		return mockProjectMap[orgId], nil
	})
	defer monkey.UnpatchAll()

	// Call the GetAllProjectIdInOrg method
	projectIds, err := audit.GetAllProjectIdInOrg(1)

	// Perform assertions on the returned values
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedProjectIds := []uint64{1, 2, 3, 4}
	if !reflect.DeepEqual(projectIds, expectedProjectIds) {
		t.Errorf("Expected project IDs %v, but got %v", expectedProjectIds, projectIds)
	}
}

func Test_constructFilterParamByReq(t *testing.T) {
	audit := &Audit{
		trans: mockTran{},
	}

	type response struct {
		filterParam *model.ListAuditParam
		err         error
	}

	// orgID - projectID
	mockProjectMap := make(map[uint64][]model.Project)

	mockProjectMap[1] = []model.Project{
		{BaseModel: model.BaseModel{ID: 1}},
		{BaseModel: model.BaseModel{ID: 2}},
		{BaseModel: model.BaseModel{ID: 3}},
		{BaseModel: model.BaseModel{ID: 4}},
	}

	mockProjectMap[2] = []model.Project{
		{BaseModel: model.BaseModel{ID: 5}},
		{BaseModel: model.BaseModel{ID: 6}},
		{BaseModel: model.BaseModel{ID: 7}},
	}

	// orgID - appID
	mockOrgAppMap := make(map[uint64][]model.Application)
	mockOrgAppMap[1] = []model.Application{
		{BaseModel: model.BaseModel{ID: 1}},
		{BaseModel: model.BaseModel{ID: 2}},
	}

	mockOrgAppMap[2] = []model.Application{
		{BaseModel: model.BaseModel{ID: 3}},
		{BaseModel: model.BaseModel{ID: 4}},
	}

	// projectID - appID
	mockProjectAppMap := make(map[uint64][]model.Application)
	mockProjectAppMap[1] = []model.Application{
		{BaseModel: model.BaseModel{ID: 1}},
	}

	mockProjectAppMap[5] = []model.Application{
		{BaseModel: model.BaseModel{ID: 3}},
		{BaseModel: model.BaseModel{ID: 4}},
	}

	// Replace the ListProjectByOrgID method with a mock implementation
	monkey.PatchInstanceMethod(reflect.TypeOf(audit.db), "ListProjectByOrgID", func(_ *dao.DBClient, orgId uint64) ([]model.Project, error) {
		return mockProjectMap[orgId], nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(audit.db), "GetApplicationsByProjectIDs", func(_ *dao.DBClient, projectIDs []uint64) ([]model.Application, error) {
		appIds := make([]model.Application, 0)
		for _, id := range projectIDs {
			if _, ok := mockProjectAppMap[id]; ok {
				appIds = append(appIds, mockProjectAppMap[id]...)
			}
		}
		return appIds, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(audit.db), "GetApplicationsByOrgId", func(_ *dao.DBClient, orgID uint64) ([]model.Application, error) {
		return mockOrgAppMap[orgID], nil
	})

	defer monkey.UnpatchAll()

	testcase := []struct {
		Name  string
		param *apistructs.AuditsListRequest
		want  response
	}{
		{
			Name: "invalid orgID,orgID is nil",
			param: &apistructs.AuditsListRequest{
				Sys:   false,
				OrgID: nil,
				AppID: []uint64{1, 2},
			},
			want: response{
				filterParam: nil,
				err:         errors.New(ErrInvalidOrg),
			},
		},
		{
			Name: "multiple orgID",
			param: &apistructs.AuditsListRequest{
				Sys:   false,
				OrgID: []uint64{1, 2, 3},
			},
			want: response{
				filterParam: nil,
				err:         errors.New(ErrInvalidOrg),
			},
		},
		{
			Name: "invalid project in org",
			param: &apistructs.AuditsListRequest{
				Sys:       false,
				OrgID:     []uint64{1},
				ProjectID: []uint64{1, 5},
			},
			want: response{
				filterParam: nil,
				err:         errors.New(ErrInvalidProjectInOrg),
			},
		},
		{
			Name: "invalid app in org",
			param: &apistructs.AuditsListRequest{
				Sys:       false,
				OrgID:     []uint64{1},
				ProjectID: []uint64{},
				AppID:     []uint64{5},
			},
			want: response{
				filterParam: nil,
				err:         errors.New(ErrInvalidAppInOrg),
			},
		},
		{
			Name: "invalid app in project",
			param: &apistructs.AuditsListRequest{
				Sys:       false,
				OrgID:     []uint64{1},
				ProjectID: []uint64{1, 2},
				AppID:     []uint64{2},
			},
			want: response{
				filterParam: nil,
				err:         errors.New(ErrInvalidAppInProject),
			},
		},
		{
			Name: "success",
			param: &apistructs.AuditsListRequest{
				Sys:       false,
				OrgID:     []uint64{1},
				ProjectID: []uint64{1, 2},
				AppID:     []uint64{1},
			},
			want: response{
				filterParam: &model.ListAuditParam{
					OrgID:     []uint64{1},
					ProjectID: []uint64{1, 2},
					AppID:     []uint64{1},
				},
				err: nil,
			},
		},
	}

	for _, tt := range testcase {
		t.Run(tt.Name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "lang_codes", i18n.LanguageCodes{})
			filterParam, err := audit.constructFilterParamByReq(ctx, tt.param)
			assert.Equal(t, filterParam, tt.want.filterParam)
			if err != nil {
				assert.Equal(t, err.Error(), tt.want.err.Error())
			} else {
				assert.Equal(t, err, tt.want.err)
			}
		})
	}
}
