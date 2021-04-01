package executortypes

import (
	"context"
	"regexp"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const kindNameFormat = `^[A-Z0-9]+$`

var formater *regexp.Regexp = regexp.MustCompile(kindNameFormat)

// Name represents a executor's name.
type Name string

func (s Name) String() string {
	return string(s)
}

func (s Name) Validate() bool {
	return formater.MatchString(string(s))
}

// Kind represents a executor's type.
type Kind string

func (s Kind) String() string {
	return string(s)
}

func (s Kind) Validate() bool {
	return formater.MatchString(string(s))
}

// Create be used to create a executor instance.
type CreateFn func(name Name, clustername string, options map[string]string, moreoptions interface{}) (Executor, error)

// return executor's event channel according to executor's name
type GetEventChanFn func(Name) (chan *eventtypes.StatusEvent, chan struct{}, *sync.Map, error)

type EventCbFn func(k string, v interface{}, t storetypes.ChangeType) error

type NodeLabelSetting struct {
	SoldierURL string
}

// Executor defines the all interfaces that must be implemented by a executor instance.
type Executor interface {
	Kind() Kind
	Name() Name

	CleanUpBeforeDelete()

	Create(ctx context.Context, spec interface{}) (interface{}, error)
	Destroy(ctx context.Context, spec interface{}) error
	Status(ctx context.Context, spec interface{}) (apistructs.StatusDesc, error)
	Remove(ctx context.Context, spec interface{}) error
	Update(ctx context.Context, spec interface{}) (interface{}, error)
	Inspect(ctx context.Context, spec interface{}) (interface{}, error)
	Cancel(ctx context.Context, spec interface{}) (interface{}, error)
	Precheck(ctx context.Context, spec interface{}) (apistructs.ServiceGroupPrecheckData, error)

	// only k8s-job executor supported
	JobVolumeCreate(ctx context.Context, spec interface{}) (string, error)

	// SetNodeLabels set schedule-labels on nodes
	// Only k8s, k8sjob, marathon, metronome executor implement this function
	SetNodeLabels(setting NodeLabelSetting, hosts []string, labels map[string]string) error

	// executor's capacity
	// 1. addonoperator
	CapacityInfo() apistructs.CapacityInfoData

	ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error)

	// only k8s executor supported
	KillPod(podname string) error
}

type TerminalExecutor interface {
	Terminal(namespace, podname, containername string, conn *websocket.Conn)
}

type ExecutorWholeConfigs struct {
	// 普通集群配置
	BasicConfig map[string]string
	// 精细化的配置
	PlusConfigs *conf.OptPlus
}

type StopEventsChans struct {
	StopWatchEventCh  chan struct{}
	StopHandleEventCh chan struct{}
}

var Factory = map[Kind]CreateFn{}
var EvFuncMap = map[Name]GetEventChanFn{}
var EvCbMap = map[Name]EventCbFn{}

// Register add a executor's create function.
func Register(kind Kind, create CreateFn) error {
	if !kind.Validate() {
		return errors.Errorf("invalid kind: %s", kind)
	}
	if _, ok := Factory[kind]; ok {
		return errors.Errorf("duplicate to register executor: %s", kind)
	}
	Factory[kind] = create
	return nil
}

// Get a GetEventChanFn according to an executor's name
func RegisterEvChan(name Name, get GetEventChanFn, cb EventCbFn) error {
	logrus.Debugf("in RegisterEvChan going to register executor: %s", name)
	if _, ok := EvFuncMap[name]; ok {
		return errors.Errorf("duplicate to register executor's event channel: %s", name)
	}
	EvFuncMap[name] = get
	EvCbMap[name] = cb
	return nil
}

func UnRegisterEvChan(name Name) {
	logrus.Debugf("in UnRegisterEvChan going to unregister executor: %s", name)
	delete(EvFuncMap, name)
	delete(EvCbMap, name)
}
