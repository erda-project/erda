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

package precheck_before_pop

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "queue" }
func (p *Plugin) Handle(ctx aoptypes.TuneContext) error {

	var httpBeforeCheckRun = HttpBeforeCheckRun{
		PipelineID: ctx.SDK.Pipeline.ID,
		DBClient:   ctx.SDK.DBClient,
		Bdl:        ctx.SDK.Bundle,
	}
	result, err := httpBeforeCheckRun.CheckRun()
	if err != nil {
		context.WithValue(ctx.Context, apistructs.PreCheckResultContextKey, apistructs.PipelineQueueValidateResult{
			Success: false,
			Reason:  err.Error(),
		})
		return err
	}

	if result.CheckResult == CheckResultSuccess {
		context.WithValue(ctx.Context, apistructs.PreCheckResultContextKey, apistructs.PipelineQueueValidateResult{
			Success: true,
		})
	} else {
		context.WithValue(ctx.Context, apistructs.PreCheckResultContextKey, apistructs.PipelineQueueValidateResult{
			Success: false,
			RetryOption: &apistructs.QueueValidateRetryOption{
				IntervalSecond: result.RetryOption.IntervalSecond,
			},
		})
	}
	return nil
}
