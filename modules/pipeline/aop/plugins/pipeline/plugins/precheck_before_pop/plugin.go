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

package precheck_before_pop

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "queue" }
func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {

	var httpBeforeCheckRun = HttpBeforeCheckRun{
		PipelineID: ctx.SDK.Pipeline.ID,
		DBClient:   ctx.SDK.DBClient,
		Bdl:        ctx.SDK.Bundle,
	}
	result, err := httpBeforeCheckRun.CheckRun()
	if err != nil {
		ctx.PutKV(apistructs.PipelinePreCheckResultContextKey, apistructs.PipelineQueueValidateResult{
			Success: false,
			Reason:  err.Error(),
			// add default retryOption if request is error
			RetryOption: &apistructs.QueueValidateRetryOption{
				IntervalSecond: 10,
			},
		})
		return err
	}

	if result.CheckResult == CheckResultSuccess {
		ctx.PutKV(apistructs.PipelinePreCheckResultContextKey, apistructs.PipelineQueueValidateResult{
			Success: true,
		})
	} else {
		var validResult = apistructs.PipelineQueueValidateResult{Success: false}
		if result.RetryOption.IntervalSecond > 0 || result.RetryOption.IntervalMillisecond > 0 {
			validResult.RetryOption = &apistructs.QueueValidateRetryOption{
				IntervalSecond:      result.RetryOption.IntervalSecond,
				IntervalMillisecond: result.RetryOption.IntervalMillisecond,
			}
		}
		validResult.Reason = result.Message

		ctx.PutKV(apistructs.PipelinePreCheckResultContextKey, validResult)
	}
	return nil
}
