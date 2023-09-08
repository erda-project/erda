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

package test_case

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestFunction_Callback(t *testing.T) {
	type fields struct {
		background *pb.Background
	}
	type args struct {
		ctx        context.Context
		arguments  json.RawMessage
		input      interface{}
		needAdjust bool
	}

	var testcaseId uint64
	testcaseId = 113523

	input := TestCaseFunctionInput{
		TestSetID: 24227,
		IssueID:   30001127564,
		Prompt:    "用户登录",
	}

	results := []apistructs.TestCaseStepAndResult{
		{
			Result: "登录成功，跳转到用户个人主页",
			Step:   "用户输入有效的用户名和密码",
		},
		{
			Result: "下次登录自动填充用户名和密码",
			Step:   "用户输入记住密码",
		},
		{
			Result: "跳转到密码重置页面",
			Step:   "用户点击忘记密码",
		},
		{
			Result: "跳转到用户注册页面",
			Step:   "用户点击注册",
		},
	}

	tc := apistructs.TestCase{
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
	}
	arguments, _ := json.Marshal(tc)
	req := apistructs.TestCaseCreateRequest{
		ProjectID:      4717,
		TestSetID:      24227,
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
		APIs:           nil,
		Desc:           "Powered by AI.\n\n对应需求:\n用户登录",
		Priority:       "P2",
		LabelIDs:       nil,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: "1003933",
		},
	}

	issue := &apistructs.Issue{
		RequirementID:    nil,
		RequirementTitle: "",
		Type:             apistructs.IssueTypeRequirement,
		Title:            "用户登录",
		Content:          "",
		Priority:         apistructs.IssuePriorityNormal,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				background: &pb.Background{
					UserID:      "1003933",
					OrgID:       633,
					OrgName:     "erda-development",
					ProjectID:   4717,
					ProjectName: "testhpa",
				},
			},
			args: args{
				arguments:  arguments,
				input:      input,
				needAdjust: true,
			},
			want: TestCaseMeta{
				Req:             req,
				RequirementName: issue.Title,
				RequirementID:   uint64(issue.ID),
			},
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				background: &pb.Background{
					UserID:      "1003933",
					OrgID:       633,
					OrgName:     "erda-development",
					ProjectID:   4717,
					ProjectName: "testhpa",
				},
			},
			args: args{
				arguments:  arguments,
				input:      input,
				needAdjust: false,
			},
			want: TestCaseMeta{
				Req:             req,
				RequirementName: issue.Title,
				RequirementID:   uint64(issue.ID),
				TestCaseID:      testcaseId,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Function{
				background: tt.fields.background,
			}

			bdl := bundle.New(bundle.WithErdaServer())
			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetIssue", func(_ *bundle.Bundle,
				id uint64) (*apistructs.Issue, error) {
				return &apistructs.Issue{
					RequirementID:    nil,
					RequirementTitle: "",
					Type:             apistructs.IssueTypeRequirement,
					Title:            "用户登录",
					Content:          "",
					Priority:         apistructs.IssuePriorityNormal,
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateTestCase", func(_ *bundle.Bundle,
				req apistructs.TestCaseCreateRequest) (apistructs.AICreateTestCaseResponse, error) {
				return apistructs.AICreateTestCaseResponse{
					TestCaseID: testcaseId,
				}, nil
			})

			defer monkey.UnpatchAll()

			got, err := f.Callback(tt.args.ctx, tt.args.arguments, tt.args.input, tt.args.needAdjust)
			if (err != nil) != tt.wantErr {
				t.Errorf("Callback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Callback() got = %v, want %v", got, tt.want)
			}
		})
	}
}
