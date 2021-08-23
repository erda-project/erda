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

package main

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/conf"
	"github.com/erda-project/erda/modules/extensions/loghub"
	"github.com/erda-project/erda/pkg/common"

	// modules
	_ "github.com/erda-project/erda/modules/core/monitor/alert/alert-apis"
	_ "github.com/erda-project/erda/modules/core/monitor/alert/details-apis"
	_ "github.com/erda-project/erda/modules/core/monitor/dataview"
	_ "github.com/erda-project/erda/modules/core/monitor/dataview/v1-chart-block"
	_ "github.com/erda-project/erda/modules/core/monitor/log/query"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/index"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query-example"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	_ "github.com/erda-project/erda/modules/core/monitor/settings"
	_ "github.com/erda-project/erda/modules/extensions/loghub/index/query"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/rules"
	_ "github.com/erda-project/erda/modules/monitor/apm/report"
	_ "github.com/erda-project/erda/modules/monitor/apm/runtime"
	_ "github.com/erda-project/erda/modules/monitor/apm/topology"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/node-topo"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/org-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/project-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/report/apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/runtime-apis"
	_ "github.com/erda-project/erda/modules/monitor/dashboard/template"
	_ "github.com/erda-project/erda/modules/monitor/notify/template/query"

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
