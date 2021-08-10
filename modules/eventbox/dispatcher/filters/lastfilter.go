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

package filters

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/goroutinepool"
)

// Filter的最后一个，实际上用来做后续对 message 的操作
type LastFilter struct {
	subscribers map[string]subscriber.Subscriber
	pools       map[string]*goroutinepool.GoroutinePool
}

func NewLastFilter(pools map[string]*goroutinepool.GoroutinePool, subscribers map[string]subscriber.Subscriber) Filter {
	return &LastFilter{
		subscribers: subscribers,
		pools:       pools,
	}
}

func (l *LastFilter) Name() string {
	return "LastFilter"
}

func (l *LastFilter) Filter(m *types.Message) *errors.DispatchError {
	derr := errors.New()
	publishErrM := make(map[string](chan []error))
	for name, sub := range l.subscribers {
		for k, v := range m.Labels {
			if k.Equal(name) {
				publishErrCh, poolErr := throttlePublish(m, l.pools[name], sub, v)
				if poolErr != nil {
					derr.BackendErrs[name] = []error{poolErr}
					continue
				} else {
					publishErrM[name] = publishErrCh
				}
			}
		}
	}
	for name, publishErr := range publishErrM {
		errs := <-publishErr
		if len(errs) > 0 {
			derr.BackendErrs[name] = errs
		}
	}
	return derr
}

func throttlePublish(m *types.Message, pool *goroutinepool.GoroutinePool, sub subscriber.Subscriber, labelV interface{}) (chan []error, error) {
	errsCh := make(chan []error, 1)
	f := func() {
		content_, err := json.Marshal(m.Content)
		if err != nil {
			errsCh <- []error{err}
			close(errsCh)
			return
		}
		labelV_, err := json.Marshal(labelV)
		if err != nil {
			errsCh <- []error{err}
			close(errsCh)
			return
		}
		errs := sub.Publish(string(labelV_), string(content_), m.Time, m)
		if len(errs) > 0 {
			logrus.Errorf("publish message: %v, err: %v", m, errs)
			errsCh <- errs
		} else {
			errsCh <- []error{}
		}
		close(errsCh)
	}

	if err := pool.Go(f); err != nil {
		if err == goroutinepool.NoMoreWorkerErr {
			rand.Seed(time.Now().Unix())
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // retry after random time
			if err = pool.Go(f); err != nil {
				close(errsCh)
				return nil, err
			}
		} else {
			close(errsCh)
			return nil, err
		}

	}
	return errsCh, nil
}
