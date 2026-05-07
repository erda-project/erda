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
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ISSUELIST = command.Command{
	ParentName: "ISSUE",
	Name:       "list",
	ShortHelp:  "list current project issues",
	Example:    "$ erda-cli issue list --type bug --iteration -1 --assignee 1001055",
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "type", Doc: "issue types, comma-separated: requirement, task, bug", DefaultValue: ""},
		command.BoolFlag{Name: "requirement", Doc: "if true, include requirement issues", DefaultValue: false},
		command.BoolFlag{Name: "task", Doc: "if true, include task issues", DefaultValue: false},
		command.BoolFlag{Name: "bug", Doc: "if true, include bug issues", DefaultValue: false},
		command.IntFlag{Name: "iteration", Doc: "iteration id; -1 means no iteration", DefaultValue: 0},
		command.StringFlag{Name: "state", Doc: "state ids, comma-separated", DefaultValue: ""},
		command.StringFlag{Name: "assignee", Doc: "assignee user ids, comma-separated", DefaultValue: ""},
		command.StringFlag{Name: "creator", Doc: "creator user ids, comma-separated", DefaultValue: ""},
		command.StringFlag{Name: "owner", Doc: "owner user ids, comma-separated", DefaultValue: ""},
		command.IntFlag{Name: "page", Doc: "page number", DefaultValue: 1},
		command.IntFlag{Name: "page-size", Doc: "page size", DefaultValue: 20},
		command.BoolFlag{Name: "wide", Doc: "if true, show name with IDs", DefaultValue: false},
		command.BoolFlag{Name: "json", Doc: "if true, output JSON", DefaultValue: false},
	},
	Run: IssueList,
}

var (
	getIssueListOrgID               = common.GetOrgID
	getIssueListProjectID           = common.GetProjectID
	listIssueResponse               = common.ListMyIssueResponse
	getIssueListStates              = common.ListState
	getIssueListUsersByIDs          = queryIssueListUsersByIDs
	getIssueListIterationByID       = queryIssueListIterationByID
	issueListStdout       io.Writer = os.Stdout
)

type issueListOutput struct {
	Total uint64             `json:"total"`
	List  []apistructs.Issue `json:"list"`
}

func IssueList(ctx *command.Context, org string, project string, issueTypes string, requirement bool, task bool, bug bool, iteration int, states string, assignees string, creators string, owners string, page int, pageSize int, wide bool, jsonOutput bool) error {
	_, orgID, err := resolveIssueScope(ctx, org, project, getIssueListOrgID, getIssueListProjectID)
	if err != nil {
		return err
	}
	projectID := ctx.CurrentProject.ProjectID

	req, err := buildIssueListRequest(orgID, projectID, issueTypes, requirement, task, bug, iteration, states, assignees, creators, owners, page, pageSize)
	if err != nil {
		return err
	}
	resp, err := listIssueResponse(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil {
		resp = &apistructs.IssuePagingResponse{}
	}
	if resp.Data == nil {
		resp.Data = &apistructs.IssuePagingResponseData{}
	}

	if !jsonOutput {
		userNames, stateNames, iterationNames := resolveIssueListDisplayNames(ctx, orgID, resp.Data.List, resp.UserInfo)
		return writeIssueListTable(issueListStdout, resp.Data.List, resp.Data.Total, page, pageSize, userNames, stateNames, iterationNames, wide)
	}

	encoder := json.NewEncoder(issueListStdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(issueListOutput{Total: resp.Data.Total, List: resp.Data.List})
}

func writeIssueListTable(w io.Writer, list []apistructs.Issue, total uint64, page int, pageSize int, userNames map[string]string, stateNames map[int64]string, iterationNames map[int64]string, wide bool) error {
	_, _ = fmt.Fprintf(w, "issues (total=%d, page=%d, pageSize=%d)\n", total, page, pageSize)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tTITLE\tTYPE\tSTATE\tASSIGNEE\tITERATION")
	for _, item := range list {
		_, _ = fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Title,
			item.Type,
			formatIssueListState(item.State, stateNames, wide),
			formatIssueListAssignee(item.Assignee, userNames, wide),
			formatIssueListIteration(item.IterationID, iterationNames, wide),
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if len(list) == 0 {
		_, _ = fmt.Fprintln(w, "No issues found.")
	}
	return nil
}

func resolveIssueListDisplayNames(ctx *command.Context, orgID uint64, list []apistructs.Issue, userInfo map[string]apistructs.UserInfo) (map[string]string, map[int64]string, map[int64]string) {
	userIDs := make([]string, 0, len(list))
	userSeen := make(map[string]struct{}, len(list))
	iterationIDs := make([]int64, 0, len(list))
	iterationSeen := make(map[int64]struct{}, len(list))
	for _, item := range list {
		if item.Assignee != "" {
			if _, ok := userSeen[item.Assignee]; !ok {
				userSeen[item.Assignee] = struct{}{}
				userIDs = append(userIDs, item.Assignee)
			}
		}
		if item.IterationID > 0 {
			if _, ok := iterationSeen[item.IterationID]; !ok {
				iterationSeen[item.IterationID] = struct{}{}
				iterationIDs = append(iterationIDs, item.IterationID)
			}
		}
	}

	userNames := make(map[string]string)
	for id, info := range userInfo {
		name := strings.TrimSpace(info.Nick)
		if name == "" {
			name = strings.TrimSpace(info.Name)
		}
		if name != "" {
			userNames[id] = name
		}
	}
	if len(userIDs) > 0 {
		needLookup := false
		for _, id := range userIDs {
			if userNames[id] == "" {
				needLookup = true
				break
			}
		}
		if needLookup {
		if users, err := getIssueListUsersByIDs(ctx, userIDs); err == nil {
				for id, name := range users {
					if strings.TrimSpace(name) != "" {
						userNames[id] = name
					}
				}
			}
		}
	}

	iterationNames := make(map[int64]string)
	for _, id := range iterationIDs {
		iteration, err := getIssueListIterationByID(ctx, orgID, id)
		if err != nil || iteration == nil {
			continue
		}
		if iteration.Title != "" {
			iterationNames[id] = iteration.Title
		}
	}

	stateNames := make(map[int64]string)
	states, err := getIssueListStates(ctx, orgID, apistructs.IssueStateRelationGetRequest{
		ProjectID: ctx.CurrentProject.ProjectID,
	})
	if err == nil {
		for _, state := range states {
			if strings.TrimSpace(state.StateName) != "" {
				stateNames[state.StateID] = state.StateName
			}
		}
	}
	return userNames, stateNames, iterationNames
}

func resolveIssueScope(ctx *command.Context, org string, project string, getOrgID func(*command.Context, string) (string, uint64, error), getProjectID func(*command.Context, uint64, string) (string, uint64, error)) (string, uint64, error) {
	org = strings.TrimSpace(org)
	project = strings.TrimSpace(project)
	if project != "" && org == "" && ctx.CurrentOrg.ID == 0 && strings.TrimSpace(ctx.CurrentOrg.Name) == "" {
		return "", 0, fmt.Errorf("--project requires --org when no workspace context is available")
	}
	orgName := ctx.CurrentOrg.Name
	orgID := ctx.CurrentOrg.ID
	if org != "" || (orgID == 0 && strings.TrimSpace(orgName) == "") {
		var err error
		orgName, orgID, err = getOrgID(ctx, org)
		if err != nil {
			return "", 0, err
		}
	}

	projectName := ctx.CurrentProject.Project
	projectID := ctx.CurrentProject.ProjectID
	if project != "" || (projectID == 0 && strings.TrimSpace(projectName) == "") {
		var err error
		projectName, projectID, err = getProjectID(ctx, orgID, project)
		if err != nil {
			return "", 0, err
		}
	}
	ctx.CurrentOrg.Name = orgName
	ctx.CurrentOrg.ID = orgID
	ctx.CurrentProject.Project = projectName
	ctx.CurrentProject.ProjectID = projectID
	return orgName, orgID, nil
}

func queryIssueListUsersByIDs(ctx *command.Context, userIDs []string) (map[string]string, error) {
	var resp apistructs.UserListResponse
	values := url.Values{"userID": userIDs}
	_, err := ctx.Get().
		Path("/core/api/users").
		Param("plaintext", "true").
		Params(values).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("failed to query users: %s", resp.Error.Msg)
	}
	result := make(map[string]string, len(resp.Data.Users))
	for _, user := range resp.Data.Users {
		if user.ID == "" {
			continue
		}
		name := strings.TrimSpace(user.Nick)
		if name == "" {
			name = strings.TrimSpace(user.Name)
		}
		if name != "" {
			result[user.ID] = name
		}
	}
	return result, nil
}

func queryIssueListIterationByID(ctx *command.Context, orgID uint64, iterationID int64) (*apistructs.Iteration, error) {
	var resp apistructs.IterationGetResponse
	_, err := ctx.Get().
		Path(fmt.Sprintf("/api/iterations/%d", iterationID)).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("failed to query iteration %d: %s", iterationID, resp.Error.Msg)
	}
	return &resp.Data, nil
}

func formatIssueListAssignee(userID string, userNames map[string]string, wide bool) string {
	if userID == "" {
		return "-"
	}
	name := strings.TrimSpace(userNames[userID])
	if name == "" {
		return fmt.Sprintf("unknown(%s)", userID)
	}
	if !wide {
		return name
	}
	return fmt.Sprintf("%s(%s)", name, userID)
}

func formatIssueListIteration(iterationID int64, iterationNames map[int64]string, wide bool) string {
	if iterationID == -1 {
		if wide {
			return "UNASSIGNED(-1)"
		}
		return "UNASSIGNED"
	}
	if iterationID == 0 {
		return "-"
	}
	name := strings.TrimSpace(iterationNames[iterationID])
	if name == "" {
		return strconv.FormatInt(iterationID, 10)
	}
	if !wide {
		return name
	}
	return fmt.Sprintf("%s(%d)", name, iterationID)
}

func formatIssueListState(stateID int64, stateNames map[int64]string, wide bool) string {
	name := strings.TrimSpace(stateNames[stateID])
	if name == "" {
		return strconv.FormatInt(stateID, 10)
	}
	if !wide {
		return name
	}
	return fmt.Sprintf("%s(%d)", name, stateID)
}

func buildIssueListRequest(orgID, projectID uint64, issueTypes string, requirement bool, task bool, bug bool, iteration int, states string, assignees string, creators string, owners string, page int, pageSize int) (*apistructs.IssuePagingRequest, error) {
	if page <= 0 {
		return nil, fmt.Errorf("--page must be greater than 0")
	}
	if pageSize <= 0 {
		return nil, fmt.Errorf("--page-size must be greater than 0")
	}

	types, err := parseIssueListTypes(issueTypes, requirement, task, bug)
	if err != nil {
		return nil, err
	}
	stateIDs, err := parseIssueListInt64s(states, "--state")
	if err != nil {
		return nil, err
	}

	return &apistructs.IssuePagingRequest{
		PageNo:     uint64(page),
		PageSize:   uint64(pageSize),
		OrgID:      int64(orgID),
		ProjectIDs: []uint64{projectID},
		IssueListRequest: apistructs.IssueListRequest{
			ProjectID:   projectID,
			IterationID: int64(iteration),
			Type:        types,
			State:       stateIDs,
			Assignees:   splitIssueListCSV(assignees),
			Creators:    splitIssueListCSV(creators),
			Owner:       splitIssueListCSV(owners),
		},
	}, nil
}

func parseIssueListTypes(raw string, requirement bool, task bool, bug bool) ([]apistructs.IssueType, error) {
	if raw == "" {
		var types []apistructs.IssueType
		if requirement {
			types = append(types, apistructs.IssueTypeRequirement)
		}
		if task {
			types = append(types, apistructs.IssueTypeTask)
		}
		if bug {
			types = append(types, apistructs.IssueTypeBug)
		}
		return types, nil
	}
	parts := splitIssueListCSV(raw)
	types := make([]apistructs.IssueType, 0, len(parts))
	for _, part := range parts {
		parsed, err := parseIssueSchemaTypes(part)
		if err != nil {
			return nil, err
		}
		types = append(types, parsed[0])
	}
	return types, nil
}

func parseIssueListInt64s(raw string, flagName string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}
	parts := splitIssueListCSV(raw)
	values := make([]int64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s value %q", flagName, part)
		}
		values = append(values, value)
	}
	return values, nil
}

func splitIssueListCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
