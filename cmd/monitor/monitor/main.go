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

package main

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom"
	"github.com/erda-project/erda/conf"

	// modules
	_ "github.com/erda-project/erda/modules/extensions/loghub/index/query"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/rules"
	_ "github.com/erda-project/erda/modules/monitor/alert/alert-apis"
	_ "github.com/erda-project/erda/modules/monitor/alert/details-apis"
	_ "github.com/erda-project/erda/modules/monitor/apm/alert"
	_ "github.com/erda-project/erda/modules/monitor/apm/report"
	_ "github.com/erda-project/erda/modules/monitor/apm/runtime"
	_ "github.com/erda-project/erda/modules/monitor/apm/topology"
	_ "github.com/erda-project/erda/modules/monitor/apm/trace"
	_ "github.com/erda-project/erda/modules/monitor/core/logs/query"
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq-example"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/node-topo"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/org-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/project-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/report/apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/report/engine"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/runtime-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/template"
	_ "github.com/erda-project/erda/modules/monitor/notify/template/query"
	_ "github.com/erda-project/erda/modules/monitor/settings"

	// providers
	_ "github.com/erda-project/erda-infra/providers/cassandra"
	_ "github.com/erda-project/erda-infra/providers/elasticsearch"
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda-infra/providers/kafka"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	_ "github.com/erda-project/erda-infra/providers/pprof"
	_ "github.com/erda-project/erda-infra/providers/redis"
)

func main() {
	// modcom.RegisterInitializer(loghub.Init)
	modcom.Run(&servicehub.RunOptions{
		ConfigFile: conf.MonitorDefaultConfig,
		Content:    conf.MonitorConfigFilePath,
	})
}
