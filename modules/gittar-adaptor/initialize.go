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

// Package adaptor gittar adaptor的初始化
package adaptor

import (
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/gittar-adaptor/conf"
	"github.com/erda-project/erda/modules/gittar-adaptor/endpoints"
	"github.com/erda-project/erda/modules/gittar-adaptor/event"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/cdp"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/filetree"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/permission"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/pipeline"
	"github.com/erda-project/erda/pkg/httpserver"
	//"terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	// 装载配置信息
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("set log level: %s", logrus.DebugLevel)
	}

	ep := initEndpoints()

	// 注册 hook
	err := ep.RegisterEvents()
	if err != nil {
		return err
	}

	server := httpserver.New(conf.ListenAddr())
	//server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("adaptor"))
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())

	return server.ListenAndServe()
}

// 初始化 Endpoints
func initEndpoints() *endpoints.Endpoints {
	// init bundle
	bdl := bundle.New(
		bundle.WithGittar(),
		bundle.WithPipeline(),
		bundle.WithEventBox(),
		bundle.WithCMDB(),
		bundle.WithMonitor(),
	)

	// init pipeline
	p := pipeline.New(pipeline.WithBundle(bdl))

	c := cdp.New(cdp.WithBundle(bdl))

	// init event
	e := event.New(event.WithBundle(bdl))

	// init permission
	perm := permission.New(permission.WithBundle(bdl))

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	fileTree := filetree.New(filetree.WithBundle(bdl))

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithBundle(bdl),
		endpoints.WithPipeline(p),
		endpoints.WithEvent(e),
		endpoints.WithCDP(c),
		endpoints.WithPermission(perm),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithGittarFileTree(fileTree),
	)

	return ep
}
