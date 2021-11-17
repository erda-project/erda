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
	"github.com/bmizerany/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func Test_handlerFactoryNewPolicyType(t *testing.T) {
	type args struct {
		task *spec.PipelineTask
		opt  PolicyHandlerOptions
	}
	tests := []struct {
		name       string
		args       args
		wantResult *spec.PipelineTask
		wantErr    bool
	}{
		{
			name: "test",
			args: args{
				task: &spec.PipelineTask{
					ID: 1,
				},
			},
			wantErr: false,
			wantResult: &spec.PipelineTask{
				ID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := NewRun{}.ResetTask(tt.args.task, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("handlerFactoryNewPolicyType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("handlerFactoryNewPolicyType() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_handlerLastSuccessResultPolicyType(t *testing.T) {
	type args struct {
		task *spec.PipelineTask

		mockSuccessPipelines []spec.Pipeline
		mockTasks            spec.PipelineTask
	}
	tests := []struct {
		name       string
		args       args
		wantResult *spec.PipelineTask
		wantErr    bool
	}{
		{
			name: "test_not_snippet",
			args: args{
				task: &spec.PipelineTask{
					IsSnippet: false,
				},
			},
			wantResult: &spec.PipelineTask{
				IsSnippet: false,
			},
			wantErr: false,
		},
		{
			name: "test_empty_snippetConfig",
			args: args{
				task: &spec.PipelineTask{
					IsSnippet: true,
				},
			},
			wantResult: &spec.PipelineTask{
				IsSnippet: true,
			},
			wantErr: false,
		},
		{
			name: "test_not_find_success_pipeline",
			args: args{
				task: &spec.PipelineTask{
					IsSnippet: true,
					Extra: spec.PipelineTaskExtra{
						Action: pipelineyml.Action{
							SnippetConfig: &pipelineyml.SnippetConfig{
								Name:   "name",
								Source: "source",
							},
						},
					},
				},
			},
			wantResult: &spec.PipelineTask{
				IsSnippet: true,
				Extra: spec.PipelineTaskExtra{
					Action: pipelineyml.Action{
						SnippetConfig: &pipelineyml.SnippetConfig{
							Name:   "name",
							Source: "source",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test_not_find_task",
			args: args{
				task: &spec.PipelineTask{
					Name:      "name",
					Type:      "type",
					IsSnippet: true,
					Extra: spec.PipelineTaskExtra{
						Action: pipelineyml.Action{
							SnippetConfig: &pipelineyml.SnippetConfig{
								Name:   "name",
								Source: "source",
							},
						},
					},
				},
				mockSuccessPipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							ID: 1,
						},
					},
				},
			},
			wantResult: &spec.PipelineTask{
				Name:      "name",
				Type:      "type",
				IsSnippet: true,
				Extra: spec.PipelineTaskExtra{
					Action: pipelineyml.Action{
						SnippetConfig: &pipelineyml.SnippetConfig{
							Name:   "name",
							Source: "source",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test_success",
			args: args{
				task: &spec.PipelineTask{
					Name:      "name",
					Type:      "type",
					IsSnippet: true,
					Extra: spec.PipelineTaskExtra{
						Action: pipelineyml.Action{
							SnippetConfig: &pipelineyml.SnippetConfig{
								Name:   "name",
								Source: "source",
							},
						},
					},
				},
				mockSuccessPipelines: []spec.Pipeline{
					{
						PipelineBase: spec.PipelineBase{
							ID: 1,
						},
					},
				},
				mockTasks: spec.PipelineTask{
					ID:        1,
					Name:      "name",
					Type:      "type",
					Status:    apistructs.PipelineStatusSuccess,
					IsSnippet: true,
					Extra: spec.PipelineTaskExtra{
						Action: pipelineyml.Action{
							SnippetConfig: &pipelineyml.SnippetConfig{
								Name:   "name",
								Source: "source",
							},
						},
					},
				},
			},
			wantResult: &spec.PipelineTask{
				Name:      "name",
				Type:      "type",
				IsSnippet: true,
				Status:    apistructs.PipelineStatusSuccess,
				Extra: spec.PipelineTaskExtra{
					Action: pipelineyml.Action{
						SnippetConfig: &pipelineyml.SnippetConfig{
							Name:   "name",
							Source: "source",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := PolicyHandlerOptions{}
			dbClient := &dbclient.Client{}
			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "PageListPipelines", func(client *dbclient.Client, req apistructs.PipelinePageListRequest, ops ...dbclient.SessionOption) ([]spec.Pipeline, []uint64, int64, int64, error) {
				return tt.args.mockSuccessPipelines, nil, 0, 0, nil
			})

			patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(dbClient), "GetPipelineTask", func(client *dbclient.Client, id interface{}) (spec.PipelineTask, error) {
				return tt.args.mockTasks, nil
			})

			opt.dbClient = dbClient
			defer patch1.Unpatch()
			defer patch2.Unpatch()

			gotResult, err := TryLastSuccessResult{}.ResetTask(tt.args.task, opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("handlerLastSuccessResultPolicyType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, gotResult.Type, tt.wantResult.Type)
			assert.Equal(t, gotResult.Name, tt.wantResult.Name)
			assert.Equal(t, gotResult.Status, tt.wantResult.Status)
			assert.Equal(t, gotResult.Extra, tt.wantResult.Extra)
		})
	}
}

func TestReconciler_adaptorPolicy(t *testing.T) {
	type args struct {
		task *spec.PipelineTask
	}
	tests := []struct {
		name       string
		args       args
		wantResult *spec.PipelineTask
		wantErr    bool
	}{
		{
			name: "test",
			args: args{
				task: &spec.PipelineTask{
					Name:      "name",
					Type:      "type",
					IsSnippet: true,
					Extra: spec.PipelineTaskExtra{
						Action: pipelineyml.Action{
							SnippetConfig: &pipelineyml.SnippetConfig{
								Name:   "name",
								Source: "source",
							},
						},
					},
				},
			},
			wantResult: &spec.PipelineTask{
				Name:      "name",
				Type:      "type",
				IsSnippet: true,
				Extra: spec.PipelineTaskExtra{
					Action: pipelineyml.Action{
						SnippetConfig: &pipelineyml.SnippetConfig{
							Name:   "name",
							Source: "source",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{}
			gotResult, err := r.adaptPolicy(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("adaptorPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("adaptorPolicy() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
