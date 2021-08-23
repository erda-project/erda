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

package ecp

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/ecp/dbclient"
	"github.com/erda-project/erda/modules/ecp/endpoints"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/dumpstack"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (p *provider) initialize() error {
	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	dumpstack.Open()
	logrus.Infoln(version.String())

	db := dbclient.Open(dbengine.MustOpen())

	// init Bundle
	bundleOpts := []bundle.Option{
		bundle.WithClusterManager(),
		bundle.WithCoreServices(),
	}
	bdl := bundle.New(bundleOpts...)

	ep := endpoints.New(
		db,
		endpoints.WithBundle(bdl),
	)

	server := httpserver.New(p.Cfg.Listen)
	server.RegisterEndpoint(append(ep.Routes()))

	logrus.Infof("start the service and listen on address [%s]", p.Cfg.Listen)
	logrus.Info("starting ecp instance")

	return server.ListenAndServe()
}
