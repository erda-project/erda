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

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smallnest/queue"
)

type Task struct {
	Key string
	Do  func()
}

type TaskQueue struct {
	tasks   *queue.LKQueue
	keys    map[string]struct{}
	maxSize int
	mtx     sync.Mutex
}

func NewTaskQueue(maxSize int) *TaskQueue {
	return &TaskQueue{
		tasks:   queue.NewLKQueue(),
		keys:    make(map[string]struct{}, maxSize),
		maxSize: maxSize,
	}
}

func (q *TaskQueue) IsFull() bool {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	return len(q.keys) >= q.maxSize
}

func (q *TaskQueue) Enqueue(task *Task) {
	q.mtx.Lock()
	if len(q.keys) >= q.maxSize {
		logrus.Warnf("queue size is full, task is ignored, key:%s", task.Key)
		q.mtx.Unlock()
		return
	}
	if _, ok := q.keys[task.Key]; ok {
		q.mtx.Unlock()
		return
	}
	q.keys[task.Key] = struct{}{}
	q.mtx.Unlock()
	q.tasks.Enqueue(task)
}

func (q *TaskQueue) Dequeue() *Task {
	ele := q.tasks.Dequeue()
	if ele == nil {
		return nil
	}
	task := ele.(*Task)
	q.mtx.Lock()
	delete(q.keys, task.Key)
	q.mtx.Unlock()
	return task
}

func (q *TaskQueue) ExecuteLoop(interval time.Duration) {
	for {
		logrus.Infof("start execute loop")
		for task := q.Dequeue(); task != nil; task = q.Dequeue() {
			task.Do()
		}
		logrus.Infof("end execute loop")
		time.Sleep(interval)
	}
}
