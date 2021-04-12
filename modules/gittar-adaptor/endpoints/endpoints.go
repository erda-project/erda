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

// Package endpoints API相关的数据信息
package endpoints

import (
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/gittar-adaptor/conf"
	"github.com/erda-project/erda/modules/gittar-adaptor/event"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/cdp"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/filetree"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/permission"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/pipeline"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	bdl                *bundle.Bundle
	pipeline           *pipeline.Pipeline
	cdp                *cdp.CDP
	event              *event.Event
	permission         *permission.Permission
	queryStringDecoder *schema.Decoder
	fileTree           *filetree.GittarFileTree
}

// Option 定义Endpoints的func类型
type Option func(*Endpoints)

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
}

type EventCallback struct {
	Name   string
	Path   string
	Events []string
}

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
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

func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithGittarFileTree(fileTree *filetree.GittarFileTree) Option {
	return func(e *Endpoints) {
		e.fileTree = fileTree
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/healthy", Method: http.MethodGet, Handler: e.Info},

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
	}
}

func (e *Endpoints) RegisterEvents() error {
	for _, callback := range eventCallbacks {
		ev := apistructs.CreateHookRequest{
			Name:   callback.Name,
			Events: callback.Events,
			URL:    strutil.Concat("http://", conf.SelfAddr(), callback.Path),
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

var queryStringDecoder *schema.Decoder

func init() {
	queryStringDecoder = schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
}
