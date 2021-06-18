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

package service

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/task"
)

type Servicer interface {
	Scale(sg *apistructs.ServiceGroup)
}

type Service struct {
	sched *task.Sched
}

func (s *Service) Scale(sg *apistructs.ServiceGroup) {
	s.sched.Send(context.Background(),
		task.TaskRequest{
			ExecutorKind: getServiceExecutorKindByName(sg.Executor),
			ExecutorName: sg.Executor,
			Action:       task.TaskScale,
			ID:           sg.ID,
			Spec:         sg,
		})
}

func getServiceExecutorKindByName(name string) string {
	e, err := executor.GetManager().Get(executortypes.Name(name))
	if err != nil {
		return conf.DefaultRuntimeExecutor()
	}
	return string(e.Kind())
}
