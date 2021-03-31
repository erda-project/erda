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
	"github.com/erda-project/erda/pkg/customhttp"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/loop"
	_ "github.com/erda-project/erda/pkg/monitor"
)

// TODO: 需要考虑marathon集群重启、异常等情况重新建连
// TODO: 定时从marathon api上获取状态（聚合计算层实现）
func (m *Marathon) WaitEvent(options map[string]string, monitor bool, killedInstanceCh chan string, stopCh chan struct{}) {
	// 等待5到30秒开始监听事件
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
			// 终止循环执行
			return true, errors.Errorf("got stop message and exit")
		default:
		}

		// 不用m.client，避开其超时时间等制约条件
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
		// marathon事件一次来两行，格式为
		// event: status_update_event
		// data: {"slaveId":"2aa23eb3-fa5c-4248-a203-ba4ddb9fb562-S2",...}
		// 除开事件本身会一直有空行进来
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

			// 先判断出事件类型
			if len(line) > 7 && string(line[:5]) == "event" {
				eventType = string(line[7:])
				eventType = strings.TrimSuffix(eventType, "\r\n")
				continue
			}

			if len(line) <= 6 || string(line[:4]) != "data" {
				continue
			}

			// 解析出事件内容
			content := line[6:]

			// 丢弃非用户服务的事件
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

				// 或者状态为 unknown 事件及 unreachable 事件
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

				// 统计经常挂掉的实例
				if monitor {
					collectDeadInstances(&ev, killedInstanceCh)
				}

			case events.INSTANCE_HEALTH_CHANGED_EVENT:
				// instance_health_changed_event 一直是两个相同事件同时发送过来
				// https://jira.mesosphere.com/browse/MARATHON-8134
				// https://jira.mesosphere.com/browse/MARATHON-8494
				// 过滤掉 instance_health_changed_event 中的 "healthy":null 的 event
				if string(content) == lastContent || strings.Contains(string(content), "\"healthy\":null") {
					continue
				}
				lastContent = string(content)

				ev := MarathonInstanceHealthChangedEvent{}
				if err := json.Unmarshal(content, &ev); err != nil {
					logrus.Errorf("unmarshal instance_health_changed_event string(%s) error: %v", string(content), err)
					continue
				}

				// instance_health_changed_event 的 instanceId 格式：
				// "instanceId":"runtimes_v1_services_mydev-590_user-service.marathon-56153805-9004-11e8-aaac-70b3d5800001"

				// status_update_event 的 taskId格式
				// "taskId":"runtimes_v1_services_mydev-590_user-service.56153805-9004-11e8-aaac-70b3d5800001"
				// instance_health_changed_event 的 instanceId的格式转化一下，即去掉"marathon-"
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

// 对etcd里的原始key做个处理
// /dice/service/services/staging-77 -> "services/staging-77"
func etcdKeyToMapKey(eKey string) string {
	fields := strings.Split(eKey, "/")
	if l := len(fields); l > 2 {
		// 过滤 addon 的服务
		if strings.HasPrefix(fields[l-2], "addon") {
			return ""
		}
		return fields[l-2] + "/" + fields[l-1]
	}
	return ""
}

func registerEventChanAndLocalStore(name string, evCh chan *eventtypes.StatusEvent, stopCh chan struct{}, lstore *sync.Map) {
	// watch 特定etcd目录的事件的处理函数
	syncRuntimeToEvent := func(key string, value interface{}, t storetypes.ChangeType) error {

		runtimeName := etcdKeyToMapKey(key)
		if len(runtimeName) == 0 {
			return nil
		}

		// 先处理delete的事件
		if t == storetypes.Del {
			// TODO: 先不删除key，删除了的话事件来了找不到对应的结构体，不好发事件
			// TODO: 后面可以置个标志位，然后开个协程定时清理
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

		// 过滤不属于本executor的事件
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
				// 当前实例信息不可知
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
			// 判断service个数是否改变，如考虑新增N个，删除M个情形
			for _, newS := range run.Services {
				found := false
				for _, oldS := range oldEvent.(events.RuntimeEvent).ServiceStatuses {
					if oldS.ServiceName == newS.Name {
						found = true
						event.ServiceStatuses = append(event.ServiceStatuses, oldS)
						// 判断是否是任何一个service里的instance个数发生变化, 只有扩容、缩容两种情况
						// 扩容
						if oldS.Replica < newS.Scale {
							i := 0
							for i < newS.Scale-oldS.Replica {
								i++
								event.ServiceStatuses[len(event.ServiceStatuses)-1].InstanceStatuses = append(
									event.ServiceStatuses[len(event.ServiceStatuses)-1].InstanceStatuses, events.InstanceStatus{})
							}
							event.ServiceStatuses[len(event.ServiceStatuses)-1].Replica = newS.Scale
						} else if len(oldS.InstanceStatuses) > newS.Scale { // 缩容
							// 根据marathon的策略"killSelection": "YOUNGEST_FIRST", 新的实例会先被杀
							// 而比较新的实例会被添加在slice的后面
							event.ServiceStatuses[len(event.ServiceStatuses)-1].Replica = newS.Scale
						}
						break
					}
				}

				// 未找到说明newS是需要新增的service, 不需考虑instance个数变化
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

	// 将注册来的executor的name及其event channel对应起来
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
	// 第一次初始化和周期性更新
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

			// 对服务的实例数被设置为 0 并且服务中无实例信息的，设置其为特定的健康状态
			if e.ServiceStatuses[i].Replica == 0 && len(r2.Services[i].InstanceInfos) == 0 {
				e.ServiceStatuses[i].ServiceStatus = string(apistructs.StatusHealthy)
			}

			if r2.Services[i].Scale != len(r2.Services[i].InstanceInfos) {
				logrus.Errorf("service(%s) scale(%v) not matched its instances number(%v)",
					r2.Services[i].Name, r2.Services[i].Scale, len(r2.Services[i].InstanceInfos))
			}

			for j, instance := range r2.Services[i].InstanceInfos {
				// instance.Id格式 runtimes_v1_services_staging-821_web.ca3113d1-9531-11e8-ad54-70b3d5800001
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

	// 每个executor被初始化的时候去jsonstore里初始化一把元数据作用：
	// 1, 拿到隶属于该executor的所有runtime的状态信息, 作为后续计算增量事件的结构基础
	// 2, 发送初始化状态事件(全量事件)
	em := events.GetEventManager()

	if err = em.MemEtcdStore.ForEach(context.Background(), "/dice/service/", apistructs.ServiceGroup{}, initRuntimeEventStore); err != nil {
		logrus.Errorf("executor(%s) foreach initRuntimeEventStore error: %v", name, err)
	}

	// 定期补偿状态
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

// Status接口通过marathon拿到实例的原生状态，转成RuntimeEvent中对应实例的状态
func convertStatus(instance *apistructs.InstanceInfo) string {
	switch instance.Status {
	case events.RUNNING:
		if instance.Alive == "false" {
			return string(apistructs.StatusUnHealthy)
		}
		// 包含实例为 Healthy 和 实例没有配置健康检查的情况
		// 当前所有服务都会配置健康检查，只有那些非常老的服务可能没有配置
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

// Status接口通过marathon拿到的服务的状态的修改
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
