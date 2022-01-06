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

package etcd

import (
	"context"
	"crypto/md5" // #nosec G501
	"encoding/json"
	"sync"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/input"
	"github.com/erda-project/erda/modules/eventbox/monitor"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/modules/pkg/etcdclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

type EtcdInput struct {
	js      jsonstore.JSONStoreWithWatch
	handler func(*types.Message) *errors.DispatchError
	// 用来通知 stop
	stopCh    chan struct{}
	runningWg sync.WaitGroup
	// 用于过滤相同事件, 使用 jsonstore lru backend, key: content-md5, value: 第一次出现的unixnano时间
	filter jsonstore.JsonStore
	// etcd client
	eclient *v3.Client
}

func New() (input.Input, error) {
	js, err := jsonstore.New()
	if err != nil {
		return nil, err
	}
	js_ := js.(jsonstore.JSONStoreWithWatch)
	if err != nil {
		return nil, err
	}
	filter, err := jsonstore.New(
		jsonstore.UseLruStore(100),
		jsonstore.UseMemStore())
	if err != nil {
		return nil, err
	}
	eclient, err := etcdclient.NewEtcdClient()
	if err != nil {
		return nil, err
	}
	return &EtcdInput{
		js:      js_,
		stopCh:  make(chan struct{}),
		filter:  filter,
		eclient: eclient,
	}, nil
}

func (e *EtcdInput) Name() string {
	return "ETCD"
}

// 处理 etcd 中已经存在的消息:
// 重新 etcd.put 一遍所有存在的消息
// 使 watch 能监听到
// NOTE: 只 put 在当前时间之前的消息
func (e *EtcdInput) restore(watching chan int64) {
	logrus.Info("Etcdinput restore start")
	defer logrus.Info("Etcdinput restore end")
	ctx := context.Background()
	var before int64
	// TODO:  better impl?
	// put a useless message to make sure etcd watch has started
	for {
		time.Sleep(500 * time.Millisecond)
		e.js.Put(ctx, constant.MessageDir+"/000",
			types.Message{
				Sender: "eventbox-self",
				Labels: map[types.LabelKey]interface{}{
					"FAKE": "restore",
				},
				Content: "restore",
				Time:    0})
		var ok bool
		before, ok = <-watching
		if ok {
			break
		}

	}
	e.js.ForEach(ctx, constant.MessageDir, types.Message{},
		func(k string, v interface{}) error {
			if !(v.(*types.Message)).Before(before) {
				return nil
			}
			if err := e.js.Put(ctx, k, v); err != nil {
				logrus.Errorf("Etcdinput restore: %v", err)
			}
			return nil
		})
}
func (e *EtcdInput) retry(f func() error) {
	err := f()
	if err != nil {
		logrus.Errorf("Etcdinput: %v", err)
		e.retry(f)
	}
}

func (e *EtcdInput) Start(handler input.Handler) error {
	e.handler = handler
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-e.stopCh
		cancel()
	}()

	watching := make(chan int64)
	var watchingOnce sync.Once
	f := func() error {
		err := e.js.Watch(ctx, constant.MessageDir, true, true, false, types.Message{},
			func(key string, value interface{}, _ storetypes.ChangeType) error {
				watchingOnce.Do(func() {
					watching <- time.Now().UnixNano()
					close(watching)
					logrus.Info("Etcdinput: start watching")
				})
				r, err := e.eclient.Txn(ctx).If(v3.Compare(v3.Version(key), "!=", 0)).Then(v3.OpDelete(key)).Commit()
				if err != nil {
					logrus.Errorf("%+v", err)
				}
				if r != nil && !r.Succeeded { // has been handled by other eventbox instance
					return nil
				}
				message := value.(*types.Message)
				monitor.Notify(monitor.MonitorInfo{Tp: monitor.EtcdInput})
				if !filter(e.filter, message) {
					// drop
					monitor.Notify(monitor.MonitorInfo{Tp: monitor.EtcdInputDrop})
					return nil
				}
				go func() {
					derr := e.handler(message)
					if derr != nil && !derr.IsOK() {
						logrus.Errorf("%v", derr)
					}
				}()
				return nil
			})
		return err
	}
	go e.restore(watching)
	e.runningWg.Add(1)
	e.retry(f)
	e.runningWg.Done()
	e.runningWg.Wait()
	logrus.Info("Etcdinput start() done")
	return nil
}

func (e *EtcdInput) Stop() error {
	e.stopCh <- struct{}{}
	logrus.Info("Etcdinput: stopping")
	e.runningWg.Wait()
	logrus.Info("Etcdinput: stopped")

	return nil
}

// 过滤 5s 内相同的 message
// true: pass
// false: filter
func filter(lru jsonstore.JsonStore, m *types.Message) bool {
	h := md5.New() // #nosec G401
	m_ := *m
	m_.Time = 0 // when compute md5, ignore Time field
	content, err := json.Marshal(m_)
	if err != nil {
		logrus.Errorf("Etcdinput: filter: %v", err)
		return true
	}
	h.Write(content)
	md5 := string(h.Sum(nil))
	t := new(int64)
	if err := lru.Get(context.Background(), md5, t); err != nil {
		// not found same message in filter, so is pass, and put it in filter
		if err := lru.Put(context.Background(), md5, time.Now().UnixNano()); err != nil {
			logrus.Errorf("Etcdinput: filter: put %s failed, %v", md5, err)
		}
		return true
	}
	t_ := time.Unix(0, *t)
	if !t_.After(time.Now().Add(-5 * time.Second)) {
		// too old , delete it
		if err := lru.Remove(context.Background(), md5, t); err != nil {
			logrus.Errorf("Etcdinput: filter: remove %s failed, %v", md5, err)
		}
		// add current
		if err := lru.Put(context.Background(), md5, time.Now().UnixNano()); err != nil {
			logrus.Errorf("Etcdinput: filter: put %s failed, %v", md5, err)
		}
		return true
	}

	// here, we found the same message in filter, so drop it
	return false
}
