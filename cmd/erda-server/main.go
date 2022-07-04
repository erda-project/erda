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

	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda-infra/providers/component-protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	_ "github.com/erda-project/erda-infra/providers/grpcclient"
	_ "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/client"
	_ "github.com/erda-project/erda-proto-go/core/org/client"
	_ "github.com/erda-project/erda-proto-go/msp/menu/client"
	_ "github.com/erda-project/erda-proto-go/msp/tenant/project/client"
	_ "github.com/erda-project/erda/internal/apps/admin"
	_ "github.com/erda-project/erda/internal/apps/admin/personal-workbench"
	_ "github.com/erda-project/erda/internal/apps/gallery"
	_ "github.com/erda-project/erda/internal/core/legacy"
	_ "github.com/erda-project/erda/internal/pkg/dingtalktest"
	"github.com/erda-project/erda/pkg/common"
	// core-services
	_ "github.com/erda-project/erda-infra/providers/pprof"
	_ "github.com/erda-project/erda-infra/providers/redis"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cms/client"
	_ "github.com/erda-project/erda/internal/core/legacy/providers/token"
	_ "github.com/erda-project/erda/internal/core/legacy/services/dingtalk/api"
	_ "github.com/erda-project/erda/internal/core/messenger/eventbox"
	_ "github.com/erda-project/erda/internal/core/messenger/notify"
	_ "github.com/erda-project/erda/internal/core/messenger/notify-channel"
	_ "github.com/erda-project/erda/internal/core/messenger/notifygroup"
	_ "github.com/erda-project/erda/internal/core/project"

	// infra
	_ "github.com/erda-project/erda-infra/providers/grpcserver"
	_ "github.com/erda-project/erda-infra/providers/httpserver"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
)

//go:embed bootstrap.yaml
var bootstrapCfg string

func main() {
	common.RegisterHubListener(cpregister.NewHubListener())
	common.Run(&servicehub.RunOptions{
		Content: bootstrapCfg,
	})
}
