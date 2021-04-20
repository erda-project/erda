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

package throttler

import (
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/enhancedqueue"
)

const defaultQueueWindow = 100

func (t *throttler) addKeyToQueue(key string, req AddKeyToQueueRequest) {
	eq, ok := t.queueByName[req.QueueName]
	if !ok {
		var queueWindow int64 = defaultQueueWindow
		if req.QueueWindow != nil {
			queueWindow = *req.QueueWindow
		}
		eq = t.addQueue(req.QueueName, queueWindow)
	} else {
		// 队列已存在，且 queueWindow 改变
		if req.QueueWindow != nil && *req.QueueWindow != eq.ProcessingWindow() {
			eq.SetProcessingWindow(*req.QueueWindow)
		}
	}

	// 队列增加 key
	eq.Add(key, req.Priority, req.CreationTime)

	// key 关联的队列
	keyQueueByName, ok := t.keyRelatedQueues[key]
	if !ok {
		keyQueueByName = make(map[string]*enhancedqueue.EnhancedQueue)
	}
	// 关联队列
	keyQueueByName[req.QueueName] = eq
	t.keyRelatedQueues[key] = keyQueueByName
}

func (t *throttler) addQueue(name string, window int64) *enhancedqueue.EnhancedQueue {
	eq := t.queueByName[name]
	if eq != nil {
		eq.SetProcessingWindow(window)
	} else {
		eq = enhancedqueue.NewEnhancedQueue(window)
	}
	t.queueByName[name] = eq
	return eq
}
