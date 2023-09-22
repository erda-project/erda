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
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
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

	params1 := params
	params1.TestSetID = 0
	reuirements1 := make([]aitestcase.TestCaseParam, 0)
	reuirements1 = append(reuirements1, aitestcase.TestCaseParam{
		IssueID: 30001127564,
		Prompt:  "用户登录",
		Req: apistructs.TestCaseCreateRequest{
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
		},
	})

	params1.Requirements = reuirements1
	jsonData, _ = json.Marshal(params1)
	value1 := structpb.Value{}
	jsonpb.UnmarshalString(string(jsonData), &value1)

	req1 := &pb.ApplyRequest{
		FunctionName:   aitestcase.Name,
		FunctionParams: &value1,
		Background: &pb.Background{
			UserID:      "1003933",
			OrgID:       633,
			OrgName:     "erda-development",
			ProjectID:   4717,
			ProjectName: "testhpa",
		},
		NeedAdjust: false,
	}

	testCaseReq := apistructs.TestCaseCreateRequest{
		ProjectID:      4717,
		TestSetID:      24227,
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
		APIs:           nil,
		Desc:           fmt.Sprintf("Powered by AI.\n\n对应需求:\n%s", "用户登录"),
		Priority:       "P2",
		LabelIDs:       nil,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: "1003933",
		},
	}

	var requirementId int64
	requirementId = 30001127564
	issue := &apistructs.Issue{
		ID: requirementId,
		//RequirementID:    nil,
		RequirementID:    &requirementId,
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
				ctx:       context.Background(),
				factory:   factory,
				req:       req,
				openaiURL: url,
			},
			want:    wantData,
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				Log:       nil,
				OpenaiURL: url,
			},
			args: args{
				ctx:       context.Background(),
				factory:   factory,
				req:       req1,
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
				ctx context.Context, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
				choices := make([]openai.ChatCompletionChoice, 0)
				choices = append(choices, openai.ChatCompletionChoice{
					Index: 0,
					Message: openai.ChatCompletionMessage{
						FunctionCall: &openai.FunctionCall{
							Arguments: arguments,
						},
					},
				})

				return &openai.ChatCompletionResponse{
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

			bdl := bundle.New(bundle.WithErdaServer())
			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetTestSets", func(_ *bundle.Bundle,
				req apistructs.TestSetListRequest) ([]apistructs.TestSet, error) {

				if tt.name == "Test_02" {
					return []apistructs.TestSet{}, nil
				}

				return []apistructs.TestSet{
					{
						ID:        10001,
						Name:      AIGeneratedTestSeName,
						ProjectID: 4717,
						ParentID:  0,
						Recycled:  false,
						Directory: "/" + AIGeneratedTestSeName,
						Order:     0,
					},
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateTestSet", func(_ *bundle.Bundle,
				req apistructs.TestSetCreateRequest) (*apistructs.TestSet, error) {
				return &apistructs.TestSet{
					ID:        24227,
					Name:      "",
					ProjectID: 0,
					ParentID:  0,
					Recycled:  false,
					Directory: "",
					Order:     0,
					CreatorID: "",
					UpdaterID: "",
				}, nil
			})

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
			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetCurrentUser", func(_ *bundle.Bundle,
				userID string) (*apistructs.UserInfo, error) {
				return &apistructs.UserInfo{
					ID:     "1003933",
					Name:   "xxx",
					Nick:   "yyy",
					Avatar: "",
					Phone:  "10111112222",
					Email:  "xxx.yyy@123.com",
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateTestCase", func(_ *bundle.Bundle,
				req apistructs.TestCaseCreateRequest) (apistructs.AICreateTestCaseResponse, error) {
				return apistructs.AICreateTestCaseResponse{
					TestCaseID: testcaseId,
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

func Test_tuningPrompt(t *testing.T) {
	type args struct {
		title   string
		content string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test_01",
			args: args{
				title:   "注册需求",
				content: "注册需求,### 用户故事/要解决的问题*\n功能需求：\n用户访问特定网站或应用程序，并点击注册按钮。\n系统显示一个注册页面，用户在用户名和密码的相应文本框中输入准确的凭据，其中用户名为英文字母，不能是中文，不能超过20个字符；密码是8-16为的大写字母、小写字母、数字组成，密码不能是包括空格，但是能包括@和$符号\n用户点击“注册”按钮或按下回车键来提交注册请求。\n如果提供的凭据是错误的，系统可能会显示错误消息提示用户重新输入准确的信息。\n安全需求��\n密码框应该显示*，密码不能够在页面源码模式下被查看，可以支持复制粘贴，\n在不登录系统的情况下，直接输入后台页面地址不能访问页面\n性能需求：\n在单用户注册场景中，系统响应时间应该小于 3 秒；在高并发场景下，用户注册的响应时间不超过 5 秒\n兼容性需求：\n不同移动设备终端和不同浏览器页面的显示以及功能正确性应该一致\n\n### 意向用户*\n\n\n### 用户体验目标*\n\n\n### 链接/参考\n\n\n",
			},
			want: "需求名称： 注册需求\n需求描述：\n1. 用户访问特定网站或应用程序，并点击注册按钮。\n2. 系统显示一个注册页面，用户在用户名和密码的相应文本框中输入准确的凭据，其中用户名为英文字母，不能是中文，不能超过20个字符；密码是8-16为的大写字母、小写字母、数字组成，密码不能是包括空格，但是能包括@和$符号\n3. 用户点击“注册”按钮或按下回车键来提交注册请求。\n4. 如果提供的凭据是错误的，系统可能会显示错误消息提示用户重新输入准确的信息。\n5. 密码框应该显示*，密码不能够在页面源码模式下被查看，可以支持复制粘贴，\n6. 在不登录系统的情况下，直接输入后台页面地址不能访问页面\n7. 在单用户注册场景中，系统响应时间应该小于 3 秒；在高并发场景下，用户注册的响应时间不超过 5 秒\n8. 不同移动设备终端和不同浏览器页面的显示以及功能正确性应该一致",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tuningPrompt(tt.args.title, tt.args.content); got != tt.want {
				t.Errorf("tuningPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
