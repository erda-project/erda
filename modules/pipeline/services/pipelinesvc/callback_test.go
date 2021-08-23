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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
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

	task := &spec.PipelineTask{
		Result: apistructs.PipelineTaskResult{
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
	task.Result.Errors = task.Result.AppendError(newTaskErrors...)

	assert.Equal(t, 3, len(task.Result.Errors))
}
