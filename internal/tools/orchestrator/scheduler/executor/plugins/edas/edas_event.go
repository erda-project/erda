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

package edas

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events/eventtypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

func (e *EDAS) registerEventChanAndLocalStore(evCh chan *eventtypes.StatusEvent, stopCh chan struct{}, lstore *sync.Map) {
	l := e.l.WithField("func", "registerEventChanAndLocalStore")

	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"

	name := string(e.name)

	l.Infof("edas registerEventChanAndLocalStore, name: %s", name)

	// watch event handler for a specific etcd directory
	syncRuntimeToLstore := func(key string, value interface{}, t storetypes.ChangeType) error {
		runtimeName := etcdKeyToMapKey(key)
		if len(runtimeName) == 0 {
			return nil
		}

		// Deal with the delete event first
		if t == storetypes.Del {
			_, ok := lstore.Load(runtimeName)
			if ok {
				var event events.RuntimeEvent
				event.RuntimeName = runtimeName
				event.IsDeleted = true
				lstore.Delete(runtimeName)
				e.notifier.Send(e, events.WithSender(name+events.SUFFIX_EDAS), events.WithDest(map[string]interface{}{"HTTP": []string{eventAddr}}))
			}
			return nil
		}

		run := value.(*apistructs.ServiceGroup)

		// Filter events that do not belong to this executor
		if run.Executor != name {
			return nil
		}

		switch t {
		// Add or update event
		case storetypes.Add, storetypes.Update:
			lstore.Store(runtimeName, *run)

		default:
			l.Errorf("unknown store type, try to skip, type: %s, name: %s", t, runtimeName)
			return nil
		}

		if strings.Contains(name, "EDAS") {
			l.Infof("edas executor(%s) lstore stored key: %s", name, key)
		}
		return nil
	}

	// Correspond the name of the registered executor and its event channel
	getEvChanFn := func(executorName executortypes.Name) (chan *eventtypes.StatusEvent, chan struct{}, *sync.Map, error) {
		l.Infof("in RegisterEvChan executor(%s)", name)
		if string(executorName) == name {
			return evCh, stopCh, lstore, nil
		}
		return nil, nil, nil, errors.Errorf("this is for %s executor, not %s", executorName, name)
	}

	executortypes.RegisterEvChan(executortypes.Name(name), getEvChanFn, syncRuntimeToLstore)
}

// Currently edas does not use an event-driven mechanism, so it uses polling to simulate each time
func (e *EDAS) WaitEvent(lstore *sync.Map, stopCh chan struct{}) {
	l := e.l.WithField("func", "WaitEvent")

	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"

	l.Infof("executor(%s) in waitEvent", e.name)

	initStore := func(k string, v interface{}) error {
		reKey := etcdKeyToMapKey(k)
		if len(reKey) == 0 {
			return nil
		}

		r := v.(*apistructs.ServiceGroup)
		if r.Executor != string(e.name) {
			return nil
		}
		lstore.Store(reKey, *r)
		return nil
	}

	em := events.GetEventManager()
	if err := em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initStore); err != nil {
		l.Errorf("executor(%s) foreach initStore error: %v", e.name, err)
	}

	keys := make([]string, 0)

	f := func(k, v interface{}) bool {
		r := v.(apistructs.ServiceGroup)

		// double check
		if r.Executor != string(e.name) {
			return true
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var err error
		c := make(chan struct{}, 1)

		go func() {
			start := time.Now()
			defer func() {
				l.Infof("edas executor(%s) get status for key(%s) took %v", e.name, k.(string), time.Since(start))
			}()
			_, err = e.Status(ctx, r)
			c <- struct{}{}
		}()

		keys = append(keys, k.(string))
		//logrus.Debugf("in edas loop inside f before status, key: %s", k.(string))
		select {
		case <-c:
			if err != nil {
				l.Errorf("executor(%s)'s key(%s) for edas get status error: %v", e.name, k, err)
				return true
			}

			var event events.RuntimeEvent
			event.EventType = events.EVENTS_TOTAL
			event.RuntimeName = k.(string)
			event.ServiceStatuses = make([]events.ServiceStatus, len(r.Services))
			for i, srv := range r.Services {
				event.ServiceStatuses[i].ServiceName = srv.Name
				event.ServiceStatuses[i].Replica = srv.Scale
				event.ServiceStatuses[i].ServiceStatus = convertServiceStatus(srv.Status)
				// Status is empty string
				if len(event.ServiceStatuses[i].ServiceStatus) == 0 {
					event.ServiceStatuses[i].ServiceStatus = convertServiceStatus(apistructs.StatusProgressing)
				}
				//Get the service status through the Status interface of edas. Currently there is no instance information attached.
				//If the upper layer sets the number of instances of a service to 0, then the edas service is set to a healthy state
				if event.ServiceStatuses[i].Replica == 0 {
					event.ServiceStatuses[i].ServiceStatus = string(apistructs.StatusHealthy)
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

			go e.notifier.Send(e, events.WithSender(string(e.name)+events.SUFFIX_EDAS), events.WithDest(map[string]interface{}{"HTTP": []string{eventAddr}}))
			return true

		case <-ctx.Done():
			l.Errorf("executor(%s)'s key(%s) get status timeout", e.name, k)
			return true
		}
	}

	for {
		select {
		case <-stopCh:
			l.Errorf("edas executor(%s) got stop chan message", e.name)
			return
		case <-time.After(10 * time.Second):
			lstore.Range(f)
		}

		l.Infof("executor(%s) edas keys list: %v", e.name, keys)
		keys = nil
	}
}

// Treat the original key in etcd
// /dice/service/services/staging-77 -> "services/staging-77"
func etcdKeyToMapKey(eKey string) string {
	fields := strings.Split(eKey, "/")
	if l := len(fields); l > 2 {
		return fields[l-2] + "/" + fields[l-1]
	}
	return ""
}

// Modification of the status of the service obtained through the Status interface through edas
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
