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
