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

package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sflink"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sjob"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sspark"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/clusterinfo"
)

// Manager for scheduler task executor
type Manager struct {
	sync.RWMutex

	factory   map[types.Kind]types.CreateFn
	executors map[types.Name]types.TaskExecutor
	clusters  map[string]apistructs.ClusterInfo
}

var mgr Manager

func GetManager() *Manager {
	return &mgr
}

func GetExecutorInfo() interface{} {
	executors := map[string]interface{}{}
	clusters := map[string]apistructs.ClusterInfo{}
	mgr.RLock()
	defer mgr.RUnlock()
	for name, exe := range mgr.executors {
		executors[name.String()] = map[string]string{
			"kind": exe.Kind().String(),
		}
	}
	for name, cluster := range mgr.clusters {
		clusters[name] = cluster
	}
	return []interface{}{executors, clusters}
}

func (m *Manager) Initialize() error {
	m.factory = types.Factory
	m.executors = make(map[types.Name]types.TaskExecutor)
	m.clusters = make(map[string]apistructs.ClusterInfo)

	logrus.Infof("pipeline scheduler task executor Inititalize ...")

	if err := m.batchUpdateExecutors(); err != nil {
		return err
	}

	triggerChan := clusterinfo.RegisterRefreshChan()
	eventChan, err := clusterinfo.RegisterClusterEvent()

	if err != nil {
		return err
	}
	go m.listenAndPatchExecutor(context.Background(), eventChan, triggerChan)

	logrus.Info("pipeline task executor manager Initialize Done .")

	return nil
}

func (m *Manager) Get(name types.Name) (types.TaskExecutor, error) {
	m.RLock()
	defer m.RUnlock()
	if len(name) == 0 {
		return nil, errors.Errorf("task executor name is empty")
	}
	e, ok := m.executors[name]
	if !ok {
		return nil, errors.Errorf("not found task executor [%s]", name)
	}
	return e, nil
}

func (m *Manager) GetCluster(clusterName string) (apistructs.ClusterInfo, error) {
	m.RLock()
	defer m.RUnlock()
	if len(clusterName) == 0 {
		return apistructs.ClusterInfo{}, errors.Errorf("clusterName is empty")
	}
	cluster, ok := m.clusters[clusterName]
	if !ok {
		return apistructs.ClusterInfo{}, errors.Errorf("failed to get cluster info by clusterName: %s", clusterName)
	}
	return cluster, nil
}

// TryGetExecutor if can`t get executor, manager try to make new executor by cluster info and return new executor
func (m *Manager) TryGetExecutor(name types.Name, cluster apistructs.ClusterInfo) (bool, types.TaskExecutor, error) {
	taskExecutor, err := m.Get(name)
	if err != nil {
		logrus.Warnf("failed to get executor: %s, err: %v, try to make executors...", name, err)
		if updateErr := m.updateClusterExecutor(cluster); updateErr != nil {
			return false, nil, updateErr
		}
		newExecutor, nErr := m.Get(name)
		if nErr != nil {
			logrus.Errorf("try to get executor failed, err: %v", nErr)
			return false, nil, nErr
		}
		return false, newExecutor, nil
	}
	return false, taskExecutor, nil
}

func (m *Manager) deleteExecutor(cluster apistructs.ClusterInfo) {
	m.Lock()
	defer m.Unlock()

	switch cluster.Type {
	case apistructs.K8S, apistructs.EDAS:
		name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
		if _, exist := m.executors[name]; exist {
			delete(m.executors, name)
			logrus.Infof("task executor kind [%s], name [%s], deleted", k8sjob.Kind, name)
		}

		name = types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sflink.Kind))
		if _, exist := m.executors[name]; exist {
			delete(m.executors, name)
			logrus.Infof("task executor kind [%s], name [%s], deleted", k8sjob.Kind, name)
		}

		name = types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sspark.Kind))
		if _, exist := m.executors[name]; exist {
			delete(m.executors, name)
			logrus.Infof("task executor kind [%s], name [%s], deleted", k8sjob.Kind, name)
		}
	default:

	}
	delete(m.clusters, cluster.Name)
}

func (m *Manager) updateClusterExecutor(cluster apistructs.ClusterInfo) error {
	var err error
	var k8sjobExecutor types.TaskExecutor
	var k8sflinkExecutor types.TaskExecutor
	var k8ssparkExecutor types.TaskExecutor
	m.Lock()
	defer func() {
		m.Unlock()
	}()

	m.clusters[cluster.Name] = cluster

	switch cluster.Type {
	case apistructs.K8S, apistructs.EDAS:
		k8sjobCreate, ok := m.factory[k8sjob.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
			k8sjobExecutor, err = k8sjobCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sjob.Kind, name, err)
				return err
			} else {
				if _, exist := m.executors[name]; exist {
					delete(m.executors, name)
				}
				m.executors[name] = k8sjobExecutor
				logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
			}
		}

		k8sflinkCreate, ok := m.factory[k8sflink.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sflink.Kind))
			k8sflinkExecutor, err = k8sflinkCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sflink.Kind, name, err)
				return err
			} else {
				if _, exist := m.executors[name]; exist {
					delete(m.executors, name)
				}
				m.executors[name] = k8sflinkExecutor
				logrus.Infof("=> kind [%s], name [%s], created", k8sflink.Kind, name)
			}
		}

		k8ssparkCreate, ok := m.factory[k8sspark.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sspark.Kind))
			k8ssparkExecutor, err = k8ssparkCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sspark.Kind, name, err)
				return err
			} else {
				if _, exist := m.executors[name]; exist {
					delete(m.executors, name)
				}
				m.executors[name] = k8ssparkExecutor
				logrus.Infof("=> kind [%s], name [%s], created", k8sspark.Kind, name)
			}
		}
	default:

	}
	return nil
}

func (m *Manager) batchUpdateExecutors() error {
	clusters, err := clusterinfo.ListAllClusters()
	if err != nil {
		return err
	}

	for i := range clusters {
		if clusters[i].Type != apistructs.K8S && clusters[i].Type != apistructs.EDAS {
			continue
		}
		if err := m.updateClusterExecutor(clusters[i]); err != nil {
			continue
		}
	}
	return nil
}

func (m *Manager) listenAndPatchExecutor(ctx context.Context, eventChan <-chan apistructs.ClusterEvent, triggerChan <-chan struct{}) {
	var err error
	interval := time.Duration(conf.ExecutorRefreshIntervalMinute())
	ticker := time.NewTicker(time.Minute * interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-triggerChan:
			if err := m.batchUpdateExecutors(); err != nil {
				logrus.Errorf("failed to refresh executors, err: %v", err)
			}
		case <-ticker.C:
			if err := m.batchUpdateExecutors(); err != nil {
				logrus.Errorf("failed to refresh executors, err: %v", err)
			}
		case event := <-eventChan:
			switch event.Action {
			case apistructs.ClusterActionCreate:
				err = m.updateClusterExecutor(event.Content)
				if err != nil {
					logrus.Errorf("failed to add task executor, err: %v", err)
				}
			case apistructs.ClusterActionUpdate:
				err = m.updateClusterExecutor(event.Content)
				if err != nil {
					logrus.Errorf("failed to update task executor, err: %v", err)
				}
			case apistructs.ClusterActionDelete:
				m.deleteExecutor(event.Content)
			default:

			}
		}
	}
}
