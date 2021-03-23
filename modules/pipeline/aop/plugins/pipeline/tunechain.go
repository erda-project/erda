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
