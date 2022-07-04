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

import (
	"time"
)

type DevFlowRuleResponse struct {
	Header
	Data *DevFlowRule `json:"data"`
}

type DevFlowRule struct {
	ID             string         `json:"id"`
	Flows          []Flow         `json:"flows"`
	BranchPolicies []BranchPolicy `json:"branchPolicies"`
	OrgID          uint64         `json:"orgID"`
	OrgName        string         `json:"orgName"`
	ProjectID      uint64         `json:"projectID"`
	ProjectName    string         `json:"projectName"`
	TimeCreated    time.Time      `json:"timeCreated"`
	TimeUpdated    time.Time      `json:"timeUpdated"`
	Creator        string         `json:"creator"`
	Updater        string         `json:"updater"`
}

type Flow struct {
	Name         string `json:"name"`
	TargetBranch string `json:"targetBranch"`
	Artifact     string `json:"artifact"`
	Environment  string `json:"environment"`
}

type (
	BranchPolicy struct {
		Branch     string        `json:"branch"`
		BranchType string        `json:"branchType"`
		Policy     *PolicyDetail `json:"policy"`
	}
	PolicyDetail struct {
		SourceBranch  string        `json:"sourceBranch"`
		CurrentBranch string        `json:"currentBranch"`
		TempBranch    string        `json:"tempBranch"`
		TargetBranch  *TargetBranch `json:"targetBranch"`
	}
	TargetBranch struct {
		MergeRequest string `json:"mergeRequest"`
		CherryPick   string `json:"cherryPick"`
	}
)

func (f *DevFlowRule) MakeBranchRules() ([]*BranchRule, error) {
	if len(f.Flows) == 0 {
		return nil, nil
	}
	rules := make([]*BranchRule, 0, len(f.Flows))
	branchWorkspaceMap := make(map[string]string, 0)
	for _, v := range f.Flows {
		rules = append(rules, &BranchRule{
			ScopeType:         "project",
			ScopeID:           int64(f.ProjectID),
			Desc:              "",
			Rule:              v.TargetBranch,
			IsProtect:         false,
			IsTriggerPipeline: false,
			NeedApproval:      false,
			Workspace:         v.Environment,
		})
		branchWorkspaceMap[v.TargetBranch] = v.Environment
	}

	for _, policy := range f.BranchPolicies {
		if policy.Policy != nil && policy.Policy.TempBranch != "" {
			rules = append(rules, &BranchRule{
				ScopeType:         "project",
				ScopeID:           int64(f.ProjectID),
				Desc:              "",
				Rule:              policy.Policy.TempBranch,
				IsProtect:         false,
				IsTriggerPipeline: false,
				NeedApproval:      false,
				Workspace:         branchWorkspaceMap[policy.Branch],
			})
		}
	}
	return rules, nil
}
