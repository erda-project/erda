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

// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cancel"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/daemon"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/engine"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/permission"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/queuemanager"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/run"
	"github.com/erda-project/erda/internal/tools/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	permissionSvc permission.Interface
	pipelineSvc   *pipelinesvc.PipelineSvc
	crondSvc      daemon.Interface

	dbClient           *dbclient.Client
	queryStringDecoder *schema.Decoder

	engine       engine.Interface
	queueManager queuemanager.Interface
	clusterInfo  clusterinfo.Interface
	edgePipeline edgepipeline.Interface
	edgeRegister edgepipeline_register.Interface
	mySQL        mysqlxorm.Interface
	run          run.Interface
	cancel       cancel.Interface
}

type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(e *Endpoints) {
		e.dbClient = dbClient
	}
}

func WithPermissionSvc(svc permission.Interface) Option {
	return func(e *Endpoints) {
		e.permissionSvc = svc
	}
}

func WithCrondSvc(svc daemon.Interface) Option {
	return func(e *Endpoints) {
		e.crondSvc = svc
	}
}

func WithPipelineSvc(svc *pipelinesvc.PipelineSvc) Option {
	return func(e *Endpoints) {
		e.pipelineSvc = svc
	}
}

func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithEngine(engine engine.Interface) Option {
	return func(e *Endpoints) {
		e.engine = engine
	}
}

func WithQueueManager(qm queuemanager.Interface) Option {
	return func(e *Endpoints) {
		e.queueManager = qm
	}
}

func WithClusterInfo(clusterInfo clusterinfo.Interface) Option {
	return func(e *Endpoints) {
		e.clusterInfo = clusterInfo
	}
}

func WithEdgePipeline(edgePipeline edgepipeline.Interface) Option {
	return func(e *Endpoints) {
		e.edgePipeline = edgePipeline
	}
}

func WithEdgeRegister(edgeRegister edgepipeline_register.Interface) Option {
	return func(e *Endpoints) {
		e.edgeRegister = edgeRegister
	}
}

func WithMysql(mysql mysqlxorm.Interface) Option {
	return func(e *Endpoints) {
		e.mySQL = mysql
	}
}

func WithRun(run run.Interface) Option {
	return func(e *Endpoints) {
		e.run = run
	}
}

func WithCancel(cancel cancel.Interface) Option {
	return func(e *Endpoints) {
		e.cancel = cancel
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		// health check
		{Path: "/ping", Method: http.MethodGet, Handler: e.healthCheck},
		// version
		{Path: "/version", Method: http.MethodGet, Handler: e.version},

		{Path: "/mysql/stats", Method: http.MethodGet, Handler: e.mysqlStats},
		{Path: "/mysql/provider/stats", Method: http.MethodGet, Handler: e.providerMysqlStats},

		// pipelines
		{Path: "/api/v2/pipelines", Method: http.MethodPost, Handler: e.pipelineCreateV2},
		{Path: "/api/pipelines", Method: http.MethodPost, Handler: e.pipelineCreate}, // TODO qa 和 adaptor 通过 bundle 调用 v1 create，需要调整后再下线
		{Path: "/api/pipelines", Method: http.MethodGet, Handler: e.pipelineList},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodGet, Handler: e.pipelineDetail},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodPut, Handler: e.pipelineOperate},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodDelete, Handler: e.pipelineDelete},
		{Path: "/api/pipelines/{pipelineID}/actions/run", Method: http.MethodPost, Handler: e.pipelineRun},
		{Path: "/api/pipelines/{pipelineID}/actions/cancel", Method: http.MethodPost, Handler: e.pipelineCancel},
		{Path: "/api/pipelines/{pipelineID}/actions/rerun", Method: http.MethodPost, Handler: e.pipelineRerun},
		{Path: "/api/pipelines/{pipelineID}/actions/rerun-failed", Method: http.MethodPost, Handler: e.pipelineRerunFailed},

		// tasks
		{Path: "/api/pipelines/{pipelineID}/tasks/{taskID}", Method: http.MethodGet, Handler: e.pipelineTaskDetail},
		{Path: "/api/pipelines/{pipelineID}/tasks/{taskID}/actions/get-bootstrap-info", Method: http.MethodGet, Handler: e.taskBootstrapInfo},

		// pipeline related actions
		{Path: "/api/pipelines/actions/batch-create", Method: http.MethodPost, Handler: e.pipelineBatchCreate},
		{Path: "/api/pipelines/actions/statistics", Method: http.MethodGet, Handler: e.pipelineStatistic},
		{Path: "/api/pipelines/actions/task-view", Method: http.MethodGet, Handler: e.pipelineTaskView},

		// platform callback
		{Path: "/api/pipelines/actions/callback", Method: http.MethodPost, Handler: e.pipelineCallback},

		// daemon
		{Path: "/_daemon/reload-action-executor-config", Method: http.MethodGet, Handler: e.reloadActionExecutorConfig},
		{Path: "/_daemon/crond/actions/reload", Method: http.MethodGet, Handler: e.crondReload},
		{Path: "/_daemon/crond/actions/snapshot", Method: http.MethodGet, Handler: e.crondSnapshot},

		{Path: "/api/pipeline-snippets/actions/query-details", Method: http.MethodPost, Handler: e.querySnippetDetails},

		// cluster info
		// TODO: clusterinfo provider provide this api directly, remove explicit declaration in endpoint.
		{Path: clusterinfo.ClusterHookApiPath, Method: http.MethodPost, Handler: e.clusterHook},

		// executor info, only for internal check executor and cluster info
		{Path: "/api/pipeline-executors", Method: http.MethodGet, Handler: e.executorInfos},
		{Path: "/api/pipeline-executors/actions/refresh", Method: http.MethodPut, Handler: e.triggerRefreshExecutors},
	}
}
