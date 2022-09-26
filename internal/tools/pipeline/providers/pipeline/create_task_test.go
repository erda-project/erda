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

package pipeline

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func TestPipelineSvc_judgeTaskExecutor(t *testing.T) {
	type args struct {
		action     *spec.PipelineTask
		actionSpec *apistructs.ActionSpec
	}
	tests := []struct {
		name    string
		args    args
		want    spec.PipelineTaskExecutorKind
		want1   spec.PipelineTaskExecutorName
		wantErr bool
	}{
		{
			name: "empty executor",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: nil,
				},
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "terminus-dev",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sJob,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sJobDefault, "terminus-dev")),
			wantErr: false,
		},
		{
			name: "empty spec",
			args: args{
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "terminus-dev",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sJob,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sJobDefault, "terminus-dev")),
			wantErr: false,
		},
		{
			name: "not match kind",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Name: spec.PipelineTaskExecutorNameEmpty.String(),
						Kind: "__other",
					},
				},
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "erda-op",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sJob,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sJobDefault, "erda-op")),
			wantErr: false,
		},
		{
			name: "not match name",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: string(spec.PipelineTaskExecutorKindMemory),
						Name: "__other",
					},
				},
				action: &spec.PipelineTask{Extra: spec.PipelineTaskExtra{ClusterName: "terminus-dev"}},
			},
			want:    spec.PipelineTaskExecutorKindK8sJob,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sJobDefault, "terminus-dev")),
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: string(spec.PipelineTaskExecutorKindAPITest),
						Name: spec.PipelineTaskExecutorNameSchedulerDefault.String(),
					},
				},
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{},
				},
			},
			want:    spec.PipelineTaskExecutorKindAPITest,
			want1:   spec.PipelineTaskExecutorNameSchedulerDefault,
			wantErr: false,
		},
		{
			name: "not find kind or name",
			args: args{
				actionSpec: &apistructs.ActionSpec{
					Executor: &apistructs.ActionExecutor{
						Kind: "__test_kind",
						Name: "__test_name",
					},
				},
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "erda-op",
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sJob,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sJobDefault, "erda-op")),
			wantErr: false,
		},
		{
			name: "k8s flink job",
			args: args{
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "erda-op",
						Action: pipelineyml.Action{
							Params: map[string]interface{}{
								"bigDataConf": "{\n    \"flinkConf\": {\"kind\": \"job\"}\n}",
							},
						},
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sFlink,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sFlinkDefault, "erda-op")),
			wantErr: false,
		},
		{
			name: "k8s spark job",
			args: args{
				action: &spec.PipelineTask{
					Extra: spec.PipelineTaskExtra{
						ClusterName: "erda-op",
						Action: pipelineyml.Action{
							Params: map[string]interface{}{
								"bigDataConf": "{\n    \"sparkConf\": {\"kind\": \"job\"}\n}",
							},
						},
					},
				},
			},
			want:    spec.PipelineTaskExecutorKindK8sSpark,
			want1:   spec.PipelineTaskExecutorName(fmt.Sprintf("%s-%s", spec.PipelineTaskExecutorNameK8sSparkDefault, "erda-op")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &pipelineService{}
			got, got1 := s.judgeTaskExecutor(tt.args.action, tt.args.actionSpec)
			if got != tt.want {
				t.Errorf("judgeTaskExecutor() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("judgeTaskExecutor() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGenSnippetTaskExtra(t *testing.T) {
	svc := pipelineService{
		bdl: bundle.New(
			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))),
			bundle.WithScheduler(),
		),
	}
	p := &spec.Pipeline{
		PipelineBase: spec.PipelineBase{},
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				Namespace:               "custom-namespace",
				NotPipelineControlledNs: true,
			},
		},
	}
	taskExtra := svc.genSnippetTaskExtra(p, &pipelineyml.Action{})
	assert.Equal(t, true, taskExtra.NotPipelineControlledNs)
}

func TestCalculateTaskTimeoutDuration(t *testing.T) {
	svc := pipelineService{
		bdl: bundle.New(
			bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))),
			bundle.WithScheduler(),
		),
	}
	duration := svc.calculateTaskTimeoutDuration(&pipelineyml.Action{
		Timeout: pipelineyml.TimeoutDuration4Forever,
	})
	assert.Equal(t, time.Duration(-1), duration)
}
