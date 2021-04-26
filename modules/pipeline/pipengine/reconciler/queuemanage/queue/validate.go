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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (q *defaultQueue) validatePipeline(p *spec.Pipeline) apistructs.PipelineQueueValidateResult {
	// capacity
	result := q.ValidateCapacity(p)
	if result.IsFailed() {
		return result
	}
	// free resources
	result = q.ValidateFreeResources(p)
	if result.IsFailed() {
		return result
	}

	// default result
	return types.SuccessValidateResult
}
