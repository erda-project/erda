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
	"github.com/erda-project/erda/conf"
	"github.com/erda-project/erda/modules/extensions/loghub"
	"github.com/erda-project/erda/pkg/common"

	_ "github.com/erda-project/erda/modules/core/monitor/alert/alert-apis"
	_ "github.com/erda-project/erda/modules/core/monitor/alert/details-apis"
	// modules
	_ "github.com/erda-project/erda/modules/core/monitor/metric/index"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query-example"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	_ "github.com/erda-project/erda/modules/extensions/loghub/index/query"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/rules"
	_ "github.com/erda-project/erda/modules/monitor/apm/report"
	_ "github.com/erda-project/erda/modules/monitor/apm/runtime"
	_ "github.com/erda-project/erda/modules/monitor/apm/topology"
	_ "github.com/erda-project/erda/modules/monitor/apm/trace"
	_ "github.com/erda-project/erda/modules/monitor/core/logs/query"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/node-topo"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/org-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/project-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/report/apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/runtime-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/template"
	_ "github.com/erda-project/erda/modules/monitor/notify/template/query"
	_ "github.com/erda-project/erda/modules/monitor/settings"

	// providers
	_ "github.com/erda-project/erda-infra/providers"
)

func main() {
	common.RegisterInitializer(loghub.Init)
	common.Run(&servicehub.RunOptions{
		ConfigFile: conf.MonitorConfigFilePath,
		Content:    conf.MonitorDefaultConfig,
	})
}
