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

package task

import (
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/autotest_cookie_keep_after"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/autotest_cookie_keep_before"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/echo"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/unit_test_report"
)

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
