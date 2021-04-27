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
	"time"

	"github.com/erda-project/erda/modules/pipeline/conf"
)

func (q *defaultQueue) Start(stopCh chan struct{}) {
	if q.started {
		return
	}
	ticket := time.NewTicker(time.Second * time.Duration(conf.QueueLoopHandleIntervalSec()))
	rangeAtOnce := make(chan struct{})
	go func() {
		for {
			select {
			case <-rangeAtOnce:
				q.RangePendingQueue()
			case <-ticket.C:
				q.RangePendingQueue()
			case <-stopCh:
				// stop handle
				return
			}
		}
	}()
	q.started = true
	rangeAtOnce <- struct{}{}
}
