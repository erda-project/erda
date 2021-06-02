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

package uc_adaptor

import (
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/uc-adaptor/conf"
	"github.com/erda-project/erda/modules/uc-adaptor/dao"
	"github.com/erda-project/erda/modules/uc-adaptor/endpoints"
	"github.com/erda-project/erda/modules/uc-adaptor/service/adaptor"
	"github.com/erda-project/erda/modules/uc-adaptor/ucclient"
	"github.com/erda-project/erda/pkg/http/httpserver"
	// "terminus.io/dice/telemetry/promxp"
)

// Initialize 初始化应用启动服务.
func Initialize() error {
	// 装载配置信息
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("set log level: %s", logrus.DebugLevel)
	}

	ep, err := initEndpoints()
	if err != nil {
		return err
	}

	server := httpserver.New(conf.ListenAddr())
	// server.Router().Path("/metrics").Methods(http.MethodGet).Handler(promxp.Handler("uc-adaptor"))
	server.Router().UseEncodedPath()
	server.RegisterEndpoint(ep.Routes())
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())

	return server.ListenAndServe()
}

// 初始化 Endpoints
func initEndpoints() (*endpoints.Endpoints, error) {
	// init bundle
	bdl := bundle.New(
		bundle.WithCMDB(),
	)

	// init uc client
	uc := ucclient.NewUCClient()

	// init db
	db, err := dao.Open()
	if err != nil {
		return nil, err
	}

	// queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)

	// ucAdaptor
	ucAdaptor := adaptor.New(
		adaptor.WithDBClient(db),
		adaptor.WithBundle(bdl),
		adaptor.WithUCClient(uc),
	)

	// compose endpoints
	ep := endpoints.New(
		endpoints.WithBundle(bdl),
		endpoints.WithQueryStringDecoder(queryStringDecoder),
		endpoints.WithUcAdaptor(ucAdaptor),
	)

	return ep, nil
}
