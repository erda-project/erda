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

package reconciler

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestParsePipelineOutputRef(t *testing.T) {
	reffedTask, reffedKey, err := parsePipelineOutputRef("${Dice 文档:OUTPUT:status}")
	spew.Dump(reffedTask, reffedKey)
	assert.NoError(t, err)
	assert.Equal(t, "Dice 文档", reffedTask)
	assert.Equal(t, "status", reffedKey)
}

func TestParsePipelineOutputRefV2(t *testing.T) {
	reffedTask, reffedKey, err := parsePipelineOutputRefV2("${{ outputs.a.b }}")
	spew.Dump(reffedTask, reffedKey)
	assert.NoError(t, err)
	assert.Equal(t, "a", reffedTask)
	assert.Equal(t, "b", reffedKey)
}

func Test_handleParentSnippetTaskOutputs(t *testing.T) {
	db := &dbclient.Client{}

	m1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTask", func(_ *dbclient.Client, id interface{}) (spec.PipelineTask, error) {
		return spec.PipelineTask{Result: &apistructs.PipelineTaskResult{}}, nil
	})
	defer m1.Unpatch()

	m2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineTaskSnippetDetail", func(_ *dbclient.Client, id uint64, snippetDetail apistructs.PipelineTaskSnippetDetail, ops ...dbclient.SessionOption) error {
		return nil
	})
	defer m2.Unpatch()

	m3 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineTaskMetadata", func(_ *dbclient.Client, id uint64, result *apistructs.PipelineTaskResult) error {
		return nil
	})
	defer m3.Unpatch()

	r := &Reconciler{dbClient: db}
	parentTaskID := uint64(1)
	snippetPipeline := &spec.Pipeline{PipelineBase: spec.PipelineBase{ParentTaskID: &parentTaskID}}
	err := r.handleParentSnippetTaskOutputs(snippetPipeline, []apistructs.PipelineOutputWithValue{{
		PipelineOutput: apistructs.PipelineOutput{Name: "pipelineID"},
		Value:          "1",
	}})
	assert.NoError(t, err)
}

func Test_calculateAndUpdatePipelineOutputValues(t *testing.T) {
	db := &dbclient.Client{}

	m1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "UpdatePipelineExtraSnapshot", func(_ *dbclient.Client, pipelineID uint64, snapshot spec.Snapshot, ops ...dbclient.SessionOption) error {
		return nil
	})
	defer m1.Unpatch()

	tasks := []*spec.PipelineTask{&spec.PipelineTask{Name: "1", Result: &apistructs.PipelineTaskResult{
		Metadata: apistructs.Metadata{{
			Name:  "pipelineID",
			Value: "1",
		}},
	}}}
	r := &Reconciler{dbClient: db}
	_, err := r.calculateAndUpdatePipelineOutputValues(&spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}}, tasks)
	assert.NoError(t, err)
}
