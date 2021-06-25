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

	"github.com/gogap/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sjob"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/clusterinfo"
)

const (
	CLUSTERTYPEK8S = "k8s"
)

// Manager for scheduler task executor
type Manager struct {
	sync.RWMutex

	factory   map[types.Kind]types.CreateFn
	executors map[types.Name]types.TaskExecutor
}

var mgr Manager

func GetManager() *Manager {
	return &mgr
}

func (m *Manager) Initialize(cfgs []apistructs.ClusterInfo) error {
	m.factory = types.Factory
	m.executors = make(map[types.Name]types.TaskExecutor)

	logrus.Infof("pipeline scheduler task executor Inititalize ...")

	for i := range cfgs {
		err := m.addExecutor(cfgs[i])
		if err != nil {
			return err
		}
	}
	eventChan, err := clusterinfo.RegisterClusterEvent()
	if err != nil {
		return err
	}
	go m.listenClusterEventSync(context.Background(), eventChan)
	m.Lock()
	defer m.Unlock()
	for i := range cfgs {
		switch cfgs[i].Type {
		case CLUSTERTYPEK8S:
			k8sjobCreate, ok := m.factory[k8sjob.Kind]
			if ok {
				name := types.Name(fmt.Sprintf("%sfor%s", cfgs[i].Name, k8sjob.Kind))
				k8sjobExecutor, err := k8sjobCreate(name, cfgs[i].Name, cfgs[i])
				if err != nil {
					logrus.Infof("=> kind [%s], name [%s], created failed, err: %v", k8sjob.Kind, name, err)
					return err
				}
				m.executors[name] = k8sjobExecutor
				logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
			}
		default:

		}
		// TODO sync load cluster info and change executor map
	}
	logrus.Info("pipengine task executor manager Initialize Done .")

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

func (m *Manager) addExecutor(cluster apistructs.ClusterInfo) error {
	m.Lock()
	defer m.Unlock()

	switch cluster.Type {
	case CLUSTERTYPEK8S:
		k8sjobCreate, ok := m.factory[k8sjob.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
			if _, exist := m.executors[name]; exist {
				return errors.Errorf("task executor name: %s already existed", name)
			}
			k8sjobExecutor, err := k8sjobCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Infof("=> kind [%s], name [%s], created failed, err: %v", k8sjob.Kind, name, err)
				return err
			}
			m.executors[name] = k8sjobExecutor
			logrus.Infof("=> kind [%s], name [%s], created", k8sjob.Kind, name)
		}
	default:

	}
	return nil
}

func (m *Manager) deleteExecutor(cluster apistructs.ClusterInfo) {
	m.Lock()
	defer m.Unlock()

	switch cluster.Type {
	case CLUSTERTYPEK8S:
		name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
		if _, exist := m.executors[name]; exist {
			delete(m.executors, name)
		}
	default:

	}
}

func (m *Manager) updateExecutor(cluster apistructs.ClusterInfo) error {
	m.Lock()
	defer m.Unlock()

	switch cluster.Type {
	case CLUSTERTYPEK8S:
		k8sjobCreate, ok := m.factory[k8sjob.Kind]
		if ok {
			name := types.Name(fmt.Sprintf("%sfor%s", cluster.Name, k8sjob.Kind))
			if _, exist := m.executors[name]; exist {
				delete(m.executors, name)
			}
			k8sjobExecutor, err := k8sjobCreate(name, cluster.Name, cluster)
			if err != nil {
				logrus.Infof("=> kind [%s], name [%s], created failed, err: %v", k8sjob.Kind, name, err)
				return err
			}
			m.executors[name] = k8sjobExecutor
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
				err = m.addExecutor(event.Content)
				if err != nil {
					logrus.Errorf("failed to add task executor, err: %v", err)
				}
			case apistructs.ClusterActionUpdate:
				err = m.updateExecutor(event.Content)
				if err != nil {
					logrus.Errorf("failed to update task executor, err: %v", err)
				}
			default:

			}
		}
	}
}
