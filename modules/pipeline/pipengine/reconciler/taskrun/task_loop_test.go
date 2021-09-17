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

package taskrun

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/bmizerany/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pexpr/pexpr_params"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestTaskRun_handleTaskLoop(t *testing.T) {
	type fields struct {
		Task                            *spec.PipelineTask
		P                               *spec.Pipeline
		EnsureFetchLatestPipelineStatus apistructs.PipelineStatus
		assertStatus                    apistructs.PipelineStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test_not_end_status",
			fields: fields{
				Task: &spec.PipelineTask{
					Status: apistructs.PipelineStatusRunning,
				},
			},
			wantErr: false,
		},
		{
			name: "test_not_end_status",
			fields: fields{
				Task: &spec.PipelineTask{
					ID:     1,
					Status: apistructs.PipelineStatusFailed,
				},
				P: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID: 1,
					},
				},
				EnsureFetchLatestPipelineStatus: apistructs.PipelineStatusFailed,
			},
			wantErr: false,
		},
		{
			name: "test_loop_was_empty",
			fields: fields{
				Task: &spec.PipelineTask{
					ID:     1,
					Status: apistructs.PipelineStatusFailed,
				},
				P: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID: 1,
					},
				},
				EnsureFetchLatestPipelineStatus: apistructs.PipelineStatusRunning,
			},
			wantErr: false,
		},
		{
			name: "test_loop_break_false",
			fields: fields{
				Task: &spec.PipelineTask{
					ID:     1,
					Status: apistructs.PipelineStatusFailed,
					Extra: spec.PipelineTaskExtra{
						LoopOptions: &apistructs.PipelineTaskLoopOptions{
							CalculatedLoop: &apistructs.PipelineTaskLoop{
								Break: "asd",
							},
						},
					},
				},
				P: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID: 1,
					},
				},
				EnsureFetchLatestPipelineStatus: apistructs.PipelineStatusRunning,
				assertStatus:                    apistructs.PipelineStatusAnalyzed,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *dbclient.Client
			var patch1 *monkey.PatchGuard
			patch1 = monkey.PatchInstanceMethod(reflect.TypeOf(client), "CreatePipelineReport", func(client *dbclient.Client, report *spec.PipelineReport, ops ...dbclient.SessionOption) error {
				return nil
			})

			var patch2 *monkey.PatchGuard
			patch2 = monkey.Patch(pexpr_params.GenerateParamsFromTask, func(pipelineID uint64, taskID uint64, taskStatus apistructs.PipelineStatus) map[string]string {
				return map[string]string{}
			})

			var patch *monkey.PatchGuard
			var tr = &TaskRun{}
			if tt.fields.EnsureFetchLatestPipelineStatus != "" {
				patch = monkey.PatchInstanceMethod(reflect.TypeOf(tr), "EnsureFetchLatestPipelineStatus", func(tr *TaskRun) {
					tr.QueriedPipelineStatus = tt.fields.EnsureFetchLatestPipelineStatus
				})
			}

			tr.Task = tt.fields.Task
			tr.P = tt.fields.P
			tr.DBClient = client

			if err := tr.handleTaskLoop(); (err != nil) != tt.wantErr {
				t.Errorf("handleTaskLoop() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.fields.assertStatus != "" {
				assert.Equal(t, tr.Task.Status, tt.fields.assertStatus)
			}

			if patch != nil {
				patch.Unpatch()
			}
			patch1.Unpatch()
			patch2.Unpatch()
		})
	}
}
