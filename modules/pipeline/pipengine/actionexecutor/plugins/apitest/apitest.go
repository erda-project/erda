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

package apitest

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/apitest/logic"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindAPITest)

type define struct {
	name     types.Name
	options  map[string]string
	dbClient *dbclient.Client
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
		return true, true, nil
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
	logic.Do(ctx, task)
	return nil, nil
}

func (d *define) Update(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, nil
}

func (d *define) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.PipelineStatusDesc, error) {
	latestTask, err := d.dbClient.GetPipelineTask(task.ID)
	if err != nil {
		return apistructs.PipelineStatusDesc{}, fmt.Errorf("failed to query latest task, err: %v", err)
	}
	*task = latestTask

	if task.Status.IsEndStatus() {
		return apistructs.PipelineStatusDesc{Status: task.Status}, nil
	}

	created, _, err := d.Exist(ctx, task)
	if err != nil {
		return apistructs.PipelineStatusDesc{}, err
	}

	if !created {
		return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusAnalyzed}, nil
	}

	// status according to api success or not
	meta := latestTask.Result.Metadata
	for _, metaField := range meta {
		if metaField.Name == logic.MetaKeyResult {
			if metaField.Value == logic.ResultSuccess {
				return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusSuccess}, nil
			}
			return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusFailed}, nil
		}
	}

	// return created status to do start step
	return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusCreated}, nil
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
		return &define{
			name:     name,
			options:  options,
			dbClient: dbClient,
		}, nil
	})
}
