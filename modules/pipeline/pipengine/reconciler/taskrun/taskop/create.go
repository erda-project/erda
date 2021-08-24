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

package taskop

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
)

type create taskrun.TaskRun

func NewCreate(tr *taskrun.TaskRun) *create {
	return (*create)(tr)
}

func (c *create) Op() taskrun.Op {
	return taskrun.Create
}

func (c *create) TaskRun() *taskrun.TaskRun {
	return (*taskrun.TaskRun)(c)
}

func (c *create) Processing() (interface{}, error) {
	_, err := c.Executor.Create(c.Ctx, c.Task)
	return nil, err
}

func (c *create) WhenDone(data interface{}) error {
	c.Task.Status = apistructs.PipelineStatusCreated
	logrus.Infof("reconciler: pipelineID: %d, task %q end create (%s -> %s)",
		c.P.ID, c.Task.Name, apistructs.PipelineStatusBorn, apistructs.PipelineStatusCreated)
	return nil
}

func (c *create) WhenLogicError(err error) error {
	c.Task.Status = apistructs.PipelineStatusCreateError
	return nil
}

func (c *create) WhenTimeout() error {
	return nil
}

func (c *create) TimeoutConfig() (<-chan struct{}, context.CancelFunc, time.Duration) {
	return nil, nil, -1
}

func (c *create) TuneTriggers() taskrun.TaskOpTuneTriggers {
	return taskrun.TaskOpTuneTriggers{
		BeforeProcessing: aoptypes.TuneTriggerTaskBeforeCreate,
		AfterProcessing:  aoptypes.TuneTriggerTaskAfterCreate,
	}
}
