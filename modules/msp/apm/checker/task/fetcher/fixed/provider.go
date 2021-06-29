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

package fixed

import (
	"context"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/task/fetcher"
)

type config struct {
	Checkers []*pb.Checker `file:"checkers"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	watchers []chan *fetcher.Event
}

func (p *provider) Run(ctx context.Context) error {
	if len(p.watchers) <= 0 {
		return nil
	}
	for _, c := range p.Cfg.Checkers {
		for _, w := range p.watchers {
			select {
			case w <- &fetcher.Event{
				Action: fetcher.ActionAdd,
				Data:   c,
			}:
			case <-ctx.Done():
				return nil
			}
		}
	}
	return nil
}

func (p *provider) Watch() <-chan *fetcher.Event {
	ch := make(chan *fetcher.Event, 16)
	p.watchers = append(p.watchers, ch)
	return ch
}

func init() {
	servicehub.Register("erda.msp.apm.checker.task.fetcher.fixed", &servicehub.Spec{
		Services:   []string{"erda.msp.apm.checker.task.fetcher"},
		Types:      []reflect.Type{reflect.TypeOf((*fetcher.Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
