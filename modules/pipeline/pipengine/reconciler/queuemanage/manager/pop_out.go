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

package manager

import (
	"strconv"
)

func (mgr *defaultManager) PopOutPipelineFromQueue(pipelineID uint64) {
	p := mgr.ensureQueryPipelineDetail(pipelineID)
	if p == nil {
		return
	}

	relatedQueueID, ok := p.GetPipelineQueueID()
	if !ok {
		return
	}

	mgr.qLock.RLock()
	defer mgr.qLock.RUnlock()
	q := mgr.queueByID[strconv.FormatUint(relatedQueueID, 10)]
	if q == nil {
		return
	}
	q.PopOutPipeline(p)
}
