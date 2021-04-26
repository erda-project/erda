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

package queue

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

func (q *defaultQueue) ValidateFreeResources(tryPopP *spec.Pipeline) apistructs.PipelineQueueValidateResult {
	// get queue total resources
	maxCPU := q.pq.MaxCPU
	maxMemoryMB := q.pq.MaxMemoryMB

	// calculate used resources
	var occupiedCPU float64
	var occupiedMemoryMB float64
	q.eq.ProcessingQueue().Range(func(item priorityqueue.Item) (stopRange bool) {
		pipelineID := parsePipelineIDFromQueueItem(item)
		existP := q.pipelineCaches[pipelineID]
		resources := existP.GetPipelineAppliedResources()
		occupiedCPU += resources.Requests.CPU
		occupiedMemoryMB += resources.Requests.MemoryMB
		return false
	})

	tryPopPResources := tryPopP.GetPipelineAppliedResources()

	var result apistructs.PipelineQueueValidateResult
	if tryPopPResources.Requests.CPU+occupiedCPU > maxCPU {
		result.Success = false
		result.Reason = fmt.Sprintf("Insufficient cpu: %s(current) + %s(apply) > %s(queue limited)",
			strutil.String(occupiedCPU), strutil.String(tryPopPResources.Requests.CPU), strutil.String(maxCPU))
		return result
	}
	if tryPopPResources.Requests.MemoryMB+occupiedMemoryMB > maxMemoryMB {
		result.Success = false
		result.Reason = fmt.Sprintf("Insufficient memory: %sMB(current) + %sMB(apply) > %sMB(queue limited)",
			strutil.String(occupiedMemoryMB), strutil.String(tryPopPResources.Requests.MemoryMB), strutil.String(maxMemoryMB))
		return result
	}

	return types.SuccessValidateResult
}
