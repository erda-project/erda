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
	"testing"

	"github.com/stretchr/testify/assert"

	spb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/apistructs"
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
