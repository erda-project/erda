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
	"fmt"
	"sync"

	"github.com/gogap/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sjob"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
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

	m.Lock()
	defer m.Unlock()
	for i := range cfgs {
		switch cfgs[i].Type {
		case CLUSTERTYPEK8S:
			k8sjobCreate, ok := m.factory[k8sjob.Kind]
			if ok {
				name := types.Name(fmt.Sprintf("%sfor%s", cfgs[i].Name, k8sjob.Kind))
				k8sjobExecutor, err := k8sjobCreate(name, cfgs[i].Name, nil)
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
