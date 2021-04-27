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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/conf"
)

func (q *defaultQueue) Start(stopCh chan struct{}) {
	if q.started {
		return
	}
	ticket := time.NewTicker(time.Second * time.Duration(conf.QueueLoopHandleIntervalSec()))
	go func() {
		for {
			select {
			case <-q.rangeAtOnceCh:
				logrus.Debugf("queueManager: queueID: %s, queueName: %s, range at once", q.ID(), q.pq.Name)
				q.RangePendingQueue()
				logrus.Debugf("queueManager: queueID: %s, queueName: %s, complete trigger by at once", q.ID(), q.pq.Name)
			case <-ticket.C:
				q.RangePendingQueue()
				logrus.Debugf("queueManager: queueID: %s, queueName: %s, complete trigger by ticker", q.ID(), q.pq.Name)
			case <-stopCh:
				// stop handle
				return
			}
		}
	}()
	q.started = true
	q.rangeAtOnceCh <- true
}
