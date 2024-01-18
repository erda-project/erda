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
	_ "github.com/erda-project/erda-infra/providers/grpcclient"
	_ "github.com/erda-project/erda-infra/providers/mysql/v2"
	_ "github.com/erda-project/erda-infra/providers/mysqlxorm"
	_ "github.com/erda-project/erda-infra/providers/pprof"
	_ "github.com/erda-project/erda-infra/providers/prometheus"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
	_ "github.com/erda-project/erda-proto-go/core/org/client"
	_ "github.com/erda-project/erda/internal/pkg/extension"
	_ "github.com/erda-project/erda/internal/pkg/profileagent"
	_ "github.com/erda-project/erda/internal/tools/pipeline"
	_ "github.com/erda-project/erda/internal/tools/pipeline/aop"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/action_runner_scheduler"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/actionagent"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/actionmgr"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/app"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/build"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/cms"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/cron"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/compensator"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/daemon"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc/definition_cleanup"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/definition"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/graph"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/label"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/lifecycle_hook_client"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/permission"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/pipeline"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/report"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/resource"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/resourcegc"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/source"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/task"
	"github.com/erda-project/erda/pkg/common"
)

//go:embed bootstrap.yaml
var bootstrapCfg string

func main() {
	common.RegisterHubListener(&servicehub.DefaultListener{
		BeforeInitFunc: func(h *servicehub.Hub, config map[string]interface{}) error {
			conf.Load()
			return nil
		},
	})
	common.Run(&servicehub.RunOptions{
		Content: bootstrapCfg,
	})
}
