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
)

type Status interface {
	LastTime() time.Time
	NextTime() time.Time
	Count() uint64
	InRunning() bool
}

type status struct {
	lastTime, nextTime time.Time
	count              uint64
	running            bool
}

func (s status) LastTime() time.Time {
	return s.lastTime
}

func (s status) NextTime() time.Time {
	return s.nextTime
}

func (s status) Count() uint64 {
	return s.count
}

func (s status) InRunning() bool {
	return s.running
}
