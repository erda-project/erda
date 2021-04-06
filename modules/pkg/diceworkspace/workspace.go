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

package diceworkspace

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
)

func GetByGitReference(ref string, branchRules []*apistructs.BranchRule) (apistructs.DiceWorkspace, error) {
	reference := GetValidBranchByGitReference(ref, branchRules)
	if reference.Workspace == "" {
		return "", fmt.Errorf("not support reference %s", ref)
	}
	return apistructs.DiceWorkspace(reference.Workspace), nil
}
func GetValidBranchByGitReference(ref string, branchRules []*apistructs.BranchRule) *apistructs.ValidBranch {
	for _, branchRule := range branchRules {
		branchFilters := strings.Split(branchRule.Rule, ",")
		for _, branchFilter := range branchFilters {
			if IsRefPatternMatch(ref, []string{branchFilter}) {
				return &apistructs.ValidBranch{
					Name:              ref,
					IsProtect:         branchRule.IsProtect,
					NeedApproval:      branchRule.NeedApproval,
					IsTriggerPipeline: branchRule.IsTriggerPipeline,
					Workspace:         branchRule.Workspace,
					ArtifactWorkspace: branchRule.ArtifactWorkspace,
				}
			}
		}
	}
	return &apistructs.ValidBranch{
		Name:              ref,
		IsProtect:         false,
		NeedApproval:      false,
		IsTriggerPipeline: false,
		Workspace:         "",
		ArtifactWorkspace: "",
	}
}

func IsRefPatternMatch(ref string, branchRulePatterns []string) bool {
	for _, branchRulePattern := range branchRulePatterns {
		if strings.HasSuffix(branchRulePattern, "*") {
			//通配符匹配
			rule := strings.TrimSuffix(branchRulePattern, "*")
			if strings.HasPrefix(ref, rule) {
				return true
			}
		} else {
			//完整路径匹配
			if ref == branchRulePattern {
				return true
			}
		}
	}
	return false
}
