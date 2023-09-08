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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	aitestcase "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type FunctionCallArguments struct {
	Name           string                             `json:"name"`           // 用例名称
	PreCondition   string                             `json:"preCondition"`   // 前置条件
	StepAndResults []apistructs.TestCaseStepAndResult `json:"stepAndResults"` // 步骤及结果
}

func TestAIFunction_createTestCaseForRequirementIDAndTestID(t *testing.T) {
	type fields struct {
		Log       logs.Logger
		OpenaiURL *url.URL
	}
	type args struct {
		ctx       context.Context
		factory   functions.FunctionFactory
		req       *pb.ApplyRequest
		openaiURL *url.URL
	}

	url, _ := url.Parse("http://ai-proxy:8081")
	reuirements := make([]aitestcase.TestCaseParam, 0)
	reuirements = append(reuirements, aitestcase.TestCaseParam{
		IssueID: 30001127564,
		Prompt:  "用户登录",
	})

	params := aitestcase.FunctionParams{
		TestSetID:    24227,
		Requirements: reuirements,
	}
	jsonData, _ := json.Marshal(params)
	value := structpb.Value{}
	jsonpb.UnmarshalString(string(jsonData), &value)

	req := &pb.ApplyRequest{
		FunctionName:   aitestcase.Name,
		FunctionParams: &value,
		Background: &pb.Background{
			UserID:      "1003933",
			OrgID:       633,
			OrgName:     "erda-development",
			ProjectID:   4717,
			ProjectName: "testhpa",
		},
		NeedAdjust: false,
	}
	factory, _ := functions.Retrieve(req.GetFunctionName())
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
	testCaseReq := apistructs.TestCaseCreateRequest{
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
	var testcaseId uint64
	testcaseId = 113523
	f := factory(nil, "", req.GetBackground())
	want := []aitestcase.TestCaseMeta{
		{
			Req:             testCaseReq,
			RequirementName: issue.Title,
			RequirementID:   uint64(issue.ID),
			TestCaseID:      testcaseId,
		},
	}
	content := httpserver.Resp{
		Success: true,
		Data:    want,
	}
	json.Marshal(content)

	wantData, _ := json.Marshal(content)

	fca := FunctionCallArguments{
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
	}
	bb, _ := json.Marshal(fca)
	arguments := fmt.Sprintf("%s", string(bb))

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
				Log:       nil,
				OpenaiURL: url,
			},
			args: args{
				ctx:       nil,
				factory:   factory,
				req:       req,
				openaiURL: url,
			},
			want:    wantData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AIFunction{
				Log:       tt.fields.Log,
				OpenaiURL: tt.fields.OpenaiURL,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(&sdk.Client{}), "CreateCompletion", func(_ *sdk.Client,
				ctx context.Context, req *sdk.CreateCompletionOptions) (*sdk.ChatCompletions, error) {
				choices := make([]*sdk.ChatChoice, 0)
				choices = append(choices, &sdk.ChatChoice{
					Index: 0,
					Message: &sdk.ChatMessage{
						FunctionCall: &sdk.FunctionCall{
							Arguments: arguments,
						},
					},
				})

				return &sdk.ChatCompletions{
					Choices: choices,
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(f), "Callback", func(_ *aitestcase.Function,
				ctx context.Context, arguments json.RawMessage, input interface{}, needAdjust bool) (any, error) {
				return aitestcase.TestCaseMeta{
					Req:             testCaseReq,
					RequirementName: issue.Title,
					RequirementID:   uint64(issue.ID),
					TestCaseID:      testcaseId,
				}, nil
			})

			defer monkey.UnpatchAll()

			got, err := h.createTestCaseForRequirementIDAndTestID(tt.args.ctx, tt.args.factory, tt.args.req, tt.args.openaiURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTestCaseForRequirementIDAndTestID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTestCaseForRequirementIDAndTestID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateParamsForCreateTestcase(t *testing.T) {
	type args struct {
		req aitestcase.FunctionParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				req: aitestcase.FunctionParams{
					TestSetID: 24227,
					Requirements: []aitestcase.TestCaseParam{
						{
							IssueID: 30001127564,
							Prompt:  "用户登录",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateParamsForCreateTestcase(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("validateParamsForCreateTestcase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
