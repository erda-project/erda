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

	"github.com/erda-project/erda/pkg/strutil"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type CapacityValidator struct {
	mgr types.QueueManager
}

func (q *defaultQueue) ValidateCapacity(tryPopP *spec.Pipeline) apistructs.PipelineQueueValidateResult {
	if int64(q.eq.ProcessingQueue().Len()) >= q.pq.Concurrency {
		var processItems []string
		q.eq.ProcessingQueue().Range(func(item priorityqueue.Item) (stopRange bool) {
			processItems = append(processItems, item.Key())
			return false
		})
		return apistructs.PipelineQueueValidateResult{
			Success: false,
			Reason: fmt.Sprintf("Insufficient processing concurrency(%d), current processing count: %d, processing items: [%s]",
				q.eq.ProcessingWindow(),
				q.eq.ProcessingQueue().Len(),
				strutil.Join(processItems, ", ", true)),
		}
	}
	return types.SuccessValidateResult
}
