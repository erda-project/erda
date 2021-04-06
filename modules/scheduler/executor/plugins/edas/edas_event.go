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

package edas

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (m *EDAS) registerEventChanAndLocalStore(evCh chan *eventtypes.StatusEvent, stopCh chan struct{}, lstore *sync.Map) {
	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"

	name := string(m.name)

	logrus.Infof("edas registerEventChanAndLocalStore, name: %s", name)

	// watch 特定etcd目录的事件的处理函数
	syncRuntimeToLstore := func(key string, value interface{}, t storetypes.ChangeType) error {
		runtimeName := etcdKeyToMapKey(key)
		if len(runtimeName) == 0 {
			return nil
		}

		// 先处理delete的事件
		if t == storetypes.Del {
			_, ok := lstore.Load(runtimeName)
			if ok {
				var e events.RuntimeEvent
				e.RuntimeName = runtimeName
				e.IsDeleted = true
				lstore.Delete(runtimeName)
				m.notifier.Send(e, events.WithSender(name+events.SUFFIX_EDAS), events.WithDest(map[string]interface{}{"HTTP": []string{eventAddr}}))
			}
			return nil
		}

		run := value.(*apistructs.ServiceGroup)

		// 过滤不属于本executor的事件
		if run.Executor != name {
			return nil
		}

		switch t {
		// Add or update event
		case storetypes.Add, storetypes.Update:
			lstore.Store(runtimeName, *run)

		default:
			logrus.Errorf("unknown store type, try to skip, type: %s, name: %s", t, runtimeName)
			return nil
		}

		if strings.Contains(name, "EDAS") {
			logrus.Infof("edas executor(%s) lstore stored key: %s", name, key)
		}
		return nil
	}

	// 将注册来的executor的name及其event channel对应起来
	getEvChanFn := func(executorName executortypes.Name) (chan *eventtypes.StatusEvent, chan struct{}, *sync.Map, error) {
		logrus.Infof("in RegisterEvChan executor(%s)", name)
		if string(executorName) == name {
			return evCh, stopCh, lstore, nil
		}
		return nil, nil, nil, errors.Errorf("this is for %s executor, not %s", executorName, name)
	}

	executortypes.RegisterEvChan(executortypes.Name(name), getEvChanFn, syncRuntimeToLstore)
}

// 目前edas没有使用事件驱动机制，因而每次用轮询去模拟
func (m *EDAS) WaitEvent(lstore *sync.Map, stopCh chan struct{}) {
	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"

	logrus.Infof("executor(%s) in waitEvent", m.name)

	initStore := func(k string, v interface{}) error {
		reKey := etcdKeyToMapKey(k)
		if len(reKey) == 0 {
			return nil
		}

		r := v.(*apistructs.ServiceGroup)
		if r.Executor != string(m.name) {
			return nil
		}
		lstore.Store(reKey, *r)
		return nil
	}

	em := events.GetEventManager()
	if err := em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initStore); err != nil {
		logrus.Errorf("executor(%s) foreach initStore error: %v", m.name, err)
	}

	keys := make([]string, 0)

	f := func(k, v interface{}) bool {
		r := v.(apistructs.ServiceGroup)

		// double check
		if r.Executor != string(m.name) {
			return true
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var err error
		c := make(chan struct{}, 1)

		go func() {
			start := time.Now()
			defer func() {
				logrus.Infof("edas executor(%s) get status for key(%s) took %v", m.name, k.(string), time.Since(start))
			}()
			_, err = m.Status(ctx, r)
			c <- struct{}{}
		}()

		keys = append(keys, k.(string))
		//logrus.Debugf("in edas loop inside f before status, key: %s", k.(string))
		select {
		case <-c:
			if err != nil {
				logrus.Errorf("executor(%s)'s key(%s) for edas get status error: %v", m.name, k, err)
				return true
			}

			var e events.RuntimeEvent
			e.EventType = events.EVENTS_TOTAL
			e.RuntimeName = k.(string)
			e.ServiceStatuses = make([]events.ServiceStatus, len(r.Services))
			for i, srv := range r.Services {
				e.ServiceStatuses[i].ServiceName = srv.Name
				e.ServiceStatuses[i].Replica = srv.Scale
				e.ServiceStatuses[i].ServiceStatus = convertServiceStatus(srv.Status)
				// 状态为空字符串
				if len(e.ServiceStatuses[i].ServiceStatus) == 0 {
					e.ServiceStatuses[i].ServiceStatus = convertServiceStatus(apistructs.StatusProgressing)
				}
				// 通过 edas 的 Status 接口拿服务状态目前没有附带实例信息，
				// 如果上层将某服务的实例数设置为 0，则将edas的服务设置为健康状态
				if e.ServiceStatuses[i].Replica == 0 {
					e.ServiceStatuses[i].ServiceStatus = string(apistructs.StatusHealthy)
				}
				/*
					e.ServiceStatuses[i].InstanceStatuses = make([]events.InstanceStatus, len(srv.InstanceInfos))
					for j := range srv.InstanceInfos {
						e.ServiceStatuses[i].InstanceStatuses[j].Ip = srv.InstanceInfos[j].Ip
						e.ServiceStatuses[i].InstanceStatuses[j].ID = srv.InstanceInfos[j].Id
						if srv.InstanceInfos[j].Status == "Running" {
							e.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = "Healthy"
						} else {
							//TODO. modify this
							e.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = srv.InstanceInfos[j].Status
						}
					}
				*/
			}

			go m.notifier.Send(e, events.WithSender(string(m.name)+events.SUFFIX_EDAS), events.WithDest(map[string]interface{}{"HTTP": []string{eventAddr}}))
			return true

		case <-ctx.Done():
			logrus.Errorf("executor(%s)'s key(%s) get status timeout", m.name, k)
			return true
		}
	}

	for {
		select {
		case <-stopCh:
			logrus.Errorf("edas executor(%s) got stop chan message", m.name)
			return
		case <-time.After(10 * time.Second):
			lstore.Range(f)
		}

		logrus.Infof("executor(%s) edas keys list: %v", m.name, keys)
		keys = nil
	}
}

// 对etcd里的原始key做个处理
// /dice/service/services/staging-77 -> "services/staging-77"
func etcdKeyToMapKey(eKey string) string {
	fields := strings.Split(eKey, "/")
	if l := len(fields); l > 2 {
		return fields[l-2] + "/" + fields[l-1]
	}
	return ""
}

// Status 接口通过 edas 拿到的服务的状态的修改
func convertServiceStatus(serviceStatus apistructs.StatusCode) string {
	switch serviceStatus {
	case apistructs.StatusReady:
		return string(apistructs.StatusHealthy)

	case apistructs.StatusProgressing:
		return string(apistructs.StatusUnHealthy)

	default:
		return string(apistructs.StatusUnknown)
	}
}
