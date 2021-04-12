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

// Package apim API 管理
package apim

import (
	"time"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/apim/bdl"
	"github.com/erda-project/erda/modules/apim/conf"
	"github.com/erda-project/erda/modules/apim/dbclient"
	"github.com/erda-project/erda/modules/apim/endpoints"
	"github.com/erda-project/erda/modules/apim/services/apidocsvc"
	"github.com/erda-project/erda/modules/apim/services/assetsvc"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httpserver"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	conf.Load()

	// init db
	if err := dbclient.Open(); err != nil {
		return err
	}
	defer dbclient.Close()

	// init bundle
	bdl.Init(
		bundle.WithCMDB(),
		bundle.WithHepa(),
		bundle.WithOrchestrator(),
		bundle.WithEventBox(),
		bundle.WithGittar(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second*15, time.Second*60), // bundle 默认 (time.Second, time.Second*3)
		)),
	)

	// init query string decoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// init service
	assetSvc := assetsvc.New()
	filetreeSvc := apidocsvc.New()

	ep := endpoints.New(
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithAssetSvc(assetSvc),
		endpoints.WithFileTreeSvc(filetreeSvc),
	)

	server := httpserver.New(conf.ListenAddr())
	server.RegisterEndpoint(ep.Routes())
	server.Router().PathPrefix("/api/apim/metrics").Handler(endpoints.InternalReverseHandler(endpoints.ProxyMetrics))

	if err := do(); err != nil {
		return err
	}

	return server.ListenAndServe()
}

// TODO 自定义初始化逻辑
func do() error {
	return nil
}
