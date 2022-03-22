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
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/common"

	// providers and modules
	_ "github.com/erda-project/erda-infra/providers/mysqlxorm"
	_ "github.com/erda-project/erda-infra/providers/prometheus"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
	_ "github.com/erda-project/erda/modules/pipeline"
	_ "github.com/erda-project/erda/modules/pipeline/aop"
	_ "github.com/erda-project/erda/modules/pipeline/providers/clusterinfo"
	_ "github.com/erda-project/erda/modules/pipeline/providers/cms"
	_ "github.com/erda-project/erda/modules/pipeline/providers/cron"
	_ "github.com/erda-project/erda/modules/pipeline/providers/dbgc"
	_ "github.com/erda-project/erda/modules/pipeline/providers/definition"
	_ "github.com/erda-project/erda/modules/pipeline/providers/source"
)

func main() {
	common.RegisterHubListener(&servicehub.DefaultListener{
		BeforeInitFunc: func(h *servicehub.Hub, config map[string]interface{}) error {
			conf.Load()
			return nil
		},
	})
	common.Run(&servicehub.RunOptions{
		ConfigFile: "conf/pipeline/pipeline.yaml",
	})
}
