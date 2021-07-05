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

package scheduled

import (
	"context"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/task/fetcher"
)

// Dispatcher
type Dispatcher struct {
	log logs.Logger

	checkers map[int64]*pb.Checker
	lister   func() (map[int64]*pb.Checker, error)
	reload   time.Duration

	putCh    chan *pb.Checker
	removeCh chan int64
	watchers []chan *fetcher.Event
}

func NewDispatcher(lister func() (map[int64]*pb.Checker, error), reload time.Duration, log logs.Logger) *Dispatcher {
	return &Dispatcher{
		log:      log,
		checkers: make(map[int64]*pb.Checker),
		lister:   lister,
		reload:   reload,
		putCh:    make(chan *pb.Checker, 8),
		removeCh: make(chan int64, 8),
	}
}

func (p *Dispatcher) Run(ctx context.Context) error {
	load := func() error {
		checkers, err := p.lister()
		if err != nil {
			return err
		}
		for _, c := range checkers {
			if exist, ok := p.checkers[c.Id]; ok {
				if compareChecker(c, exist) {
					continue
				}
				p.notify(&fetcher.Event{
					Action: fetcher.ActionUpdate,
					Data:   copyChecker(c),
				})
				p.log.Debugf("update checker %v to %v", c, exist)
			} else {
				p.notify(&fetcher.Event{
					Action: fetcher.ActionAdd,
					Data:   copyChecker(c),
				})
			}
			p.checkers[c.Id] = c
		}
		for id, c := range p.checkers {
			if _, ok := checkers[c.Id]; !ok {
				p.notify(&fetcher.Event{
					Action: fetcher.ActionDelete,
					Data:   copyChecker(c),
				})
				delete(p.checkers, id)
			}
		}
		return nil
	}

	err := load()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case c := <-p.putCh:
			if exist, ok := p.checkers[c.Id]; ok {
				if compareChecker(c, exist) {
					continue
				}
				p.notify(&fetcher.Event{
					Action: fetcher.ActionUpdate,
					Data:   copyChecker(c),
				})
				p.log.Debugf("update checker %v to %v", c, exist)
			} else {
				p.notify(&fetcher.Event{
					Action: fetcher.ActionAdd,
					Data:   copyChecker(c),
				})
			}
			p.checkers[c.Id] = c
		case id := <-p.removeCh:
			if c, ok := p.checkers[id]; ok {
				p.notify(&fetcher.Event{
					Action: fetcher.ActionDelete,
					Data:   copyChecker(c),
				})
				delete(p.checkers, id)
			}
		case <-time.After(p.reload):
		}

		err := load()
		if err != nil {
			p.log.Errorf("fail to load checkers: %s", err)
		}
	}
}

func (p *Dispatcher) Watch() <-chan *fetcher.Event {
	ch := make(chan *fetcher.Event, 16)
	p.watchers = append(p.watchers, ch)
	return ch
}

func (p *Dispatcher) notify(e *fetcher.Event) {
	for _, w := range p.watchers {
		w <- e
	}
}

func (p *Dispatcher) Put(c *pb.Checker) {
	p.putCh <- c
}

func (p *Dispatcher) Remove(id int64) {
	p.removeCh <- id
}

func compareChecker(a, b *pb.Checker) bool {
	if a.Id != b.Id || a.Name != b.Name || a.Type != b.Type {
		return false
	}
	if !reflect.DeepEqual(a.Config, b.Config) || !reflect.DeepEqual(a.Tags, b.Tags) {
		return false
	}
	return true
}

func copyChecker(c *pb.Checker) *pb.Checker {
	ck := &pb.Checker{
		Id:   c.Id,
		Name: c.Name,
		Type: c.Type,
	}
	if c.Config != nil {
		ck.Config = make(map[string]string)
		for k, v := range c.Config {
			ck.Config[k] = v
		}
	}
	if c.Tags != nil {
		ck.Tags = make(map[string]string)
		for k, v := range c.Tags {
			ck.Tags[k] = v
		}
	}
	return ck
}
