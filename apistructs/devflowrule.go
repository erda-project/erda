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
	"encoding/json"
	"time"
)

type DevFlowRuleResponse struct {
	Header
	Data *DevFlowRule `json:"data"`
}

type DevFlowRule struct {
	ID          string    `json:"id"`
	Flows       string    `json:"flows"`
	OrgID       uint64    `json:"orgID"`
	OrgName     string    `json:"orgName"`
	ProjectID   uint64    `json:"projectID"`
	ProjectName string    `json:"projectName"`
	TimeCreated time.Time `json:"timeCreated"`
	TimeUpdated time.Time `json:"timeUpdated"`
	Creator     string    `json:"creator"`
	Updater     string    `json:"updater"`
}

type Flow struct {
	Name               string              `json:"name"`
	FlowType           string              `json:"flowType"`
	TargetBranch       string              `json:"targetBranch"`
	ChangeFromBranch   string              `json:"changeFromBranch"`
	ChangeBranch       string              `json:"changeBranch"`
	EnableAutoMerge    bool                `json:"enableAutoMerge"`
	AutoMergeBranch    string              `json:"autoMergeBranch"`
	Artifact           string              `json:"artifact"`
	Environment        string              `json:"environment"`
	StartWorkflowHints []StartWorkflowHint `json:"startWorkflowHints"`
}

type StartWorkflowHint struct {
	Place            string
	ChangeBranchRule string
}

func (f *DevFlowRule) MakeBranchRules() ([]*BranchRule, error) {
	if f.Flows == "" {
		return nil, nil
	}
	var flows []Flow
	err := json.Unmarshal([]byte(f.Flows), &flows)
	if err != nil {
		return nil, err
	}
	rules := make([]*BranchRule, 0, len(flows))
	for i := range flows {
		var rule string
		if flows[i].FlowType == "single_branch" {
			rule = flows[i].TargetBranch
		} else if flows[i].FlowType == "two_branch" {
			rule = flows[i].ChangeBranch
		} else if flows[i].FlowType == "three_branch" {
			rule = flows[i].ChangeBranch
		}
		rules = append(rules, &BranchRule{
			ScopeType:         "project",
			ScopeID:           int64(f.ProjectID),
			Desc:              "",
			Rule:              rule,
			IsProtect:         false,
			IsTriggerPipeline: false,
			NeedApproval:      false,
			Workspace:         flows[i].Environment,
			ArtifactWorkspace: "",
		})
	}
	return rules, nil
}
