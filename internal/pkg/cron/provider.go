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
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/servicehub"
)

var (
	name = "easy-cron"
	spec = servicehub.Spec{
		Services:    []string{"easy-cron", "easy-cron-client"},
		Types:       []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		Description: "easy-corn",
		ConfigFunc:  func() interface{} { return new(struct{}) },
		Creator:     func() servicehub.Provider { return &provider{} },
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type Interface interface {
	Add(cronExpr string, task Task) (TaskStopper, error)
	AddFunc(cronExpr, name string, f func() bool) (TaskStopper, error)
	Cancel()
	GetTasks() map[Task]Status
	GetRunning() map[Task]Status
}

type provider struct{}

func (p *provider) Run(_ context.Context) error {
	Run()
	return nil
}

func (p *provider) Add(cron string, t Task) (TaskStopper, error) {
	return Add(cron, t.Name(), t.Do)
}

// AddFunc add the function as task to run by the crontab expression.
// if f returns true, the task scheduled exit.
func (p *provider) AddFunc(cron, name string, f func() bool) (TaskStopper, error) {
	return Add(cron, name, f)
}

func (p *provider) Cancel() {
	Cancel()
}

func (p *provider) GetTasks() map[Task]Status {
	var m = make(map[Task]Status)
	crontabsM.Range(func(key, value interface{}) bool {
		m[key.(Task)] = value.(Status)
		return true
	})
	return m
}

func (p *provider) GetRunning() map[Task]Status {
	var m = make(map[Task]Status)
	crontabsM.Range(func(key, value interface{}) bool {
		if status := value.(status); status.InRunning() {
			m[key.(Task)] = status
		}
		return true
	})
	return m
}
