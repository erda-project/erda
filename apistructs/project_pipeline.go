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

package apistructs

import "path/filepath"

type PipelineCategory string

type ProjectPipelineType string

const (
	CategoryBuildDeploy          = "build-deploy"
	CategoryBuildArtifact        = "build-artifact"
	CategoryBuildCombineArtifact = "build-combine-artifact"
	CategoryBuildIntegration     = "build-integration"
	CategoryOthers               = "others"
)

const (
	PipelineTypeCICD ProjectPipelineType = "cicd"
)

const (
	SourceTypeErda = "erda"
)

func (p ProjectPipelineType) String() string {
	return string(p)
}

func MakeLocation(app *ApplicationDTO, t ProjectPipelineType) string {
	return filepath.Join(t.String(), app.OrgName, app.ProjectName)
}

func (c PipelineCategory) String() string {
	return string(c)
}

var CategoryKeyRuleMap = map[PipelineCategory][]string{
	CategoryBuildDeploy:          {"pipeline.yml", ".erda/pipelines/ci-deploy.yml"},
	CategoryBuildArtifact:        {".erda/pipelines/ci-artifact.yml"},
	CategoryBuildCombineArtifact: {".erda/pipelines/combine-artifact.yml"},
	CategoryBuildIntegration:     {".erda/pipelines/integration.yml"},
}

func GetRuleCategoryKeyMap() map[string]PipelineCategory {
	m := make(map[string]PipelineCategory, 0)
	for k, rules := range CategoryKeyRuleMap {
		for _, v := range rules {
			m[v] = k
		}
	}
	return m
}
