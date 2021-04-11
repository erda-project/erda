// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	// providers and modules

	"github.com/erda-project/erda-infra/modcom"
	_ "github.com/erda-project/erda/modules/monitor/core/logs/query"

	// _ "github.com/erda-project/erda/modules/business/alert/alert-apis"
	// _ "github.com/erda-project/erda/modules/business/alert/details-apis"
	// _ "github.com/erda-project/erda/modules/business/dashboard/chart-block"
	// _ "github.com/erda-project/erda/modules/business/dashboard/node-topo"
	// _ "github.com/erda-project/erda/modules/business/dashboard/org-apis"
	// _ "github.com/erda-project/erda/modules/business/dashboard/project-apis"
	// _ "github.com/erda-project/erda/modules/business/dashboard/runtime-apis"
	// _ "github.com/erda-project/erda/modules/business/dashboard/template"
	// _ "github.com/erda-project/erda/modules/business/logs/loghub/index/query"
	// _ "github.com/erda-project/erda/modules/business/logs/loghub/metrics/rules"
	// _ "github.com/erda-project/erda/modules/business/report/apis"
	_ "github.com/erda-project/erda/modules/monitor/settings"
	_ "github.com/erda-project/erda/modules/tools/admin-tools"

	// _ "github.com/erda-project/erda/modules/domain/metrics/index"
	// _ "github.com/erda-project/erda/modules/domain/metrics/metricq"
	// _ "github.com/erda-project/erda/modules/domain/metrics/metricq-example"

	//notify
	// _ "github.com/erda-project/erda/modules/domain/notify/template/query"

	// apm
	// _ "github.com/erda-project/erda/modules/business/apm/alert"
	// _ "github.com/erda-project/erda/modules/business/apm/report"
	// _ "github.com/erda-project/erda/modules/business/apm/runtime"
	// _ "github.com/erda-project/erda/modules/business/apm/topology"
	// _ "github.com/erda-project/erda/modules/business/apm/trace"

	_ "github.com/erda-project/erda-infra/providers/cassandra"
	_ "github.com/erda-project/erda-infra/providers/elasticsearch"
	_ "github.com/erda-project/erda-infra/providers/kafka"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	// _ "github.com/erda-project/erda-infra/providers/redis"
	// _ "github.com/erda-project/erda-infra/providers/telemetry"
)

//go:generate sh -c "cd ${PROJ_PATH} && go generate -v -x github.com/erda-project/erda/modules/monitor/tools/admin-tools"
func main() {
	// modcom.RegisterInitializer(loghub.Init)
	modcom.RunWithCfgDir("conf/monitor/monitor", "monitor")
}
