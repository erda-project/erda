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

package tasks

import (
	"fmt"
	"time"
)

type ExitError struct {
	Msg string
}

func (e ExitError) Error() string {
	return e.Msg
}

type Task func() error

type Ticker struct {
	Name     string
	Interval time.Duration
	Task     func() (stop bool, err error)
	done     chan bool
}

func New(interval time.Duration, task func() (bool, error)) *Ticker {
	return &Ticker{
		Interval: interval,
		Task:     task,
		done:     make(chan bool),
	}
}

func (d *Ticker) Run() error {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	var (
		err  error
		stop bool
	)
	fmt.Printf("the interval task %s is running right now: %s\n", d.Name, time.Now().Format(time.RFC3339))
	stop, err = d.Task()
	fmt.Printf("the interval task %s is complete this time, err: %v\n", d.Name, err)
	if stop {
		d.Close()
		return err
	}

	for {
		select {
		case <-d.done:
			fmt.Printf("the interval task %s is done!\n", d.Name)
			return err
		case t := <-ticker.C:
			fmt.Printf("the interval task %s is running at: %s\n", d.Name, t.Format(time.RFC3339))
			stop, err = d.Task()
			fmt.Printf("the interval task %s is complete this time, err: %v\n", d.Name, err)
			if stop {
				d.Close()
			}
		}
	}
}

func (d *Ticker) Close() {
	close(d.done)
}
