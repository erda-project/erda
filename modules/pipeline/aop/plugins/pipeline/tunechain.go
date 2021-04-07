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

package pipeline

import (
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/apitest_report"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/basic"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/echo"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/project"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/scene_after"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline/plugins/scene_before"
)

// TuneTriggerChains 保存流水线所有触发时机下的调用链
var TuneTriggerChains = map[aoptypes.TuneTrigger]aoptypes.TuneChain{
	// pipeline 执行前
	aoptypes.TuneTriggerPipelineBeforeExec: []aoptypes.TunePoint{
		echo.New(),
		project.New(),
		scene_before.New(),
	},
	// pipeline 执行后
	aoptypes.TuneTriggerPipelineAfterExec: []aoptypes.TunePoint{
		echo.New(),
		basic.New(),
		apitest_report.New(),
		scene_after.New(),
	},
}
