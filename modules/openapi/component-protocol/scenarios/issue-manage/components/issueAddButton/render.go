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

package issueAddButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type AddButtonCandidateOp struct {
	Click struct {
		Reload bool   `json:"reload"`
		Key    string `json:"key"`
	} `json:"click"`
}
type AddButtonCandidate struct {
	Disabled    bool                 `json:"disabled"`
	DisabledTip string               `json:"disabledTip"`
	Key         string               `json:"key"`
	Operations  AddButtonCandidateOp `json:"operations"`
	PrefixIcon  string               `json:"prefixIcon"`
	Text        string               `json:"text"`
}
type Props struct {
	Menu       []AddButtonCandidate `json:"menu"`
	SuffixIcon string               `json:"suffixIcon"`
	Text       string               `json:"text"`
	Type       string               `json:"type"`
	Operations AddButtonCandidateOp `json:"operations"`
	Disabled   bool                 `json:"disabled"`
}

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	ctxBdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	fixedIssueType := ctxBdl.InParams["fixedIssueType"].(string)

	isGuest, err := ca.CheckUserPermission(ctxBdl)
	if err != nil {
		return err
	}

	requirementCandidate := AddButtonCandidate{
		Disabled:    isGuest,
		DisabledTip: "",
		Key:         "requirement",
		Operations: AddButtonCandidateOp{Click: struct {
			Reload bool   `json:"reload"`
			Key    string `json:"key"`
		}{Reload: false, Key: "createRequirement"}},
		PrefixIcon: "ISSUE_ICON.issue.REQUIREMENT",
		Text:       "需求",
	}
	taskCandidate := AddButtonCandidate{
		Disabled:    isGuest,
		DisabledTip: "",
		Key:         "task",
		Operations: AddButtonCandidateOp{Click: struct {
			Reload bool   `json:"reload"`
			Key    string `json:"key"`
		}{Reload: false, Key: "createTask"}},
		PrefixIcon: "ISSUE_ICON.issue.TASK",
		Text:       "任务",
	}
	bugCandidate := AddButtonCandidate{
		Disabled:    isGuest,
		DisabledTip: "",
		Key:         "bug",
		Operations: AddButtonCandidateOp{Click: struct {
			Reload bool   `json:"reload"`
			Key    string `json:"key"`
		}{Reload: false, Key: "createBug"}},
		PrefixIcon: "ISSUE_ICON.issue.BUG",
		Text:       "缺陷",
	}
	props := Props{
		Menu:       nil,
		SuffixIcon: "di",
		Text:       "新建事项",
		Type:       "primary",
		Disabled:   isGuest,
	}

	prop := Props{
		Type: "primary",
	}

	var menu []AddButtonCandidate

	switch fixedIssueType {
	case "ALL":
		menu = []AddButtonCandidate{requirementCandidate, taskCandidate, bugCandidate}
		props.Menu = menu
		c.Props = props
	case apistructs.IssueTypeRequirement.String():
		prop.Text = "新建需求"
		c.Operations = make(apistructs.ComponentOps)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createRequirement", Disabled: isGuest}
		c.Props = prop
	case apistructs.IssueTypeTask.String():
		prop.Text = "新建任务"
		c.Operations = make(apistructs.ComponentOps)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createTask", Disabled: isGuest}
		c.Props = prop
	case apistructs.IssueTypeBug.String():
		prop.Text = "新建缺陷"
		c.Operations = make(apistructs.ComponentOps)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createBug", Disabled: isGuest}
		c.Props = prop
	}

	return nil
}

// GetUserPermission  check Guest permission
func (ca *ComponentAction) CheckUserPermission(bdl protocol.ContextBundle) (bool, error) {
	isGuest := false
	projectID := bdl.InParams["projectId"].(string)
	scopeRole, err := bdl.Bdl.ScopeRoleAccess(bdl.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   projectID,
		},
	})
	if err != nil {
		return false, err
	}
	for _, role := range scopeRole.Roles {
		if role == "Guest" {
			isGuest = true
		}
	}
	return isGuest, nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
