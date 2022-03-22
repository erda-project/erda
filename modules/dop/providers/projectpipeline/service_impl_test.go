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

	"github.com/erda-project/erda/apistructs"
)

func TestGetRulesByCategoryKey(t *testing.T) {
	tt := []struct {
		key  apistructs.PipelineCategory
		want []string
	}{
		{apistructs.CategoryBuildDeploy, apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildDeploy]},
		{apistructs.CategoryBuildArtifact, apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildArtifact]},
		{apistructs.CategoryOthers, append(apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildDeploy], apistructs.CategoryKeyRuleMap[apistructs.CategoryBuildArtifact]...)},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, getRulesByCategoryKey(v.key.String()))
	}
}
