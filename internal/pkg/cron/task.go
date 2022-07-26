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

package cron

import (
	"time"

	cronPkg "github.com/erda-project/erda/pkg/cron"
)

type Task interface {
	Name() string
	Do() (stop bool)
	NextTime(from time.Time) time.Time
}

type TaskStopper interface {
	Task

	Stop()
}

type innerTask interface {
	Task

	waitForStarting() <-chan time.Time
}

type defaultTaskImpl struct {
	name     string
	do       func() bool
	schedule cronPkg.Schedule
}

func (impl *defaultTaskImpl) Name() string {
	return impl.name
}

func (impl *defaultTaskImpl) Do() bool {
	return impl.do()
}

func (impl *defaultTaskImpl) NextTime(from time.Time) time.Time {
	return impl.schedule.Next(from)
}

type defaultInnerTaskImpl struct {
	Task

	stopped bool
}

func (d *defaultInnerTaskImpl) Stop() {
	d.stopped = true
}

func (d *defaultInnerTaskImpl) Do() bool {
	if d.stopped {
		return true
	}
	return d.Task.Do()
}

func (d *defaultInnerTaskImpl) NextTime(from time.Time) time.Time {
	if d.stopped {
		return ZeroTime
	}
	return d.Task.NextTime(from)
}

func (d *defaultInnerTaskImpl) waitForStarting() <-chan time.Time {
	var now = time.Now()
	sub := d.NextTime(now).Sub(now)
	if d.stopped || sub < 0 {
		var c = make(chan time.Time, 1)
		c <- ZeroTime
		close(c)
		return c
	}
	return time.After(sub)
}
