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
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/cms"
	"github.com/erda-project/erda/modules/pipeline/providers/definition_client/deftype"
	"github.com/erda-project/erda/modules/pkg/gitflowutil"
)

func TestGetBranch(t *testing.T) {
	ss := []struct {
		ref  string
		Want string
	}{
		{"", ""},
		{"refs/heads/", ""},
		{"refs/heads/master", "master"},
		{"refs/heads/feature/test", "feature/test"},
	}
	for _, v := range ss {
		assert.Equal(t, v.Want, getBranch(v.ref))
	}
}

func TestIsPipelineYmlPath(t *testing.T) {
	ss := []struct {
		path string
		want bool
	}{
		{"pipeline.yml", true},
		{".dice/pipelines/a.yml", true},
		{"", false},
		{"dice/pipeline.yml", false},
	}
	for _, v := range ss {
		assert.Equal(t, v.want, isPipelineYmlPath(v.path))
	}

}

type process struct{}

func (process) ProcessPipelineDefinition(ctx context.Context, req deftype.ClientDefinitionProcessRequest) (*deftype.ClientDefinitionProcessResponse, error) {
	return &deftype.ClientDefinitionProcessResponse{
		ID:              1,
		PipelineSource:  req.PipelineSource,
		PipelineYmlName: req.PipelineYmlName,
		PipelineYml:     req.PipelineYml,
		VersionLock:     req.VersionLock,
	}, nil
}

func TestPipeline_deletePipelineDefinition(t *testing.T) {
	type args struct {
		name   string
		appID  uint64
		branch string
		userID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				name:   "test",
				appID:  1,
				branch: "test",
				userID: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				ds: process{},
			}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertPipelineToV2", func(p *Pipeline, pv1 *apistructs.PipelineCreateRequest) (*apistructs.PipelineCreateRequestV2, error) {
				return &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: pv1.PipelineYmlName,
					PipelineSource:  apistructs.PipelineSource(pv1.PipelineYmlSource),
					PipelineYml:     pv1.PipelineYmlContent,
				}, nil
			})
			if err := p.deletePipelineDefinition(tt.args.name, tt.args.appID, tt.args.branch, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("deletePipelineDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			patch.Unpatch()
		})
	}
}

func TestPipeline_reportPipelineDefinition(t *testing.T) {
	type args struct {
		appDto      *apistructs.ApplicationDTO
		userID      string
		branch      string
		name        string
		pipelineYml string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				appDto: &apistructs.ApplicationDTO{
					Name:        "test",
					ProjectID:   1,
					ProjectName: "test",
					OrgID:       1,
				},
				branch:      "test",
				userID:      "1",
				name:        "test",
				pipelineYml: "version: 1.1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				ds: process{},
			}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertPipelineToV2", func(p *Pipeline, pv1 *apistructs.PipelineCreateRequest) (*apistructs.PipelineCreateRequestV2, error) {
				return &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: pv1.PipelineYmlName,
					PipelineSource:  apistructs.PipelineSource(pv1.PipelineYmlSource),
					PipelineYml:     pv1.PipelineYmlContent,
				}, nil
			})

			if err := p.reportPipelineDefinition(tt.args.appDto, tt.args.userID, tt.args.branch, tt.args.name, tt.args.pipelineYml); (err != nil) != tt.wantErr {
				t.Errorf("reportPipelineDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			patch.Unpatch()
		})
	}
}

func Test_getWorkspaceMainBranch(t *testing.T) {
	type args struct {
		workspace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "invalid workspace",
			args: args{
				workspace: "xxx",
			},
			want: "",
		},
		{
			name: "dev",
			args: args{
				workspace: "dev",
			},
			want: "feature",
		},
		{
			name: "Dev",
			args: args{
				workspace: "Dev",
			},
			want: "feature",
		},
		{
			name: "test",
			args: args{
				workspace: "test",
			},
			want: "develop",
		},
		{
			name: "staging",
			args: args{
				workspace: "staging",
			},
			want: "release",
		},
		{
			name: "prOD",
			args: args{
				workspace: "prOD",
			},
			want: "master",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getWorkspaceMainBranch(tt.args.workspace); got != tt.want {
				t.Errorf("getWorkspaceMainBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeAppDefaultCmsNs(t *testing.T) {
	type args struct {
		appID uint64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "app 1",
			args: args{
				appID: 1,
			},
			want: "app-1-default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeAppDefaultCmsNs(tt.args.appID); got != tt.want {
				t.Errorf("makeAppDefaultCmsNs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeAppWorkspaceCmsNs(t *testing.T) {
	type args struct {
		appID     uint64
		workspace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "app 1 dev",
			args: args{
				appID:     1,
				workspace: gitflowutil.DevWorkspace,
			},
			want: "app-1-dev",
		},
		{
			name: "app 1 prod",
			args: args{
				appID:     1,
				workspace: gitflowutil.ProdWorkspace,
			},
			want: "app-1-prod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeAppWorkspaceCmsNs(tt.args.appID, tt.args.workspace); got != tt.want {
				t.Errorf("makeAppWorkspaceCmsNs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeBranchWorkspaceLevelCmsNs(t *testing.T) {
	type args struct {
		appID     uint64
		workspace string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "invalid workspace",
			args: args{
				appID:     1,
				workspace: "xxx",
			},
			want: []string{cms.MakeAppDefaultSecretNamespace("1")},
		},
		{
			name: "staging",
			args: args{
				appID:     1,
				workspace: "STAGING",
			},
			want: []string{cms.MakeAppDefaultSecretNamespace("1"), cms.MakeAppBranchPrefixSecretNamespaceByBranchPrefix("1", "release")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeBranchWorkspaceLevelCmsNs(tt.args.appID, tt.args.workspace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeBranchWorkspaceLevelCmsNs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeAppWorkspaceLevelCmsNs(t *testing.T) {
	type args struct {
		appID     uint64
		workspace string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "invalid workspace",
			args: args{
				appID:     1,
				workspace: "xxx",
			},
			want: []string{"app-1-default", "app-1-xxx"},
		},
		{
			name: "staging",
			args: args{
				appID:     1,
				workspace: "staging",
			},
			want: []string{"app-1-default", "app-1-staging"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeAppWorkspaceLevelCmsNs(tt.args.appID, tt.args.workspace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeAppWorkspaceLevelCmsNs() = %v, want %v", got, tt.want)
			}
		})
	}
}


