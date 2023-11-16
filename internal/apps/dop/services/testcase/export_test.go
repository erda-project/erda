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

package testcase

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	issuedao "github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func newStepAndResults() []apistructs.TestCaseStepAndResult {
	tcsar := make([]apistructs.TestCaseStepAndResult, 0)
	tcsar = append(tcsar, apistructs.TestCaseStepAndResult{
		Step:   "用户进入列表页面",
		Result: "列表页面成功加载",
	})
	return tcsar
}

func newTestCaseMetas() []apistructs.TestCasesMeta {
	tcms := make([]apistructs.TestCasesMeta, 0)
	tcms = append(tcms, apistructs.TestCasesMeta{
		Reqs: []apistructs.TestCaseCreateRequest{
			{
				ProjectID:      6665,
				TestSetID:      24452,
				Name:           "列表搜索",
				PreCondition:   "用户已登录并具有搜索列表的权限",
				StepAndResults: newStepAndResults(),
				APIs:           nil,
				Desc:           "Powered by AI.",
				Priority:       "P2",
				IdentityInfo:   apistructs.IdentityInfo{},
			},
		},
		RequirementName: "列表搜索",
		RequirementID:   30001128102,
	})
	return tcms
}

func TestService_ExportAIGeneratedx(t *testing.T) {
	type fields struct {
		db              *dao.DBClient
		bdl             *bundle.Bundle
		hc              *httpclient.HTTPClient
		org             org.Interface
		issueDBClient   *issuedao.DBClient
		pipelineSvc     pb.PipelineServiceServer
		CreateTestSetFn func(apistructs.TestSetCreateRequest) (*apistructs.TestSet, error)
	}
	type args struct {
		req apistructs.TestCaseExportRequest
	}

	var fileRecordID uint64 = 10001
	tcms := newTestCaseMetas()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				db: &dao.DBClient{
					DBEngine: &dbengine.DBEngine{},
				},
				bdl: bundle.New(),
			},
			args: args{
				req: apistructs.TestCaseExportRequest{
					TestCasePagingRequest: apistructs.TestCasePagingRequest{
						PageNo:    1,
						ProjectID: 6665,
						TestSetID: 24452,
						Recycled:  false,
						IdentityInfo: apistructs.IdentityInfo{
							UserID: "10000",
						},
					},
					FileType:          "xmind",
					Locale:            "",
					TestSetCasesMetas: tcms,
				},
			},
			want:    fileRecordID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:              tt.fields.db,
				bdl:             tt.fields.bdl,
				hc:              tt.fields.hc,
				org:             tt.fields.org,
				issueDBClient:   tt.fields.issueDBClient,
				pipelineSvc:     tt.fields.pipelineSvc,
				CreateTestSetFn: tt.fields.CreateTestSetFn,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(svc), "CreateFileRecord", func(_ *Service, req apistructs.TestFileRecordRequest) (uint64, error) {
				return fileRecordID, nil
			})
			defer monkey.UnpatchAll()

			got, err := svc.ExportAIGenerated(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportAIGenerated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExportAIGenerated() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_generateTestCasePagingResponseData(t *testing.T) {
	type fields struct {
		db              *dao.DBClient
		bdl             *bundle.Bundle
		hc              *httpclient.HTTPClient
		org             org.Interface
		issueDBClient   *issuedao.DBClient
		pipelineSvc     pb.PipelineServiceServer
		CreateTestSetFn func(apistructs.TestSetCreateRequest) (*apistructs.TestSet, error)
	}
	type args struct {
		req *apistructs.TestCaseExportRequest
	}

	tcs := make([]apistructs.TestCase, 0)
	recycled := false
	tcs = append(tcs, apistructs.TestCase{
		ID:             0,
		Name:           "列表搜索",
		Priority:       "P2",
		PreCondition:   "用户已登录并具有搜索列表的权限",
		Desc:           "Powered by AI.",
		Recycled:       &recycled,
		TestSetID:      24452,
		ProjectID:      6665,
		StepAndResults: newStepAndResults(),
	})

	tss := make([]apistructs.TestSetWithCases, 0)
	tss = append(tss, apistructs.TestSetWithCases{
		TestSetID: 24452,
		Recycled:  false,
		TestCases: tcs,
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apistructs.TestCasePagingResponseData
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				db: &dao.DBClient{
					DBEngine: &dbengine.DBEngine{},
				},
			},
			args: args{
				req: &apistructs.TestCaseExportRequest{
					TestCasePagingRequest: apistructs.TestCasePagingRequest{
						PageNo:       0,
						PageSize:     0,
						ProjectID:    0,
						TestSetID:    0,
						NoSubTestSet: false,
						IdentityInfo: apistructs.IdentityInfo{
							UserID: "10000",
						},
					},
					TestSetCasesMetas: newTestCaseMetas(),
				},
			},
			want: &apistructs.TestCasePagingResponseData{
				Total:    1,
				TestSets: tss,
				UserIDs:  []string{"10000"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:              tt.fields.db,
				bdl:             tt.fields.bdl,
				hc:              tt.fields.hc,
				org:             tt.fields.org,
				issueDBClient:   tt.fields.issueDBClient,
				pipelineSvc:     tt.fields.pipelineSvc,
				CreateTestSetFn: tt.fields.CreateTestSetFn,
			}
			monkey.PatchInstanceMethod(reflect.TypeOf(svc.db), "GetTestSetByID", func(_ *dao.DBClient, id uint64) (*dao.TestSet, error) {
				return &dao.TestSet{
					BaseModel: dbengine.BaseModel{},
					Name:      "",
					ParentID:  0,
					Recycled:  false,
					ProjectID: 0,
					Directory: "/AI_Generated",
					OrderNum:  0,
					CreatorID: "",
					UpdaterID: "",
				}, nil
			})
			defer monkey.UnpatchAll()

			got, err := svc.generateTestCasePagingResponseData(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateTestCasePagingResponseData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateTestCasePagingResponseData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
