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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
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

type ComponentAction struct{ base.DefaultProvider }

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	fixedIssueType := sdk.InParams["fixedIssueType"].(string)

	isGuest, err := ca.CheckUserPermission(ctx)
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
		Text:       cputil.I18n(ctx, "requirement"),
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
		Text:       cputil.I18n(ctx, "task"),
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
		Text:       cputil.I18n(ctx, "bug"),
	}
	props := Props{
		Menu:       nil,
		SuffixIcon: "di",
		Text:       cputil.I18n(ctx, "create-issue"),
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
		c.Props = cputil.MustConvertProps(props)
	case apistructs.IssueTypeRequirement.String():
		prop.Text = cputil.I18n(ctx, "create-requirement")
		c.Operations = make(cptype.ComponentOperations)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createRequirement", Disabled: isGuest}
	case apistructs.IssueTypeTask.String():
		prop.Text = cputil.I18n(ctx, "create-task")
		c.Operations = make(cptype.ComponentOperations)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createTask", Disabled: isGuest}
	case apistructs.IssueTypeBug.String():
		prop.Text = cputil.I18n(ctx, "create-bug")
		c.Operations = make(cptype.ComponentOperations)
		c.Operations["click"] = struct {
			Reload   bool   `json:"reload"`
			Key      string `json:"key"`
			Disabled bool   `json:"disabled"`
		}{Reload: false, Key: "createBug", Disabled: isGuest}
	}
	c.Props = cputil.MustConvertProps(prop)

	return nil
}

// GetUserPermission  check Guest permission
func (ca *ComponentAction) CheckUserPermission(ctx context.Context) (bool, error) {
	sdk := cputil.SDK(ctx)
	isGuest := false
	projectID := cputil.GetInParamByKey(ctx, "projectId").(string)
	scopeRole, err := bdl.Bdl.ScopeRoleAccess(sdk.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
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

func init() {
	base.InitProviderWithCreator("issue-manage", "issueAddButton",
		func() servicehub.Provider { return &ComponentAction{} },
	)
}
