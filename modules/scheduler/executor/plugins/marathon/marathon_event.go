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

package marathon

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/marathon/instanceinfosync"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/customhttp"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/loop"
	_ "github.com/erda-project/erda/pkg/monitor"
)

// TODO: Re-establish connections in consideration of marathon cluster restarts, exceptions, etc.
// TODO: Regularly obtain status from marathon api (implemented by aggregation computing layer)
func (m *Marathon) WaitEvent(options map[string]string, monitor bool, killedInstanceCh chan string, stopCh chan struct{}) {
	// Wait for 5 to 30 seconds to start listening for events
	waitSeconds := 2 + rand.Intn(5)
	logrus.Infof("executor(name:%s, addr:%s) in WaitEvent, sleep %vs", m.name, m.addr, waitSeconds)
	time.Sleep(time.Duration(waitSeconds) * time.Second)

	defer logrus.Infof("executor(%s) WaitEvent exited", m.name)

	useHttps := false
	var clientCrt string
	var clientKey string
	var caCrt string

	f := func(s string) string { return strings.Replace(s, "\n", "\\n", -1) }

	if _, ok := options["CA_CRT"]; ok {
		useHttps = true
		caCrt = options["CA_CRT"]
		clientCrt = options["CLIENT_CRT"]
		clientKey = options["CLIENT_KEY"]
		logrus.Infof("init executor(%s) ca.crt: %s, client.crt: %s, client.key: %s",
			m.name, f(caCrt), f(clientCrt), f(clientKey))
	}

	eventHandler := func() (bool, error) {
		select {
		case <-stopCh:
			logrus.Errorf("executor(%s) event got stop chan message", m.name)
			// Terminate loop execution
			return true, errors.Errorf("got stop message and exit")
		default:
		}

		// Don't use m.client, avoid its timeout time and other constraints
		var (
			req *http.Request
			err error
		)

		client := http.DefaultClient
		url := m.addr + "/v2/events"

		if useHttps {
			if !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}
			keyPair, ca := withHttpsCertFromJSON([]byte(clientCrt),
				[]byte(clientKey),
				[]byte(caCrt))

			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      ca,
					Certificates: []tls.Certificate{keyPair},
				},
			}
		} else if !strings.HasPrefix(url, "inet://") {
			if !strings.HasPrefix(url, "http://") {
				url = "http://" + url
			}
		}

		req, err = customhttp.NewRequest("GET", url, nil)
		if err != nil {
			logrus.Errorf("[alert] construct WaitEvent request err: %v", err)
			return false, err
		}

		q := req.URL.Query()
		q.Add("event_type", events.STATUS_UPDATE_EVENT)
		q.Add("event_type", events.INSTANCE_HEALTH_CHANGED_EVENT)
		req.URL.RawQuery = q.Encode()

		basicAuth := options["BASICAUTH"]
		if len(basicAuth) > 0 {
			basticStr := base64.StdEncoding.EncodeToString([]byte(basicAuth))
			req.Header.Set("Authorization", "Basic "+basticStr)
		}

		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Transfer-Encoding", "chunked")
		req.Header.Set("Portal-SSE", "on")

		// default client is no timeout
		resp, err := client.Do(req)
		if err != nil {
			logrus.Errorf("[alert] executor(%s) failed to request the marathon events (%v)", m.name, err)
			return false, err
		}

		reader := bufio.NewReader(resp.Body)
		// The marathon event comes two lines at a time, the format is
		// event: status_update_event
		// data: {"slaveId":"2aa23eb3-fa5c-4248-a203-ba4ddb9fb562-S2",...}
		// Except for the event itself, there will always be time to travel
		var eventType string
		var lastContent string
		for {
			select {
			case <-stopCh:
				logrus.Errorf("executor(%s) got stop chan message for sending incremental events", m.name)
				// retrun error and the loop exits
				return true, errors.Errorf("got stop message and exit")
			default:
			}

			line, err := reader.ReadBytes('\n')
			if err != nil {
				logrus.Errorf("[alert] executor(%s) failed to read data from marathon (%v)", m.name, err)
				break
			}

			// First determine the type of event
			if len(line) > 7 && string(line[:5]) == "event" {
				eventType = string(line[7:])
				eventType = strings.TrimSuffix(eventType, "\r\n")
				continue
			}

			if len(line) <= 6 || string(line[:4]) != "data" {
				continue
			}

			// Parse out the content of the event
			content := line[6:]

			// Discard non-user service events
			if !strings.Contains(string(content), "runtimes") {
				continue
			}

			switch eventType {
			case events.STATUS_UPDATE_EVENT:
				ev := MarathonStatusUpdateEvent{}
				if err := json.Unmarshal(content, &ev); err != nil {
					logrus.Errorf("unmarshal event string(%s) error: %v", string(content), err)
					continue
				}

				// Or the status is unknown event and unreachable event
				if ev.TaskStatus == events.UNKNOWN || ev.TaskStatus == events.UNREACHABLE {
					continue
				}

				statusEvent := eventtypes.StatusEvent{
					Type:    eventType,
					ID:      ev.AppId,
					Status:  ev.TaskStatus,
					TaskId:  ev.TaskId,
					Host:    ev.Host,
					Cluster: string(m.name),
					Message: ev.Message,
				}
				if len(ev.IpAddresses) > 0 {
					statusEvent.IP = ev.IpAddresses[0].IpAddress
				}
				logrus.Debugf("going to send status update ev: %+v", statusEvent)
				m.evCh <- &statusEvent

				// Count the instances that frequently hang
				if monitor {
					collectDeadInstances(&ev, killedInstanceCh)
				}

			case events.INSTANCE_HEALTH_CHANGED_EVENT:
				// instance_health_changed_event Two identical events have always been sent at the same time
				// https://jira.mesosphere.com/browse/MARATHON-8134
				// https://jira.mesosphere.com/browse/MARATHON-8494
				// Filter out "healthy":null events in instance_health_changed_event
				if string(content) == lastContent || strings.Contains(string(content), "\"healthy\":null") {
					continue
				}
				lastContent = string(content)

				ev := MarathonInstanceHealthChangedEvent{}
				if err := json.Unmarshal(content, &ev); err != nil {
					logrus.Errorf("unmarshal instance_health_changed_event string(%s) error: %v", string(content), err)
					continue
				}

				// instanceId format of instance_health_changed_event ：
				// "instanceId":"runtimes_v1_services_mydev-590_user-service.marathon-56153805-9004-11e8-aaac-70b3d5800001"

				// taskId format of status_update_event
				// "taskId":"runtimes_v1_services_mydev-590_user-service.56153805-9004-11e8-aaac-70b3d5800001"
				// Convert the format of instanceId of instance_health_changed_event to remove "marathon-"
				statusEvent := eventtypes.StatusEvent{
					Type:    eventType,
					TaskId:  modifyHealthEventId(ev.InstanceId),
					Cluster: string(m.name),
				}
				if ev.Healthy {
					statusEvent.Status = "Healthy"
				} else {
					statusEvent.Status = "UnHealthy"
				}
				logrus.Debugf("going to send instance_health_changed_event: %+v", statusEvent)
				m.evCh <- &statusEvent
				instances, err := m.db.InstanceReader().ByTaskID(statusEvent.TaskId).Do()
				if err != nil {
					logrus.Errorf("failed to get instance: %v", err)
					continue
				}

				if len(instances) == 0 {
					if err := m.db.InstanceWriter().Create(&instanceinfo.InstanceInfo{
						TaskID: statusEvent.TaskId,
						Phase: map[string]instanceinfo.InstancePhase{
							"Healthy":   instanceinfo.InstancePhaseHealthy,
							"UnHealthy": instanceinfo.InstancePhaseUnHealthy,
						}[statusEvent.Status],
					}); err != nil {
						logrus.Errorf("failed to create instance: %v", err)
						continue
					}
				} else {
					for _, ins := range instances {
						ins.Phase = instanceinfosync.InstancestatusStateMachine(
							ins.Phase,
							map[string]instanceinfo.InstancePhase{
								"Healthy":   instanceinfo.InstancePhaseHealthy,
								"UnHealthy": instanceinfo.InstancePhaseUnHealthy,
							}[statusEvent.Status])
						if err := m.db.InstanceWriter().Update(ins); err != nil {
							logrus.Errorf("failed to update instance: %v", err)
						}
					}
				}

			}
		}

		resp.Body.Close()
		return false, nil
	}

	loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*60)).Do(eventHandler)
}

func modifyHealthEventId(eid string) string {
	if strings.Contains(eid, "marathon-") {
		return strings.Replace(eid, "marathon-", "", 1)
	}
	logrus.Errorf("health_status_changed_event instanceId(%s) not contains marathon- ", eid)
	return eid
}

func withHttpsCertFromJSON(certFile, keyFile, cacrt []byte) (tls.Certificate, *x509.CertPool) {
	pair, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		logrus.Fatalf("LoadX509KeyPair: %v", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cacrt)

	return pair, pool
}

// Treat the original key in etcd
// /dice/service/services/staging-77 -> "services/staging-77"
func etcdKeyToMapKey(eKey string) string {
	fields := strings.Split(eKey, "/")
	if l := len(fields); l > 2 {
		// Filter addon services
		if strings.HasPrefix(fields[l-2], "addon") {
			return ""
		}
		return fields[l-2] + "/" + fields[l-1]
	}
	return ""
}

func registerEventChanAndLocalStore(name string, evCh chan *eventtypes.StatusEvent, stopCh chan struct{}, lstore *sync.Map) {
	// watch event handler for a specific etcd directory
	syncRuntimeToEvent := func(key string, value interface{}, t storetypes.ChangeType) error {

		runtimeName := etcdKeyToMapKey(key)
		if len(runtimeName) == 0 {
			return nil
		}

		// Deal with the delete event first
		if t == storetypes.Del {
			// TODO: Do not delete the key first, if the event is deleted, the corresponding structure cannot be found, so it is not easy to send the event
			// TODO: You can set a flag bit later, and then open a coroutine to clean up regularly
			// lstore.Delete(runtimeName)
			de_, ok := lstore.Load(runtimeName)
			if ok {
				de := de_.(events.RuntimeEvent)
				de.IsDeleted = true
				lstore.Store(runtimeName, de)
			}
			return nil
		}

		run := value.(*apistructs.ServiceGroup)

		// Filter events that do not belong to this executor
		if run.Executor != name {
			return nil
		}

		event := events.RuntimeEvent{
			RuntimeName: runtimeName,
		}

		switch t {
		// Added event
		case storetypes.Add:
			logrus.Debugf("key(%s) in added event", key)
			event.ServiceStatuses = make([]events.ServiceStatus, len(run.Services))
			for i, srv := range run.Services {
				event.ServiceStatuses[i].ServiceName = srv.Name
				event.ServiceStatuses[i].Replica = srv.Scale
				// Unknowable current instance information
				event.ServiceStatuses[i].InstanceStatuses = make([]events.InstanceStatus, len(srv.InstanceInfos))
			}

			lstore.Store(runtimeName, event)

		// Updated event
		case storetypes.Update:
			logrus.Debugf("key(%s) in updated event", key)
			oldEvent, ok := lstore.Load(runtimeName)
			if !ok {
				logrus.Errorf("key(%s) updated but not found related key in lstore", key)
				return nil
			}
			// Determine whether the number of services has changed, such as considering adding N and deleting M cases
			for _, newS := range run.Services {
				found := false
				for _, oldS := range oldEvent.(events.RuntimeEvent).ServiceStatuses {
					if oldS.ServiceName == newS.Name {
						found = true
						event.ServiceStatuses = append(event.ServiceStatuses, oldS)
						// Determine whether the number of instances in any service has changed, and there are only two cases: expansion and contraction
						// scale up
						if oldS.Replica < newS.Scale {
							i := 0
							for i < newS.Scale-oldS.Replica {
								i++
								event.ServiceStatuses[len(event.ServiceStatuses)-1].InstanceStatuses = append(
									event.ServiceStatuses[len(event.ServiceStatuses)-1].InstanceStatuses, events.InstanceStatus{})
							}
							event.ServiceStatuses[len(event.ServiceStatuses)-1].Replica = newS.Scale
						} else if len(oldS.InstanceStatuses) > newS.Scale { // scale down
							// According to marathon's strategy "killSelection": "YOUNGEST_FIRST", new instances will be killed first
							// The newer instance will be added after the slice
							event.ServiceStatuses[len(event.ServiceStatuses)-1].Replica = newS.Scale
						}
						break
					}
				}

				// Not found, indicating that newS is a service that needs to be added, no need to consider the number of instances
				if !found {
					event.ServiceStatuses = append(event.ServiceStatuses, events.ServiceStatus{
						ServiceName:      newS.Name,
						Replica:          newS.Scale,
						InstanceStatuses: make([]events.InstanceStatus, newS.Scale),
					})
				}
			}

			lstore.Store(runtimeName, event)
		}

		if t != storetypes.Del {
			logrus.Debugf("executor(%s) stored runtime(%s) event to lstore: %+v", name, runtimeName, event)
		}

		return nil
	}

	// Correspond the name of the registered executor and its event channe
	getEvChanFn := func(executorName executortypes.Name) (chan *eventtypes.StatusEvent, chan struct{}, *sync.Map, error) {
		logrus.Infof("in RegisterEvChan executor(%s)", name)
		if string(executorName) == name {
			return evCh, stopCh, lstore, nil
		}
		return nil, nil, nil, errors.Errorf("this is for %s executor, not %s", executorName, name)
	}

	executortypes.RegisterEvChan(executortypes.Name(name), getEvChanFn, syncRuntimeToEvent)
}

func (m *Marathon) initEventAndPeriodSync(name string, lstore *sync.Map, stopCh chan struct{}) error {
	start := time.Now()
	defer func() {
		logrus.Infof("in initEventAndPeriodSync executor(%s) took %v", name, time.Since(start))
	}()
	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"
	notifier, err := events.New(name+events.SUFFIX_INIT, map[string]interface{}{"HTTP": []string{eventAddr}})
	if err != nil {
		return errors.Errorf("new eventbox api error when executor(%s) init and send batch events", name)
	}

	isInit := true
	// First initialization and periodic update
	initRuntimeEventStore := func(k string, v interface{}) error {
		r := v.(*apistructs.ServiceGroup)
		if r.Executor != string(name) {
			return nil
		}
		r2_, err := m.Inspect(context.Background(), *r)
		if err != nil {
			logrus.Errorf("executor(%s)'s key(%s) get status error: %v", name, k, err)
			return nil
		}
		r2 := r2_.(*apistructs.ServiceGroup)
		reKey := etcdKeyToMapKey(k)
		if len(reKey) == 0 {
			return nil
		}

		var e events.RuntimeEvent
		e.RuntimeName = reKey
		e.ServiceStatuses = make([]events.ServiceStatus, len(r2.Services))
		for i := range r2.Services {
			e.ServiceStatuses[i].ServiceName = r2.Services[i].Name
			e.ServiceStatuses[i].Replica = r2.Services[i].Scale
			e.ServiceStatuses[i].ServiceStatus = convertServiceStatus(r2.Services[i].Status)
			e.ServiceStatuses[i].InstanceStatuses = make([]events.InstanceStatus, len(r2.Services[i].InstanceInfos))

			// If the number of instances of the service is set to 0 and there is no instance information in the service, set it to a specific health state
			if e.ServiceStatuses[i].Replica == 0 && len(r2.Services[i].InstanceInfos) == 0 {
				e.ServiceStatuses[i].ServiceStatus = string(apistructs.StatusHealthy)
			}

			if r2.Services[i].Scale != len(r2.Services[i].InstanceInfos) {
				logrus.Errorf("service(%s) scale(%v) not matched its instances number(%v)",
					r2.Services[i].Name, r2.Services[i].Scale, len(r2.Services[i].InstanceInfos))
			}

			for j, instance := range r2.Services[i].InstanceInfos {
				// instance.Id format:  runtimes_v1_services_staging-821_web.ca3113d1-9531-11e8-ad54-70b3d5800001
				// runtimes_v1_services_staging-821_web.ca3113d1-9531-11e8-ad54-70b3d5800001.1
				e.ServiceStatuses[i].InstanceStatuses[j].ID = instance.Id
				e.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = convertStatus(&instance)
				e.ServiceStatuses[i].InstanceStatuses[j].Ip = instance.Ip
				e.ServiceStatuses[i].InstanceStatuses[j].Extra = make(map[string]interface{})
			}

			if r2.Services[i].NewHealthCheck != nil {
				if r2.Services[i].NewHealthCheck.ExecHealthCheck != nil {
					e.ServiceStatuses[i].HealthCheckDuration = r2.Services[i].NewHealthCheck.ExecHealthCheck.Duration
				} else if r2.Services[i].NewHealthCheck.HttpHealthCheck != nil {
					e.ServiceStatuses[i].HealthCheckDuration = r2.Services[i].NewHealthCheck.HttpHealthCheck.Duration
				}
			}
		}

		e.EventType = events.EVENTS_TOTAL

		if isInit {
			go notifier.Send(e)
		} else {
			go notifier.Send(e, events.WithSender(string(name)+events.SUFFIX_PERIOD))
		}

		lstore.Store(reKey, e)
		return nil
	}

	// When each executor is initialized, go to jsonstore to initialize a metadata function:
	// 1, Get the status information of all runtimes belonging to the executor, as the structural basis for subsequent calculation of incremental events
	// 2, Send initialization status event (full event)
	em := events.GetEventManager()

	if err = em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initRuntimeEventStore); err != nil {
		logrus.Errorf("executor(%s) foreach initRuntimeEventStore error: %v", name, err)
	}

	// Periodic compensation status
	go func() {
		isInit = false
		for {
			select {
			case <-stopCh:
				logrus.Errorf("executor(%s) got stop chan message for sending total events", name)
				return

			case <-time.After(10 * time.Minute):
				logrus.Infof("executor(%s) periodically sync status form Marathon Status interface", name)

				if err = em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initRuntimeEventStore); err != nil {
					logrus.Errorf("executor(%s) foreach initRuntimeEventStore error: %v", name, err)
				}
			}
		}
	}()

	return nil
}

// The Status interface gets the native state of the instance through marathon, and converts it to the state of the corresponding instance in RuntimeEvent
func convertStatus(instance *apistructs.InstanceInfo) string {
	switch instance.Status {
	case events.RUNNING:
		if instance.Alive == "false" {
			return string(apistructs.StatusUnHealthy)
		}
		// Including the case where the instance is Healthy and the instance is not configured with health check
		// All current services will be configured with health checks, only those very old services may not be configured
		return string(apistructs.StatusHealthy)

	case events.KILLED:
		return "Killed"

	case events.FINISHED:
		return "Finished"

	case events.FAILED:
		return "Failed"

	case events.KILLING:
		return "Killing"

	case events.STARTING:
		return "Starting"

	case events.STAGING:
		return "Staging"

	default:
		if len(instance.Status) > 0 {
			return instance.Status
		}
		return "Unknown"
	}
}

// Modification of the status of the service obtained through the Status interface through marathon
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

func collectDeadInstances(ev *MarathonStatusUpdateEvent, ch chan string) {
	if ev.TaskStatus == events.FAILED || ev.TaskStatus == events.KILLED || ev.TaskStatus == events.FINISHED {
		select {
		case ch <- ev.AppId:
		case <-time.After(200 * time.Millisecond):
			logrus.Errorf("timeout sending dead instance's ev(%s) to judge", ev.AppId)
		}
	}
}
