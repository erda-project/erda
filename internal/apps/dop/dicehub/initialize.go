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

// Package dicehub Dicehub module
package dicehub

import (
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/conf"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/dbclient"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/endpoints"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/service/release"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/service/release_rule"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Initialize 初始化应用启动服务.
func Initialize(p *provider) error {
	// 加载环境变量配置
	conf.Load()
	// 设置日志级别
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	// init endpoints
	ep, err := initEndpoints(p)
	if err != nil {
		return err
	}

	server := httpserver.New("")
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("dicehub"))
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())

	// TODO pull操作用于获取diceyml不合适，需预留给更通用的操作,为了保持兼容，暂时先保留,推动业务方使用get-dice API
	server.Router().Path("/api/releases/{releaseId}/actions/pull").Methods(http.MethodGet).HandlerFunc(ep.GetDiceYAML)
	server.Router().Path("/api/releases/{releaseId}/actions/get-dice").Methods(http.MethodGet).HandlerFunc(ep.GetDiceYAML)

	return server.RegisterToNewHttpServerRouter(p.Router)
}

// 初始化 Endpoints
func initEndpoints(p *provider) (*endpoints.Endpoints, error) {
	// 数据库初始化
	db, err := dbclient.Open()
	if err != nil {
		return nil, err
	}

	// init bundle
	bundleOpts := []bundle.Option{
		bundle.WithErdaServer(),
		bundle.WithCMP(),
		bundle.WithMonitor(),
		bundle.WithPipeline(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)
	rl := release.New(
		release.WithDBClient(db),
	)

	// 3.20 灰度逻辑迁移，3.21删除
	// publishItem.Migration320()

	releaseRule := release_rule.New(release_rule.WithDBClient(db))

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithDBClient(db),
		endpoints.WithBundle(bdl),
		endpoints.WithRelease(rl),
		endpoints.WithReleaseRule(releaseRule),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithOrg(p.Org),
	)

	return ep, nil
}
