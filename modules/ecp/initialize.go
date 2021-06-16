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
		bundle.WithCMDB(),
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
