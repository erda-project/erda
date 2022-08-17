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

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/dbclient"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/service/release"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/service/release_rule"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	db                 *dbclient.DBClient
	bdl                *bundle.Bundle
	release            *release.Release
	releaseRule        *release_rule.ReleaseRule
	queryStringDecoder *schema.Decoder
	org                org.Interface
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

// WithDBClient 配置 db
func WithDBClient(db *dbclient.DBClient) Option {
	return func(e *Endpoints) {
		e.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// WithRelease 配置 release service
func WithRelease(release *release.Release) Option {
	return func(e *Endpoints) {
		e.release = release
	}
}

func WithReleaseRule(rule *release_rule.ReleaseRule) Option {
	return func(e *Endpoints) {
		e.releaseRule = rule
	}
}

// WithQueryStringDecoder 配置 queryStringDecoder
func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithOrg(org org.Interface) Option {
	return func(e *Endpoints) {
		e.org = org
	}
}

// Release 获取 release service
func (e *Endpoints) Release() *release.Release {
	return e.release
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/healthz", Method: http.MethodGet, Handler: e.Info},

		// Release相关
		{Path: "/api/releases/{releaseId}/actions/download", Method: http.MethodGet, WriterHandler: e.DownloadRelease},

		//模板市场
		//{Path: "/api/pipeline-templates/{scopeType}/{scopeId}", Method: http.MethodPost, Handler: e.CreatePipelineTemplate},

		// 分支 release 规则
		{Path: "/api/release-rules", Method: http.MethodPost, Handler: httpserver.Wrap(e.CreateRule, e.ReleaseRuleMiddleware)},
		{Path: "/api/release-rules", Method: http.MethodGet, Handler: httpserver.Wrap(e.ListRules, e.ReleaseRuleMiddleware)},
		{Path: "/api/release-rules/{id}", Method: http.MethodPut, Handler: httpserver.Wrap(e.UpdateRule, e.ReleaseRuleMiddleware)},
		{Path: "/api/release-rules/{id}", Method: http.MethodDelete, Handler: httpserver.Wrap(e.DeleteRule, e.ReleaseRuleMiddleware)},
	}
}
