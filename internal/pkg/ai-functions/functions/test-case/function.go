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
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	aiHandlerUtils "github.com/erda-project/erda/internal/pkg/ai-functions/handler/utils"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name                               = "create-test-case"
	AIGeneratedTestSetName             = "AI_Generated"
	AIGeneratedLevel2TestSetNamePrefix = "AI_"

	// OperationTypeGenerate 表示当前调用接口触发测试用例 生成 操作
	OperationTypeGenerate = "Generate"
	// OperationTypeSave 表示当前调用接口触发测试用例 保存 操作
	OperationTypeSave = "Save"
)

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
	SystemPrompt string          `json:"systemPrompt,omitempty"`
	Requirements []TestCaseParam `json:"requirements,omitempty"`
}
type TestCaseParam struct {
	IssueID          uint64                             `json:"issueID,omitempty"`
	Prompt           string                             `json:"prompt,omitempty"`
	ParentTestSetID  uint64                             `json:"parentTestSetID,omitempty"`  // 所属测试集的 Parent 测试集 ID
	ParentTestSetDir string                             `json:"parentTestSetDir,omitempty"` // 所属测试集的 Parent 测试集 Directory
	Reqs             []apistructs.TestCaseCreateRequest `json:"testCaseCreateReqs,omitempty"`
}

// TestCaseFunctionInput 用于为单个需求生成测试用例的输入
type TestCaseFunctionInput struct {
	TestSetParentID  uint64 // 生成的测试用例所在的测试集的父测试集 ID
	TestSetParentDir string // 生成的测试用例所在的测试集的父测试集 Directory
	TestSetID        uint64 // 生成的测试用例所在的测试集测 ID
	IssueID          uint64 // 生成的测试用例对应的需求的 ID
	Prompt           string // 为本次生成测试用例输入的 Prompt

	Name      string //生成的测试用例对应的按需求内容分组名称
	UserId    string
	ProjectId uint64
}

// TestcasesDirsInfo 用于返回生成的测试集目录以及测试集中对应的测试用例的数量
type TestcasesDirsInfo struct {
	RootDir string    `json:"rootDir"`
	SubDirs []SubDirs `json:"subdirs"`
}

type SubDirs struct {
	Dir           string    `json:"dir"`
	RequirementID uint64    `json:"requirementId,omitempty"`
	SubDirs       []SubDirs `json:"subDirs,omitempty"`
	Count         int       `json:"count,omitempty"`
}

// AI 生成测试用例返回数据
type AICreateTestCasesResult struct {
	IsSaveTestCasesSave bool              `json:"isTestCasesSaved,omitempty"`
	TestCases           interface{}       `json:"testcases,omitempty"`
	TestSetsInfo        TestcasesDirsInfo `json:"testSetsInfo,omitempty"`
}

// 保存测试用例时，用于构建需求与测试用例的索引关联关系
type RequireTestCaseIndex struct {
	RequirementIndex int // aitestcase.FunctionParams.Requirements  的 index
	TestCaseIndex    int // aitestcase.FunctionParams.Requirements[].Reqs 的 index 列表
}

// createRootTestSetIfNecessary 根据请求参数确定创建的测试集的 Root 测试集(不存在则创建)
func createRootTestSetIfNecessary(fps *FunctionParams, bdl *bundle.Bundle, userId string, projectId uint64) error {
	// 1. 指定了 Root TestSet ID，无需创建
	if fps.TestSetID > 0 {
		return nil
	}

	// 2. 未指定 Root TestSet ID，判断测试集 0 下的名称为 AI_Generated 的测试集是否创建，未创建则创建名称为 AI_Generated 的测试集
	var parentTestSetId uint64 = 0
	testSets, err := bdl.GetTestSets(apistructs.TestSetListRequest{
		ParentID:  &parentTestSetId,
		ProjectID: &projectId,
	})
	if err != nil {
		return errors.Wrap(err, "get  testSets by project failed")
	}

	needCreate := true
	for _, testSet := range testSets {
		if testSet.ParentID == 0 && testSet.Name == AIGeneratedTestSetName {
			fps.TestSetID = testSet.ID
			needCreate = false
			break
		}
	}

	if needCreate {
		testSet, err := bdl.CreateTestSet(apistructs.TestSetCreateRequest{
			ProjectID: &projectId,
			ParentID:  &parentTestSetId,
			Name:      AIGeneratedTestSetName,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: userId,
			},
		})
		if err != nil {
			return errors.Wrap(err, "create TestSet failed")
		}
		fps.TestSetID = testSet.ID
	}
	return nil
}

// getOperationType 根据请求参数获取当前 API 调用触发的操作类型
func getOperationType(fps FunctionParams) string {
	for _, r := range fps.Requirements {
		for _, tc := range r.Reqs {
			if len(tc.StepAndResults) > 0 {
				// 表示是批量应用应用生成的测试用例，之前的二测试集都已生产，无需无需再次生成
				// 对应后续操作是测试用例存储到测试集，因此先要判断测试集有没有删除
				return OperationTypeSave
			}
		}
	}
	return OperationTypeGenerate
}

// execOperationTypeSave 执行实际的保存操作
func execOperationTypeSave(fps *FunctionParams, bdl *bundle.Bundle, userId string, projectId uint64) (AICreateTestCasesResult, error) {
	results := make([]apistructs.TestCaseMeta, 0)
	// 保存:  保存每个需求的  N 个分组生成 N 个测试用例
	// 获取要创建的 3 级测试集
	//map[2级测试集ID]map[3级测试集ID]3级测试集名称
	secondToThirdTestSetIds := make(map[uint64]map[uint64]string)
	// make(map[3级测试集Id][]RequireTestCaseIndex)
	thirdTestSetIdToTestCaseIndexies := make(map[uint64][]RequireTestCaseIndex)
	for idx, tp := range fps.Requirements {
		if _, ok := secondToThirdTestSetIds[tp.ParentTestSetID]; !ok {
			secondToThirdTestSetIds[tp.ParentTestSetID] = make(map[uint64]string)
		}
		for tcIndex, tcr := range tp.Reqs {
			secondToThirdTestSetIds[tp.ParentTestSetID][tcr.TestSetID] = tcr.TestSetName
			if _, ok := thirdTestSetIdToTestCaseIndexies[tcr.TestSetID]; !ok {
				thirdTestSetIdToTestCaseIndexies[tcr.TestSetID] = make([]RequireTestCaseIndex, 0)
			}
			thirdTestSetIdToTestCaseIndexies[tcr.TestSetID] = append(thirdTestSetIdToTestCaseIndexies[tcr.TestSetID], RequireTestCaseIndex{
				RequirementIndex: idx,
				TestCaseIndex:    tcIndex,
			})
		}
	}

	// 创建 3 级测试集，并将创建的测试集 ID 更新到对应的测试用例创建请求的 TestSetID 字段
	for secondId, thirdIds := range secondToThirdTestSetIds {
		for thirdId, name := range thirdIds {
			if name == "" || secondId == 0 {
				continue
			}
			ts, err := bdl.CreateTestSet(apistructs.TestSetCreateRequest{
				ProjectID: &projectId,
				ParentID:  &secondId,
				Name:      name,
				IdentityInfo: apistructs.IdentityInfo{
					UserID: userId,
				},
			})
			if err != nil {
				return AICreateTestCasesResult{
					IsSaveTestCasesSave: false,
				}, errors.Errorf("create third level TestSet name [%s] failed, err: %v", name, err)
			}

			for _, rtcr := range thirdTestSetIdToTestCaseIndexies[thirdId] {
				fps.Requirements[rtcr.RequirementIndex].Reqs[rtcr.TestCaseIndex].TestSetID = ts.ID
				fps.Requirements[rtcr.RequirementIndex].Reqs[rtcr.TestCaseIndex].TestSetDir = ts.Directory
			}
		}
	}

	for _, tcp := range fps.Requirements {
		issue, err := bdl.GetIssue(tcp.IssueID)
		if err != nil {
			return AICreateTestCasesResult{
				IsSaveTestCasesSave: false,
			}, errors.Wrap(err, "get requirement info failed when create testcase")
		}
		for _, tcr := range tcp.Reqs {
			// 表示是批量应用应用生成的测试用例，直接调用创建接口，无需再次生成
			aiCreateTestCaseResponse, err := bdl.CreateTestCase(tcr)
			if err != nil {
				err = errors.Errorf("create testcase with req %+v failed: %v", tcr, err)
				return AICreateTestCasesResult{
					IsSaveTestCasesSave: false,
				}, errors.Wrap(err, "bundle CreateTestCase failed for ")
			}
			results = append(results, apistructs.TestCaseMeta{
				Req:             tcr,
				RequirementName: issue.Title,
				RequirementID:   tcp.IssueID,
				TestCaseID:      aiCreateTestCaseResponse.TestCaseID,
			})
		}
	}

	return AICreateTestCasesResult{
		IsSaveTestCasesSave: true,
	}, nil
}

// execOperationTypeGenerate 执行实际的生成操作
func execOperationTypeGenerate(fps *FunctionParams, ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, xAIProxyModelId string, requirements []*apistructs.Issue, dirsInfo TestcasesDirsInfo) (AICreateTestCasesResult, error) {
	results := make([]any, 0)
	var wg sync.WaitGroup

	// 根据需求产生分组信息
	requirementIdToGroups, err := GenerateGroupsForRequirements(ctx, requirements, openaiURL, xAIProxyModelId)
	if err != nil {
		return AICreateTestCasesResult{}, errors.Wrap(err, "create TestSet GenerateGroupsForRequirements failed")
	}

	// 生成:  为每个需求的 N 个分组生成 N 个测试用例
	for idx, tp := range fps.Requirements {
		if groups, ok := requirementIdToGroups[tp.IssueID]; ok {
			for _, groupName := range groups {
				if groupName == "" {
					continue
				}
				wg.Add(1)
				go processSingleTestCase(ctx, factory, req, openaiURL, &wg, tp, 0, fps.Requirements[idx].ParentTestSetID, fps.Requirements[idx].ParentTestSetDir, groupName, groups, fps.SystemPrompt, &results)
			}
		}
	}
	wg.Wait()

	// 异步操作可能请求参数、Header 错误或者后端 ai-proxy 服务不可用,倒是实际无结果数据
	err = verifyAIGenerateResults(results)
	if err != nil {
		return AICreateTestCasesResult{}, err
	}

	// 生成测试用例结果的统计信息
	l2DirToL3Dirs := make(map[string]map[string]int)
	issueToTestCasesMetas := make(map[uint64][]apistructs.TestCasesMeta)
	for _, r := range results {
		tcsm, ok := r.(apistructs.TestCasesMeta)
		if !ok {
			continue
		}
		if len(tcsm.Reqs) == 0 {
			continue
		}

		if _, exist := issueToTestCasesMetas[tcsm.RequirementID]; !exist {
			issueToTestCasesMetas[tcsm.RequirementID] = make([]apistructs.TestCasesMeta, 0)
		}
		issueToTestCasesMetas[tcsm.RequirementID] = append(issueToTestCasesMetas[tcsm.RequirementID], tcsm)

		for _, tcr := range tcsm.Reqs {
			if _, exist := l2DirToL3Dirs[tcr.ParentTestSetDir]; !exist {
				l2DirToL3Dirs[tcr.ParentTestSetDir] = make(map[string]int)
				l2DirToL3Dirs[tcr.ParentTestSetDir][tcr.TestSetDir] = 0
			}
			l2DirToL3Dirs[tcr.ParentTestSetDir][tcr.TestSetDir] += 1
		}
	}

	for idx := range dirsInfo.SubDirs {
		for l3dir, count := range l2DirToL3Dirs[dirsInfo.SubDirs[idx].Dir] {
			dirsInfo.SubDirs[idx].SubDirs = append(dirsInfo.SubDirs[idx].SubDirs, SubDirs{
				Dir:   l3dir,
				Count: count,
			})
		}
	}

	res := make([]TestCaseParam, 0)
	for issuId, tcms := range issueToTestCasesMetas {
		tcrs := make([]apistructs.TestCaseCreateRequest, 0)
		for _, tcm := range tcms {
			tcrs = append(tcrs, tcm.Reqs...)
		}
		res = append(res, TestCaseParam{
			IssueID: issuId,
			Reqs:    tcrs,
		})
	}

	return AICreateTestCasesResult{
		IsSaveTestCasesSave: false,
		TestCases:           res,
		TestSetsInfo:        dirsInfo,
	}, nil
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
		// 改变温度参数会改变模型的输出。 温度参数可以设置为 0 到 2。 较高的值（例如 0.7）将使输出更随机，并产生更多发散的响应，而较小的值（例如 0.2）将使输出更加集中和具体。
		sdk.PathOptionWithTemperature(0.5),
	}
}

type TestCaseCreateRequestList struct {
	List []apistructs.TestCaseCreateRequest `json:"list"`
}

func (f *Function) Callback(ctx context.Context, arguments json.RawMessage, input interface{}) (any, error) {
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

	var reqs TestCaseCreateRequestList
	if err := json.Unmarshal(arguments, &reqs); err != nil {
		return nil, errors.Wrap(err, "Unmarshal arguments to TestCaseCreateRequest failed")
	}

	// 创建 分组对应的测试集
	// 生成操作  不需要真实创建 3 级 测试集
	var fakeId uint64
	fakeId = uint64(time.Now().Nanosecond())
	testSet := &apistructs.TestSet{
		ID:        fakeId,
		Name:      testCaseInput.Name,
		ProjectID: testCaseInput.ProjectId,
		ParentID:  testCaseInput.TestSetParentID,
		Directory: testCaseInput.TestSetParentDir + "/" + testCaseInput.Name,
	}

	for idx := range reqs.List {
		f.adjustTestCaseCreateRequest(&reqs.List[idx], testSet.ID, testSet.Name, testCaseInput.TestSetParentID, testCaseInput.TestSetParentDir, issue)
	}

	// 返回创建测试用例的请求 apistructs.TestCaseCreateRequest
	return apistructs.TestCasesMeta{
		Reqs:            reqs.List,
		RequirementName: issue.Title,
		RequirementID:   uint64(issue.ID),
	}, nil
}

func (f *Function) adjustTestCaseCreateRequest(req *apistructs.TestCaseCreateRequest, testSetID uint64, testSetName string, pTestSetID uint64, pTestSetDir string, issue *apistructs.Issue) {
	req.ProjectID = f.background.ProjectID
	req.Desc = fmt.Sprintf("Powered by AI.")
	req.ParentTestSetID = pTestSetID
	req.ParentTestSetDir = pTestSetDir
	req.TestSetID = testSetID
	req.TestSetDir = pTestSetDir + "/" + testSetName
	req.TestSetName = testSetName
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

func (f *Function) Handler(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, xAIProxyModelId string) (any, error) {
	var functionParams FunctionParams
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

	// 【Step 1】确定 Root 测试集
	projectId := req.GetBackground().GetProjectID()
	userId := req.GetBackground().GetUserID()
	bdl := bundle.New(bundle.WithErdaServer())
	// 用户未指定测试集可能需要创建测试集
	err = createRootTestSetIfNecessary(&functionParams, bdl, userId, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "create or get root testset faild")
	}

	// 【Step 2】确定是【生成】 操作 还是 【保存】 操作
	// 生成测试集目录信息
	operationType := getOperationType(functionParams)

	// 【Step 3】根据【生成】 操作 或 【保存】 操作 的情况生成 Fake 或 Real 的 二级测试集
	var dirsInfo TestcasesDirsInfo
	dirsInfo.SubDirs = make([]SubDirs, 0)

	// 为每个需求生成测试集，并关联到分组信息
	requirements := make([]*apistructs.Issue, 0)
	dateStr := time.Now().Local().Format("20060102150405")
	for idx, r := range functionParams.Requirements {
		issue, err := bdl.GetIssue(r.IssueID)
		if err != nil {
			return nil, errors.Wrap(err, "generateGroupsFromRequirement get requirement info failed")
		}

		parentTestSetId := functionParams.TestSetID
		var testSet *apistructs.TestSet

		switch operationType {
		case OperationTypeSave:
			// 保存操作  需要真实创建 二级 测试集
			testSet, err = bdl.CreateTestSet(apistructs.TestSetCreateRequest{
				ProjectID: &projectId,
				ParentID:  &parentTestSetId,
				Name:      AIGeneratedLevel2TestSetNamePrefix + fmt.Sprintf("%d_%s", issue.ID, dateStr),
				IdentityInfo: apistructs.IdentityInfo{
					UserID: userId,
				},
			})
			if err != nil {
				return nil, errors.Errorf("create second level TestSet %s failed, err: %v", AIGeneratedLevel2TestSetNamePrefix+fmt.Sprintf("%d_%s", issue.ID, dateStr), err)
			}
		default:
			// 生成操作  不需要真实创建 二级 测试集
			//OperationTypeGenerate
			// 获取 一级 测试集
			parentTestSet, err := bdl.GetTestSetById(apistructs.TestSetGetRequest{
				ID: functionParams.TestSetID,
				IdentityInfo: apistructs.IdentityInfo{
					UserID: userId,
				},
			})
			if err != nil {
				return nil, errors.Errorf("get root testset by id %d failed, err: %v\n", functionParams.TestSetID, err)
			}

			var fakeId uint64
			fakeId = uint64(idx) + 1
			testSet = &apistructs.TestSet{
				ID:        fakeId,
				Name:      AIGeneratedLevel2TestSetNamePrefix + fmt.Sprintf("%d_%s", issue.ID, dateStr),
				ProjectID: projectId,
				ParentID:  parentTestSetId,
				Directory: parentTestSet.Directory + "/" + AIGeneratedLevel2TestSetNamePrefix + fmt.Sprintf("%d_%s", issue.ID, dateStr),
			}
		}

		// 用于后续获取三级目录
		requirementIDToTestSetID := make(map[uint64]uint64)
		// 情况 2: 调用接口实际操作是生成

		requirementIDToTestSetID[r.IssueID] = testSet.ID
		dirsInfo.SubDirs = append(dirsInfo.SubDirs, SubDirs{
			Dir:           testSet.Directory,
			RequirementID: r.IssueID,
		})

		functionParams.Requirements[idx].ParentTestSetID = testSet.ID
		functionParams.Requirements[idx].ParentTestSetDir = testSet.Directory
		requirements = append(requirements, issue)
	}

	// 【Step 4】进行 【生成】 操作 或 【保存】 操作
	data := AICreateTestCasesResult{}
	switch operationType {
	case OperationTypeSave:
		// 保存操作
		// 【Step 4】【情况 1】 保存操作
		data, err = execOperationTypeSave(&functionParams, bdl, userId, projectId)
		if err != nil {
			return nil, errors.Wrap(err, "execOperationTypeSave failed")
		}

	default:
		// 生成操作
		//【Step 4】【情况 2】 生成操作
		data, err = execOperationTypeGenerate(&functionParams, ctx, factory, req, openaiURL, xAIProxyModelId, requirements, dirsInfo)
		if err != nil {
			return nil, errors.Wrap(err, "execOperationTypeGenerate failed")
		}
	}

	content := httpserver.Resp{
		Success: true,
		Data:    data,
	}
	return json.Marshal(content)
}

// processSingleTestCase 仅处理生成操作
func processSingleTestCase(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, wg *sync.WaitGroup, tp TestCaseParam, testSetId uint64, parentTestSetId uint64, parentTestSetDir string, groupName string, groups []string, systemPrompt string, results *[]any) error {
	defer wg.Done()
	callbackInput := TestCaseFunctionInput{
		TestSetParentID:  parentTestSetId,
		TestSetParentDir: parentTestSetDir,
		TestSetID:        testSetId,
		IssueID:          tp.IssueID,
		Prompt:           tp.Prompt,
		Name:             groupName,
		UserId:           req.GetBackground().GetUserID(),
		ProjectId:        req.GetBackground().GetProjectID(),
	}

	bdl := bundle.New(bundle.WithErdaServer())
	issue, err := bdl.GetIssue(tp.IssueID)
	if err != nil {
		return errors.Wrap(err, "get requirement info failed when create testcase")
	}
	// 根据需求内容生成 prompt
	if hasDetailInfoInRequirementContent(issue.Content) {
		callbackInput.Prompt = issue.Content
	} else {
		callbackInput.Prompt = issue.Title
	}
	tp.Prompt = callbackInput.Prompt

	// 1. 生成 操作
	var f = factory(ctx, "", req.GetBackground())
	messages := make([]openai.ChatCompletionMessage, 0)
	// 添加系统提示词
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: f.SystemMessage(), // 系统提示语
		Name:    "system",
	})

	// 添加需求名称
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "requirement name:" + issue.Title,
		Name:    "erda",
	})

	// 添加需求信息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "requirement description (markdown format): \n" + issue.Content,
		Name:    "erda",
	})

	// 添加任务相关信息
	tasks := make([]string, 0)
	issueRelations, err := bdl.GetIssueRelations(tp.IssueID)
	if err != nil {
		return errors.Wrap(err, "processSingleTestCase get requirement related tasks info failed")
	}
	for _, include := range issueRelations.IssueInclude {
		if include.Type == apistructs.IssueTypeTask && (include.TaskType == "dev" || include.TaskType == "开发") {
			tasks = append(tasks, include.Title)
		}
	}
	if len(tasks) > 0 {
		taskContent := "需求和任务相关联，一个需求事项包含多个任务事项，这是我所有的任务标题:"
		for idx, task := range tasks {
			tasks[idx] = "\ntask name:" + task
		}
		taskContent = taskContent + strings.Join(tasks, ",")

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: taskContent,
		})
	}

	// 添加分组信息
	groupContent := "这是我的功能分组: \n"
	groupContent = groupContent + strings.Join(groups, "\n") + "\n"
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: groupContent,
	})

	// 最后生成指令
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: fmt.Sprintf("请根据 '%s' 这个功能分组，基于需求名称、需求描述和任务名称设计一系列高质量的功能测试用例。测试用例的名称应该以对应的功能点作为命名。请确保生成的测试用例能够充分覆盖该功能分组，并包括清晰的输入条件、操作步骤和期望的输出结果。", groupName),
	})

	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: systemPrompt,
		})
	}

	result, err := aiHandlerUtils.GetChatMessageFunctionCallArguments(ctx, factory, req, openaiURL, messages, callbackInput)
	if err != nil {
		return err
	}

	*results = append(*results, result)
	return nil
}

// validateParamsForCreateTestcase 校验创建测试用例对应的参数配置
func validateParamsForCreateTestcase(req FunctionParams) error {
	if req.TestSetID < 0 {
		return errors.Errorf("AI function functionParams testSetID for %s invalid", Name)
	}

	// 由于生成操作调用 AI 返回时间较长，每个需求的测试用例生成是异步并行执行,因此根据 CPU 数量做输入需求数量的限制
	if len(req.Requirements) == 0 || len(req.Requirements) > runtime.NumCPU() {
		return errors.Errorf("AI function functionParams requirements for %s invalid, limit by CPU cores, must between [1, %d]", Name, runtime.NumCPU())
	}

	for idx, tp := range req.Requirements {
		if tp.IssueID <= 0 {
			return errors.Errorf("AI function functionParams requirements[%d].issueID for %s invalid", idx, Name)
		}
	}

	return nil
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

	if len(result) == 4 {
		return false
	}

	return true
}

func verifyAIGenerateResults(results interface{}) error {
	// AI 生成测试用例结果校验
	if tcs, ok := results.([]any); ok {
		if len(tcs) == 0 {
			return errors.Errorf("0 testcases generated by AI, maybe AI-Proxy is out of service.")
		}
	}

	// AI 生成需求分组结果校验
	if rgs, ok := results.([]RequirementGroup); ok {
		if len(rgs) == 0 {
			return errors.Errorf("0 groups info generated by AI, maybe AI-Proxy is out of service.")
		}

		for _, rmg := range rgs {
			if len(rmg.Groups) == 0 {
				return errors.Errorf("no groups info generated by AI for requirement ID %d, maybe AI-Proxy is out of service.", rmg.ID)
			}
		}
	}

	return nil
}
