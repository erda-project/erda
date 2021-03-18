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
