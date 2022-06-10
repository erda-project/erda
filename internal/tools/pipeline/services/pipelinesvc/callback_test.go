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

package pipelinesvc

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/events"
	spec2 "github.com/erda-project/erda/internal/tools/pipeline/spec"

	"github.com/erda-project/erda/apistructs"
)

func TestAppendPipelineTaskResult(t *testing.T) {
	cb := apistructs.ActionCallback{
		Errors: []apistructs.ErrorResponse{
			apistructs.ErrorResponse{Msg: "a"},
			apistructs.ErrorResponse{Msg: "b"},
			apistructs.ErrorResponse{Msg: "a"},
			apistructs.ErrorResponse{Msg: "a"},
		},
	}

	task := &spec2.PipelineTask{
		Inspect: apistructs.PipelineTaskInspect{
			Errors: []*apistructs.PipelineTaskErrResponse{
				&apistructs.PipelineTaskErrResponse{Msg: "a"},
			},
		},
	}

	newTaskErrors := make([]*apistructs.PipelineTaskErrResponse, 0)
	for _, e := range cb.Errors {
		newTaskErrors = append(newTaskErrors, &apistructs.PipelineTaskErrResponse{
			Msg: e.Msg,
		})
	}
	task.Inspect.Errors = task.Inspect.AppendError(newTaskErrors...)

	assert.Equal(t, 3, len(task.Inspect.Errors))
}

func TestDealPipelineCallbackOfAction(t *testing.T) {
	db := &dbclient.Client{}

	m1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTask", func(_ *dbclient.Client, id interface{}) (spec2.PipelineTask, error) {
		return spec2.PipelineTask{PipelineID: 1}, nil
	})
	defer m1.Unpatch()

	m2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipeline", func(_ *dbclient.Client, id interface{}, ops ...dbclient.SessionOption) (spec2.Pipeline, error) {
		return spec2.Pipeline{PipelineBase: spec2.PipelineBase{ID: 1}}, nil
	})
	defer m2.Unpatch()

	m3 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineTaskMetadata", func(_ *dbclient.Client, id uint64, result *apistructs.PipelineTaskResult) error {
		return nil
	})
	defer m3.Unpatch()

	m4 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineTaskInspect", func(_ *dbclient.Client, id uint64, inspect apistructs.PipelineTaskInspect) error {
		return nil
	})
	defer m4.Unpatch()

	m5 := monkey.Patch(events.EmitTaskEvent, func(task *spec2.PipelineTask, p *spec2.Pipeline) {
		return
	})
	defer m5.Unpatch()
	data := []byte("{\"metadata\":[{\"name\":\"pipelineID\",\"value\":\"1\"}],\"errors\":[{\"code\":\"\",\"msg\":\"network error\",\"ctx\":null}],\"pipelineID\":1,\"pipelineTaskID\":1}")
	pSvc := PipelineSvc{dbClient: db}
	err := pSvc.DealPipelineCallbackOfAction(data)
	assert.NoError(t, err)
}
