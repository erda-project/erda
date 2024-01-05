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
	"fmt"
	"reflect"
	"testing"
	"time"

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
		ctx       context.Context
		arguments json.RawMessage
		input     interface{}
	}

	var testcaseId, testSetId uint64
	testcaseId = 113523
	//testSetId = 25111

	input := TestCaseFunctionInput{
		TestSetID: 24227,
		IssueID:   30001127564,
		Prompt:    "用户登录",
		Name:      "groupName",
		UserId:    "1003933",
		ProjectId: 4717,
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

	list := make([]apistructs.TestCaseCreateRequest, 0)
	list = append(list, apistructs.TestCaseCreateRequest{
		TestCaseID:     0,
		ProjectID:      0,
		TestSetID:      0,
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
		APIs:           nil,
		Desc:           "",
		Priority:       "",
		LabelIDs:       nil,
		IdentityInfo:   apistructs.IdentityInfo{},
	})
	tcrl := TestCaseCreateRequestList{
		List: list,
	}

	arguments, _ := json.Marshal(tcrl)
	reqs := make([]apistructs.TestCaseCreateRequest, 0)
	//reqs2 := make([]apistructs.TestCaseCreateRequest, 0)
	req := apistructs.TestCaseCreateRequest{
		ProjectID:      4717,
		TestSetID:      testSetId,
		TestSetDir:     "/groupName",
		TestSetName:    "groupName",
		Name:           "用户登录",
		PreCondition:   "用户打开登录页面",
		StepAndResults: results,
		APIs:           nil,
		Desc:           "Powered by AI.",
		Priority:       "P2",
		LabelIDs:       nil,
		IdentityInfo: apistructs.IdentityInfo{
			UserID: "1003933",
		},
	}
	reqs = append(reqs, req)
	//req.TestCaseID = testcaseId
	//reqs2 = append(reqs2, req)
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
				arguments: arguments,
				input:     input,
			},
			want: apistructs.TestCasesMeta{
				Reqs:            reqs,
				RequirementName: issue.Title,
				RequirementID:   uint64(issue.ID),
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

			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CreateTestSet", func(_ *bundle.Bundle,
				req apistructs.TestSetCreateRequest) (*apistructs.TestSet, error) {
				return &apistructs.TestSet{
					ID: testSetId,
				}, nil
			})

			defer monkey.UnpatchAll()

			got, err := f.Callback(tt.args.ctx, tt.args.arguments, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Callback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var fakeId uint64
			fakeId = uint64(time.Now().Nanosecond())
			gotStruct, _ := got.(apistructs.TestCasesMeta)
			for idx := range gotStruct.Reqs {
				gotStruct.Reqs[idx].TestSetID = fakeId
			}
			wantStruct, _ := tt.want.(apistructs.TestCasesMeta)
			for idx := range wantStruct.Reqs {
				wantStruct.Reqs[idx].TestSetID = fakeId
			}

			if !reflect.DeepEqual(gotStruct, wantStruct) {
				t.Errorf("Callback() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateParamsForCreateTestcase(t *testing.T) {
	type args struct {
		req FunctionParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				req: FunctionParams{
					TestSetID: 24227,
					Requirements: []TestCaseParam{
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

func Test_createRootTestSetIfNecessary(t *testing.T) {
	type args struct {
		fps       *FunctionParams
		bdl       *bundle.Bundle
		userId    string
		projectId uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				fps: &FunctionParams{
					TestSetID:    0,
					SystemPrompt: "xxx",
					Requirements: []TestCaseParam{
						{
							IssueID: 30001127564,
							Prompt:  "用户登录",
						},
					},
				},
				bdl:       bundle.New(),
				userId:    "10001",
				projectId: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.bdl), "GetTestSets", func(_ *bundle.Bundle,
				req apistructs.TestSetListRequest) ([]apistructs.TestSet, error) {
				return []apistructs.TestSet{
					{
						ID:        1,
						Name:      "",
						ProjectID: 1,
						ParentID:  0,
						Recycled:  false,
						Directory: "/",
						Order:     0,
						CreatorID: "10001",
						UpdaterID: "10001",
					},
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.bdl), "CreateTestSet", func(_ *bundle.Bundle,
				req apistructs.TestSetCreateRequest) (*apistructs.TestSet, error) {
				return &apistructs.TestSet{
					ID:        2,
					Name:      AIGeneratedTestSetName,
					ProjectID: 1,
					ParentID:  1,
					Recycled:  false,
					Directory: "/" + AIGeneratedTestSetName,
					Order:     1,
					CreatorID: "10001",
					UpdaterID: "10001",
				}, nil
			})

			defer monkey.UnpatchAll()

			if err := createRootTestSetIfNecessary(tt.args.fps, tt.args.bdl, tt.args.userId, tt.args.projectId); (err != nil) != tt.wantErr {
				t.Errorf("createRootTestSetIfNecessary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getOperationType(t *testing.T) {
	type args struct {
		fps FunctionParams
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test_01",
			args: args{
				fps: FunctionParams{
					TestSetID:    0,
					SystemPrompt: "xxx",
					Requirements: []TestCaseParam{
						{
							IssueID: 30001127564,
							Prompt:  "用户登录",
						},
					},
				},
			},
			want: OperationTypeGenerate,
		},
		{
			name: "Test_02",
			args: args{
				fps: FunctionParams{
					TestSetID:    0,
					SystemPrompt: "xxx",
					Requirements: []TestCaseParam{
						{
							IssueID: 30001127564,
							Prompt:  "用户登录",
							Reqs: []apistructs.TestCaseCreateRequest{
								{
									StepAndResults: []apistructs.TestCaseStepAndResult{
										{
											Step:   "Step",
											Result: "Result",
										},
									},
								},
							},
						},
					},
				},
			},
			want: OperationTypeSave,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOperationType(tt.args.fps); got != tt.want {
				t.Errorf("getOperationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_execOperationTypeSave(t *testing.T) {
	type args struct {
		fps       *FunctionParams
		bdl       *bundle.Bundle
		userId    string
		projectId uint64
	}

	var issueId, testcaseId, AIGeneratedTestSetId uint64
	issueId = 30001127564
	testcaseId = 111001
	AIGeneratedTestSetId = 2
	issueName := "用户登录"

	results := make([]apistructs.TestCaseMeta, 0)
	results = append(results, apistructs.TestCaseMeta{
		Req: apistructs.TestCaseCreateRequest{
			//TestCaseID:       111,
			ProjectID:        1,
			ParentTestSetID:  1,
			ParentTestSetDir: "/1",
			TestSetID:        AIGeneratedTestSetId,
			TestSetDir:       "/" + AIGeneratedTestSetName,
			TestSetName:      "1",
			Name:             "",
			PreCondition:     "",
			StepAndResults: []apistructs.TestCaseStepAndResult{
				{
					Step:   "Step",
					Result: "Result",
				},
			},
			Desc:     "",
			Priority: "",

			IdentityInfo: apistructs.IdentityInfo{
				UserID: "10001",
			},
		},
		RequirementName: issueName,
		RequirementID:   issueId,
		TestCaseID:      testcaseId,
	})
	tests := []struct {
		name    string
		args    args
		want    AICreateTestCasesResult
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				fps: &FunctionParams{
					TestSetID:    1,
					SystemPrompt: "xxx",
					Requirements: []TestCaseParam{
						{
							IssueID:          issueId,
							Prompt:           issueName,
							ParentTestSetID:  1,
							ParentTestSetDir: "/1",
							Reqs: []apistructs.TestCaseCreateRequest{
								{
									TestSetID:        111,
									ProjectID:        1,
									ParentTestSetID:  1,
									ParentTestSetDir: "/1",
									TestSetDir:       "/" + AIGeneratedTestSetName,
									TestSetName:      "1",
									StepAndResults: []apistructs.TestCaseStepAndResult{
										{
											Step:   "Step",
											Result: "Result",
										},
									},
									IdentityInfo: apistructs.IdentityInfo{
										UserID: "10001",
									},
								},
							},
						},
					},
				},
				bdl:       bundle.New(),
				userId:    "10001",
				projectId: 1,
			},
			want: AICreateTestCasesResult{
				IsSaveTestCasesSave: true,
				//TestCases: results,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.bdl), "CreateTestSet", func(_ *bundle.Bundle,
				req apistructs.TestSetCreateRequest) (*apistructs.TestSet, error) {
				return &apistructs.TestSet{
					ID:        AIGeneratedTestSetId,
					Name:      AIGeneratedTestSetName,
					ProjectID: 1,
					ParentID:  1,
					Recycled:  false,
					Directory: "/" + AIGeneratedTestSetName,
					Order:     1,
					CreatorID: "10001",
					UpdaterID: "10001",
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.bdl), "GetIssue", func(_ *bundle.Bundle,
				id uint64) (*apistructs.Issue, error) {
				return &apistructs.Issue{
					ID:    int64(issueId),
					Title: issueName,
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.bdl), "CreateTestCase", func(_ *bundle.Bundle,
				req apistructs.TestCaseCreateRequest) (apistructs.AICreateTestCaseResponse, error) {
				return apistructs.AICreateTestCaseResponse{
					TestCaseID: testcaseId,
				}, nil
			})

			defer monkey.UnpatchAll()

			got, err := execOperationTypeSave(tt.args.fps, tt.args.bdl, tt.args.userId, tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("execOperationTypeSave() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("execOperationTypeSave() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verifyAIGenerateResults(t *testing.T) {
	type args struct {
		results interface{}
	}

	groups := make([]string, 0)
	groups = append(groups, "group")
	results01 := make([]RequirementGroup, 0)
	results02 := make([]RequirementGroup, 0)
	results03 := make([]RequirementGroup, 0)
	results02 = append(results01, RequirementGroup{
		ID:     101,
		Groups: nil,
	})
	results03 = append(results01, RequirementGroup{
		ID:     101,
		Groups: groups,
	})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			args: args{
				results: results01,
			},
			wantErr: true,
		},
		{
			name: "Test_02",
			args: args{
				results: results02,
			},
			wantErr: true,
		},
		{
			name: "Test_03",
			args: args{
				results: results03,
			},
			wantErr: false,
		},
		{
			name: "Test_04",
			args: args{
				results: make([]any, 0),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyAIGenerateResults(tt.args.results); (err != nil) != tt.wantErr {
				t.Errorf("verifyAIGenerateResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFunction_SystemMessage(t *testing.T) {
	type fields struct {
		background *pb.Background
	}
	type args struct {
		lang string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Test_01",
			fields: fields{
				background: &pb.Background{},
			},
			args: args{
				lang: I18nLang_en_US,
			},
			want: systemMessage + "\n    - Return in English",
		},
		{
			name: "Test_02",
			fields: fields{
				background: &pb.Background{},
			},
			args: args{
				lang: I18nLang_zh_CN,
			},
			want: systemMessage + "\n    - Return in Chinese",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Function{
				background: tt.fields.background,
			}
			if got := f.SystemMessage(tt.args.lang); got != tt.want {
				t.Errorf("SystemMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_adjustMessageByLanguage(t *testing.T) {
	type args struct {
		lang      string
		groupName string
	}
	groupName := "Test"
	tests := []struct {
		name string
		args args
		want MessageByLanguage
	}{
		{
			name: "Test_01",
			args: args{
				lang:      I18nLang_zh_CN,
				groupName: groupName,
			},
			want: MessageByLanguage{
				TaskContent:   "需求和任务相关联，一个需求事项包含多个任务事项，这是我所有的任务标题:",
				GroupContent:  "这是我的功能分组: \n",
				GenerateTC:    fmt.Sprintf("请根据 '%s' 这个功能分组，基于需求名称、需求描述和任务名称设计一系列高质量的功能测试用例。测试用例的名称应该以对应的功能点作为命名。请确保生成的测试用例能够充分覆盖该功能分组，并包括清晰的输入条件、操作步骤和期望的输出结果。", groupName),
				GenerateGroup: "请根据需求标题、需求内容和任务标题，帮助我生成一系列高质量的测试用例功能分组。生成测试用例分组的规则参考我上面给出的案例。测试用例功能分组应该基于需求的主题和任务的关联性。并使用功能点对每个功能分组进行命名。不要出现含义相同或者重复的测试用例功能分组。",
			},
		},
		{
			name: "Test_02",
			args: args{
				lang:      I18nLang_en_US,
				groupName: groupName,
			},
			want: MessageByLanguage{
				TaskContent:   "Requirements are related to tasks. A requirement item contains multiple task items. This is the title of all my tasks:",
				GroupContent:  "This is my function grouping result: \n",
				GenerateTC:    fmt.Sprintf("Please design a series of high-quality functional test cases based on the requirement name, requirement description and task name according to the functional grouping '%s'. The name of the test case should be named after the corresponding function point. Please ensure that the generated test cases can fully cover the functional grouping and include clear input conditions, operation steps and expected output results. The results must be returned in English.", groupName),
				GenerateGroup: "Please help me generate a series of high-quality test case functional groupings based on the requirement title, requirement content and task title. The rules for generating test case groups refer to the case I gave above. Functional grouping of test cases should be based on the topic of the requirements and the relevance of the tasks. And use function points to name each function group. The naming results must be expressed in English. There should be no test case function groups with the same meaning or duplicates.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adjustMessageByLanguage(tt.args.lang, tt.args.groupName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("adjustMessageByLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}
