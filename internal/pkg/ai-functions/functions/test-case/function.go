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
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/strutil"
)

const Name = "create-test-case"

//go:embed schema.yaml
var Schema json.RawMessage

//go:embed system-message.txt
var systemMessage string

var (
	_ functions.Function = (*Function)(nil)
)

func init() {
	functions.Register(Name, New)
}

type Function struct {
	background *pb.Background
}

// FunctionParams 解析 *pb.ApplyRequest 字段 FunctionParams
type FunctionParams struct {
	TestSetID    uint64          `json:"testSetID,omitempty"`
	Requirements []TestCaseParam `json:"requirements,omitempty"`
}
type TestCaseParam struct {
	IssueID uint64                           `json:"issueID,omitempty"`
	Prompt  string                           `json:"prompt,omitempty"`
	Req     apistructs.TestCaseCreateRequest `json:"testCaseCreateReq,omitempty"`
}

// TestCaseFunctionInput 用于为单个需求生成测试用例的输入
type TestCaseFunctionInput struct {
	TestSetParentID uint64
	TestSetID       uint64
	IssueID         uint64
	Prompt          string
}

// TestCaseMeta 用于关联生成的测试用例与对应的需求
type TestCaseMeta struct {
	Req             apistructs.TestCaseCreateRequest `json:"testCaseCreateReq,omitempty"` // 当前项目 ID 对应的创建测试用例请求
	RequirementName string                           `json:"requirementName"`             // 需求对应的 issue 的 Title
	RequirementID   uint64                           `json:"requirementID"`               // 需求对应的 issueID
	TestCaseID      uint64                           `json:"testcaseID,omitempty"`        // 创建测试用例成功返回的测试用例 ID
}

func New(ctx context.Context, prompt string, background *pb.Background) functions.Function {
	return &Function{background: background}
}

func (f *Function) Name() string {
	return Name
}

func (f *Function) Description() string {
	return "create test case"
}

func (f *Function) SystemMessage() string {
	return systemMessage
}

func (f *Function) UserMessage() string {
	return "Not really implemented."
}

func (f *Function) Schema() (json.RawMessage, error) {
	schema, err := strutil.YamlOrJsonToJson(Schema)
	return schema, err
}

func (f *Function) RequestOptions() []sdk.RequestOption {
	return []sdk.RequestOption{
		sdk.RequestOptionWithResetAPIVersion("2023-07-01-preview"),
	}
}

func (f *Function) CompletionOptions() []sdk.PatchOption {
	return []sdk.PatchOption{
		sdk.PathOptionWithModel("gpt-35-turbo-16k"),
		sdk.PathOptionWithTemperature(1),
	}
}

func (f *Function) Callback(ctx context.Context, arguments json.RawMessage, input interface{}, needAdjust bool) (any, error) {
	testCaseInput, ok := input.(TestCaseFunctionInput)
	if !ok {
		err := errors.Errorf("input %v with type %T is not valid for AI Function %s", input, input, Name)
		return nil, errors.Wrap(err, "bad request: invalid input")
	}

	bdl := bundle.New(bundle.WithErdaServer())
	// 根据 issueID 获取对应的需求 Title
	issue, err := bdl.GetIssue(testCaseInput.IssueID)
	if err != nil {
		return nil, errors.Wrap(err, "get requirement info failed")
	}
	if issue.Type != apistructs.IssueTypeRequirement {
		return nil, errors.Wrap(err, "bad request: issue is not type REQUIREMENT")
	}

	var req apistructs.TestCaseCreateRequest
	if err := json.Unmarshal(arguments, &req); err != nil {
		return nil, errors.Wrap(err, "Unmarshal arguments to TestCaseCreateRequest failed")
	}
	f.adjustTestCaseCreateRequest(&req, testCaseInput.TestSetID, testCaseInput.Prompt, issue)

	// 需要调整，则返回创建测试用例的请求 apistructs.TestCaseCreateRequest
	if needAdjust {
		return TestCaseMeta{
			Req:             req,
			RequirementName: issue.Title,
			RequirementID:   uint64(issue.ID),
		}, nil
	}

	// 无需调整，则返回创建测试用例的请求 apistructs.TestCaseCreateRequest 以及创建成功之后对应的 testcaseID
	aiCreateTestCaseResponse, err := bdl.CreateTestCase(req)
	if err != nil {
		return nil, errors.Wrap(err, "bundle CreateTestCase failed")
	}

	return TestCaseMeta{
		Req:             req,
		RequirementName: issue.Title,
		RequirementID:   uint64(issue.ID),
		TestCaseID:      aiCreateTestCaseResponse.TestCaseID,
	}, nil
}

func (f *Function) adjustTestCaseCreateRequest(req *apistructs.TestCaseCreateRequest, testSetID uint64, prompt string, issue *apistructs.Issue) {
	req.Name = issue.Title
	req.ProjectID = f.background.ProjectID
	req.Desc = fmt.Sprintf("Powered by AI.\n\n对应需求:\n%s", prompt)
	req.TestSetID = testSetID
	// 根据需求优先级相应设置测试用例优先级
	switch issue.Priority {
	case apistructs.IssuePriorityUrgent:
		req.Priority = apistructs.TestCasePriorityP0
	case apistructs.IssuePriorityHigh:
		req.Priority = apistructs.TestCasePriorityP1
	case apistructs.IssuePriorityNormal:
		req.Priority = apistructs.TestCasePriorityP2
	default:
		req.Priority = apistructs.TestCasePriorityP3
	}
	req.UserID = f.background.UserID
	return
}
