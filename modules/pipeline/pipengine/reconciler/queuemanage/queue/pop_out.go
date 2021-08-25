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

package queue

import "github.com/erda-project/erda/modules/pipeline/spec"

func (q *defaultQueue) PopOutPipeline(p *spec.Pipeline) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.eq.PopPendingKey(makeItemKey(p))
	q.eq.PopProcessing(makeItemKey(p))
	// delete from caches
	delete(q.pipelineCaches, p.ID)
	// send popped signal to channel
	ch, ok := q.doneChanByPipelineID[p.ID]
	if ok {
		ch <- struct{}{}
		close(ch)
		delete(q.doneChanByPipelineID, p.ID)
	}
}
