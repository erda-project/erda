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

package enhancedqueue

import (
	"fmt"
	"sync"
	"time"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
)

// EnhancedQueue 在优先队列基础上进行了封装，功能增强
type EnhancedQueue struct {
	// pending 等待处理的优先队列
	// 该队列可以无限追加，没有大小限制
	pending *priorityqueue.PriorityQueue

	// processing 处理中的优先队列
	processing *priorityqueue.PriorityQueue
	// processingWindow 处理中的优先队列窗口大小，即同时处理的并发度
	// window 大小可以调整，不影响当前正在处理中的任务
	// 例如之前 window=2，且有两个任务正在处理中，此时缩小 window=1，不会影响已经在处理中的两个任务
	processingWindow int64

	lock sync.RWMutex
}

func NewEnhancedQueue(window int64) *EnhancedQueue {
	return &EnhancedQueue{
		pending:          priorityqueue.NewPriorityQueue(),
		processing:       priorityqueue.NewPriorityQueue(),
		processingWindow: window,
		lock:             sync.RWMutex{},
	}
}

func (eq *EnhancedQueue) PendingQueue() *priorityqueue.PriorityQueue {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	return eq.pending
}

func (eq *EnhancedQueue) ProcessingQueue() *priorityqueue.PriorityQueue {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	return eq.processing
}

// InProcessing 返回指定 key 是否正在处理中
func (eq *EnhancedQueue) InProcessing(key string) bool {
	eq.lock.RLock()
	defer eq.lock.RUnlock()

	return eq.inProcessing(key)
}

// InPending 返回指定 key 是否在等待中
func (eq *EnhancedQueue) InPending(key string) bool {
	eq.lock.RLock()
	defer eq.lock.RUnlock()

	return eq.inPending(key)
}

// InQueue 返回指定 key 是否在某一具体队列中
func (eq *EnhancedQueue) InQueue(key string) bool {
	eq.lock.RLock()
	defer eq.lock.RUnlock()

	return eq.inPending(key) || eq.inProcessing(key)
}

// Add 将指定 key 插入 pending 队列
func (eq *EnhancedQueue) Add(key string, priority int64, creationTime time.Time) {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	eq.pending.Add(priorityqueue.NewItem(key, priority, creationTime))
}

// PopPending 将 pending 中的第一个 key 推进到 processing
// 返回被 pop 的 key
func (eq *EnhancedQueue) PopPending(dryRun ...bool) string {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	// 查看 pending 中是否有可以 push 的
	peeked := eq.pending.Peek()
	if peeked == nil {
		return ""
	}
	// 确认窗口大小
	if int64(eq.processing.Len()) >= eq.processingWindow {
		return ""
	}
	// 小于窗口，可以 doPop

	// dryRun
	if len(dryRun) > 0 && dryRun[0] {
		return peeked.Key()
	}
	// 真实处理
	popped := eq.pending.Pop()
	if peeked.Key() != popped.Key() {
		panic(fmt.Errorf("should be same, peeked: %s, popped: %s", peeked.Key(), popped.Key()))
	}
	eq.processing.Add(popped)
	return popped.Key()
}

// PopProcessing 将指定 key 从 processing 队列中移除，表示完成
func (eq *EnhancedQueue) PopProcessing(key string, dryRun ...bool) string {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	// 查看 key 是否在 processing queue 中
	if !eq.inProcessing(key) {
		return ""
	}

	// dryRun
	if len(dryRun) > 0 && dryRun[0] {
		return key
	}
	// 真实处理
	eq.processing.Remove(key)
	return key
}

func (eq *EnhancedQueue) ProcessingWindow() int64 {
	eq.lock.RLock()
	defer eq.lock.RUnlock()

	return eq.processingWindow
}

func (eq *EnhancedQueue) SetProcessingWindow(newWindow int64) {
	eq.lock.Lock()
	defer eq.lock.Unlock()

	eq.processingWindow = newWindow
}
