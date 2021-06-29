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

package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/msp/apm/checker/plugins"
	"github.com/erda-project/erda/modules/msp/apm/checker/task/fetcher"
	"github.com/erda-project/erda/providers/metrics/report"
)

type config struct {
	DefaultPeriodicWorkerInterval time.Duration `file:"default_periodic_worker_interval" default:"30s"`
}

// +provider
type provider struct {
	Cfg     *config
	Log     logs.Logger
	Report  report.MetricReport `autowired:"metric-report-client" optional:"true"`
	Fetcher fetcher.Interface   `autowired:"erda.msp.apm.checker.task.fetcher"`
	events  <-chan *fetcher.Event
	plugins map[string]plugins.Interface
	workers WorkerManager
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.DefaultPeriodicWorkerInterval > 0 {
		defaultPeriodicInterval = p.Cfg.DefaultPeriodicWorkerInterval
	}
	p.workers = newSimpleWorkerManager(p.Log)
	p.plugins = make(map[string]plugins.Interface)
	ctx.Hub().ForeachServices(func(service string) bool {
		if strings.HasPrefix(service, "erda.msp.apm.checker.task.plugins.") {
			plugin, ok := ctx.Service(service).(plugins.Interface)
			if !ok {
				panic(fmt.Errorf("service %s is not checker plugins", service))
			}
			name := service[len("erda.msp.apm.checker.task.plugins."):]
			p.Log.Debugf("load checker plugin %q", name)
			p.plugins[name] = plugin
		}
		return true
	})
	p.events = p.Fetcher.Watch()
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	taskCtx := newTaskContext(ctx, p.Report)
	for {
		select {
		case event, ok := <-p.events:
			if !ok {
				return nil
			}
			switch event.Action {
			case fetcher.ActionAdd, fetcher.ActionUpdate:
				plugin, ok := p.plugins[event.Data.Type]
				if !ok {
					p.Log.Debugf("invalid checker plugin type %q", event.Data.Type)
					continue
				}
				err := plugin.Validate(event.Data)
				if err != nil {
					p.Log.Debugf("invalid checker checker: %v", err)
					continue
				}

				// setup tags
				if event.Data.Tags == nil {
					event.Data.Tags = make(map[string]string)
				}
				event.Data.Tags["_meta"] = "true"
				event.Data.Tags["type"] = event.Data.Type

				// new checker handler
				checker, err := plugin.New(event.Data)
				if err != nil {
					p.Log.Debugf("fail to create %s checker: %v", event.Data.Type, err)
					continue
				}
				w, err := NewWorker(p.Log, event.Data, taskCtx, checker)
				if err != nil {
					p.Log.Debugf("fail to create %s checker worker: %v", event.Data.Type, err)
					continue
				}
				p.Log.Infof("start worker for checker(id=%d type=%s)", event.Data.Id, event.Data.Type)
				go w.Run(ctx)
				p.workers.Put(event.Data.Id, w)
			case fetcher.ActionDelete:
				p.workers.Remove(event.Data.Id)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func init() {
	servicehub.Register("erda.msp.apm.checker.task", &servicehub.Spec{
		DependenciesFunc: func(hub *servicehub.Hub) (list []string) {
			hub.ForeachServices(func(service string) bool {
				if strings.HasPrefix(service, "erda.msp.apm.checker.task.plugins.") {
					list = append(list, service)
				}
				return true
			})
			return list
		},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
