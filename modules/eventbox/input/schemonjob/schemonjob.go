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

package schemonjob

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/input"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

type ScheMonJob struct {
	handler input.Handler
	js      jsonstore.JSONStoreWithWatch
	// map[<namespace/name>]context_cancel
	ctxMap map[string]context.CancelFunc
	stopCh chan struct{}
	stopWg sync.WaitGroup
	lock   *dlock.DLock
}

func New() (input.Input, error) {
	lock, err := dlock.New(constant.ScheMonJobLockKey, func() {})
	if err != nil {
		return nil, err
	}
	js_, err := jsonstore.New()
	if err != nil {
		return nil, err
	}
	js := js_.(jsonstore.JSONStoreWithWatch)

	return &ScheMonJob{
		js:     js,
		ctxMap: make(map[string]context.CancelFunc),
		stopCh: make(chan struct{}),
		lock:   lock,
	}, nil
}

func (*ScheMonJob) Name() string {
	return "SCHEMONJOB"
}

func (s *ScheMonJob) Start(handler input.Handler) error {
	s.handler = handler
	s.stopWg.Add(1)
	defer s.stopWg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-s.stopCh
		cancel()
	}()

	cleanup, err := input.OnlyOne(ctx, s.lock)
	defer cleanup()

	if err != nil {
		logrus.Errorf("ScheMonJob lock: %v", err)
	}

	f := func() error {
		err := s.js.Watch(ctx, constant.ScheMonJobWatchDir, true, false, true, nil, func(k string, _ interface{}, t storetypes.ChangeType) error {
			other, name := filepath.Split(k)
			_, namespace := filepath.Split(other[:len(other)-1])
			if t == storetypes.Del {
				logrus.Infof("ScheMonJob: watch delete [%s:%s]", namespace, name)
				cancel, ok := s.ctxMap[filepath.Join(namespace, name)]
				if !ok {
					logrus.Warnf("ScheMonJob: no cancel for [%s:%s]", namespace, name)
					return nil
				}
				cancel()
				delete(s.ctxMap, filepath.Join(namespace, name))
				return nil
			} else {
				logrus.Infof("ScheMonJob: watch [%s:%s]", namespace, name)
				if _, ok := s.ctxMap[filepath.Join(namespace, name)]; ok {
					logrus.Warnf("ScheMonJob: already watching [%s:%s]", namespace, name)
					return nil
				}
				ctx_, cancel_ := context.WithCancel(ctx)
				go loopQueryStatus(ctx_, namespace, name, handler)
				s.ctxMap[filepath.Join(namespace, name)] = cancel_
				return nil
			}
		})
		return err
	}
	if err := s.recover(); err != nil {
		return err
	}
	if err := f(); err != nil {
		return err
	}
	return nil
}

func (s *ScheMonJob) Stop() error {
	s.stopCh <- struct{}{}
	logrus.Info("ScheMonJob: stopping")
	s.stopWg.Wait()
	logrus.Info("ScheMonJob: stopped")
	return nil
}

func (s *ScheMonJob) recover() error {
	logrus.Info("ScheMonJob recover start")
	defer logrus.Info("ScheMonJob recover done")
	keys, err := s.js.ListKeys(context.Background(), constant.ScheMonJobWatchDir)
	if err != nil {
		return err
	}
	for _, key := range keys {
		other, name := filepath.Split(key)
		_, namespace := filepath.Split(other[:len(other)-1])
		path := fmt.Sprintf(constant.ScheMonJobQueryURL, namespace, name)
		var jbody StatusForEventbox
		resp, err := httpclient.New().Get(discover.Scheduler()).Path(path).Do().JSON(&jbody)
		if err != nil {
			logrus.Errorf("ScheMonJob recover: %v, key: %s", err, key)
			continue
		}
		if !resp.IsOK() {
			logrus.Errorf("ScheMonJob recover: resp code: %d, key: %s", resp.StatusCode(), key)
			continue
		}
		if err := syncToCallBackURL(&jbody, s.handler); err != nil {
			logrus.Errorf("ScheMonJob recover: %v, key: %s", err, key)
			continue
		}
	}
	return nil
}

// 如果请求 constant.ScheMonJobQueryURL 失败，则按规律增加下一次 query 前的 sleep 时间
func sleepDurationBeforeNext(t int, errcase bool) int {
	if errcase {
		return t + 5
	}
	return rand.Intn(5)
}

func loopQueryStatus(ctx context.Context, namespace, name string, handler input.Handler) {
	lastStatus := &StatusForEventbox{}
	next := sleepDurationBeforeNext(0, false)
	path := fmt.Sprintf(constant.ScheMonJobQueryURL, namespace, name)
	for {
		timer := time.NewTimer(time.Duration(next) * time.Second)
		select {
		case <-ctx.Done():
			return
		case <-timer.C: // go on
		}
		originNext := next
		var jbody StatusForEventbox
		resp, err := httpclient.New().Get(discover.Scheduler()).Path(path).Do().JSON(&jbody)
		if err != nil {
			logrus.Errorf("loopQueryStatus: %v", err)
			next = sleepDurationBeforeNext(originNext, true)
			continue
		}
		if !resp.IsOK() {
			logrus.Errorf("loopQueryStatus resp code:%d", resp.StatusCode())
			next = sleepDurationBeforeNext(originNext, true)
			continue

		}
		if Diff(&jbody, lastStatus) {
			continue
		}
		err = syncToCallBackURL(&jbody, handler)
		next = sleepDurationBeforeNext(originNext, err != nil)
		if err == nil {
			lastStatus = &jbody
		}
	}
}
func syncToCallBackURL(stat *StatusForEventbox, handler input.Handler) error {
	m := types.Message{
		Sender:  "eventbox-schemon",
		Content: stat,
		Labels:  map[types.LabelKey]interface{}{"HTTP": stat.Addr},
		Time:    time.Now().UnixNano(),
	}

	derr := handler(&m)
	if !derr.IsOK() {
		logrus.Errorf("syncToCallBackURL:%v", derr.String())
		return errors.New(derr.String())
	}
	return nil
}
