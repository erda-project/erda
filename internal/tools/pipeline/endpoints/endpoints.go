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
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/daemon"
	"github.com/erda-project/erda/internal/tools/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	pipelineSvc *pipelinesvc.PipelineSvc
	crondSvc    daemon.Interface

	dbClient           *dbclient.Client
	queryStringDecoder *schema.Decoder

	clusterInfo clusterinfo.Interface
	mySQL       mysqlxorm.Interface
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

func WithClusterInfo(clusterInfo clusterinfo.Interface) Option {
	return func(e *Endpoints) {
		e.clusterInfo = clusterInfo
	}
}

func WithMysql(mysql mysqlxorm.Interface) Option {
	return func(e *Endpoints) {
		e.mySQL = mysql
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

		// daemon
		{Path: "/_daemon/reload-action-executor-config", Method: http.MethodGet, Handler: e.reloadActionExecutorConfig},
		{Path: "/_daemon/crond/actions/reload", Method: http.MethodGet, Handler: e.crondReload},
		{Path: "/_daemon/crond/actions/snapshot", Method: http.MethodGet, Handler: e.crondSnapshot},

		// cluster info
		// TODO: clusterinfo provider provide this api directly, remove explicit declaration in endpoint.
		{Path: clusterinfo.ClusterHookApiPath, Method: http.MethodPost, Handler: e.clusterHook},

		// executor info, only for internal check executor and cluster info
		{Path: "/api/pipeline-executors", Method: http.MethodGet, Handler: e.executorInfos},
		{Path: "/api/pipeline-executors/actions/refresh", Method: http.MethodPut, Handler: e.triggerRefreshExecutors},
	}
}
