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
	Task     func() error
	done     chan bool
}

func New(interval time.Duration, task func() error) *Ticker {
	return &Ticker{
		Interval: interval,
		Task:     task,
		done:     make(chan bool),
	}
}

func (d *Ticker) Run() error {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	var err error
	for {
		select {
		case <-d.done:
			fmt.Printf("the interval task %s is done!", d.Name)
			return err
		case t := <-ticker.C:
			fmt.Printf("the interval task %s is running at: %s", d.Name, t.Format(time.RFC3339))
			err = d.Task()
			fmt.Printf("the interval task %s is complete this time", d.Name)
			switch err.(type) {
			case nil:
			case *ExitError, ExitError:
				_ = d.Close()
				fmt.Printf("the interval task %s is breaking: %v", d.Name, err)
			default:
				fmt.Printf("the interval task %s is complete with err: %v", d.Name, err)
			}
		}
	}
}

func (d *Ticker) Close() error {
	close(d.done)
	return nil
}
