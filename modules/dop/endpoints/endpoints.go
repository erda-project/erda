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

// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/event"
	"github.com/erda-project/erda/modules/dop/services/apidocsvc"
	"github.com/erda-project/erda/modules/dop/services/assetsvc"
	"github.com/erda-project/erda/modules/dop/services/autotest"
	atv2 "github.com/erda-project/erda/modules/dop/services/autotest_v2"
	"github.com/erda-project/erda/modules/dop/services/cdp"
	"github.com/erda-project/erda/modules/dop/services/cq"
	"github.com/erda-project/erda/modules/dop/services/filetree"
	"github.com/erda-project/erda/modules/dop/services/migrate"
	"github.com/erda-project/erda/modules/dop/services/permission"
	"github.com/erda-project/erda/modules/dop/services/pipeline"
	"github.com/erda-project/erda/modules/dop/services/sceneset"
	"github.com/erda-project/erda/modules/dop/services/sonar_metric_rule"
	"github.com/erda-project/erda/modules/dop/services/testcase"
	"github.com/erda-project/erda/modules/dop/services/testplan"
	"github.com/erda-project/erda/modules/dop/services/testset"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// ReleaseCallbackPath ReleaseCallback的路径
	ReleaseCallbackPath     = "/api/actions/release-callback"
	CDPCallbackPath         = "/api/actions/cdp-callback"
	GitCreateMrCallback     = "/api/actions/git-mr-create-callback"
	GitMergeMrCallback      = "/api/actions/git-mr-merge-callback"
	GitCloseMrCallback      = "/api/actions/git-mr-close-callback"
	GitCommentMrCallback    = "/api/actions/git-mr-comment-callback"
	GitDeleteBranchCallback = "/api/actions/git-branch-delete-callback"
	GitDeleteTagCallback    = "/api/actions/git-tag-delete-callback"
	IssueCallback           = "/api/actions/issue-callback"
	MrCheckRunCallback      = "/api/actions/check-run-callback"
)

type EventCallback struct {
	Name   string
	Path   string
	Events []string
}

var eventCallbacks = []EventCallback{
	{Name: "git_push_release", Path: ReleaseCallbackPath, Events: []string{"git_push"}},
	{Name: "cdp_pipeline", Path: CDPCallbackPath, Events: []string{"pipeline"}},
	{Name: "git_create_mr", Path: GitCreateMrCallback, Events: []string{"git_create_mr"}},
	{Name: "git_merge_mr", Path: GitMergeMrCallback, Events: []string{"git_merge_mr"}},
	{Name: "git_close_mr", Path: GitCloseMrCallback, Events: []string{"git_close_mr"}},
	{Name: "git_comment_mr", Path: GitCommentMrCallback, Events: []string{"git_comment_mr"}},
	{Name: "git_delete_branch", Path: GitDeleteBranchCallback, Events: []string{"git_delete_branch"}},
	{Name: "git_delete_tag", Path: GitDeleteTagCallback, Events: []string{"git_delete_tag"}},
	{Name: "issue", Path: IssueCallback, Events: []string{"issue"}},
	{Name: "check-run", Path: MrCheckRunCallback, Events: []string{"check-run"}},
	{Name: "qa_git_mr_create", Path: "/api/callbacks/git-mr-create", Events: []string{"git_create_mr"}},
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/health", Method: http.MethodGet, Handler: e.Health},

		{Path: "/api/api-assets", Method: http.MethodPost, Handler: e.CreateAPIAsset},
		{Path: "/api/api-assets", Method: http.MethodGet, Handler: e.PagingAPIAssets},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodGet, Handler: e.GetAPIAsset},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodPut, Handler: e.UpdateAPIAsset},
		{Path: "/api/api-assets/{assetID}", Method: http.MethodDelete, Handler: e.DeleteAPIAsset},

		{Path: "/api/api-assets/{assetID}/api-gateways", Method: http.MethodGet, Handler: e.ListAPIGateways},
		{Path: "/api/api-gateways/{projectID}", Method: http.MethodGet, Handler: e.ListProjectAPIGateways},

		{Path: "/api/api-assets/{assetID}/versions", Method: http.MethodGet, Handler: e.PagingAPIAssetVersions},
		{Path: "/api/api-assets/{assetID}/versions", Method: http.MethodPost, Handler: e.CreateAPIVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodGet, Handler: e.GetAPIAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodPut, Handler: e.UpdateAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}", Method: http.MethodDelete, Handler: e.DeleteAPIAssetVersion},
		{Path: "/api/api-assets/{assetID}/versions/{versionID}/export", Method: http.MethodGet, WriterHandler: e.DownloadSpecText},

		{Path: "/api/api-assets/{assetID}/swagger-versions", Method: http.MethodGet, Handler: e.ListSwaggerVersions},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/clients", Method: http.MethodGet, Handler: e.ListSwaggerClient},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/attempt-test", Method: http.MethodPost, Handler: e.ExecuteAttemptTest},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations", Method: http.MethodPost, Handler: e.CreateInstantiation},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations", Method: http.MethodGet, Handler: e.GetInstantiations},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/minors/{minor}/instantiations/{instantiationID}", Method: http.MethodPut, Handler: e.UpdateInstantiation},

		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas", Method: http.MethodGet, Handler: e.ListSLAs},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas", Method: http.MethodPost, Handler: e.CreateSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodGet, Handler: e.GetSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodDelete, Handler: e.DeleteSLA},
		{Path: "/api/api-assets/{assetID}/swagger-versions/{swaggerVersion}/slas/{slaID}", Method: http.MethodPut, Handler: e.UpdateSLA},

		{Path: "/api/api-clients", Method: http.MethodPost, Handler: e.CreateClient},
		{Path: "/api/api-clients", Method: http.MethodGet, Handler: e.ListMyClients},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodGet, Handler: e.GetClient},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodPut, Handler: e.UpdateClient},
		{Path: "/api/api-clients/{clientID}", Method: http.MethodDelete, Handler: e.DeleteClient},

		{Path: "/api/api-clients/{clientID}/contracts", Method: http.MethodPost, Handler: e.CreateContract},
		{Path: "/api/api-clients/{clientID}/contracts", Method: http.MethodGet, Handler: e.ListContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodGet, Handler: e.GetContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodPut, Handler: e.UpdateContract},
		{Path: "/api/api-clients/{clientID}/contracts/{contractID}", Method: http.MethodDelete, Handler: e.DeleteContract},

		{Path: "/api/api-clients/{clientID}/contracts/{contractID}/operation-records", Method: http.MethodGet, Handler: e.ListContractRecords},

		{Path: "/api/api-access", Method: http.MethodPost, Handler: e.CreateAccess},
		{Path: "/api/api-access", Method: http.MethodGet, Handler: e.ListAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodGet, Handler: e.GetAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodPut, Handler: e.UpdateAccess},
		{Path: "/api/api-access/{accessID}", Method: http.MethodDelete, Handler: e.DeleteAccess},

		{Path: "/api/api-app-services/{appID}", Method: http.MethodGet, Handler: e.ListRuntimeServices},

		{Path: "/api/apim-ws/api-docs/filetree/{inode}", Method: http.MethodGet, WriterHandler: e.APIDocWebsocket},
		{Path: "/api/apim/{treeName}/filetree", Method: http.MethodPost, Handler: e.CreateNode},
		{Path: "/api/apim/{treeName}/filetree", Method: http.MethodGet, Handler: e.ListChildrenNodes},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodDelete, Handler: e.DeleteNode},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodPut, Handler: e.UpdateNode},
		{Path: "/api/apim/{treeName}/filetree/{inode}", Method: http.MethodGet, Handler: e.GetNodeDetail},
		{Path: "/api/apim/{treeName}/filetree/{inode}/actions/{action}", Method: http.MethodPost, Handler: e.MvCpNode},

		{Path: "/api/apim/operations", Method: http.MethodGet, Handler: e.SearchOperations},
		{Path: "/api/apim/operations/{id}", Method: http.MethodGet, Handler: e.GetOperation},

		{Path: "/api/apim/validate-swagger", Method: http.MethodPost, Handler: e.ValidateSwagger},

		// gittar 事件回调
		{Path: ReleaseCallbackPath, Method: http.MethodPost, Handler: e.ReleaseCallback},
		{Path: MrCheckRunCallback, Method: http.MethodPost, Handler: e.checkrunCreate},

		// cdp 事件回调
		{Path: CDPCallbackPath, Method: http.MethodPost, Handler: e.CDPCallback},
		{Path: GitCreateMrCallback, Method: http.MethodPost, Handler: e.RepoMrEventCallback},
		{Path: GitMergeMrCallback, Method: http.MethodPost, Handler: e.RepoMrEventCallback},
		{Path: GitCloseMrCallback, Method: http.MethodPost, Handler: e.RepoMrEventCallback},
		{Path: GitCommentMrCallback, Method: http.MethodPost, Handler: e.RepoMrEventCallback},
		{Path: GitDeleteBranchCallback, Method: http.MethodPost, Handler: e.RepoBranchEventCallback},
		{Path: GitDeleteTagCallback, Method: http.MethodPost, Handler: e.RepoTagEventCallback},

		{Path: IssueCallback, Method: http.MethodPost, Handler: e.IssueCallback},

		// cicd
		{Path: "/api/cicd/{pipelineID}/tasks/{taskID}/logs", Method: http.MethodGet, Handler: e.CICDTaskLog},
		{Path: "/api/cicd/{pipelineID}/tasks/{taskID}/logs/actions/download", Method: http.MethodGet, ReverseHandler: e.ProxyCICDTaskLogDownload},

		// pipeline
		{Path: "/api/cicds", Method: http.MethodPost, Handler: e.pipelineCreate},
		{Path: "/api/cicds", Method: http.MethodGet, Handler: e.pipelineList},
		{Path: "/api/cicds/actions/pipelineYmls", Method: http.MethodGet, Handler: e.pipelineYmlList},
		{Path: "/api/cicds/actions/app-invoked-combos", Method: http.MethodGet, Handler: e.pipelineAppInvokedCombos},
		{Path: "/api/cicds/actions/fetch-pipeline-id", Method: http.MethodGet, Handler: e.fetchPipelineByAppInfo},
		{Path: "/api/cicds/actions/app-all-valid-branch-workspaces", Method: http.MethodGet, Handler: e.branchWorkspaceMap},
		{Path: "/api/cicds/{pipelineID}/actions/run", Method: http.MethodPost, Handler: e.pipelineRun},
		{Path: "/api/cicds/{pipelineID}/actions/cancel", Method: http.MethodPost, Handler: e.pipelineCancel},
		{Path: "/api/cicds/{pipelineID}/actions/rerun", Method: http.MethodPost, Handler: e.pipelineRerun},
		{Path: "/api/cicds/{pipelineID}/actions/rerun-failed", Method: http.MethodPost, Handler: e.pipelineRerunFailed},
		{Path: "/api/cicds/{pipelineID}", Method: http.MethodPut, Handler: e.pipelineOperate},

		{Path: "/api/cicds/{pipelineID}/actions/get-branch-rule", Method: http.MethodGet, Handler: e.pipelineGetBranchRule},

		// pipeline cron
		{Path: "/api/cicd-crons", Method: http.MethodGet, Handler: e.pipelineCronPaging},
		{Path: "/api/cicd-crons/{cronID}/actions/start", Method: http.MethodPut, Handler: e.pipelineCronStart},
		{Path: "/api/cicd-crons/{cronID}/actions/stop", Method: http.MethodPut, Handler: e.pipelineCronStop},
		{Path: "/api/cicd-crons", Method: http.MethodPost, Handler: e.pipelineCronCreate},
		{Path: "/api/cicd-crons/{cronID}", Method: http.MethodDelete, Handler: e.pipelineCronDelete},

		// project pipeline
		{Path: "/api/cicds-project", Method: http.MethodPost, Handler: e.projectPipelineCreate},

		// cms
		{Path: "/api/cicds/configs", Method: http.MethodPost, Handler: e.createOrUpdateCmsNsConfigs},
		{Path: "/api/cicds/configs", Method: http.MethodDelete, Handler: e.deleteCmsNsConfigs},
		{Path: "/api/cicds/multinamespace/configs", Method: http.MethodPost, Handler: e.getCmsNsConfigs},
		{Path: "/api/cicds/actions/fetch-config-namespaces", Method: http.MethodGet, Handler: e.getConfigNamespaces},
		{Path: "/api/cicds/actions/list-workspaces", Method: http.MethodGet, Handler: e.listConfigWorkspaces},

		{Path: "/api/pipeline-snippets/actions/query-snippet-yml", Method: http.MethodPost, Handler: e.querySnippetYml},

		{Path: "/api/cicd-pipeline/filetree/{inode}/actions/find-ancestors", Method: http.MethodGet, Handler: e.FindGittarFileTreeNodeAncestors},
		{Path: "/api/cicd-pipeline/filetree/actions/get-inode-by-pipeline", Method: http.MethodGet, Handler: e.GetGittarFileByPipelineId},
		{Path: "/api/cicd-pipeline/filetree", Method: http.MethodPost, Handler: e.CreateGittarFileTreeNode},
		{Path: "/api/cicd-pipeline/filetree/{inode}", Method: http.MethodDelete, Handler: e.DeleteGittarFileTreeNode},
		{Path: "/api/cicd-pipeline/filetree", Method: http.MethodGet, Handler: e.ListGittarFileTreeNodes},
		{Path: "/api/cicd-pipeline/filetree/{inode}", Method: http.MethodGet, Handler: e.GetGittarFileTreeNode},
		{Path: "/api/cicd-pipeline/filetree/actions/fuzzy-search", Method: http.MethodGet, Handler: e.FuzzySearchGittarFileTreeNodes},

		// gittar
		{Path: "/callback/gittar", Method: http.MethodPost, Handler: e.GittarWebHookCallback},
		{Path: "/api/callbacks/git-mr-create", Method: http.MethodPost, Handler: e.GittarMRCreateCallback},

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
		{Path: "/api/autotests/spaces/actions/copy", Method: http.MethodPost, Handler: e.CopyAutoTestSpaceV2},
		{Path: "/api/autotests/spaces/actions/export", Method: http.MethodPost, WriterHandler: e.ExportAutoTestSpace},
		{Path: "/api/autotests/spaces/actions/import", Method: http.MethodPost, Handler: e.ImportAutotestSpace},

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

func NotImplemented(ctx context.Context, request *http.Request, m map[string]string) (httpserver.Responser, error) {
	return httpserver.ErrResp(http.StatusNotImplemented, "", "not implemented")
}

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	queryStringDecoder *schema.Decoder

	assetSvc    *assetsvc.Service
	fileTreeSvc *apidocsvc.Service

	bdl        *bundle.Bundle
	pipeline   *pipeline.Pipeline
	cdp        *cdp.CDP
	event      *event.Event
	permission *permission.Permission
	fileTree   *filetree.GittarFileTree

	db              *dao.DBClient
	testcase        *testcase.Service
	testset         *testset.Service
	testPlan        *testplan.TestPlan
	autotest        *autotest.Service
	autotestV2      *atv2.Service
	sonarMetricRule *sonar_metric_rule.Service
	cq              *cq.CQ
	sceneset        *sceneset.Service
	migrate         *migrate.Service
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

func WithAssetSvc(svc *assetsvc.Service) Option {
	return func(e *Endpoints) {
		e.assetSvc = svc
	}
}

func WithFileTreeSvc(svc *apidocsvc.Service) Option {
	return func(e *Endpoints) {
		e.fileTreeSvc = svc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// WithPipeline 配置 pipeline
func WithPipeline(p *pipeline.Pipeline) Option {
	return func(e *Endpoints) {
		e.pipeline = p
	}
}

// WithCDP 配置 cdp
func WithCDP(c *cdp.CDP) Option {
	return func(e *Endpoints) {
		e.cdp = c
	}
}

// WithEvent 配置 event
func WithEvent(ev *event.Event) Option {
	return func(e *Endpoints) {
		e.event = ev
	}
}

// WithPermission 配置 permission
func WithPermission(perm *permission.Permission) Option {
	return func(e *Endpoints) {
		e.permission = perm
	}
}

func WithGittarFileTree(fileTree *filetree.GittarFileTree) Option {
	return func(e *Endpoints) {
		e.fileTree = fileTree
	}
}

func (e *Endpoints) RegisterEvents() error {
	fmt.Println(discover.DOP())
	for _, callback := range eventCallbacks {
		ev := apistructs.CreateHookRequest{
			Name:   callback.Name,
			Events: callback.Events,
			URL:    strutil.Concat("http://", discover.DOP(), callback.Path),
			Active: true,
			HookLocation: apistructs.HookLocation{
				Org:         "-1",
				Project:     "-1",
				Application: "-1",
			},
		}
		if err := e.bdl.CreateWebhook(ev); err != nil {
			logrus.Errorf("failed to register %s event to eventbox, (%v)", callback.Name, err)
			return err
		}
		logrus.Infof("register release event to eventbox, event:%+v", ev)
	}
	return nil
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

var queryStringDecoder *schema.Decoder

func init() {
	queryStringDecoder = schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
}
