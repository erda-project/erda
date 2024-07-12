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

package project_report

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/erda-project/erda/apistructs"
)

func TestWrapBadRequest(t *testing.T) {
	p := &provider{}

	rw := httptest.NewRecorder()
	err := errors.New("Bad request")

	p.wrapBadRequest(rw, err)
	expectedBody := rw.Body.String()
	assert.Equal(t, expectedBody, rw.Body.String())
}

func TestGenLastValueWhereSql(t *testing.T) {
	req := &apistructs.ProjectReportRequest{
		ProjectIDs:   []uint64{1, 2, 3},
		IterationIDs: []uint64{10, 20, 30},
		LabelQuerys: []apistructs.ReportLabelOperation{
			{
				Key:       "project_name",
				Val:       "example",
				Operation: "like",
			},
			{
				Key:       "status",
				Val:       "done",
				Operation: "=",
			},
		},
	}

	expectedSQL := "AND tag_values[indexOf(tag_keys,'project_id')] IN ('1','2','3') " +
		"AND tag_values[indexOf(tag_keys,'iteration_id')] IN ('10','20','30') " +
		"AND (tag_values[indexOf(tag_keys,'project_name')] like '%example%' or tag_values[indexOf(tag_keys,'project_display_name')] like '%example%') " +
		"AND tag_values[indexOf(tag_keys,'status')] = 'done' "
	p := &provider{
		DB: &gorm.DB{
			Config: &gorm.Config{
				Dialector: &mysql.Dialector{},
			},
		},
	}

	result := p.genLastValueWhereSql(req)
	assert.Equal(t, expectedSQL, result)
}

func TestGenBasicWhereSql(t *testing.T) {
	req := &apistructs.ProjectReportRequest{
		Operations: []apistructs.ReportFilterOperation{
			{
				Key:       "requirementTotal",
				Val:       100,
				Operation: ">=",
			},
			{
				Key:       "bugCount",
				Val:       10,
				Operation: "<",
			},
		},
	}

	expectedSQL := "AND 'requirementTotal' >= 100.000000 AND 'bugCount' < 10.000000 "

	p := &provider{
		DB: &gorm.DB{
			Config: &gorm.Config{
				Dialector: &mysql.Dialector{},
			},
		},
	}
	result := p.genBasicWhereSql(req)
	assert.Equal(t, expectedSQL, result)
}

func Test_checkQueryRequest(t *testing.T) {
	cases := []struct {
		name    string
		arg     *apistructs.ProjectReportRequest
		wantErr bool
	}{
		{
			name: "missing org id",
			arg: &apistructs.ProjectReportRequest{
				Start: "now-1h",
				End:   "now",
			},
			wantErr: true,
		},
		{
			name: "missing start",
			arg: &apistructs.ProjectReportRequest{
				OrgID: 1,
				End:   "now",
			},
			wantErr: true,
		},
		{
			name: "missing end",
			arg: &apistructs.ProjectReportRequest{
				OrgID: 1,
				Start: "now-1h",
			},
			wantErr: true,
		},
		{
			name: "invalid operation",
			arg: &apistructs.ProjectReportRequest{
				OrgID: 1,
				Start: "now-1h",
				End:   "now",
				Operations: []apistructs.ReportFilterOperation{
					{
						Key:       "requirementTotal",
						Val:       100,
						Operation: "*",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid query",
			arg: &apistructs.ProjectReportRequest{
				OrgID: 1,
				Start: "now-1h",
				End:   "now",
				Operations: []apistructs.ReportFilterOperation{
					{
						Key:       "requirementTotal",
						Val:       100,
						Operation: ">=",
					},
				},
			},
		},
		{
			name: "invalid label operation",
			arg: &apistructs.ProjectReportRequest{
				OrgID: 1,
				Start: "now-1h",
				End:   "now",
				LabelQuerys: []apistructs.ReportLabelOperation{
					{
						Key:       "project_name",
						Val:       "100",
						Operation: ">=",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkQueryRequest(tc.arg); (err != nil) != tc.wantErr {
				t.Errorf("checkQueryRequest() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
