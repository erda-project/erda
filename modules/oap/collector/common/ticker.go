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

package common

import (
	"context"
	"math/rand"
	"time"
)

// RunningTimer will always running.
// Trigger event when timeout or reset timer by other routine
// The real interval duration = interval + random(jitter)
type RunningTimer struct {
	interval time.Duration
	jitter   time.Duration
	timer    *time.Timer
	ch       chan time.Time
}

func NewRunningTimer(interval, jitter time.Duration) *RunningTimer {
	rt := &RunningTimer{
		interval: interval,
		jitter:   jitter,
		ch:       make(chan time.Time, 1),
	}
	rt.timer = time.NewTimer(rt.randomDuration())
	return rt
}

func (rt *RunningTimer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			rt.timer.Stop()
			return
		case now := <-rt.timer.C:
			select {
			case rt.ch <- now:
			default:
			}
			rt.Reset()
		}
	}
}

func (rt *RunningTimer) Elapsed() <-chan time.Time {
	return rt.ch
}

func (rt *RunningTimer) Reset() {
	rt.timer.Stop()
	rt.timer.Reset(rt.randomDuration())
}

func (rt *RunningTimer) randomDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return rt.interval + time.Duration(rand.Int63n(rt.jitter.Nanoseconds()))
}
