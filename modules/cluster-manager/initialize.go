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

package cluster_manager

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/cluster-manager/conf"
	"github.com/erda-project/erda/modules/cluster-manager/dbclient"
	"github.com/erda-project/erda/modules/cluster-manager/endpoints"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func initialize(cfg *conf.Conf) error {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ep := endpoints.New(
		endpoints.WithDBClient(dbclient.Open()),
	)

	server := httpserver.New(cfg.Listen)
	server.RegisterEndpoint(append(ep.Routes()))

	logrus.Infof("start the service and listen on address [%s]", cfg.Listen)
	logrus.Info("starting cluster-manager instance")

	return server.ListenAndServe()
}
