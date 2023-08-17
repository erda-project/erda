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
	"testing"

	"github.com/erda-project/erda/apistructs"
)

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
