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

package projectpipeline

import (
	"reflect"
	"testing"
)

func Test_parseSourceDicePipelineYmlName(t *testing.T) {
	type args struct {
		ymlName string
		branch  string
	}
	tests := []struct {
		name string
		args args
		want *pipelineYmlName
	}{
		{
			name: "test with other source",
			args: args{
				ymlName: "pipeline.yml",
				branch:  "master",
			},
			want: nil,
		},
		{
			name: "test with dice source1",
			args: args{
				ymlName: "200/DEV/develop/pipeline.yml",
				branch:  "develop",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "develop",
				fileName:  "pipeline.yml",
			},
		},
		{
			name: "test with dice source2",
			args: args{
				ymlName: "200/DEV/develop/.erda/pipeline.yml",
				branch:  "develop",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "develop",
				fileName:  ".erda/pipeline.yml",
			},
		},
		{
			name: "test with dice source3",
			args: args{
				ymlName: "200/DEV/feature/erda/.erda/pipelines/pipeline.yml",
				branch:  "feature/erda",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "feature/erda",
				fileName:  ".erda/pipelines/pipeline.yml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSourceDicePipelineYmlName(tt.args.ymlName, tt.args.branch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSourceDicePipelineYmlName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeProjectPipelineName(t *testing.T) {
	type args struct {
		pipelineYml string
		fileName    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with from pipelineYml",
			args: args{
				pipelineYml: `version: "1.1"
name: pipeline-deploy
stages:
`,
				fileName: "pipeline.yml",
			},
			want: "pipeline-deploy",
		},
		{
			name: "test with from fileName",
			args: args{
				pipelineYml: `version: "1.1"
stages:
`,
				fileName: "erda-test-pipeline.yml",
			},
			want: "erda-test-pipeline.yml",
		},
		{
			name: "test with from fileName2",
			args: args{
				pipelineYml: `version: "1.1"
stages:
`,
				fileName: "erda-project-develop-test-pipeline.yml",
			},
			want: "pipeline.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeProjectPipelineName(tt.args.pipelineYml, tt.args.fileName); got != tt.want {
				t.Errorf("MakeProjectPipelineName() = %v, want %v", got, tt.want)
			}
		})
	}
}
