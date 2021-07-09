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

package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
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

func (m *Manager) Initialize(cfgs []apistructs.ClusterInfo) error {
	m.factory = types.Factory
	m.executors = make(map[types.Name]types.TaskExecutor)
	m.clusters = make(map[string]apistructs.ClusterInfo)

	logrus.Infof("pipeline scheduler task executor Inititalize ...")

	for i := range cfgs {
		if cfgs[i].Type != apistructs.K8S {
			continue
		}
		if err := m.updateClusterExecutor(cfgs[i]); err != nil {
			continue
		}
	}

	eventChan, err := clusterinfo.RegisterClusterEvent()
	if err != nil {
		return err
	}
	go m.listenClusterEventSync(context.Background(), eventChan)

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

func (m *Manager) deleteExecutor(cluster apistructs.ClusterInfo) {
	m.Lock()
	defer m.Unlock()

	switch cluster.Type {
	case apistructs.K8S:
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
		if err == nil {
			m.clusters[cluster.Name] = cluster
		}
		m.Unlock()
	}()

	switch cluster.Type {
	case apistructs.K8S:
		k8sjobCreate, ok := m.factory[k8sjob.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
			if _, exist := m.executors[name]; exist {
				delete(m.executors, name)
			}
			k8sjobExecutor, err = k8sjobCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sjob.Kind, name, err)
				return err
			}
			m.executors[name] = k8sjobExecutor
			logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
		}

		k8sflinkCreate, ok := m.factory[k8sflink.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sflink.Kind))
			if _, exist := m.executors[name]; exist {
				delete(m.executors, name)
			}
			k8sflinkExecutor, err = k8sflinkCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sflink.Kind, name, err)
				return err
			}
			m.executors[name] = k8sflinkExecutor
			logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
		}

		k8ssparkCreate, ok := m.factory[k8sspark.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sspark.Kind))
			if _, exist := m.executors[name]; exist {
				delete(m.executors, name)
			}
			k8ssparkExecutor, err = k8ssparkCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Errorf("=> kind [%s], name [%s], created failed, err: %v", k8sspark.Kind, name, err)
				return err
			}
			m.executors[name] = k8ssparkExecutor
			logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
		}
	default:

	}
	return nil
}

func (m *Manager) listenClusterEventSync(ctx context.Context, eventChan <-chan apistructs.ClusterEvent) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
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
