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
	"github.com/sirupsen/logrus"
)

type Task struct {
	Key string
	Do  func()
}

type TaskQueue struct {
	tasks chan *Task
}

func NewTaskQueue(maxSize int) *TaskQueue {
	return &TaskQueue{
		tasks: make(chan *Task, maxSize),
	}
}

func (q *TaskQueue) Enqueue(task *Task) {
	select {
	case q.tasks <- task:
	default:
		logrus.Warnf("queue size is full, task is ignored, key:%s", task.Key)
	}
}

func (q *TaskQueue) ExecuteLoop() {
	for {
		task := <-q.tasks
		logrus.Infof("start execute loop")
		task.Do()
		logrus.Infof("end execute loop")
	}
}
