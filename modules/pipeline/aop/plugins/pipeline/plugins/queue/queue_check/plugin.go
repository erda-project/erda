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

package queue_check

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
)

var (
	SkipResult = apistructs.PipelineQueueValidateResult{
		Success: true,
		Reason:  "no queue found, skip validate, treat as success",
	}
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "queue" }
func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	// get queue from ctx
	queueI, ok := ctx.TryGet("queue")
	if !ok {
		// no queue, skip check
		ctx.PutKV("queue_result", SkipResult)
		return nil
	}
	queue, ok := queueI.(types.Queue)
	if !ok {
		// not queue, skip check
		ctx.PutKV("queue_result", SkipResult)
		return nil
	}

	_ = queue

	// TODO invoke fdp

	return nil
}
