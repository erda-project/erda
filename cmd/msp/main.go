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
	"github.com/erda-project/erda/pkg/common"

	// modules and providers
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda-infra/providers/cassandra"
	_ "github.com/erda-project/erda-proto-go/core/monitor/alert/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/metric/client"
	_ "github.com/erda-project/erda/modules/msp/apm/adapter"
	_ "github.com/erda-project/erda/modules/msp/apm/alert"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/apis"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/plugins/certificate"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/plugins/dns"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/plugins/http"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/plugins/page"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/plugins/tcp"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/storage/sync-cache"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/task"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/task/fetcher/fixed"
	_ "github.com/erda-project/erda/modules/msp/apm/checker/task/fetcher/scheduled"
	_ "github.com/erda-project/erda/modules/msp/apm/exception"
	_ "github.com/erda-project/erda/modules/msp/apm/metric"
	_ "github.com/erda-project/erda/modules/msp/apm/trace"
	_ "github.com/erda-project/erda/modules/msp/configcenter"
	_ "github.com/erda-project/erda/modules/msp/instance/permission"
	_ "github.com/erda-project/erda/modules/msp/menu"
	_ "github.com/erda-project/erda/modules/msp/registercenter"
	_ "github.com/erda-project/erda/modules/msp/resource"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/coordinator"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/apigateway"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/configcenter"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/etcd"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/generalability"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/jvmprofiler"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/loganalytics"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/loges"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/logexporter"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/monitor"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/monitorcollector"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/monitorkafka"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/monitorzk"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/mysql"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/nacos"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/postgresql"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/registercenter"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/servicemesh"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/zkproxy"
	_ "github.com/erda-project/erda/modules/msp/resource/deploy/handlers/zookeeper"
	_ "github.com/erda-project/erda/modules/msp/tenant"
	_ "github.com/erda-project/erda/modules/msp/tenant/project"
	_ "github.com/erda-project/erda/pkg/common/permission"
)

func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: conf.MSPConfigFilePath,
		Content:    conf.MSPDefaultConfig,
	})
}
