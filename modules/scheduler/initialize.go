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

package scheduler

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/i18n"
	"github.com/erda-project/erda/modules/scheduler/server"
	"github.com/erda-project/erda/pkg/dumpstack"
)

// Initialize Application-related initialization operations
func Initialize() error {
	logrus.Infof(version.String())
	conf.Load()
	// control log's level.
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	})
	// open the function of dump stack
	dumpstack.Open()
	logrus.Infof("start the service and listen on address: \"%s\"", conf.ListenAddr())
	logrus.Errorf("[alert] starting scheduler instance")
	i18n.InitI18N()

	return server.NewServer(conf.ListenAddr()).ListenAndServe()
}
