package executor

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/goroutinepool"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

const defaultGoPool = "defaultpool"

var (
	mgr      Manager
	initonce sync.Once
)

func GetManager() *Manager {
	err := mgr.initialize()
	if err != nil {
		// init failed, just panic here
		panic(err)
	}
	return &mgr
}

// Manager is a executor manager, it holds the all executor instances.
type Manager struct {
	factory   map[executortypes.Kind]executortypes.CreateFn
	executors map[executortypes.Name]executortypes.Executor
	pools     map[executortypes.Name]*goroutinepool.GoroutinePool
	evChanMap map[executortypes.Name]executortypes.GetEventChanFn
	evCbMap   map[executortypes.Name]executortypes.EventCbFn

	executorConfigs map[executortypes.Name]*executortypes.ExecutorWholeConfigs
	executorStopCh  map[executortypes.Name]executortypes.StopEventsChans
}

func (m *Manager) initialize() error {
	var r error
	initonce.Do(func() {
		m.pools = make(map[executortypes.Name]*goroutinepool.GoroutinePool)
		m.pools[defaultGoPool] = goroutinepool.New(conf.PoolSize())
		m.pools[defaultGoPool].Start()

		m.executors = make(map[executortypes.Name]executortypes.Executor)
		m.factory = executortypes.Factory
		m.evChanMap = executortypes.EvFuncMap
		m.evCbMap = executortypes.EvCbMap
		m.executorConfigs = make(map[executortypes.Name]*executortypes.ExecutorWholeConfigs)
		m.executorStopCh = make(map[executortypes.Name]executortypes.StopEventsChans)

		f := func(k string, v interface{}, t storetypes.ChangeType) {
			config, ok := v.(*conf.ExecutorConfig)
			if !ok {
				logrus.Errorf("not executorConfig type, key: %s, value: %+v", k, v)
				return
			}
			switch t {
			case storetypes.Del:
				logrus.Infof("watched executor(%s) deleted, key: %s", config.Name, k)
				deleteOneExecutor(m, config)
			case storetypes.Update:
				logrus.Infof("watched executor(%s) updated, key: %s", config.Name, k)
				conf.GetConfStore().ExecutorStore.Store(config.Name, config)
				if err := updateOneExecutor(m, config); err != nil {
					logrus.Errorf("updating executor(name: %s, key: %s) error: %v", config.Name, k, err)
				}
			case storetypes.Add:
				logrus.Infof("watched executor(%s) created, key: %s", config.Name, k)
				conf.GetConfStore().ExecutorStore.Store(config.Name, config)
				if err := createOneExecutor(m, config); err != nil {
					logrus.Errorf("creating executor(name: %s, key: %s) error: %v", config.Name, k, err)
				}
			}
		}
		option := jsonstore.UseMemEtcdStore(context.Background(), conf.CLUSTERS_CONFIG_PATH, f, conf.ExecutorConfig{})
		if _, err := jsonstore.New(option); err != nil {
			r = err
		}
		logrus.Infof("executor factories:%+v", m.factory)
		logrus.Infof("executors:%+v", m.executors)
	})
	return r
}

func createOneExecutor(m *Manager, config *conf.ExecutorConfig) error {
	logrus.Infof("[createOneExecutor] config: %+v", config)
	create, ok := m.factory[executortypes.Kind(config.Kind)]
	if !ok {
		return errors.Errorf("executor kind (%s) not found, factory: %v", config.Kind, m.factory)
	}

	name := executortypes.Name(config.Name)

	executor, err := create(name, config.ClusterName, config.Options, config.OptionsPlus)
	if err != nil {
		return err
	}
	m.executors[name] = executor
	m.pools[name] = goroutinepool.New(conf.PoolSize())
	m.pools[name].Start()

	// TODO: dont create manually
	wholeConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}
	m.executorConfigs[name] = wholeConfigs
	conf.GetConfStore().ExecutorStore.Store(config.Name, config)

	logrus.Infof("created executor: %s", name)

	if getEvChanFn, ok := m.evChanMap[name]; ok {
		eventCh, stopEventWatchCh, lstore, err := getEvChanFn(name)
		if err != nil {
			logrus.Errorf("getEvChanFn for executor(%s) err: %v", name, err)
		}

		logrus.Infof("executor(%s) found event channel, lstore addr: %p", name, lstore)
		cb, ok := m.evCbMap[name]
		if !ok {
			logrus.Errorf("executor(%s) cannot find event cb", name)
		}
		stopEventHandleCh := make(chan struct{}, 1)
		m.executorStopCh[name] = executortypes.StopEventsChans{StopWatchEventCh: stopEventWatchCh, StopHandleEventCh: stopEventHandleCh}
		go events.HandleOneExecutorEvent(string(name), eventCh, lstore, cb, stopEventHandleCh)
	}
	return nil
}

func deleteOneExecutor(m *Manager, config *conf.ExecutorConfig) {
	name := executortypes.Name(config.Name)
	if chs, ok := m.executorStopCh[name]; ok {
		close(chs.StopWatchEventCh)
		close(chs.StopHandleEventCh)
	}

	if p, ok := m.pools[name]; ok {
		p.Stop()
	}

	executortypes.UnRegisterEvChan(executortypes.Name(config.Name))
	events.GetEventManager().UnRegisterEventCallback(config.Name)

	if executor, ok := m.executors[executortypes.Name(config.Name)]; ok {
		executor.CleanUpBeforeDelete()
	}

	// 删除 executor 对应的创建函数
	delete(m.executors, executortypes.Name(config.Name))
	// 删除 executor 的配置
	delete(m.executorConfigs, executortypes.Name(config.Name))
	//
	//delete(m.executorStopCh, executortypes.Name(config.Name))
	// 删除 conf 中对应 executor name 与 cluster name 关系的 map
	conf.GetConfStore().ExecutorStore.Delete(config.Name)
}

func updateOneExecutor(m *Manager, config *conf.ExecutorConfig) error {
	deleteOneExecutor(m, config)
	logrus.Infof("updating executor: %s", config.Name)
	return createOneExecutor(m, config)
}

func (m *Manager) Pool(name executortypes.Name) *goroutinepool.GoroutinePool {
	p, ok := m.pools[name]
	if !ok {
		return m.pools[defaultGoPool]
	}
	return p
}

// Get returns the executor with name.
func (m *Manager) Get(name executortypes.Name) (executortypes.Executor, error) {
	c, ok := m.executors[name]
	if !ok {
		return nil, errors.Errorf("not found executor: %s", name)
	}
	return c, nil
}

// GetByKind returns the executor instances with specify kind.
func (m *Manager) GetByKind(kind executortypes.Kind) []executortypes.Executor {
	executors := make([]executortypes.Executor, 0, len(m.executors))
	for _, c := range m.executors {
		if c.Kind() == kind {
			executors = append(executors, c)
		}
	}
	return executors
}

// ListExecutors returns the all executor instances.
func (m *Manager) ListExecutors() []executortypes.Executor {
	executors := make([]executortypes.Executor, 0, len(m.executors))
	for _, c := range m.executors {
		executors = append(executors, c)
	}
	return executors
}

func (m *Manager) PrintPoolUsage() {
	for exc, pool := range m.pools {
		stat := pool.Statistics()
		total := stat[1]
		inuse := total - stat[0]
		logrus.Infof("[%s] pool: total worker num: %d, inuse worker num: %d", exc, total, inuse)
	}
}

// GetJobExecutorKindByName return executor Kind, e.g. METRONOME, according to user-defined job specification.
func GetJobExecutorKindByName(name string) string {
	e, err := GetManager().Get(executortypes.Name(name))
	if err != nil {
		logrus.Errorf("[alert] failed to get executor by name %s", name)
		return ""
	}
	return string(e.Kind())
}

func (m *Manager) GetExecutorConfigs(name executortypes.Name) (*executortypes.ExecutorWholeConfigs, error) {
	config, ok := m.executorConfigs[name]
	if !ok {
		return nil, errors.Errorf("[GetExecutorConfigs] not found executor: %s", name)
	}
	if config == nil {
		return nil, errors.Errorf("get executor(%s) configs null", name)
	}
	return config, nil
}
