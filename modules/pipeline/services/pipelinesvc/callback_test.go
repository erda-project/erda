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
