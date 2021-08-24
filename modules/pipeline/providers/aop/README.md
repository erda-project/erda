## 使用方式

- 明确需要调节的类型 (tune type)，目前支持 pipeline / task
- 进入对应插件目录
- 在 plugins 目录下新建目录，开发你的插件，参考 echo 插件
- 在 `tunechain.go` 对应的触发时机下编排你的插件

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

// TuneTriggerChains 保存任务所有触发时机下的调用链
var TuneTriggerChains = map[aoptypes.TuneTrigger]aoptypes.TuneChain{
	// 执行前
	aoptypes.TuneTriggerTaskBeforeExec: []aoptypes.TunePoint{
		echo.New(),
	},
	// 准备前
	aoptypes.TuneTriggerTaskBeforePrepare: []aoptypes.TunePoint{
		echo.New(),
	},
	// 准备后
	aoptypes.TuneTriggerTaskAfterPrepare: []aoptypes.TunePoint{
		echo.New(),
	},
	// 创建前
	aoptypes.TuneTriggerTaskBeforeCreate: []aoptypes.TunePoint{
		echo.New(),
		autotest_cookie_keep_before.New(),
	},
	// 创建后
	aoptypes.TuneTriggerTaskAfterCreate: []aoptypes.TunePoint{
		echo.New(),
	},
	// 启动前
	aoptypes.TuneTriggerTaskBeforeStart: []aoptypes.TunePoint{
		echo.New(),
	},
	// 启动后
	aoptypes.TuneTriggerTaskAfterStart: []aoptypes.TunePoint{
		echo.New(),
	},
	// 排队前
	aoptypes.TuneTriggerTaskBeforeQueue: []aoptypes.TunePoint{
		echo.New(),
	},
	// 排队后
	aoptypes.TuneTriggerTaskAfterQueue: []aoptypes.TunePoint{
		echo.New(),
	},
	// 等待前
	aoptypes.TuneTriggerTaskBeforeWait: []aoptypes.TunePoint{
		echo.New(),
	},
	// 等待后
	aoptypes.TuneTriggerTaskAfterWait: []aoptypes.TunePoint{
		echo.New(),
	},
	// 执行后
	aoptypes.TuneTriggerTaskAfterExec: []aoptypes.TunePoint{
		echo.New(),
		unit_test_report.New(),
		autotest_cookie_keep_after.New(),
	},
}
