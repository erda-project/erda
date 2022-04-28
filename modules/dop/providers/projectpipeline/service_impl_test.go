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

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs"
	spb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestGetRulesByCategoryKey(t *testing.T) {
	tt := []struct {
		key  apistructs.PipelineCategory
		want []string
	}{
		{apistructs.CategoryBuildDeploy, apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildDeploy]},
		{apistructs.CategoryBuildArtifact, apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildArtifact]},
		{apistructs.CategoryOthers, append(append(append(apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildDeploy],
			apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildArtifact]...), apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildCombineArtifact]...),
			apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildIntegration]...)},
	}
	for _, v := range tt {
		for i := range v.want {
			if !strutil.InSlice(v.want[i], getRulesByCategoryKey(v.key)) {
				t.Errorf("fail")
			}
		}
	}
}

func TestPipelineYmlsFilterIn(t *testing.T) {
	pipelineYmls := []string{"pipeline.yml", ".erda/pipelines/ci-artifact.yml", "a.yml", "b.yml"}
	uncategorizedPipelineYmls := pipelineYmlsFilterIn(pipelineYmls, func(yml string) bool {
		for k := range apistructs.GetRuleCategoryKeyMap() {
			if k == yml {
				return false
			}
		}
		return true
	})
	if len(uncategorizedPipelineYmls) != 2 {
		t.Errorf("fail")
	}
}

func TestGetFilePath(t *testing.T) {
	tt := []struct {
		path string
		want string
	}{
		{
			path: "pipeline.yml",
			want: "",
		},
		{
			path: ".erda/pipelines/pipeline.yml",
			want: ".erda/pipelines",
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, getFilePath(v.path))
	}
}

func TestIsSameSourceInApp(t *testing.T) {
	tt := []struct {
		source *spb.PipelineSource
		params *pb.UpdateProjectPipelineRequest
		want   bool
	}{
		{
			source: &spb.PipelineSource{
				Ref:  "master",
				Path: "",
				Name: "pipeline.yml",
			},
			params: &pb.UpdateProjectPipelineRequest{
				ProjectPipelineSource: &pb.ProjectPipelineSource{
					Ref:      "master",
					Path:     "",
					FileName: "pipeline.yml",
				},
			},
			want: true,
		},
		{
			source: &spb.PipelineSource{
				Ref:  "master",
				Path: "",
				Name: "pipeline.yml",
			},
			params: &pb.UpdateProjectPipelineRequest{
				ProjectPipelineSource: &pb.ProjectPipelineSource{
					Ref:      "master",
					Path:     ".erda/pipelines",
					FileName: "pipeline.yml",
				},
			},
			want: false,
		},
		{
			source: &spb.PipelineSource{
				Ref:  "master",
				Path: "",
				Name: "pipeline.yml",
			},
			params: &pb.UpdateProjectPipelineRequest{
				ProjectPipelineSource: &pb.ProjectPipelineSource{
					Ref:      "develop",
					Path:     "",
					FileName: "pipeline.yml",
				},
			},
			want: false,
		},
		{
			source: &spb.PipelineSource{
				Ref:  "master",
				Path: "",
				Name: "pipeline.yml",
			},
			params: &pb.UpdateProjectPipelineRequest{
				ProjectPipelineSource: &pb.ProjectPipelineSource{
					Ref:      "master",
					Path:     "",
					FileName: "ci.yml",
				},
			},
			want: false,
		},
	}

	for _, v := range tt {
		assert.Equal(t, v.want, isSameSourceInApp(v.source, v.params))
	}

}

func Test_getRemotes(t *testing.T) {
	type args struct {
		appNames    []string
		orgName     string
		projectName string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test remote",
			args: args{
				appNames:    []string{"erda-release", "dice"},
				orgName:     "org",
				projectName: "erda",
			},
			want: []string{"org/erda/erda-release", "org/erda/dice"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRemotes(tt.args.appNames, tt.args.orgName, tt.args.projectName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRemotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectPipelineService_fetchRemotePipeline(t *testing.T) {
	type fields struct {
		logger logs.Logger
	}
	type args struct {
		source *spb.PipelineSource
		orgID  string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "test fetch fetchRemotePipeline",
			fields: fields{},
			args: args{
				source: &spb.PipelineSource{
					Remote: "org/erda/erda-release",
				},
				orgID:  "",
				userID: "",
			},
			want: `version: "1.1"
stages:
  - stage:
      - git-checkout:
          alias: git-checkout
          description: 代码仓库克隆`,
			wantErr: false,
		},
	}

	var bdl *bundle.Bundle
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetGittarBlobNode",
		func(bdl *bundle.Bundle, repo, orgID, userID string) (string, error) {
			return `version: "1.1"
stages:
  - stage:
      - git-checkout:
          alias: git-checkout
          description: 代码仓库克隆`, nil
		})
	defer pm.Unpatch()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectPipelineService{
				bundle: bdl,
			}
			got, err := p.fetchRemotePipeline(tt.args.source, tt.args.orgID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchRemotePipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchRemotePipeline() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNameByRemote(t *testing.T) {
	type args struct {
		remote string
	}
	tests := []struct {
		name string
		args args
		want RemoteName
	}{
		{
			name: "test getNameByRemote",
			args: args{remote: "org/erda/erda-release"},
			want: RemoteName{
				OrgName:     "org",
				ProjectName: "erda",
				AppName:     "erda-release",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNameByRemote(tt.args.remote); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNameByRemote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectPipelineService_makeLocationByAppID(t *testing.T) {
	type fields struct {
		bundle *bundle.Bundle
	}
	type args struct {
		appID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test makeLocationByAppID",
			fields:  fields{},
			args:    args{appID: 1},
			want:    "cicd/org/erda",
			wantErr: false,
		},
	}

	var bdl *bundle.Bundle
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp",
		func(bdl *bundle.Bundle, appID uint64) (*apistructs.ApplicationDTO, error) {
			return &apistructs.ApplicationDTO{
				ID:          1,
				Name:        "erda-release",
				OrgName:     "org",
				ProjectName: "erda",
			}, nil
		})
	defer pm.Unpatch()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectPipelineService{
				bundle: bdl,
			}
			got, err := p.makeLocationByAppID(tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeLocationByAppID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeLocationByAppID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectPipelineService_makeLocationByProjectID(t *testing.T) {
	type fields struct {
		bundle *bundle.Bundle
	}
	type args struct {
		projectID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test makeLocationByProjectID",
			fields:  fields{},
			args:    args{},
			want:    "cicd/org/erda",
			wantErr: false,
		},
	}

	var bdl *bundle.Bundle
	pm := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(bdl *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				ID:   1,
				Name: "erda",
			}, nil
		})
	defer pm.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg",
		func(bdl *bundle.Bundle, id interface{}) (*apistructs.OrgDTO, error) {
			return &apistructs.OrgDTO{
				ID:   1,
				Name: "org",
			}, nil
		})
	defer pm2.Unpatch()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectPipelineService{
				bundle: bdl,
			}
			got, err := p.makeLocationByProjectID(tt.args.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeLocationByProjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeLocationByProjectID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectPipelineService_checkDataPermission(t *testing.T) {
	type fields struct {
		logger logs.Logger
	}
	type args struct {
		project *apistructs.ProjectDTO
		org     *apistructs.OrgDTO
		source  *spb.PipelineSource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test has permission",
			fields: fields{},
			args: args{
				project: &apistructs.ProjectDTO{Name: "erda"},
				org:     &apistructs.OrgDTO{Name: "org"},
				source:  &spb.PipelineSource{Remote: "org/erda/erda-release"},
			},
			wantErr: false,
		},
		{
			name:   "test no permission",
			fields: fields{},
			args: args{
				project: &apistructs.ProjectDTO{Name: "dice"},
				org:     &apistructs.OrgDTO{Name: "org"},
				source:  &spb.PipelineSource{Remote: "org/erda/erda-release"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectPipelineService{
				logger: tt.fields.logger,
			}
			if err := p.checkDataPermission(tt.args.project, tt.args.org, tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("checkDataPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
