// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package migrate

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

// taskName -> stepID
// 老语法 -> 新语法
var (
	oldOutputRegex = regexp.MustCompile(`\${([^{}]+):OUTPUT:([^{}]+)}`)
	oldParamsRegex = regexp.MustCompile(`\${params.([^{}]+)}`)
)

func replacePipelineYmlOutputsForTaskStepID(ymlContent string, taskNameStepIDRelations map[string]*dao.AutoTestSceneStep, caseInode string) string {
	return strutil.ReplaceAllStringSubmatchFunc(oldOutputRegex, ymlContent, func(subs []string) string {
		taskName := subs[1]
		metaKey := subs[2]

		step, ok := taskNameStepIDRelations[taskName]
		if !ok {
			logrus.Warnf("failed to replace pipeline yml for task stepID, taskName %q doesn't have corresponding step, caseInode: %s", taskName, caseInode)
			// 返回原数据，调整格式
			newFormat := fmt.Sprintf("${{ outputs.%s.%s }}", taskName, metaKey)
			return newFormat
		}

		newStr := fmt.Sprintf(`${{ outputs.%d.%s }}`, step.ID, metaKey)
		return newStr
	})
}

func replacePipelineYmlParams(ymlContent string, taskNameStepIDRelations map[string]*dao.AutoTestSceneStep, caseInode string) string {
	return strutil.ReplaceAllStringSubmatchFunc(oldParamsRegex, ymlContent, func(subs []string) string {
		paramKey := subs[1]

		newFormat := fmt.Sprintf(`${{ params.%s }}`, paramKey)
		return newFormat
	})
}
