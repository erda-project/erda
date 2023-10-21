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
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	aitestcase "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
	aiHandlerUtils "github.com/erda-project/erda/internal/pkg/ai-functions/handler/utils"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

const (
	AIGeneratedTestSeName = "AI_Generated"
)

func (h *AIFunction) createTestCaseForRequirementIDAndTestID(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL) (any, error) {
	results := make([]any, 0)
	var wg sync.WaitGroup
	var functionParams aitestcase.FunctionParams

	FunctionParamsBytes, err := req.GetFunctionParams().MarshalJSON()
	if err != nil {
		return nil, errors.Wrapf(err, "MarshalJSON for req.FunctionParams failed.")
	}
	if err = json.Unmarshal(FunctionParamsBytes, &functionParams); err != nil {
		return nil, errors.Wrapf(err, "Unmarshal req.FunctionParams to struct FunctionParams failed.")
	}
	logrus.Debugf("parse createTestCase functionParams=%+v", functionParams)

	if err := validateParamsForCreateTestcase(functionParams); err != nil {
		return nil, errors.Wrapf(err, "validateParamsForCreateTestcase faild")
	}

	// 用户未指定测试集可能需要创建测试集
	if functionParams.TestSetID == 0 {
		bdl := bundle.New(bundle.WithErdaServer())
		var parentTestSetId uint64 = 0
		projectId := req.GetBackground().GetProjectID()
		userId := req.GetBackground().GetUserID()

		testSets, err := bdl.GetTestSets(apistructs.TestSetListRequest{
			ParentID:  &parentTestSetId,
			ProjectID: &projectId,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get  testSets by project failed")
		}

		needCreate := true
		for _, testSet := range testSets {
			if testSet.ParentID == 0 && testSet.Name == AIGeneratedTestSeName {
				functionParams.TestSetID = testSet.ID
				needCreate = false
				break
			}
		}

		if needCreate {
			testSet, err := bdl.CreateTestSet(apistructs.TestSetCreateRequest{
				ProjectID: &projectId,
				ParentID:  &parentTestSetId,
				Name:      AIGeneratedTestSeName,
				IdentityInfo: apistructs.IdentityInfo{
					UserID: userId,
				},
			})
			if err != nil {
				return nil, errors.Wrap(err, "create TestSet failed")
			}
			functionParams.TestSetID = testSet.ID
		}
	}

	for _, tp := range functionParams.Requirements {
		wg.Add(1)
		var err error
		if err = processSingleTestCase(ctx, factory, req, openaiURL, &wg, tp, functionParams.TestSetID, functionParams.SystemPrompt, &results); err != nil {
			return nil, errors.Wrapf(err, "process single testCase create faild")
		}
	}
	wg.Wait()

	content := httpserver.Resp{
		Success: true,
		Data:    results,
	}
	return json.Marshal(content)
}

func processSingleTestCase(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, wg *sync.WaitGroup, tp aitestcase.TestCaseParam, testSetId uint64, systemPrompt string, results *[]any) error {
	defer wg.Done()

	callbackInput := aitestcase.TestCaseFunctionInput{
		TestSetID: testSetId,
		IssueID:   tp.IssueID,
		Prompt:    tp.Prompt,
	}

	bdl := bundle.New(bundle.WithErdaServer())
	issue, err := bdl.GetIssue(tp.IssueID)
	if err != nil {
		return errors.Wrap(err, "get requirement info failed when create testcase")
	}
	// 根据需求内容生成 prompt
	if hasDetailInfoInRequirementContent(issue.Content) {
		callbackInput.Prompt = issue.Content
		// callbackInput.Prompt = tuningPrompt(issue.Title, issue.Content)
	} else {
		callbackInput.Prompt = issue.Title
	}
	tp.Prompt = callbackInput.Prompt

	if len(tp.Req.StepAndResults) > 0 {
		// 表示是修改后批量应用应用生成的测试用例，直接调用创建接口，无需再次生成
		aiCreateTestCaseResponse, err := bdl.CreateTestCase(tp.Req)
		if err != nil {
			err = errors.Errorf("create testcase with req %+v failed: %v", tp.Req, err)
			return errors.Wrap(err, "bundle CreateTestCase failed for ")
		}

		*results = append(*results, aitestcase.TestCaseMeta{
			Req:             tp.Req,
			RequirementName: issue.Title,
			RequirementID:   tp.IssueID,
			TestCaseID:      aiCreateTestCaseResponse.TestCaseID,
		})
	} else {

		var (
			f          = factory(ctx, "", req.GetBackground())
			systemMsg0 = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: f.SystemMessage(), // 系统提示语
				Name:    "system",
			}
			userMsg0 = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "requirement name:" + issue.Title,
				Name:    "erda",
			}
			userMsg1 = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "requirement description (markdown format): \n" + issue.Content,
				Name:    "erda",
			}
			systemMsg1 = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Please generate a high-quality, complete test case based on the user's requirement name and description, including functional dimensions, boundary values, exception scenarios, performance scenarios, security scenarios and compatibility scenarios.",
			}
		)
		if systemPrompt != "" {
			systemMsg1.Content = systemPrompt
		}

		messages := []openai.ChatCompletionMessage{systemMsg0, userMsg0, userMsg1, systemMsg1}

		result, err := aiHandlerUtils.GetChatMessageFunctionCallArguments(ctx, factory, req, openaiURL, messages, callbackInput)
		if err != nil {
			return err
		}

		*results = append(*results, result)
	}

	return nil
}

// validateParamsForCreateTestcase 校验创建测试用例对应的参数配置
func validateParamsForCreateTestcase(req aitestcase.FunctionParams) error {
	if req.TestSetID < 0 {
		return errors.Errorf("AI function functionParams testSetID for %s invalid", aitestcase.Name)
	}

	if len(req.Requirements) == 0 {
		return errors.Errorf("AI function functionParams requirements for %s invalid", aitestcase.Name)
	}

	for idx, tp := range req.Requirements {
		if tp.IssueID <= 0 {
			return errors.Errorf("AI function functionParams requirements[%d].issueID for %s invalid", idx, aitestcase.Name)
		}
	}

	return nil
}

// tuningPrompt 提取需求的内容,整理成标准格式，如:
// 需求名称： 注册需求
// 需求描述：
// 1. xxx
// 2. xxx
func tuningPrompt(title, content string) string {
	result := make([]string, 0)
	inputs := strings.Split(content, "###")
	result = append(result, "需求名称： "+title)
	result = append(result, "需求描述：")
	item := 1

	for i := 1; i < len(inputs); i++ {
		// 意向用户 和 链接/参考 不作为需求的详细描述内容
		if strings.Contains(inputs[i], "意向用户*") || strings.Contains(inputs[i], "链接/参考") {
			continue
		}

		for _, r := range strings.Split(inputs[i], "\n") {
			if r == "" || strings.Contains(r, "用户故事/要解决的问题*") || strings.Contains(r, "用户体验目标*") {
				continue
			}
			r = strings.TrimSpace(r)
			// 删除常见标题性质的内容
			if (strings.HasPrefix(r, "功能需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "安全需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "性能需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "兼容性需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "业务需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "用户需求") && len(r) <= 21) ||
				(strings.HasPrefix(r, "系统需求") && len(r) <= 21) {
				continue
			}
			result = append(result, fmt.Sprintf("%d. ", item)+r)
			item++
		}
	}

	return strings.Join(result, "\n")
}

func hasDetailInfoInRequirementContent(input string) bool {
	result := make([]string, 0)
	inputs := strings.Split(input, "\n")

	for i := 0; i < len(inputs); i++ {
		if inputs[i] == "" {
			continue
		}
		result = append(result, inputs[i])
	}

	input = strings.Join(result, "\n")
	fmt.Printf("\nInput now:\n%s\n", input)

	if len(result) == 4 {
		return false
	}

	return true
}
