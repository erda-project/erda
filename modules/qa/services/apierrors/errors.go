// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package apierrors 定义了错误列表
package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

const (
	MissingRequestBody = "request body"
)

var (
	ErrCheckPermission = err("ErrCheckPermission", "权限校验失败")

	ErrDoGittarWebHookCallback = err("ErrDoGittarWebHookCallback", "处理 Gittar WebHook 回调失败")
	ErrDoGitMrCreateCallback   = err("ErrDoGitMrCreateCallback", "处理 Gittar MR 创建 Webhook 失败")
	ErrDoTestCallback          = err("ErrDoTestCallback", "测试回调失败")

	ErrPagingTestRecords = err("ErrPagingTestRecords", "测试记录分页查询失败")
	ErrGetTestRecord     = err("ErrGetTestRecord", "查询测试记录详情失败")

	ErrCreateAPITestEnv = err("ErrCreateAPITestEnv", "创建接口测试环境失败")
	ErrUpdateAPITestEnv = err("ErrUpdateAPITestEnv", "更新接口测试环境失败")
	ErrGetAPITestEnv    = err("ErrGetAPITestEnv", "查询接口测试环境失败")
	ErrListAPITestEnvs  = err("ErrListAPITestEnvs", "查询接口测试环境列表失败")
	ErrDeleteAPITestEnv = err("ErrDeleteAPITestEnv", "删除接口测试环境失败")

	ErrCreateAPITest         = err("ErrCreateAPITest", "创建接口测试失败")
	ErrUpdateAPITest         = err("ErrUpdateAPITest", "更新接口测试失败")
	ErrGetAPITest            = err("ErrGetAPITest", "查询接口测试失败")
	ErrListAPITests          = err("ErrListAPITests", "查询接口测试列表失败")
	ErrDeleteAPITest         = err("ErrDeleteAPITest", "删除接口测试失败")
	ErrExecuteAPITest        = err("ErrExecuteAPITest", "执行接口测试失败")
	ErrAttemptExecuteAPITest = err("ErrAttemptExecuteAPITest", "尝试执行接口测试失败")
	ErrCancelAPITests        = err("ErrCancelAPITests", "取消执行测试计划失败")
	ErrGetStatisticResults   = err("ErrGetStatisticResults", "查询 API 测试结果统计失败")

	ErrGetPipelineDetail = err("ErrGetPipelineDetail", "查询流水线详情失败")
	ErrGetPipelineLog    = err("ErrGetPipelineLog", "查询流水线日志失败")

	ErrStoreSonarIssue = err("ErrStoreSonarIssue", "保存 Sonar 分析结果失败")
	ErrGetSonarIssue   = err("ErrGetSonarIssue", "查询 Sonar 分析结果失败")

	ErrPagingTestCases                   = err("ErrPagingTestCases", "分页查询测试用例失败")
	ErrListTestCases                     = err("ErrListTestCases", "获取测试用例列表失败")
	ErrGetTestCase                       = err("ErrGetTestCase", "获取指定测试用例失败")
	ErrCreateTestCase                    = err("ErrCreateTestCase", "创建测试用例失败")
	ErrBatchCreateTestCases              = err("ErrBatchCreateTestCases", "批量创建测试用例失败")
	ErrUpdateTestCase                    = err("ErrUpdateTestCase", "更新测试用例失败")
	ErrBatchUpdateTestCases              = err("ErrBatchUpdateTestCases", "批量更新测试用例失败")
	ErrBatchCopyTestCases                = err("ErrBatchCopyTestCases", "批量复制测试用例失败")
	ErrDeleteTestCase                    = err("ErrDeleteTestCase", "删除测试用例失败")
	ErrExportTestCases                   = err("ErrExportTestCases", "导出测试用例失败")
	ErrImportTestCases                   = err("ErrImportTestCases", "导入测试用例失败")
	ErrInvalidTestCaseExcelFormat        = err("ErrInvalidTestCaseExcelFormat", "文件格式不正确，请对比 Excel 导入模板")
	ErrGetApiTestInfo                    = err("ErrErrGetApiTestInfo", "查询接口测试信息失败")
	ErrBatchCleanTestCasesFromRecycleBin = err("ErrBatchCleanTestCasesFromRecycleBin", "从回收站批量删除测试用例失败")
	ErrExportTestPlanCaseRels            = err("ErrExportTestPlanCaseRels", "导出测试计划下的测试用例失败")
	ErrGenerateTestPlanReport            = err("ErrGenerateTestPlanReport", "生成测试计划报告失败")
	ErrExecuteTestPlanReport             = err("ErrExecuteTestPlanReport", "执行测试计划失败")
	ErrCancelTestPlanReport              = err("ErrCancelTestPlanReport", "取消执行测试计划失败")

	ErrListTestSets                 = err("ErrListTestSets", "获取测试集列表失败")
	ErrCreateTestSet                = err("ErrCreateTestSet", "创建测试集失败")
	ErrUpdateTestSet                = err("ErrUpdateTestSet", "更新测试集失败")
	ErrDeleteTestSet                = err("ErrDeleteTestSet", "删除测试集失败")
	ErrCopyTestSet                  = err("ErrCopyTestSet", "复制测试集失败")
	ErrGetTestSet                   = err("ErrGetTestSet", "获取指定测试集失败")
	ErrRecycleTestSet               = err("ErrRecycleTestSet", "回收测试集失败")
	ErrCleanTestSetFromRecycleBin   = err("ErrCleanTestSetFromRecycleBin", "从回收站彻底删除测试集失败")
	ErrRecoverTestSetFromRecycleBin = err("ErrRecoverTestSetFromRecycleBin", "从回收站恢复测试集失败")

	ErrCreateTestPlan                     = err("ErrCreateTestPlan", "创建测试计划失败")
	ErrUpdateTestPlan                     = err("ErrUpdateTestPlan", "更新测试计划失败")
	ErrDeleteTestPlan                     = err("ErrDeleteTestPlan", "删除测试计划失败")
	ErrGetTestPlan                        = err("ErrGetTestPlan", "获取测试计划详情失败")
	ErrAddTestPlanStep                    = err("ErrAddTestPlanStep", "添加测试计划步骤失败")
	ErrDeleteTestPlanStep                 = err("ErrDeleteTestPlanStep", "删除测试计划步骤失败")
	ErrUpdateTestPlanStep                 = err("ErrUpdateTestPlanStep", "更新测试计划步骤失败")
	ErrCreateTestPlanMember               = err("ErrCreateTestPlanMember", "测试计划关联成员失败")
	ErrUpdateTestPlanMember               = err("ErrUpdateTestPlanMember", "测试计划更新成员失败")
	ErrListTestPlanMembers                = err("ErrListTestPlanMembers", "查询测试计划关联成员列表失败")
	ErrPagingTestPlans                    = err("ErrPagingTestPlans", "分页查询测试计划失败")
	ErrPagingTestPlanCaseRels             = err("ErrPagingTestPlanCaseRels", "获取测试计划内测试用例列表失败")
	ErrTestPlanExecuteAPITest             = err("ErrTestPlanExecuteAPITest", "执行测试计划接口测试失败")
	ErrTestPlanCancelAPITest              = err("ErrTestPlanCancelAPITest", "取消测试计划接口测试失败")
	ErrCreateTestPlanCaseRel              = err("ErrCreateTestPlanCaseRel", "引用测试用例失败")
	ErrBatchUpdateTestPlanCaseRels        = err("ErrBatchUpdateTestPlanCaseRels", "批量更新测试用例引用失败")
	ErrRemoveTestPlanCaseRelIssueRelation = err("ErrRemoveTestPlanCaseRelIssueRelation", "解除测试计划用例与缺陷关联关系失败")
	ErrAddTestPlanCaseRelIssueRelation    = err("ErrAddTestPlanCaseRelIssueRelation", "新增测试计划用例与缺陷关联关系失败")
	ErrDeleteTestPlanUsecaseRel           = err("ErrDeleteTestPlanUsecaseRel", "删除测试用例引用失败")
	ErrGetTestPlanCaseRel                 = err("ErrGetTestPlanCaseRel", "查询测试计划引用失败")
	ErrUpdateTestPlanCaseRel              = err("ErrUpdateTestPlanCaseRel", "更新测试计划引用失败")
	ErrListTestPlanTestSets               = err("ErrListTestPlanTestSets", "获取测试计划下的测试集列表失败")

	ErrCreateIssueRelation         = err("ErrCreateIssueRelation", "添加关联事件失败")
	ErrGetIssueRelations           = err("ErrGetIssueRelations", "查看关联事件失败")
	ErrDeleteIssueRelation         = err("ErrDeleteIssueRelation", "删除关联事件失败")
	ErrBatchCreateIssueTestCaseRel = err("ErrBatchCreateIssueTestCaseRel", "事件批量关联测试计划用例失败")
	ErrDeleteIssueTestCaseRel      = err("ErrDeleteIssueTestCaseRel", "事件取消关联测试计划用例失败")
	ErrListIssueTestCaseRels       = err("ErrListIssueTestCaseRels", "查询事件用例关联列表失败")

	ErrCreateAutoTestFileTreeNode        = err("ErrCreateAutoTestFileTreeNode", "创建自动化测试目录树节点失败")
	ErrDeleteAutoTestFileTreeNode        = err("ErrDeleteAutoTestFileTreeNode", "删除自动化测试目录树节点失败")
	ErrUpdateAutoTestSetBasicInfo        = err("ErrUpdateAutoTestSetBasicInfo", "更新自动化测试目录树节点基础信息失败")
	ErrMoveAutoTestFileTreeNode          = err("ErrMoveAutoTestFileTreeNode", "移动自动化测试目录树节点失败")
	ErrCopyAutoTestFileTreeNode          = err("ErrCopyAutoTestFileTreeNode", "复制自动化测试目录树节点失败")
	ErrGetAutoTestFileTreeNode           = err("ErrGetAutoTestFileTreeNode", "查询自动化测试目录树节点详情失败")
	ErrListAutoTestFileTreeNodes         = err("ErrListAutoTestFileTreeNodes", "查询自动化测试目录树节点列表失败")
	ErrListAutoTestFileTreeNodeHistory   = err("ErrListAutoTestFileTreeNodeHistory", "查询自动化测试目录树节点历史列表失败")
	ErrFuzzySearchAutoTestFileTreeNodes  = err("ErrFuzzySearchAutoTestFileTreeNodes", "模糊搜索自动化测试目录树节点失败")
	ErrQueryPipelineSnippetYaml          = err("ErrQueryPipelineSnippetYaml", "查询自动化测试用例流水线文件失败")
	ErrSaveAutoTestFileTreeNodePipeline  = err("ErrSaveAutoTestFileTreeNodePipeline", "保存自动化测试用例流水线失败")
	ErrFindAutoTestFileTreeNodeAncestors = err("ErrFindAutoTestFileTreeNodeAncestors", "自动化测试目录树节点寻祖失败")
	ErrCreateAutoTestGlobalConfig        = err("ErrCreateAutoTestGlobalConfig", "创建自动化测试全局配置失败")
	ErrUpdateAutoTestGlobalConfig        = err("ErrUpdateAutoTestGlobalConfig", "更新自动化测试全局配置失败")
	ErrDeleteAutoTestGlobalConfig        = err("ErrDeleteAutoTestGlobalConfig", "删除自动化测试全局配置失败")
	ErrListAutoTestGlobalConfigs         = err("ErrListAutoTestGlobalConfigs", "查询自动化测试全局配置列表失败")

	ErrCreateAutoTestSpace = err("ErrCreateAutoTestSpace", "创建自动化测试空间失败")
	ErrUpdateAutoTestSpace = err("ErrUpdateAutoTestSpace", "更新自动化测试空间失败")
	ErrDeleteAutoTestSpace = err("ErrDeleteAutoTestSpace", "删除自动化测试空间失败")
	ErrCopyAutoTestSpace   = err("ErrCopyAutoTestSpace", "复制自动化测试空间失败")
	ErrGetAutoTestSpace    = err("ErrGetAutoTestSpace", "获取自动化测试空间失败")
	ErrListAutoTestSpace   = err("ErrListAutoTestSpace", "获取自动化测试空间列表失败")

	ErrCreateAutoTestScene      = err("ErrCreateAutoTestScene", "创建自动化测试场景失败")
	ErrUpdateAutoTestScene      = err("ErrUpdateAutoTestScene", "更新自动化测试场景失败")
	ErrDeleteAutoTestScene      = err("ErrDeleteAutoTestScene", "删除自动化测试场景失败")
	ErrGetAutoTestScene         = err("ErrGetAutoTestScene", "获取自动化测试场景失败")
	ErrListAutoTestScene        = err("ErrListAutoTestScene", "获取自动化测试场景列表失败")
	ErrExecuteAutoTestScene     = err("ErrExecuteAutoTestScene", "执行自动化测试场景失败")
	ErrExecuteAutoTestSceneStep = err("ErrExecuteAutoTestSceneStep", "执行自动化测试场景步骤失败")
	ErrCancelAutoTestScene      = err("ErrCancelAutoTestScene", "取消执行自动化测试场景失败")
	ErrMoveAutoTestScene        = err("ErrMoveAutoTestScene", "拖动自动化测试场景失败")
	ErrCopyAutoTestScene        = err("ErrCopyAutoTestScene", "复制自动化测试场景失败")

	ErrCreateAutoTestSceneInput = err("ErrCreateAutoTestSceneInput", "创建自动化测试场景入参失败")
	ErrUpdateAutoTestSceneInput = err("ErrUpdateAutoTestSceneInput", "更新自动化测试场景入参失败")
	ErrDeleteAutoTestSceneInput = err("ErrDeleteAutoTestSceneInput", "删除自动化测试场景入参失败")
	ErrListAutoTestSceneInput   = err("ErrListAutoTestSceneInput", "获取自动化测试场景入参列表失败")

	ErrCreateAutoTestSceneOutput = err("ErrCreateAutoTestSceneOutput", "创建自动化测试场景出参失败")
	ErrUpdateAutoTestSceneOutput = err("ErrUpdateAutoTestSceneOutput", "更新自动化测试场景出参失败")
	ErrDeleteAutoTestSceneOutput = err("ErrDeleteAutoTestSceneOutput", "删除自动化测试场景出参失败")
	ErrListAutoTestSceneOutput   = err("ErrListAutoTestSceneOutput", "获取自动化测试场景出参列表失败")

	ErrCreateAutoTestSceneStep     = err("ErrCreateAutoTestSceneStep", "创建自动化测试场景步骤失败")
	ErrUpdateAutoTestSceneStep     = err("ErrUpdateAutoTestSceneStep", "更新自动化测试场景步骤失败")
	ErrDeleteAutoTestSceneStep     = err("ErrDeleteAutoTestSceneStep", "删除自动化测试场景步骤失败")
	ErrListAutoTestSceneStep       = err("ErrListAutoTestSceneStep", "获取自动化测试场景步骤失败")
	ErrListAutoTestSceneStepOutPut = err("ErrListAutoTestSceneStepOutPut", "获取自动化测试场景步骤出参失败")

	ErrPagingSonarMetricRules          = err("ErrPagingSonarMetricRules", "分页查询指标规则失败")
	ErrQuerySonarMetricRules           = err("ErrQuerySonarMetricRules", "查询指标规则失败")
	ErrBatchCreateSonarMetricRules     = err("ErrBatchCreateSonarMetricRules", "批量创建指标规则失败")
	ErrUpdateSonarMetricRules          = err("ErrUpdateSonarMetricRules", "更新指标规则失败")
	ErrDeleteSonarMetricRules          = err("ErrDeleteSonarMetricRules", "删除指标规则失败")
	ErrQuerySonarMetricRuleDefinitions = err("ErrQuerySonarMetricRuleDefinitions", "查询未添加的指标规则失败")

	ErrCreateAutoTestSceneSet = err("ErrCreateAutoTestSceneSet", "创建自动化测试场景集失败")
	ErrUpdateAutoTestSceneSet = err("ErrUpdateAutoTestSceneSet", "更新自动化测试场景集失败")
	ErrDeleteAutoTestSceneSet = err("ErrDeleteAutoTestSceneSet", "删除自动化测试场景集失败")
	ErrGetAutoTestSceneSet    = err("ErrGetAutoTestSceneSet", "获取自动化测试场景集失败")
	ErrListAutoTestSceneSet   = err("ErrListAutoTestSceneSet", "获取自动化测试场景集列表失败")
	ErrDragAutoTestSceneSet   = err("ErrDragAutoTestSceneSet", "拖动自动化测试场景集失败")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
