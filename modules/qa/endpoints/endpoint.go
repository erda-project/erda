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

package endpoints

import (
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/autotest"
	atv2 "github.com/erda-project/erda/modules/qa/services/autotest_v2"
	"github.com/erda-project/erda/modules/qa/services/cq"
	"github.com/erda-project/erda/modules/qa/services/migrate"
	"github.com/erda-project/erda/modules/qa/services/sceneset"
	"github.com/erda-project/erda/modules/qa/services/sonar_metric_rule"
	"github.com/erda-project/erda/modules/qa/services/testcase"
	"github.com/erda-project/erda/modules/qa/services/testplan"
	"github.com/erda-project/erda/modules/qa/services/testset"
	"github.com/erda-project/erda/pkg/httpserver"
)

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/health", Method: http.MethodGet, Handler: e.Health},

		// gittar
		{Path: "/callback/gittar", Method: http.MethodPost, Handler: e.GittarWebHookCallback},
		{Path: qaGitMRCreateCallback.URLPath, Method: http.MethodPost, Handler: e.GittarMRCreateCallback},

		// sonar issues.
		{Path: "/api/qa/actions/sonar-results-store", Method: http.MethodPost, Handler: e.SonarIssuesStore},
		{Path: "/api/qa", Method: http.MethodGet, Handler: e.SonarIssues},

		// sonar metric key
		{Path: "/api/sonar-metric-rules", Method: http.MethodGet, Handler: e.PagingSonarMetricRules},
		{Path: "/api/sonar-metric-rules/{id}", Method: http.MethodGet, Handler: e.GetSonarMetricRules},
		{Path: "/api/sonar-metric-rules/{id}", Method: http.MethodPut, Handler: e.UpdateSonarMetricRules},
		{Path: "/api/sonar-metric-rules/actions/batch-insert", Method: http.MethodPost, Handler: e.BatchInsertSonarMetricRules},
		{Path: "/api/sonar-metric-rules/actions/batch-delete", Method: http.MethodDelete, Handler: e.BatchDeleteSonarMetricRules},
		{Path: "/api/sonar-metric-rules/{id}", Method: http.MethodDelete, Handler: e.DeleteSonarMetricRules},
		{Path: "/api/sonar-metric-rules/actions/query-metric-definition", Method: http.MethodGet, Handler: e.QuerySonarMetricRulesDefinition},
		{Path: "/api/sonar-metric-rules/actions/query-list", Method: http.MethodGet, Handler: e.QuerySonarMetricRules},

		// test platform
		{Path: "/api/qa/actions/all-test-type", Method: http.MethodGet, Handler: e.GetTestTypes},
		{Path: "/api/qa/actions/test-list", Method: http.MethodGet, Handler: e.GetRecords},
		{Path: "/api/qa/test/{id}", Method: http.MethodGet, Handler: e.GetTestRecord},
		{Path: "/api/qa/actions/test-callback", Method: http.MethodPost, Handler: e.TestCallback},
		{Path: "/api/qa/actions/get-sonar-credential", Method: http.MethodGet, Handler: e.GetSonarCredential},

		// pmp api test
		{Path: "/api/apitests", Method: http.MethodPost, Handler: e.CreateAPITest},
		{Path: "/api/apitests/{id}", Method: http.MethodPut, Handler: e.UpdateApiTest},
		{Path: "/api/apitests/{id}", Method: http.MethodGet, Handler: e.GetApiTests},
		{Path: "/api/apitests/{id}", Method: http.MethodDelete, Handler: e.DeleteApiTestsByApiID},
		{Path: "/api/apitests/actions/list-apis", Method: http.MethodGet, Handler: e.ListApiTests},

		// api test env
		{Path: "/api/testenv", Method: http.MethodPost, Handler: e.CreateAPITestEnv},
		{Path: "/api/testenv/{id}", Method: http.MethodPut, Handler: e.UpdateAPITestEnv},
		{Path: "/api/testenv/{id}", Method: http.MethodGet, Handler: e.GetAPITestEnv},
		{Path: "/api/testenv/{id}", Method: http.MethodDelete, Handler: e.DeleteAPITestEnvByEnvID},
		{Path: "/api/testenv/actions/list-envs", Method: http.MethodGet, Handler: e.ListAPITestEnvs},

		{Path: "/api/apitests/actions/execute-tests", Method: http.MethodPost, Handler: e.ExecuteApiTests},
		{Path: "/api/apitests/actions/cancel-testplan", Method: http.MethodPost, Handler: e.CancelApiTests},
		{Path: "/api/apitests/actions/attempt-test", Method: http.MethodPost, Handler: e.ExecuteAttemptTest},
		{Path: "/api/apitests/actions/statistic-results", Method: http.MethodPost, Handler: e.StatisticResults},
		{Path: "/api/apitests/pipeline/{pipelineID}", Method: http.MethodGet, Handler: e.GetPipelineDetail},
		{Path: "/api/apitests/pipeline/{pipelineID}/task/{taskID}/logs", Method: http.MethodGet, Handler: e.GetPipelineTaskLogs},

		// 测试用例
		{Path: "/api/testcases", Method: http.MethodPost, Handler: e.CreateTestCase},
		{Path: "/api/testcases/{testCaseID}", Method: http.MethodGet, Handler: e.GetTestCase},
		{Path: "/api/testcases/actions/batch-create", Method: http.MethodPost, Handler: e.BatchCreateTestCases},
		{Path: "/api/testcases", Method: http.MethodGet, Handler: e.PagingTestCases},
		{Path: "/api/testcases/{testCaseID}", Method: http.MethodPut, Handler: e.UpdateTestCase},
		{Path: "/api/testcases/actions/batch-update", Method: http.MethodPost, Handler: e.BatchUpdateTestCases},
		{Path: "/api/testcases/actions/batch-copy", Method: http.MethodPost, Handler: e.BatchCopyTestCases},
		{Path: "/api/testcases/actions/batch-clean-from-recycle-bin", Method: http.MethodDelete, Handler: e.BatchCleanTestCasesFromRecycleBin},
		{Path: "/api/testcases/actions/export", Method: http.MethodGet, WriterHandler: e.ExportTestCases},
		{Path: "/api/testcases/actions/import", Method: http.MethodPost, Handler: e.ImportTestCases},

		// 测试集 管理
		{Path: "/api/testsets", Method: http.MethodPost, Handler: e.CreateTestSet},
		{Path: "/api/testsets", Method: http.MethodGet, Handler: e.ListTestSets},
		{Path: "/api/testsets/{testSetID}", Method: http.MethodGet, Handler: e.GetTestSet},
		{Path: "/api/testsets/{testSetID}", Method: http.MethodPut, Handler: e.UpdateTestSet},
		{Path: "/api/testsets/{testSetID}/actions/copy", Method: http.MethodPost, Handler: e.CopyTestSet},
		{Path: "/api/testsets/{testSetID}/actions/recycle", Method: http.MethodPost, Handler: e.RecycleTestSet},
		{Path: "/api/testsets/{testSetID}/actions/clean-from-recycle-bin", Method: http.MethodDelete, Handler: e.CleanTestSetFromRecycleBin},
		{Path: "/api/testsets/{testSetID}/actions/recover-from-recycle-bin", Method: http.MethodPost, Handler: e.RecoverTestSetFromRecycleBin},

		// 测试计划
		{Path: "/api/testplans", Method: http.MethodPost, Handler: e.CreateTestPlan},
		{Path: "/api/testplans", Method: http.MethodGet, Handler: e.PagingTestPlans},
		{Path: "/api/testplans/{testPlanID}", Method: http.MethodGet, Handler: e.GetTestPlan},
		{Path: "/api/testplans/{testPlanID}", Method: http.MethodPut, Handler: e.UpdateTestPlan},
		{Path: "/api/testplans/{testPlanID}", Method: http.MethodDelete, Handler: e.DeleteTestPlan},
		{Path: "/api/testplans/{testPlanID}/testcase-relations", Method: http.MethodPost, Handler: e.CreateTestPlanCaseRelations},
		{Path: "/api/testplans/{testPlanID}/testcase-relations", Method: http.MethodGet, Handler: e.PagingTestPlanCaseRelations},
		{Path: "/api/testplans/testcase-relations/actions/internal-list", Method: http.MethodGet, Handler: e.InternalListTestPlanCaseRels},
		{Path: "/api/testplans/{testPlanID}/testcase-relations/{relationID}", Method: http.MethodGet, Handler: e.GetTestPlanCaseRel},
		{Path: "/api/testplans/{testPlanID}/testcase-relations/{relationID}/actions/add-issue-relations", Method: http.MethodPost, Handler: e.AddTestPlanCaseRelIssueRelations},
		{Path: "/api/testplans/{testPlanID}/testcase-relations/{relationID}/actions/remove-issue-relations", Method: http.MethodPost, Handler: e.RemoveTestPlanCaseRelIssueRelations},
		{Path: "/api/testplans/testcase-relations/actions/internal-remove-issue-relations", Method: http.MethodDelete, Handler: e.InternalRemoveTestPlanCaseRelIssueRelations},
		{Path: "/api/testplans/{testPlanID}/testcase-relations/actions/batch-update", Method: http.MethodPost, Handler: e.BatchUpdateTestPlanCaseRelations},
		{Path: "/api/testplans/{testPlanID}/actions/execute-apitest", Method: http.MethodPost, Handler: e.ExecuteTestPlanAPITest},
		{Path: "/api/testplans/{testPlanID}/actions/cancel-apitest/{pipelineID}", Method: http.MethodPost, Handler: e.CancelApiTestPipeline},
		{Path: "/api/testplans/{testPlanID}/actions/export", Method: http.MethodGet, WriterHandler: e.ExportTestPlanCaseRels},
		{Path: "/api/testplans/{testPlanID}/testsets", Method: http.MethodGet, Handler: e.ListTestPlanTestSets},
		{Path: "/api/testplans/{testPlanID}/actions/generate-report", Method: http.MethodGet, Handler: e.GenerateTestPlanReport},

		// 自动化测试 - 测试集
		{Path: "/api/autotests/filetree", Method: http.MethodPost, Handler: e.CreateAutoTestFileTreeNode},
		{Path: "/api/autotests/filetree", Method: http.MethodGet, Handler: e.ListAutoTestFileTreeNodes},
		{Path: "/api/autotests/filetree/actions/fuzzy-search", Method: http.MethodGet, Handler: e.FuzzySearchAutoTestFileTreeNodes},
		{Path: "/api/autotests/filetree/{inode}", Method: http.MethodDelete, Handler: e.DeleteAutoTestFileTreeNode},
		{Path: "/api/autotests/filetree/{inode}", Method: http.MethodPut, Handler: e.UpdateAutoTestFileTreeNodeBasicInfo},
		{Path: "/api/autotests/filetree/{inode}", Method: http.MethodGet, Handler: e.GetAutoTestFileTreeNode},
		{Path: "/api/autotests/filetree/{inode}/actions/move", Method: http.MethodPost, Handler: e.MoveAutoTestFileTreeNode},
		{Path: "/api/autotests/filetree/{inode}/actions/copy", Method: http.MethodPost, Handler: e.CopyAutoTestFileTreeNode},
		{Path: "/api/autotests/filetree/{inode}/actions/find-ancestors", Method: http.MethodGet, Handler: e.FindAutoTestFileTreeNodeAncestors},
		{Path: "/api/autotests/filetree/{inode}/actions/save-pipeline", Method: http.MethodPost, Handler: e.SaveAutoTestFileTreeNodePipeline},
		{Path: "/api/autotests/filetree/{inode}/actions/get-histories", Method: http.MethodGet, Handler: e.ListAutoTestFileTreeNodeHistory},
		{Path: "/api/autotests/pipeline-snippets/actions/query-snippet-yml", Method: http.MethodPost, Handler: e.QueryPipelineSnippetYaml},
		{Path: "/api/autotests/pipeline-snippets/actions/batch-query-snippet-yml", Method: http.MethodPost, Handler: e.BatchQueryPipelineSnippetYaml},
		{Path: "/api/autotests/global-configs", Method: http.MethodPost, Handler: e.CreateAutoTestGlobalConfig},
		{Path: "/api/autotests/global-configs/{ns}", Method: http.MethodPut, Handler: e.UpdateAutoTestGlobalConfig},
		{Path: "/api/autotests/global-configs/{ns}", Method: http.MethodDelete, Handler: e.DeleteAutoTestGlobalConfig},
		{Path: "/api/autotests/global-configs", Method: http.MethodGet, Handler: e.ListAutoTestGlobalConfigs},

		// 自动化测试 - 测试空间
		{Path: "/api/autotests/spaces", Method: http.MethodPost, Handler: e.CreateAutoTestSpace},
		{Path: "/api/autotests/spaces", Method: http.MethodPut, Handler: e.UpdateAutoTestSpace},
		{Path: "/api/autotests/spaces", Method: http.MethodGet, Handler: e.GetAutoTestSpaceList},
		{Path: "/api/autotests/spaces/{id}", Method: http.MethodGet, Handler: e.GetAutoTestSpace},
		{Path: "/api/autotests/spaces/{id}", Method: http.MethodDelete, Handler: e.DeleteAutoTestSpace},
		{Path: "/api/autotests/spaces/actions/copy", Method: http.MethodPost, Handler: e.CopyAutoTestSpace},

		// 自动化测试 - 场景
		{Path: "/api/autotests/scenes", Method: http.MethodPost, Handler: e.CreateAutoTestScene},
		{Path: "/api/autotests/scenes/{sceneID}", Method: http.MethodPut, Handler: e.UpdateAutoTestScene},
		{Path: "/api/autotests/scenes/actions/move-scene", Method: http.MethodPut, Handler: e.MoveAutoTestScene},
		{Path: "/api/autotests/scenes", Method: http.MethodGet, Handler: e.ListAutoTestScene},
		{Path: "/api/autotests/scenes/modal", Method: http.MethodGet, Handler: e.ListAutoTestScenes},
		{Path: "/api/autotests/scenes/{sceneID}", Method: http.MethodGet, Handler: e.GetAutoTestScene},
		{Path: "/api/autotests/scenes/{sceneID}", Method: http.MethodDelete, Handler: e.DeleteAutoTestScene},
		{Path: "/api/autotests/scenes/actions/copy", Method: http.MethodPost, Handler: e.CopyAutoTestScene},

		// 自动化测试 - 入参
		{Path: "/api/autotests/scenes/{sceneID}/actions/add-input", Method: http.MethodPost, Handler: e.CreateAutoTestSceneInput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/delete-input", Method: http.MethodDelete, Handler: e.DeleteAutoTestSceneInput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/update-input", Method: http.MethodPut, Handler: e.UpdateAutoTestSceneInput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/list-input", Method: http.MethodGet, Handler: e.ListAutoTestSceneInput},

		// 自动化测试 - 出参
		{Path: "/api/autotests/scenes/{sceneID}/actions/add-output", Method: http.MethodPost, Handler: e.CreateAutoTestSceneOutput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/delete-output", Method: http.MethodDelete, Handler: e.DeleteAutoTestSceneOutput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/update-output", Method: http.MethodPut, Handler: e.UpdateAutoTestSceneOutput},
		{Path: "/api/autotests/scenes/{sceneID}/actions/list-output", Method: http.MethodGet, Handler: e.ListAutoTestSceneOutput},

		// 自动化测试 - 步骤
		{Path: "/api/autotests/scenes/{sceneID}/actions/add-step", Method: http.MethodPost, Handler: e.CreateAutoTestSceneStep},
		{Path: "/api/autotests/scenes-step/{stepID}", Method: http.MethodDelete, Handler: e.DeleteAutoTestSceneStep},
		{Path: "/api/autotests/scenes-step/{stepID}", Method: http.MethodPut, Handler: e.UpdateAutoTestSceneStep},
		{Path: "/api/autotests/scenes-step/actions/move", Method: http.MethodPut, Handler: e.MoveAutoTestSceneStep},
		{Path: "/api/autotests/scenes/{sceneID}/actions/get-step", Method: http.MethodGet, Handler: e.ListAutoTestSceneStep},
		{Path: "/api/autotests/scenes-step-output", Method: http.MethodGet, Handler: e.ListAutoTestSceneStepOutPut},
		{Path: "/api/autotests/scenes-step/{stepID}", Method: http.MethodGet, Handler: e.GetAutoTestSceneStep},

		// 自动化测试 - 测试计划
		{Path: "/api/autotests/testplans", Method: http.MethodPost, Handler: e.CreateTestPlanV2},
		{Path: "/api/autotests/testplans/{testPlanID}", Method: http.MethodDelete, Handler: e.DeleteTestPlanV2},
		{Path: "/api/autotests/testplans/{testPlanID}", Method: http.MethodPut, Handler: e.UpdateTestPlanV2},
		{Path: "/api/autotests/testplans", Method: http.MethodGet, Handler: e.PagingTestPlansV2},
		{Path: "/api/autotests/testplans/{testPlanID}", Method: http.MethodGet, Handler: e.GetTestPlanV2},
		{Path: "/api/autotests/testplans/{testPlanID}/actions/add-step", Method: http.MethodPost, Handler: e.AddTestPlanV2Step},
		{Path: "/api/autotests/testplans/{testPlanID}/actions/delete-step", Method: http.MethodDelete, Handler: e.DeleteTestPlanV2Step},
		{Path: "/api/autotests/testplans/{testPlanID}/actions/move-step", Method: http.MethodPut, Handler: e.MoveTestPlanV2Step},
		{Path: "/api/autotests/testplans-step/{stepID}", Method: http.MethodGet, Handler: e.GetTestPlanV2Step},
		{Path: "/api/autotests/testplans-step/{stepID}", Method: http.MethodPut, Handler: e.UpdateTestPlanV2Step},

		{Path: "/api/reportsets/{pipelineID}", Method: http.MethodGet, Handler: e.queryReportSets},

		// 场景 执行取消
		{Path: "/api/autotests/scenes-step/{stepID}/actions/execute", Method: http.MethodPost, Handler: e.ExecuteDiceAutotestSceneStep},
		{Path: "/api/autotests/scenes/{sceneID}/actions/execute", Method: http.MethodPost, Handler: e.ExecuteDiceAutotestScene},
		{Path: "/api/autotests/scenes/{sceneID}/actions/cancel", Method: http.MethodPost, Handler: e.CancelDiceAutotestScene},

		// 计划 执行取消
		{Path: "/api/autotests/testplans/{testPlanID}/actions/execute", Method: http.MethodPost, Handler: e.ExecuteDiceAutotestTestPlans},
		{Path: "/api/autotests/testplans/{testPlanID}/actions/cancel", Method: http.MethodPost, Handler: e.CancelDiceAutotestTestPlans},

		// 自动化测试v2
		//{Path: "/api/autotests/testplans/actions/query-snippet-yml", Method: http.MethodPost, Handler: e.QueryPipelineSnippetYamlV2},

		//场景集
		{Path: "/api/autotests/scenesets/{setID}", Method: http.MethodGet, Handler: e.GetSceneSet},
		{Path: "/api/autotests/scenesets", Method: http.MethodGet, Handler: e.GetSceneSets},
		{Path: "/api/autotests/scenesets", Method: http.MethodPost, Handler: e.CreateSceneSet},
		{Path: "/api/autotests/scenesets/{setID}", Method: http.MethodPut, Handler: e.UpdateSceneSet},
		{Path: "/api/autotests/scenesets/{setID}", Method: http.MethodDelete, Handler: e.DeleteSceneSet},
		{Path: "/api/autotests/scenesets/actions/drag", Method: http.MethodPut, Handler: e.DragSceneSet},
		{Path: "/api/autotests/scenesets/actions/copy", Method: http.MethodPost, Handler: e.CopySceneSet},

		// migrate
		{Path: "/api/autotests/actions/migrate-from-autotestv1", Method: http.MethodGet, Handler: e.MigrateFromAutoTestV1},
	}
}

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	queryStringDecoder *schema.Decoder
	bdl                *bundle.Bundle
	db                 *dao.DBClient

	testcase *testcase.Service
	testset  *testset.Service
	testPlan *testplan.TestPlan

	autotest   *autotest.Service
	autotestV2 *atv2.Service

	sonarMetricRule *sonar_metric_rule.Service

	cq *cq.CQ

	sceneset *sceneset.Service

	migrate *migrate.Service
}

type Option func(*Endpoints)

func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

func WithDB(db *dao.DBClient) Option {
	return func(e *Endpoints) {
		e.db = db
	}
}

func WithTestcase(svc *testcase.Service) Option {
	return func(e *Endpoints) {
		e.testcase = svc
	}
}

func WithTestSet(svc *testset.Service) Option {
	return func(e *Endpoints) {
		e.testset = svc
	}
}

func WithSonarMetricRule(sonarMetricRule *sonar_metric_rule.Service) Option {
	return func(e *Endpoints) {
		e.sonarMetricRule = sonarMetricRule
	}
}

// WithTestplan 设置 testplan endpoint
func WithTestplan(testPlan *testplan.TestPlan) Option {
	return func(e *Endpoints) {
		e.testPlan = testPlan
	}
}

func WithCQ(cq *cq.CQ) Option {
	return func(e *Endpoints) {
		e.cq = cq
	}
}

func WithAutoTest(svc *autotest.Service) Option {
	return func(e *Endpoints) {
		e.autotest = svc
	}
}

func WithAutoTestV2(svc *atv2.Service) Option {
	return func(e *Endpoints) {
		e.autotestV2 = svc
	}
}

func WithSceneSet(svc *sceneset.Service) Option {
	return func(e *Endpoints) {
		e.sceneset = svc
	}
}

func WithMigrate(svc *migrate.Service) Option {
	return func(e *Endpoints) {
		e.migrate = svc
	}
}
