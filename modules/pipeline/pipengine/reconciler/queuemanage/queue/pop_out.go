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
