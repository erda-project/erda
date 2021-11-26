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

package mysql_assert

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/mysql_assert/execute"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindMysqlConfigSheet)

type define struct {
	name        types.Name
	options     map[string]string
	dbClient    *dbclient.Client
	bdl         *bundle.Bundle
}

func (d *define) Kind() types.Kind { return Kind }
func (d *define) Name() types.Name { return d.name }

func (d *define) Exist(ctx context.Context, task *spec.PipelineTask) (created bool, started bool, err error) {
	status := task.Status
	switch true {
	case status == apistructs.PipelineStatusAnalyzed, status == apistructs.PipelineStatusBorn:
		return false, false, nil
	case status == apistructs.PipelineStatusCreated:
		return true, false, nil
	case status == apistructs.PipelineStatusQueue, status == apistructs.PipelineStatusRunning:
		return true, false, nil
	case status.IsEndStatus():
		return true, true, nil
	default:
		return false, false, fmt.Errorf("invalid status when query task exist")
	}
}

func (d *define) Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) Start(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) Update(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.PipelineStatusDesc, error) {
	created, _, err := d.Exist(ctx, task)
	if err != nil {
		return apistructs.PipelineStatusDesc{}, err
	}

	if !created {
		return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusAnalyzed}, nil
	}

	if task.Status.IsEndStatus() {
		return apistructs.PipelineStatusDesc{Status: task.Status}, nil
	}

	if task.Status == apistructs.RunnerTaskStatusRunning {
		var status = apistructs.PipelineStatusFailed
		execute.Do(ctx, task)

		latestTask, err := d.dbClient.GetPipelineTask(task.ID)
		if err != nil {
			logrus.Errorf("failed to query latest task, err: %v \n", err)
			return apistructs.PipelineStatusDesc{Status: status}, nil
		}
		meta := latestTask.GetMetadata()
		for _, metaField := range meta {
			if metaField.Name == execute.MetaAssertResultKey {
				if metaField.Value == execute.ResultSuccess {
					status = apistructs.PipelineStatusSuccess
				}
			}
		}
		return apistructs.PipelineStatusDesc{Status: status}, nil
	}

	return apistructs.PipelineStatusDesc{Status: apistructs.RunnerTaskStatusRunning}, nil
}

func (d *define) Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error) {
	return apistructs.TaskInspect{}, nil
}

func (d *define) Cancel(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) BatchDelete(ctx context.Context, actions []*spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func init() {
	types.MustRegister(Kind, func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		dbClient, err := dbclient.New()
		if err != nil {
			return nil, fmt.Errorf("failed to init dbclient, err: %v", err)
		}
		bdl := bundle.New(bundle.WithAllAvailableClients())
		return &define{
			name:        name,
			options:     options,
			dbClient:    dbClient,
			bdl: bdl,
		}, nil
	})
}
