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

package k8s

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events"
	eventboxapi "github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

type Event struct {
	Type   watch.EventType `json:"type"`
	Object apiv1.Event     `json:"object"`
}

func (k *Kubernetes) sendEvent(localStore *sync.Map, stopCh chan struct{}, notifier eventboxapi.Notifier) {
	time.Sleep(5 * time.Second)

	logrus.Infof("executor in k8s sendEvent, name: %s", k.name)

	// 处理增量事件
	eventHandler := func() (bool, error) {

		urlPath := "/api/v1/watch/events"

		body, resp, err := k.client.Get(k.addr).
			Path(urlPath).
			Header("Portal-SSE", "on").
			Param("fieldSelector", "involvedObject.kind=Pod,"+
				"involvedObject.namespace!=default,"+
				"involvedObject.namespace!=kube-system").
			Do().
			StreamBody()

		if err != nil {
			logrus.Errorf("failed to get resp from k8s event, name: %s, (%v)", k.name, err)
			return false, err
		}

		if !resp.IsOK() {
			errMsg := fmt.Sprintf("failed to get resp from k8s event, name: %s, resp is not OK", k.name)
			logrus.Errorf(errMsg)
			return false, errors.New(errMsg)
		}

		logrus.Infof("k8s event from executor: %s, req.URL: %s", k.name, urlPath)

		defer body.Close()
		reader := bufio.NewReader(body)
		var event Event
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				logrus.Errorf("failed to read data from k8s event, name: %s, (%v)", k.name, err)
				break
			}

			if err := json.Unmarshal(line, &event); err != nil {
				logrus.Errorf("failed to unmarshal k8s event data, bs: %v, (%v)", string(line), err)
				break
			}

			status := ConvertEventStatus(event.Object.Reason)
			// todo: 去掉 addon 的事件
			if status == "" || strings.Contains(event.Object.Namespace, "addon-") {
				continue
			}

			paths := strings.Split(event.Object.Namespace, "--")
			if len(paths) != 2 {
				//logrus.Errorf("failed to parse k8s event namespace: %s", event.Object.Namespace)
				continue
			}

			// 实例事件
			runtimeName := strutil.Concat(paths[0], "/", paths[1])
			go k.InstanceEvent(event, runtimeName, notifier)

			// 对于中间件多个 statefulset 的情况, 多个 statefulset 共用一个 namespace, 在传递给调度器的 namespace 上
			// 加上了前缀 group-, 因此对于这种情况, 需要去掉前缀 group-, 才能在 etcd 中找到相应的记录
			paths[0] = strings.TrimPrefix(paths[0], "group-")

			key := strutil.Concat("/dice/service/", paths[0], "/", paths[1])
			var sg apistructs.ServiceGroup
			if err := events.GetEventManager().MemEtcdStore.Get(context.Background(), key, &sg); err != nil {
				logrus.Errorf("failed to get k8s servicegroup from etcd, key: %s, (%v)", key, err)
				continue
			}
			if _, err := k.Status(context.Background(), sg); err != nil {
				logrus.Errorf("failed to get k8s servicegroup status in event, namespace: %s, name: %s",
					paths[0], paths[1])
				continue
			}

		}
		return false, nil
	}

	if err := loop.New(loop.WithDeclineRatio(4), loop.WithDeclineLimit(time.Second*60)).Do(eventHandler); err != nil {
		return
	}
}

func (k *Kubernetes) InstanceEvent(event Event, runtimeName string, notifier eventboxapi.Notifier) {
	name := event.Object.Name
	pieces := strings.Split(name, "-")
	var serviceName string
	if len(pieces) <= 2 {
		logrus.Errorf("failed to parse name from event, name: %s", name)
		return
	}

	for i := 0; i < len(pieces)-2; i++ {
		serviceName = strutil.Concat(serviceName, "-", pieces[i])
	}

	var ie apistructs.InstanceStatusData
	if status := ConvertEventStatus(event.Object.Reason); status != "" {
		ie = apistructs.InstanceStatusData{
			ClusterName:    k.options["cluster"],
			RuntimeName:    runtimeName,
			ServiceName:    serviceName[1:],
			InstanceStatus: status,
			Host:           event.Object.Source.Host,
			Message:        event.Object.Message,
			Timestamp:      time.Now().UnixNano(),
		}
	}

	// wait for pod status updated
	time.Sleep(1 * time.Second)

	pod, err := k.pod.Get(event.Object.Namespace, event.Object.InvolvedObject.Name)
	if err != nil {
		logrus.Errorf("failed to get pod status, namespace: %s, name: %s",
			event.Object.Namespace, event.Object.InvolvedObject.Name)
		return
	}
	ie.IP = pod.Status.PodIP
	if len(pod.Status.ContainerStatuses) == 0 {
		logrus.Infof("[alert] empty containerStatuses list: %v", pod.Status)
		// 不存在 containerStatuses, 不发事件
		return
	}
	// "containerID": "docker://c894809fa10635e455be2bfec5c151a23ac9d27ec6cfc5948444ff01b6836819"
	// 去掉 prefix "docker://"
	ie.ID = strutil.TrimPrefixes(pod.Status.ContainerStatuses[0].ContainerID, "docker://")

	// 兼容 edas v2
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == "DICE_SERVICE_NAME" {
			ie.ServiceName = env.Value
			break
		}
	}

	if err := notifier.Send(ie, eventboxapi.WithDest(map[string]interface{}{"WEBHOOK": apistructs.EventHeader{
		Event:     "instances-status",
		Action:    "changed",
		OrgID:     "-1",
		ProjectID: "-1",
	}})); err != nil {
		logrus.Errorf("failed to send instances-status event, executor: %s", k.name)
	}
}

// 定期发送全量事件
func (k *Kubernetes) totalEvent(localStore *sync.Map, notifier eventboxapi.Notifier, eventAddr string) {
	initStore := func(key string, v interface{}) error {
		reKey := etcdKeyToMapKey(key)
		if len(reKey) == 0 {
			return nil
		}

		sg := v.(*apistructs.ServiceGroup)
		if sg.Executor != string(k.name) {
			return nil
		}
		if _, err := k.Status(context.Background(), *sg); err != nil {
			logrus.Errorf("failed to init k8s event in totalEvent, name: %s", sg.ID)
			return nil
		}
		e := GenerateEvent(sg)
		localStore.Store(reKey, e)
		logrus.Infof("k8s executor in initStore , key: %v, event: %v", key, e)
		return nil
	}

	em := events.GetEventManager()
	if err := em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initStore); err != nil {
		logrus.Errorf("executor(%s) foreach initStore error: %v", k.name, err)
	}

	isInitEvent := true
	f := func(key, val interface{}) bool {
		logrus.Infof("in totalEvent f, key: %v, value: %v", key, val)

		_, ok := val.(events.RuntimeEvent)
		if !ok {
			logrus.Errorf("failed to parse val to runtime event in totalEvent, key: %v, value: %v", key, val)
			return true
		}

		var err error
		c := make(chan struct{}, 1)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		paths := strings.Split(key.(string), "/")
		if len(paths) != 2 {
			logrus.Errorf("failed to parse key to two parts in k8s totalEvent, key: %v", key)
			return true
		}
		etcdKey := strutil.Concat("/dice/service/", paths[0], "/", paths[1])
		var sg apistructs.ServiceGroup
		if err := events.GetEventManager().MemEtcdStore.Get(context.Background(), etcdKey, &sg); err != nil {
			logrus.Errorf("failed to get servicegroup from etcd in totalEvent, key: %s, (%v)", key, err)
			return true
		}

		go func() {
			_, err = k.Status(ctx, sg)
			c <- struct{}{}
		}()

		select {
		case <-c:
			if err != nil {
				logrus.Errorf("failed to get executor(%s)'s key(%s) status, (%v)", k.name, key, err)
				return true
			}

			e := GenerateEvent(&sg)

			go func() {
				sender := strutil.Concat(string(k.name), events.SUFFIX_K8S_PERIOD)
				if isInitEvent {
					sender = strutil.Concat(string(k.name), events.SUFFIX_K8S_INIT)
				}
				err := notifier.Send(e, eventboxapi.WithSender(sender),
					eventboxapi.WithDest(map[string]interface{}{"HTTP": []string{eventAddr}}))
				if err != nil {
					logrus.Errorf("failed to send k8s period event, executor: %s, runtime: %s", k.name, key)
				}
			}()

		case <-ctx.Done():
			logrus.Errorf("failed to get executor(%s)'s key(%s), get status timeout", k.name, key)
		}
		return true
	}

	localStore.Range(f)
	isInitEvent = false
	for range time.Tick(5 * time.Minute) {
		localStore.Range(f)
	}
}

func GenerateEvent(sg *apistructs.ServiceGroup) events.RuntimeEvent {
	var e events.RuntimeEvent
	e.EventType = events.EVENTS_TOTAL
	e.RuntimeName = strutil.Concat(sg.Type, "/", sg.ID)
	e.ServiceStatuses = make([]events.ServiceStatus, len(sg.Services))
	for i, srv := range sg.Services {
		e.ServiceStatuses[i].ServiceName = srv.Name
		e.ServiceStatuses[i].Replica = srv.Scale
		e.ServiceStatuses[i].ServiceStatus = string(srv.Status)
		// 状态为空字符串
		if len(e.ServiceStatuses[i].ServiceStatus) == 0 {
			e.ServiceStatuses[i].ServiceStatus = convertServiceStatus(apistructs.StatusProgressing)
		}
		if e.ServiceStatuses[i].Replica == 0 {
			e.ServiceStatuses[i].ServiceStatus = convertServiceStatus(apistructs.StatusHealthy)
		}
	}
	return e
}

// 用于增量服务事件中判断缓存服务状态是否变更
func isStatusCached(localStore *sync.Map, name, status string) bool {
	if v, ok := localStore.Load(name); ok && v.(string) == status {
		return true
	}

	localStore.Store(name, status)
	return false
}

func (k *Kubernetes) registerEvent(localStore *sync.Map, stopCh chan struct{}, notifier eventboxapi.Notifier) error {

	name := string(k.name)

	logrus.Infof("in k8s registerEvent, executor: %s", name)

	// watch 特定etcd目录的事件的处理函数
	syncRuntimeToLstore := func(key string, value interface{}, t storetypes.ChangeType) error {

		runtimeName := etcdKeyToMapKey(key)
		if len(runtimeName) == 0 {
			return nil
		}

		// 先处理delete的事件
		if t == storetypes.Del {
			_, ok := localStore.Load(runtimeName)
			if ok {
				var e events.RuntimeEvent
				e.RuntimeName = runtimeName
				e.IsDeleted = true
				localStore.Delete(runtimeName)

			}
			return nil
		}

		sg, ok := value.(*apistructs.ServiceGroup)
		if !ok {
			logrus.Errorf("failed to parse value to servicegroup, key: %v, value: %v", key, value)
			return nil
		}

		// 过滤不属于本executor的事件
		if sg.Executor != name {
			return nil
		}

		switch t {
		// Add or update event
		case storetypes.Add, storetypes.Update:
			if _, err := k.Status(context.Background(), *sg); err != nil {
				logrus.Errorf("failed to get k8s status in event, name: %s", sg.ID)
				return nil
			}
			e := GenerateEvent(sg)
			localStore.Store(runtimeName, e)

		default:
			logrus.Errorf("failed to get watch type, type: %s, name: %s", t, runtimeName)
			return nil
		}

		logrus.Infof("succeed to stored key: %s, executor: %s", key, name)
		return nil
	}

	// 将注册来的executor的name及其event channel对应起来
	getEventFn := func(executorName executortypes.Name) (chan *eventtypes.StatusEvent, chan struct{}, *sync.Map, error) {
		logrus.Infof("in RegisterEvChan executor(%s)", name)
		if string(executorName) == name {
			return k.evCh, stopCh, localStore, nil
		}
		return nil, nil, nil, errors.Errorf("this is for %s executor, not %s", executorName, name)
	}

	return executortypes.RegisterEvChan(executortypes.Name(name), getEventFn, syncRuntimeToLstore)
}

// ConvertEventStatus convert k8s status
func ConvertEventStatus(reason string) string {
	switch reason {
	case "Started", "Healthy":
		return "Healthy"
	case "Killing":
		return "Killed"
	case "Unhealthy", "UnHealthy":
		return "UnHealthy"
	}

	return ""
}

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

// todo: refactor this function
// e.g. /dice/service/services/staging-99 -> services/staging-99
func etcdKeyToMapKey(eKey string) string {
	fields := strings.Split(eKey, "/")
	if l := len(fields); l > 2 {
		return fields[l-2] + "/" + fields[l-1]
	}
	return ""
}
