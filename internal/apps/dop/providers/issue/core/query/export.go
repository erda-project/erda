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

package query

import (
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
)

// GetStageMap return a map,the key is the struct of dice_issue_stage.Value and dice_issue_stage.IssueType,
// the value is dice_issue_stage.Name
// example: name: 代码研发, value: codeDevelopment
func GetStageMap(stages []dao.IssueStage) map[IssueStage]string {
	stageMap := make(map[IssueStage]string, len(stages))
	for _, v := range stages {
		if v.Value != "" && v.IssueType != "" {
			stage := IssueStage{
				Type:  v.IssueType,
				Value: v.Value,
			}
			stageMap[stage] = v.Name
		}
	}
	return stageMap
}
