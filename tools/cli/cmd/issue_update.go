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

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ISSUEUPDATE = command.Command{
	Name:       "update",
	ParentName: "ISSUE",
	ShortHelp:  "update issue fields such as state",
	Example:    "$ erda-cli issue update <issue-id> --state 进行中",
	Args: []command.Arg{
		command.IntArg{}.Name("issue-id"),
	},
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "state", Doc: "target state name", DefaultValue: ""},
		command.IntFlag{Name: "state-id", Doc: "target state ID", DefaultValue: 0},
	},
	ValidArgsFunction: UpdateIssueCompletion,
	Run:               IssueUpdate,
}

var (
	getIssueUpdateIssue  = common.GetIssue
	getIssueUpdateStates = common.ListState
	updateIssueState     = common.UpdateIssue
)

func UpdateIssueCompletion(ctx *cobra.Command, args []string, toComplete string, issueID int, org string, project string, state string, stateID int) []string {
	return IssueCompletion(ctx, args, toComplete, issueID)
}

func IssueUpdate(ctx *command.Context, issueID int, org string, project string, state string, stateID int) error {
	if issueID <= 0 {
		return fmt.Errorf("issue-id must be greater than 0")
	}
	if (state == "") == (stateID == 0) {
		return fmt.Errorf("exactly one of --state or --state-id is required")
	}

	_, orgID, err := resolveIssueScope(ctx, org, project, common.GetOrgID, common.GetProjectID)
	if err != nil {
		return err
	}
	projectID := ctx.CurrentProject.ProjectID

	issue, err := getIssueUpdateIssue(ctx, orgID, projectID, uint64(issueID))
	if err != nil {
		return err
	}
	stateMap, err := loadIssueStateMap(ctx, projectID, issue.Type)
	if err != nil {
		return err
	}

	current, ok := stateMap[issue.State]
	if !ok {
		return fmt.Errorf("current issue state %d not found in project workflow", issue.State)
	}

	var target *apistructs.IssueStateRelation
	if stateID > 0 {
		target = stateMap[int64(stateID)]
		if target == nil {
			return fmt.Errorf("target state-id %d not found in project workflow", stateID)
		}
	} else {
		target = findTransitionStateByName(current, stateMap, state)
		if target == nil {
			return buildInvalidTransitionError(current, stateMap, fmt.Sprintf("state %q", state))
		}
	}

	if !isReachableState(current, target.StateID) {
		return buildInvalidTransitionError(current, stateMap, fmt.Sprintf("state %q (%d)", target.StateName, target.StateID))
	}

	req := &apistructs.IssueUpdateRequest{
		ID:    uint64(issue.ID),
		State: &target.StateID,
	}
	if err := updateIssueState(ctx, orgID, req); err != nil {
		return err
	}

	ctx.Succ("Issue %d state updated: %s(%d) -> %s(%d)", issue.ID, current.StateName, current.StateID, target.StateName, target.StateID)
	return nil
}

func loadIssueStateMap(ctx *command.Context, projectID uint64, issueType apistructs.IssueType) (map[int64]*apistructs.IssueStateRelation, error) {
	states, err := getIssueUpdateStates(ctx, ctx.CurrentOrg.ID, apistructs.IssueStateRelationGetRequest{
		ProjectID: projectID,
		IssueType: issueType,
	})
	if err != nil {
		return nil, err
	}

	stateMap := make(map[int64]*apistructs.IssueStateRelation, len(states))
	for i := range states {
		stateMap[states[i].StateID] = &states[i]
	}
	return stateMap, nil
}

func isReachableState(current *apistructs.IssueStateRelation, targetStateID int64) bool {
	if current == nil {
		return false
	}
	for _, nextID := range current.StateRelation {
		if nextID == targetStateID {
			return true
		}
	}
	return false
}

func findTransitionStateByName(current *apistructs.IssueStateRelation, stateMap map[int64]*apistructs.IssueStateRelation, wanted string) *apistructs.IssueStateRelation {
	name := strings.ToLower(strings.TrimSpace(wanted))
	for _, nextID := range current.StateRelation {
		next := stateMap[nextID]
		if next == nil {
			continue
		}
		if strings.ToLower(next.StateName) == name {
			return next
		}
	}
	return nil
}

func buildInvalidTransitionError(current *apistructs.IssueStateRelation, stateMap map[int64]*apistructs.IssueStateRelation, target string) error {
	allowed := make([]string, 0, len(current.StateRelation))
	for _, nextID := range current.StateRelation {
		if next, ok := stateMap[nextID]; ok {
			allowed = append(allowed, fmt.Sprintf("%s(%d)", next.StateName, next.StateID))
		}
	}
	return fmt.Errorf(
		"invalid transition from %s(%d) to %s, allowed targets: %s",
		current.StateName,
		current.StateID,
		target,
		strings.Join(allowed, ", "),
	)
}

func findTransitionByStateBelong(current *apistructs.IssueStateRelation, stateMap map[int64]*apistructs.IssueStateRelation, belong apistructs.IssueStateBelong) *apistructs.IssueStateRelation {
	for _, nextID := range current.StateRelation {
		next := stateMap[nextID]
		if next != nil && next.StateBelong == belong {
			return next
		}
	}
	return nil
}

func formatStateDebug(state *apistructs.IssueStateRelation) string {
	return state.StateName + "(" + strconv.FormatInt(state.StateID, 10) + ")"
}
