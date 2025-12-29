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
	_ "embed"

	// apps
	_ "github.com/erda-project/erda/internal/apps/admin"
	_ "github.com/erda-project/erda/internal/apps/admin/personal-workbench"
	_ "github.com/erda-project/erda/internal/apps/cmp"
	_ "github.com/erda-project/erda/internal/apps/gallery"

	// pkg
	_ "github.com/erda-project/erda-infra/providers/clickhouse"
	_ "github.com/erda-project/erda-infra/providers/prometheus"
	_ "github.com/erda-project/erda/internal/pkg/audit"
	_ "github.com/erda-project/erda/internal/pkg/dingtalktest"
	_ "github.com/erda-project/erda/internal/pkg/profileagent"
	_ "github.com/erda-project/erda/internal/pkg/service-discover/erda-discover"
	_ "github.com/erda-project/erda/internal/pkg/service-discover/fixed-discover"
	"github.com/erda-project/erda/pkg/common"
	_ "github.com/erda-project/erda/pkg/common/permission"

	// core
	_ "github.com/erda-project/erda-infra/providers/pprof"
	_ "github.com/erda-project/erda-infra/providers/redis"
	_ "github.com/erda-project/erda/internal/apps/dop/dicehub"
	_ "github.com/erda-project/erda/internal/apps/dop/dicehub/image"
	_ "github.com/erda-project/erda/internal/apps/dop/dicehub/release"
	_ "github.com/erda-project/erda/internal/core/file"
	_ "github.com/erda-project/erda/internal/core/legacy"
	_ "github.com/erda-project/erda/internal/core/legacy/providers/token"
	_ "github.com/erda-project/erda/internal/core/legacy/services/dingtalk/api"
	_ "github.com/erda-project/erda/internal/core/messenger/eventbox"
	_ "github.com/erda-project/erda/internal/core/messenger/notify"
	_ "github.com/erda-project/erda/internal/core/messenger/notify-channel"
	_ "github.com/erda-project/erda/internal/core/messenger/notifygroup"
	_ "github.com/erda-project/erda/internal/core/project"
	_ "github.com/erda-project/erda/internal/core/user"
	_ "github.com/erda-project/erda/internal/core/user/auth/credstore/iam"
	_ "github.com/erda-project/erda/internal/core/user/auth/identity/iam"
	_ "github.com/erda-project/erda/internal/core/user/auth/oauth/iam"

	// infra
	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/cassandra"
	_ "github.com/erda-project/erda-infra/providers/component-protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	_ "github.com/erda-project/erda-infra/providers/etcd"
	_ "github.com/erda-project/erda-infra/providers/etcd-election"
	_ "github.com/erda-project/erda-infra/providers/grpcserver"
	_ "github.com/erda-project/erda-infra/providers/health"
	_ "github.com/erda-project/erda-infra/providers/httpserver"
	_ "github.com/erda-project/erda-infra/providers/i18n"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	_ "github.com/erda-project/erda-infra/providers/mysql/v2"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"

	// grpc
	_ "github.com/erda-project/erda-infra/providers/grpcclient"
	_ "github.com/erda-project/erda-proto-go/apps/gallery/client"
	_ "github.com/erda-project/erda-proto-go/cmp/dashboard/client"
	_ "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/client"
	_ "github.com/erda-project/erda-proto-go/core/dicehub/release/client"
	_ "github.com/erda-project/erda-proto-go/core/messenger/notify/client"
	_ "github.com/erda-project/erda-proto-go/core/messenger/notifygroup/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/alert/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/diagnotor/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/event/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/log/query/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/metric/client"
	_ "github.com/erda-project/erda-proto-go/core/monitor/settings/client"
	_ "github.com/erda-project/erda-proto-go/core/org/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cms/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cron/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/definition/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/graph/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/queue/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/source/client"
	_ "github.com/erda-project/erda-proto-go/core/services/errorbox/client"
	_ "github.com/erda-project/erda-proto-go/core/token/client"
	_ "github.com/erda-project/erda-proto-go/core/user/client"
	_ "github.com/erda-project/erda-proto-go/msp/menu/client"
	_ "github.com/erda-project/erda-proto-go/msp/tenant/project/client"
	_ "github.com/erda-project/erda-proto-go/oap/entity/client"
	_ "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/client"

	// openapi
	_ "github.com/erda-project/erda/internal/core/openapi/legacy"
	_ "github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/provider"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/compatibility"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/over_permission/org_id"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/over_permission/org_name"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/password"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/session"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/token"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/backend"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/custom-register"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/custom-route-source"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/publish"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/external-openapi"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/common"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/csrf"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/dump"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/response"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/user-info"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/custom"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/dynamic/register"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/dynamic/temporary"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/dynamic/watcher"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/openapi-v1"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/proto"

	// dop
	_ "github.com/erda-project/erda/internal/apps/devflow/flow"
	_ "github.com/erda-project/erda/internal/apps/devflow/issuerelation"
	_ "github.com/erda-project/erda/internal/apps/dop"
	_ "github.com/erda-project/erda/internal/apps/dop/component-protocol/components"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/api-management"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/autotest/testplan"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/cms"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/contribution"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/devflowrule"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/efficiency_measure"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/guide"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/core"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/core"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/sync"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/pipelinetemplate"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/project/home"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/project_report"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/projectpipeline"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/publishitem"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/queue"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/rule"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/search"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/taskerror"

	// cmp
	_ "github.com/erda-project/erda/internal/apps/cmp/component-protocol/components"

	// msp
	_ "github.com/erda-project/erda/internal/apps/msp/apm/adapter"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/alert"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/alert/components"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/browser/components"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/apis"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/plugins/certificate"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/plugins/dns"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/plugins/http"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/plugins/page"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/plugins/tcp"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/storage/sync-cache"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/task"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/task/fetcher/fixed"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/checker/task/fetcher/scheduled"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/diagnotor"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/exception/query"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/log-service/query"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/log-service/rules"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/metric"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/notifygroup"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/service"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/service/components"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/service/datasources"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/trace/query"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/trace/query/components"
	_ "github.com/erda-project/erda/internal/apps/msp/apm/trace/storage/elasticsearch"
	_ "github.com/erda-project/erda/internal/apps/msp/configcenter"
	_ "github.com/erda-project/erda/internal/apps/msp/credential"
	_ "github.com/erda-project/erda/internal/apps/msp/instance/permission"
	_ "github.com/erda-project/erda/internal/apps/msp/member"
	_ "github.com/erda-project/erda/internal/apps/msp/menu"
	_ "github.com/erda-project/erda/internal/apps/msp/registercenter"
	_ "github.com/erda-project/erda/internal/apps/msp/resource"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/coordinator"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/apigateway"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/configcenter"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/etcd"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/externalprovider"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/jvmprofiler"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/loganalytics"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/loges"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/logexporter"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/logservice"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/monitor"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/monitorcollector"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/monitorkafka"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/monitorzk"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/mysql"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/nacos"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/postgresql"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/registercenter"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/servicemesh"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/zkproxy"
	_ "github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers/zookeeper"
	_ "github.com/erda-project/erda/internal/apps/msp/tenant"
	_ "github.com/erda-project/erda/internal/apps/msp/tenant/project"
	_ "github.com/erda-project/erda/internal/pkg/extension"
	_ "github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	_ "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/cleaner"
	_ "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/loader"
	_ "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/extensions/loghub"
	_ "github.com/erda-project/erda/internal/tools/monitor/extensions/loghub/index/query"

	// ai-function
	_ "github.com/erda-project/erda/internal/pkg/ai-functions"
)

//go:embed bootstrap.yaml
var bootstrapCfg string

func main() {
	common.RegisterInitializer(loghub.Init) // msp
	common.RegisterHubListener(cpregister.NewHubListener())
	common.Run(&servicehub.RunOptions{
		Content: bootstrapCfg,
	})
}
