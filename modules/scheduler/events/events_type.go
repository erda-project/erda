package events

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/pkg/jsonstore"
)

const (
	// status_update_event 定义的状态
	KILLING     = "TASK_KILLING"
	KILLED      = "TASK_KILLED"
	RUNNING     = "TASK_RUNNING"
	FINISHED    = "TASK_FINISHED"
	FAILED      = "TASK_FAILED"
	STARTING    = "TASK_STARTING"
	STAGING     = "TASK_STAGING"
	DROPPED     = "TASK_DROPPED"
	UNKNOWN     = "TASK_UNKNOWN"
	UNREACHABLE = "TASK_UNREACHABLE"

	// instance 状态
	INSTANCE_RUNNING  = "Running"
	INSTANCE_FAILED   = "Failed"
	INSTANCE_FINISHED = "Finished"
	INSTANCE_KILLED   = "Killed"
	// instance_health_changed_event 封装后的状态
	// instance 和 service 共用
	HEALTHY   = "Healthy"
	UNHEALTHY = "UnHealthy"

	WATCHED_DIR = "/dice/service/"

	// 调用eventbox时发送方名字的后缀, 用于区分发事件的阶段
	// 初始化executor阶段发送的事件
	SUFFIX_INIT = "_INIT"
	// 周期性补偿阶段发的事件
	SUFFIX_PERIOD = "_PERIOD"
	// 其他时期的普通时间到的事件
	SUFFIX_NORMAL = "_NORMAL"
	// 暂时为edas事件指定的前缀
	SUFFIX_EDAS = "_EDAS"
	// EDASV2 周期性补偿阶段发的事件
	SUFFIX_EDASV2_PERIOD = "_EDASV2_PERIOD"
	// EDASV2 增量事件
	SUFFIX_EDASV2_NORMAL = "_EDASV2_NORMAL"
	// EDASV2 初始事件
	SUFFIX_EDASV2_INIT = "_EDASV2_INIT"
	// K8S 周期性补偿阶段发的事件
	SUFFIX_K8S_PERIOD = "_K8S_PERIOD"
	// K8S 增量事件
	SUFFIX_K8S_NORMAL = "_K8S_NORMAL"
	// K8S 初始事件
	SUFFIX_K8S_INIT = "_K8S_INIT"

	// 事件类型
	// 计算得出的事件，对应SUFFIX_NORMAL
	EVENTS_INCR = "increment"
	// 初始化或者周期性补偿的事件，对应SUFFIX_INIT和SUFFIX_PERIOD
	EVENTS_TOTAL = "total"

	// instance 后缀, 作为key一部分存储于lstore, 标识实例开始健康检查的时间
	START_HC_TIME_SUFFIX = "_start_hc_time"

	// 判断健康检查超时的左窗口边缘, 包含 interval(15s) 的时间
	LEFT_EDGE int64 = -20
	// 判断健康检查超时的右窗口边缘, 考虑 sigterm 干不掉容器, 等待一段时间(20s~30s)后收到 sigkill 退出
	RIGHT_DEGE int64 = 40

	// marathon 实例状态变化事件类型
	STATUS_UPDATE_EVENT = "status_update_event"
	// marathon 实例健康变化类型
	INSTANCE_HEALTH_CHANGED_EVENT = "instance_health_changed_event"

	// 调度器实例变化事件，根据上述两类事件处理得出
	INSTANCE_STATUS = "instances-status"
)

type LabelKey string

type Op struct {
	labels map[string]interface{}
	dest   map[string]interface{}
	sender string
}
type OpOperation func(*Op)

type Message struct {
	Sender  string                   `json:"sender"`
	Content interface{}              `json:"content"`
	Labels  map[LabelKey]interface{} `json:"labels"`
	Time    int64                    `json:"time,omitempty"` // UnixNano

	originContent interface{} `json:"-"`
}

type NotifierImpl struct {
	sender string
	labels map[string]interface{} // optional, can also assign `dest' as `Send' option
	dir    string
	js     jsonstore.JsonStore
}

type GetExecutorJsonStoreFn func(string) (*sync.Map, error)
type GetExecutorLocalStoreFn func(string) (*sync.Map, error)
type SetCallbackFn func(string) error

type Notifier interface {
	Send(content interface{}, options ...OpOperation) error
	SendRaw(message *Message) error
}

type EventMgr struct {
	ctx context.Context
	// 记录所有executor的callback函数的map
	executorCbMap sync.Map
	// watch WATCHED_DIR 路径，同步到 executor 的本地缓存中
	// 1, 用于计算增量状态事件
	// 2, 用于初始化和定期补偿事件
	MemEtcdStore jsonstore.JsonStore

	notifier Notifier
}

// 塞到eventbox里的结构体
type RuntimeEvent struct {
	RuntimeName     string          `json:"runtimeName"`
	ServiceStatuses []ServiceStatus `json:"serviceStatuses,omitempty"`
	// 临时字段，标识对应的runtime是否被删除
	IsDeleted bool   `json:"isDeleted,omitempty"`
	EventType string `json:"eventType,omitempty"`
}

type ServiceStatus struct {
	ServiceName      string           `json:"serviceName"`
	ServiceStatus    string           `json:"serviceStatus,omitempty"`
	Replica          int              `json:"replica,omitempty"`
	InstanceStatuses []InstanceStatus `json:"instanceStatuses,omitempty"`

	HealthCheckDuration int `json:"healthCheckDuration,omitempty"`
}

type InstanceStatus struct {
	ID             string `json:"id,omitempty"`
	Ip             string `json:"ip,omitempty"`
	InstanceStatus string `json:"instanceStatus,omitempty"`
	// stage记录容器退出的阶段, 只体现在增量事件中的退出(Killed, Failed, Finished)事件
	// 当前分的阶段为：
	// a) 容器启动阶段（健康检查超时之前退出）,"BeforeHealthCheckTimeout"
	// b) 健康检查超时阶段（被健康检查所杀）,"HealthCheckTimeout"
	// c) 后健康检查阶段（健康检查完成后退出）,"AfterHealthCheckTimeout"
	Stage string `json:"stage,omitempty"`
	// 扩展字段
	Extra map[string]interface{} `json:"extra,omitempty"`
}

type EventLayer struct {
	InstanceId  string
	ServiceName string
	RuntimeName string
}

type WindowStatus struct {
	key        string
	statusType string
}

type InstancesWindow struct {
	key string
	e   *eventtypes.StatusEvent
}

func convert(m map[string]interface{}) map[LabelKey]interface{} {
	m_ := map[LabelKey]interface{}{}
	for k, v := range m {
		m_[LabelKey(k)] = v
	}
	return m_
}

func genMessagePath(dir string, timestamp int64) string {
	return filepath.Join(dir, strconv.FormatInt(timestamp, 10))
}

func genMessage(sender string, content interface{}, timestamp int64, labels map[string]interface{}) (*Message, error) {
	return &Message{
		Sender:  sender,
		Content: content,
		Labels:  convert(labels),
		Time:    timestamp,
	}, nil
}

func mergeMap(main map[string]interface{}, minor map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range minor {
		m[k] = v
	}

	for k, v := range main {
		m[k] = v
	}
	return m
}

// 将 message 写入 etcd
func (n *NotifierImpl) Send(content interface{}, options ...OpOperation) error {
	option := &Op{}
	for _, op := range options {
		op(option)
	}
	timestamp := time.Now().UnixNano()
	labels := n.labels
	sender := n.sender
	if option.dest != nil {
		labels = mergeMap(option.dest, labels)
	}
	if option.labels != nil {
		labels = mergeMap(option.labels, labels)
	}
	if option.sender != "" {
		sender = option.sender
	}

	message, err := genMessage(sender, content, timestamp, labels)
	if err != nil {
		return err
	}
	return n.SendRaw(message)
}

func (n *NotifierImpl) SendRaw(message *Message) error {
	messagePath := genMessagePath(n.dir, message.Time)
	ctx := context.Background()
	return n.js.Put(ctx, messagePath, message)
}
