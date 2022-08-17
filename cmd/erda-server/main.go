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
	_ "github.com/erda-project/erda/internal/apps/gallery"

	// pkg
	_ "github.com/erda-project/erda/internal/pkg/audit"
	_ "github.com/erda-project/erda/internal/pkg/dingtalktest"
	_ "github.com/erda-project/erda/internal/pkg/service-discover/erda-discover"
	_ "github.com/erda-project/erda/internal/pkg/service-discover/fixed-discover"
	"github.com/erda-project/erda/pkg/common"

	// core
	_ "github.com/erda-project/erda-infra/providers/pprof"
	_ "github.com/erda-project/erda-infra/providers/redis"
	_ "github.com/erda-project/erda/internal/apps/dop/dicehub"
	_ "github.com/erda-project/erda/internal/apps/dop/dicehub/extension"
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

	// infra
	"github.com/erda-project/erda-infra/base/servicehub"
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
	_ "github.com/erda-project/erda-proto-go/core/org/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cms/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cron/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/definition/client"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/source/client"
	_ "github.com/erda-project/erda-proto-go/core/services/errorbox/client"
	_ "github.com/erda-project/erda-proto-go/core/token/client"
	_ "github.com/erda-project/erda-proto-go/core/user/client"
	_ "github.com/erda-project/erda-proto-go/msp/menu/client"
	_ "github.com/erda-project/erda-proto-go/msp/tenant/project/client"
	_ "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/client"

	// openapi
	_ "github.com/erda-project/erda/internal/core/openapi/legacy"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/compatibility"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/ory-kratos"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/password"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/token"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/uc-session"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/backend"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/custom-register"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/custom-route-source"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/example/publish"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/common"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/csrf"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/dump"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/interceptors/user-info"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/custom"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/dynamic"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/dynamic/temporary"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/openapi-v1"
	_ "github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes/proto"

	// uc-adaptor
	_ "github.com/erda-project/erda/internal/core/user/impl/uc/uc-adaptor"

	_ "github.com/erda-project/erda/internal/apps/devflow/flow"
	_ "github.com/erda-project/erda/internal/apps/devflow/issuerelation"

	// dop
	_ "github.com/erda-project/erda/internal/apps/dop"
	_ "github.com/erda-project/erda/internal/apps/dop/component-protocol/components"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/api-management"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/autotest/testplan"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/cms"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/contribution"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/devflowrule"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/guide"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/core"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/stream/core"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/issue/sync"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/pipelinetemplate"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/project/home"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/projectpipeline"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/publishitem"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/rule"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/search"
	_ "github.com/erda-project/erda/internal/apps/dop/providers/taskerror"
)

//go:embed bootstrap.yaml
var bootstrapCfg string

func main() {
	common.RegisterHubListener(cpregister.NewHubListener())
	common.Run(&servicehub.RunOptions{
		Content: bootstrapCfg,
	})
}
