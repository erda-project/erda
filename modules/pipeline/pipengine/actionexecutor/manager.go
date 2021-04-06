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

package actionexecutor

import (
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// Manager is an executor manager, it holds the all executor instances.
type Manager struct {
	factory   map[types.Kind]types.CreateFn
	executors map[types.Name]types.ActionExecutor
}

var mgr Manager

func GetManager() *Manager {
	return &mgr
}

func (m *Manager) Initialize(cfgs chan spec.ActionExecutorConfig) error {
	m.factory = types.Factory
	m.executors = make(map[types.Name]types.ActionExecutor)

	logrus.Info("pipengine action executor manager Initialize ...")

	for len(cfgs) > 0 {
		c := <-cfgs

		create, ok := m.factory[types.Kind(c.Kind)]
		if !ok {
			return errors.Errorf("unregistered action executor kind: %v", c.Kind)
		}

		name := types.Name(c.Name)
		for k, v := range c.Options {
			logrus.Infof("=> kind [%s], name [%s], option: %s=%s", c.Kind, c.Name, k, v)
		}

		actionExecutor, err := create(name, c.Options)
		if err != nil {
			logrus.Infof("=> kind [%s], name [%s], created failed, err: %v", c.Kind, c.Name, err)
			return err
		}

		m.executors[name] = actionExecutor
		logrus.Infof("=> kind [%s], name [%s], created", c.Kind, c.Name)
	}

	logrus.Info("pipengine action executor manager Initialize Done .")

	return nil
}

// Get returns the executor with name.
func (m *Manager) Get(name types.Name) (types.ActionExecutor, error) {
	if len(name) == 0 {
		return nil, errors.Errorf("executor name is empty")
	}
	e, ok := m.executors[name]
	if !ok {
		return nil, errors.Errorf("not found action executor [%s]", name)
	}
	return e, nil
}

// GetByKind returns the executor instances with specify kind.
func (m *Manager) GetByKind(kind types.Kind) []types.ActionExecutor {
	executors := make([]types.ActionExecutor, 0, len(m.executors))
	for _, e := range m.executors {
		if e.Kind() == kind {
			executors = append(executors, e)
		}
	}
	return executors
}

// ListExecutors returns the all executor instances.
func (m *Manager) ListExecutors() []types.ActionExecutor {
	executors := make([]types.ActionExecutor, 0, len(m.executors))
	for _, e := range m.executors {
		executors = append(executors, e)
	}
	return executors
}

// GetActionExecutorKindByName return executor Kind, e.g. MEMORY
func (m *Manager) GetActionExecutorKindByName(name string) (string, error) {
	e, err := m.Get(types.Name(name))
	if err != nil {
		return "", errors.Wrapf(err, "failed to get action executor by name [%s]", name)
	}
	return string(e.Kind()), nil
}
