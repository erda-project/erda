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

package pipeline

import (
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/apitest_report"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/basic"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/echo"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/precheck_before_pop"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/project"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/scene_after"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/scene_before"
)

// TuneTriggerChains 保存流水线所有触发时机下的调用链
var TuneTriggerChains = map[aoptypes.TuneTrigger]aoptypes.TuneChain{
	// pipeline before exec
	aoptypes.TuneTriggerPipelineBeforeExec: []aoptypes.TunePoint{
		echo.New(),
		project.New(),
		scene_before.New(),
	},
	// pipeline in queue precheck before pop
	aoptypes.TuneTriggerPipelineInQueuePrecheckBeforePop: []aoptypes.TunePoint{
		precheck_before_pop.New(),
	},
	// pipeline after exec
	aoptypes.TuneTriggerPipelineAfterExec: []aoptypes.TunePoint{
		echo.New(),
		basic.New(),
		apitest_report.New(),
		scene_after.New(),
	},
}
