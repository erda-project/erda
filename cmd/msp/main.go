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
