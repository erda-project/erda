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

package scheduler

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/i18n"
	"github.com/erda-project/erda/modules/scheduler/server"
	"github.com/erda-project/erda/pkg/dumpstack"
)

// Initialize Application-related initialization operations
func Initialize() error {
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
	logrus.Infof("[alert] starting scheduler instance")
	i18n.InitI18N()

	return server.NewServer(conf.ListenAddr()).ListenAndServe()
}
