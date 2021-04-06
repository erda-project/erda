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

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/uuid"
)

const (
	eDir = "/websocket/events"
	ttl  = 60
)

type Subscriber struct {
	js        jsonstore.JSONStoreWithWatch
	eventC    chan *Event
	stopC     chan struct{}
	runningWg sync.WaitGroup
}

func NewSubscriber(ec chan *Event) (*Subscriber, error) {
	js, err := jsonstore.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create json storage instance")
	}

	js_ := js.(jsonstore.JSONStoreWithWatch)

	return &Subscriber{
		js:     js_,
		eventC: ec,
		stopC:  make(chan struct{}),
	}, nil
}

func (s *Subscriber) retry(f func() error) {
	err := f()
	if err != nil {
		s.retry(f)
	}
}

func (s *Subscriber) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-s.stopC
		cancel()
	}()

	f := func() error {
		err := s.js.Watch(ctx, eDir, true, true, false, Event{},
			func(key string, value interface{}, _ storetypes.ChangeType) error {
				event := value.(*Event)
				go func() {
					logrus.Debugf("watch event %v", event)
					s.eventC <- event
				}()
				return nil
			})
		return err
	}

	s.retry(f)
}

func (s *Subscriber) Stop() {
	s.stopC <- struct{}{}
	s.runningWg.Wait()
	logrus.Warn("stop subscriber")
}

type Publisher struct {
	es *etcd.Store
}

func NewPublisher() (*Publisher, error) {
	es, err := etcd.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etcd store instance")
	}
	return &Publisher{es: es}, nil
}

func (p *Publisher) EmitEvent(ctx context.Context, e Event) error {
	path := generatePath()
	lease := clientv3.NewLease(p.es.GetClient())
	r, err := lease.Grant(ctx, ttl)
	if err != nil {
		return errors.Wrap(err, "failed to grant etcd lease")
	}
	b, err := json.Marshal(e)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal event %v", e)
	}
	if _, err := p.es.PutWithOption(ctx, path, string(b), []interface{}{clientv3.WithLease(r.ID)}); err != nil {
		return err
	}
	logrus.Debugf("emit ws event %+v", e)
	return nil
}

func generatePath() string {
	uuid := fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.UUID())
	return filepath.Join(eDir, uuid)
}
