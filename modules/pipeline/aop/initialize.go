package aop

import (
	"sync"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/pipeline"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
)

// tuneGroup 保存所有 tune chain
// 根据 类型、触发时机 初始化所有场景下的调用链
var tuneGroup aoptypes.TuneGroup
var once sync.Once
var initialized bool
var globalSDK aoptypes.SDK

func Initialize(bdl *bundle.Bundle, dbClient *dbclient.Client, report *reportsvc.ReportSvc) {
	once.Do(func() {
		initialized = true

		globalSDK.Bundle = bdl
		globalSDK.DBClient = dbClient
		globalSDK.Report = report

		tuneGroup = aoptypes.TuneGroup{
			// pipeline level
			aoptypes.TuneTypePipeline: pipeline.TuneTriggerChains,
			// task level
			aoptypes.TuneTypeTask: task.TuneTriggerChains,
		}
	})
}
