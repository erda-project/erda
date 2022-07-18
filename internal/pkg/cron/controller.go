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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	cronPkg "github.com/erda-project/erda/pkg/cron"
)

var (
	ZeroTime    = time.Unix(0, 0)
	DefaultTime = time.Time{}
)

var (
	once             = sync.Once{}
	crontabs         = make(chan innerTask, 1<<10)
	goroutinesLimits = make(chan struct{}, 1<<10)
	crontabsM        = sync.Map{}
	canceled         = false
)

func Run() {
	once.Do(func() {
		go run()
	})
}

func Add(cron, name string, f func() bool) (TaskStopper, error) {
	var (
		schedule cronPkg.Schedule
		err      error
	)
	switch fields := strings.Fields(cron); len(fields) {
	case 7:
		// trim year
		cron = strings.Join(fields[:len(fields)-1], " ")
		fallthrough
	case 6:
		schedule, err = cronPkg.Parse(cron)
	default:
		schedule, err = cronPkg.ParseStandard(cron)
	}
	if err != nil {
		return nil, err
	}

	var task Task = &defaultTaskImpl{
		name:     name,
		do:       f,
		schedule: schedule,
	}
	return AddTask(task)
}

func AddTask(task Task) (TaskStopper, error) {
	// check if exists
	var exist = false
	crontabsM.Range(func(key, value interface{}) bool {
		if exist = key.(Task).Name() == task.Name(); exist {
			return false
		}
		return true
	})
	if exist {
		return nil, errors.New("same name task already exists")
	}

	// add crontab
	var crontab = &defaultInnerTaskImpl{
		Task: task,
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	select {
	case crontabs <- crontab:
		fmt.Printf("[%s] add task %s\n", time.Now().Format(time.RFC3339), crontab.Name())
		crontabsM.Store(crontab, status{
			lastTime: ZeroTime,
			nextTime: crontab.NextTime(time.Now()),
		})
	case <-ticker.C:
		return nil, errors.New("add task timeout, may there be too many tasks")
	}
	return crontab, nil
}

func Cancel() {
	canceled = true
}

func run() {
	var sec = time.NewTimer(time.Second)
	defer sec.Stop()
	for {
		if canceled {
			return
		}
		crontab := <-crontabs
		select {
		case goroutinesLimits <- struct{}{}:
			go func() {
				defer func() { go func() { <-goroutinesLimits }() }()
				thisTime := <-crontab.waitForStarting()
				if thisTime == ZeroTime || thisTime == DefaultTime {
					fmt.Printf("[%s] the cron task exit because of invalid next time: %s\n",
						time.Now().Format(time.RFC3339), thisTime.Format(time.RFC3339))
					return
				}
				var theStatus status
				value, ok := crontabsM.Load(crontab)
				if ok {
					theStatus = value.(status)
				}
				theStatus.running = true
				crontabsM.Store(crontab, theStatus)
				stop := crontab.Do()
				theStatus.running = false
				theStatus.lastTime = thisTime
				theStatus.nextTime = crontab.NextTime(thisTime)
				theStatus.count++
				if !stop {
					crontabs <- crontab
				} else {
					crontabsM.Delete(crontab)
				}
			}()
		case <-sec.C:
			crontabs <- crontab
		}
	}
}
