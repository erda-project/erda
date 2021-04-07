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

package events

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	EventboxDir = "/eventbox"
	MessageDir  = filepath.Join(EventboxDir, "messages")
)
var eventMgr EventMgr

// GetEventManager returns global instance
func GetEventManager() *EventMgr {
	return &eventMgr
}

// RegisterEventCallback 将插件中事件处理函数注册到事件控制面
func (m *EventMgr) RegisterEventCallback(name string, cb executortypes.EventCbFn) error {
	if _, ok := m.executorCbMap.Load(name); ok {
		return errors.Errorf("duplicate executor registering for executor(%s)", name)
	}

	m.executorCbMap.Store(name, cb)
	logrus.Debugf("register executor(%s) event cb", name)
	return nil
}

// UnRegisterEventCallback 在事件控制面中清理某插件的事件处理函数
func (m *EventMgr) UnRegisterEventCallback(name string) {
	if _, ok := m.executorCbMap.Load(name); !ok {
		return
	}
	m.executorCbMap.Delete(name)
	logrus.Debugf("unregister executor(%s) event cb", name)
}

var js jsonstore.JsonStore

func New(sender string, dest map[string]interface{}) (Notifier, error) {
	var err error
	if js == nil {
		if js, err = jsonstore.New(); err != nil {
			return nil, err
		}
	}
	return &NotifierImpl{
		sender: sender,
		labels: dest,
		dir:    MessageDir,
		js:     js,
	}, nil
}

func WithDest(dest map[string]interface{}) OpOperation {
	return func(op *Op) {
		op.dest = dest
	}
}

func WithSender(sender string) OpOperation {
	return func(op *Op) {
		op.sender = sender
	}
}

func OnlyOne(ctx context.Context, lock *dlock.DLock) (func(), error) {
	var isCanceled bool
	var locked bool

	cleanup := func() {
		if isCanceled {
			return
		}
		logrus.Infof("Onlyone: etcdlock: unlock, key: %s", lock.Key())
		if err := lock.Unlock(); err != nil {
			logrus.Errorf("Onlyone: etcdlock unlock err: %v", err)
			return
		}

	}

	go func() {
		time.Sleep(3 * time.Second)
		if !locked && !isCanceled {
			logrus.Warnf("Onlyone: not get lock yet after 3s")
		}
	}()
	if err := lock.Lock(ctx); err != nil {
		if err == context.Canceled {
			isCanceled = true
			logrus.Infof("Onlyone: etcdlock: %v", err)
			return cleanup, nil
		}
		return cleanup, err
	}
	locked = true
	logrus.Infof("Onlyone: etcdlock: lock, key: %s", lock.Key())

	return cleanup, nil
}
func init() {
	// TODO: need set timeout
	eventMgr.ctx = context.Background()
	eventMgr.executorCbMap = sync.Map{}

	allExecutorCallback := func(key string, value interface{}, t storetypes.ChangeType) error {
		f := func(k, v interface{}) bool {
			v.(executortypes.EventCbFn)(key, value, t)
			return true
		}
		eventMgr.executorCbMap.Range(f)
		return nil
	}

	store, err := jsonstore.New()
	if err != nil {
		panic(fmt.Sprintf("new jsonstore: %v", err))
	}
	watchstore := store.IncludeWatch()
	if watchstore == nil {
		panic(fmt.Sprintf("failed to new jsonstore with watch"))
	}
	go func() {
		for {
			select {
			case <-time.After(time.Duration(5) * time.Second):
			}
			if err := watchstore.Watch(eventMgr.ctx, WATCHED_DIR, true, false, false, apistructs.ServiceGroup{}, allExecutorCallback); err != nil {
				logrus.Errorf("watchstore watch err: %v", err)
			}
		}
	}()
	eventMgr.MemEtcdStore = watchstore

	notifier, err := New("", nil)
	if err != nil {
		panic(fmt.Sprintf("new eventbox api error: %v", err))
	}
	eventMgr.notifier = notifier
}

func getLayerInfoFromEvent(id, eventType string) (*EventLayer, error) {
	err := errors.Errorf("event's taskId(%s) format error", id)
	// status_update_event的id(taskId)格式: runtimes_v1_services_staging-773_web.e622bf15-9300-11e8-ad54-70b3d5800001
	// instance_health_changed_event的id(instanceId)格式: runtimes_v1_services_staging-790_web.marathon-0ad6d3ce-946c-11e8-ad54-70b3d5800001
	strs := strings.Split(id, ".")
	if len(strs) < 2 {
		return nil, err
	}
	var e EventLayer
	fields := strings.Split(strs[0], "_")
	if l := len(fields); l > 3 {
		if eventType == INSTANCE_HEALTH_CHANGED_EVENT {
			e.InstanceId = strings.Replace(strs[1], "marathon-", "", 1)
		} else if eventType == STATUS_UPDATE_EVENT {
			e.InstanceId = strs[1]
		}
		e.ServiceName = fields[l-1]
		e.RuntimeName = fields[l-3] + "/" + fields[l-2]
		return &e, nil
	}
	return nil, err
}

// HandleOneExecutorEvent 处理单个插件 executor 的事件逻辑
func HandleOneExecutorEvent(name string, ch chan *eventtypes.StatusEvent, lstore *sync.Map, cb executortypes.EventCbFn, stopCh chan struct{}) {
	evm := GetEventManager()
	if err := evm.RegisterEventCallback(name, cb); err != nil {
		logrus.Errorf("register executor failed, err: %v", err)
		return
	}

	if strings.Contains(name, "EDAS") {
		return
	}
	o11 := discover.Orchestrator()
	eventAddr := "http://" + o11 + "/api/events/runtimes/actions/sync"
	defer func() {
		// TODO: clear lstore
		logrus.Infof("executor(%s) event handler exit", name)
	}()

	// cache window's keys
	var eventsInWindow []WindowStatus
	nowTime := time.Now()
	for {
		select {
		case <-stopCh:
			logrus.Errorf("executor(%s) got stop chan message for receiving events", name)
			return
		default:
		}
		// 时间窗口 + 统计状态
		if time.Since(nowTime) >= 10*time.Second {
			nowTime = time.Now()

			// firstly send event
			for _, k := range eventsInWindow {
				re, ok := lstore.Load(k.key)
				if !ok {
					logrus.Errorf("in statistics cannot get key(%s) from lstore, type: %v", k.key, k.statusType)
					continue
				}
				ev := re.(RuntimeEvent)
				computeServiceStatus(&ev)
				ev.EventType = EVENTS_INCR

				if err := evm.notifier.Send(ev, WithSender(name+SUFFIX_NORMAL), WithDest(map[string]interface{}{"HTTP": []string{eventAddr}})); err != nil {
					logrus.Errorf("eventbox send api err: %v", err)
					continue
				}
				// 清理该事件中记录的KILLED/FINISHED/FAILED的实例，避免重复发送
				for i := range ev.ServiceStatuses {
					for j := len(ev.ServiceStatuses[i].InstanceStatuses) - 1; j >= 0; j-- {
						is := ev.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus
						if is == "Killed" || is == "Finished" || is == "Failed" {
							ev.ServiceStatuses[i].InstanceStatuses = append(ev.ServiceStatuses[i].InstanceStatuses[:j],
								ev.ServiceStatuses[i].InstanceStatuses[j+1:]...)
						}
					}
				}
				lstore.Store(k, ev)
			}
			// clear the window
			if len(eventsInWindow) != 0 {
				eventsInWindow = nil
			}
		}

		select {
		case e := <-ch:
			if e.Cluster != name {
				logrus.Infof("event's executor(%s) not matched this executor(%s)", e.Cluster, name)
				break
			}
			// 过滤KILLING/STARTING事件
			if e.Status == KILLING || e.Status == STARTING || e.Status == STAGING || e.Status == DROPPED {
				break
			}
			// 通过 eventbox hook 发送实例变化事件
			if err := handleInstanceStatusChangedEvents(e, lstore); err != nil {
				logrus.Errorf("handle instance status with err: %v", err)
			}

			// 对 addons 的事件不递送到 orchestrator
			if strings.HasPrefix(e.TaskId, "runtimes_v1_addon") {
				break
			}

			var eKey string
			var err error
			if e.Type == INSTANCE_HEALTH_CHANGED_EVENT {
				if eKey, err = handleHealthStatusChanged(e, lstore); err != nil {
					logrus.Errorf("executor(%s) handle health status changed event err: %v", name, err)
					break
				}
			} else if e.Type == STATUS_UPDATE_EVENT {
				if eKey, err = handleStatusUpdateEvent(e, lstore); err != nil {
					logrus.Errorf("executor(%s) handle status update event err: %v", name, err)
					break
				}
			}

			// 记录该时间窗口内的事件，用于时间到期后发送某类状态事件
			// eventsInWindow 是以 runtime 为单位，发送整个 runtime 下的事件
			evExisted := false
			for i, ev := range eventsInWindow {
				if ev.key != eKey {
					continue
				}
				eventsInWindow[i].statusType = e.Status
				evExisted = true
				break

			}
			if !evExisted {
				eventsInWindow = append(eventsInWindow, WindowStatus{eKey, e.Status})
			}

		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func handleStatusUpdateEvent(e *eventtypes.StatusEvent, lstore *sync.Map) (string, error) {
	layer, err := getLayerInfoFromEvent(e.TaskId, e.Type)
	if err != nil {
		return "", errors.Errorf("got layer info from event error: %v", err)
	}
	eKey := layer.RuntimeName
	eInstanceID := layer.InstanceId
	srvName := layer.ServiceName

	re, ok := lstore.Load(eKey)
	if !ok {
		return "", errors.Errorf("key(%s) not found in lstore, event type: %v, status: %s, drop it as not found in lstore", eKey, e.Type, e.Status)
	}

	run := re.(RuntimeEvent)

	foundEvInSrv := false
	for i, srv := range run.ServiceStatuses {
		// 事件隶属于该service下的某个instance
		// logrus.Infof("srv.SericeName: %v, srvName: %v", srv.ServiceName, srvName)
		// 用户定义的service如果是大写，对应到marathon中去创建时候会被转化成小写
		if !strings.EqualFold(srv.ServiceName, srvName) {
			continue
		}
		foundEvInSrv = true
		foundInstance := false
		// 记录下如果是健康检查超时导致的将实例杀死的情况
		recordInstanceExited := func(insKey string, status string, instance *InstanceStatus) {
			// 容器已退出，缓存中的 key 需要清除
			defer lstore.Delete(insKey)
			// Failed 或者 Finished 必定是用户自己原因导致的容器退出, 默认标记在容器启动阶段
			if status == INSTANCE_FAILED || status == INSTANCE_FINISHED {
				instance.Stage = "BeforeHealthCheckTimeout"
				return
			}
			v, ok := lstore.Load(insKey)
			if !ok {
				return
			}
			// 处理容器被 Killed 的情况

			d := run.ServiceStatuses[i].HealthCheckDuration
			if d < apistructs.HealthCheckDuration {
				logrus.Infof("instance(%s) health check duration change from %v to %v", insKey, d, apistructs.HealthCheckDuration)
				d = apistructs.HealthCheckDuration
			}
			startHcTime := v.(int64)
			current := time.Now().Unix()
			expectedKilledTime := startHcTime + int64(d)

			// 判断这个实例是被健康检查超时所杀的, 被杀时间在 (开始健康时间 + duration) 的时间之后, 且没有超过太久
			// startHcTime 在该实例已健康的情况下被置为 0
			if startHcTime > 0 &&
				current-expectedKilledTime > LEFT_EDGE &&
				current-expectedKilledTime < RIGHT_DEGE {
				instance.Stage = "HealthCheckTimeout"
			} else if current-expectedKilledTime < LEFT_EDGE {
				instance.Stage = "BeforeHealthCheckTimeout"
			} else if current-expectedKilledTime > RIGHT_DEGE {
				instance.Stage = "AfterHealthCheckTimeout"
			}

		}
		for j := range srv.InstanceStatuses {
			instance := &(run.ServiceStatuses[i].InstanceStatuses[j])
			if instance.ID != e.TaskId {
				continue
			}
			foundInstance = true
			switch e.Status {
			case KILLED:
				// KILLED 有可能是健康检查超时被杀, 区分出这种情况出来
				instance.InstanceStatus = INSTANCE_KILLED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_KILLED, instance)

			case RUNNING:
				// 禁止实例从 healthy / unhealthy 退化到 running, 因为是否是 healthy 得由 healthy 事件来决定
				if instance.InstanceStatus != HEALTHY && instance.InstanceStatus != UNHEALTHY {
					instance.InstanceStatus = INSTANCE_RUNNING
				}

				// 如果健康检查设置了 delaySeconds, 会有两个 TASK_RUNNING 事件
				// 后一次 TASK_RUNNING 才开始做健康检查
				// 所以记录的是开始健康检查的时间
				lstore.Store(e.TaskId+START_HC_TIME_SUFFIX, time.Now().Unix())

			case FINISHED:
				// FINISHED 必定是业务方自己原因退出
				instance.InstanceStatus = INSTANCE_FINISHED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_FINISHED, instance)

			case FAILED:
				// FAILED 必定是业务自己原因退出
				instance.InstanceStatus = INSTANCE_FAILED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_FAILED, instance)

			default:
				instance.InstanceStatus = e.Status
				logrus.Debugf("service(%s) instance(%s) has default status: %s",
					srv.ServiceName, instance.ID, e.Status)
			}

			// marathon bug: marathon instance KILLED 事件中容器 ip 绝大多数情况下为宿主机 ip
			if e.Status != KILLED {
				instance.Ip = e.IP
			}
			break
		}

		// instanceId 已记录在当前service的status中
		if foundInstance {
			lstore.Store(eKey, run)
			break
		}

		// instanceId 未记录在当前service的status中
		hasVacancy := false
		for j := range run.ServiceStatuses[i].InstanceStatuses {
			instance := &(run.ServiceStatuses[i].InstanceStatuses[j])
			if instance.ID != "" {
				continue
			}
			hasVacancy = true
			instance.ID = e.TaskId
			instance.Ip = e.IP

			switch e.Status {
			case RUNNING:
				instance.InstanceStatus = INSTANCE_RUNNING
				lstore.Store(e.TaskId+START_HC_TIME_SUFFIX, time.Now().Unix())

			case FAILED:
				instance.InstanceStatus = INSTANCE_FAILED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_FAILED, instance)

			case FINISHED:
				instance.InstanceStatus = INSTANCE_FINISHED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_FINISHED, instance)

			case KILLED:
				// 正常情况下KILLING/KILLED的instance之前会被记录
				logrus.Errorf("event taskID(%s) not found in previous status but its status is %s", eInstanceID, e.Status)
				instance.InstanceStatus = INSTANCE_KILLED
				recordInstanceExited(e.TaskId+START_HC_TIME_SUFFIX, INSTANCE_KILLED, instance)

			default:
				//continue
				instance.InstanceStatus = e.Status
				logrus.Debugf("service(%s) instance(%s) has other status: %s",
					srv.ServiceName, instance.ID, e.Status)
			}

			break
		}

		// 未有空位表明从 etcd 元数据信息中获知某服务A有B个副本(scale), 当前记录的不同实例已超过了副本数
		// 经常在重启，滚动升级中可见
		if !hasVacancy {
			var status string
			ins := InstanceStatus{
				ID:    e.TaskId,
				Ip:    e.IP,
				Extra: make(map[string]interface{}),
			}

			switch e.Status {
			case RUNNING:
				status = INSTANCE_RUNNING
				lstore.Store(ins.ID+START_HC_TIME_SUFFIX, time.Now().Unix())
			case FAILED:
				status = INSTANCE_FAILED
				recordInstanceExited(ins.ID+START_HC_TIME_SUFFIX, INSTANCE_FAILED, &ins)
			case FINISHED:
				status = INSTANCE_FINISHED
				recordInstanceExited(ins.ID+START_HC_TIME_SUFFIX, INSTANCE_FINISHED, &ins)
			case KILLED:
				status = INSTANCE_KILLED
				recordInstanceExited(ins.ID+START_HC_TIME_SUFFIX, INSTANCE_KILLED, &ins)
			}
			ins.InstanceStatus = status
			run.ServiceStatuses[i].InstanceStatuses = append(run.ServiceStatuses[i].InstanceStatuses, ins)
		}

		lstore.Store(eKey, run)
		break
	}

	// 理论上该情况不会出现
	if !foundEvInSrv {
		return "", errors.Errorf("event taskId(%s) instance(%s) not found in runtime(%s)'s any service",
			e.TaskId, eInstanceID, run.RuntimeName)
	}
	return eKey, nil
}

func handleHealthStatusChanged(e *eventtypes.StatusEvent, lstore *sync.Map) (string, error) {
	layer, err := getLayerInfoFromEvent(e.TaskId, e.Type)
	if err != nil {
		return "", errors.Errorf("got layer info from event error: %v", err)
	}

	eKey := layer.RuntimeName
	srvName := layer.ServiceName
	re, ok := lstore.Load(eKey)
	if !ok {
		return "", errors.Errorf("key(%s) not found in lstore, event type: %v, status: %v, drop it as not found in lstore", eKey, e.Type, e.Status)
	}

	run := re.(RuntimeEvent)

	foundEvInSrv := false
	for i, srv := range run.ServiceStatuses {
		// 事件隶属于该service下的某个instance
		if !strings.EqualFold(srv.ServiceName, srvName) {
			continue
		}
		foundEvInSrv = true
		foundInstance := false
		for j, instance := range srv.InstanceStatuses {
			//if instance.ID == eInstanceID {
			if instance.ID != e.TaskId {
				continue
			}
			foundInstance = true
			switch e.Status {
			case HEALTHY:
				run.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = HEALTHY
				// 置为0表明该实例至少有通过过健康检查
				lstore.Store(instance.ID+START_HC_TIME_SUFFIX, int64(0))
			case UNHEALTHY:
				run.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = UNHEALTHY
			}
			break
		}

		// instanceId 未记录在当前service的status中,理论上不会发生
		if !foundInstance {
			return "", errors.Errorf("healthy instance(%s) not found in service(%s)", e.TaskId, srvName)
		}
		lstore.Store(eKey, run)
		break
	}

	// 理论上该情况不会出现
	if !foundEvInSrv {
		return "", errors.Errorf("event taskId(%s) not found in runtime(%s)'s any service",
			e.TaskId, run.RuntimeName)
	}
	return eKey, nil
}

func computeServiceStatus(e *RuntimeEvent) {
	for i, srv := range e.ServiceStatuses {
		runningReplica := 0
		healthyReplica := 0

		// 表明runtime被删除
		if e.IsDeleted {
			e.ServiceStatuses[i].ServiceStatus = "Deleted"
			continue
		}

		for j, instance := range srv.InstanceStatuses {
			switch instance.InstanceStatus {
			case INSTANCE_RUNNING:
				runningReplica++
				e.ServiceStatuses[i].InstanceStatuses[j].InstanceStatus = string(apistructs.StatusStarting)
			case UNHEALTHY:
				runningReplica++
			case HEALTHY:
				runningReplica++
				healthyReplica++
			}
		}
		// 为了便于上层展示的方便，约定对实例数为0并且无实例在跑的服务，也设置其状态为 Healthy
		if healthyReplica == srv.Replica && runningReplica == srv.Replica {
			//e.ServiceStatuses[i].ServiceStatus = string(spec.StatusReady)
			e.ServiceStatuses[i].ServiceStatus = HEALTHY
		} else {
			// spec.StatusProgressing || spec.StatusRunning -> spec.StatusUnHealthy
			e.ServiceStatuses[i].ServiceStatus = UNHEALTHY
		}
	}
}
