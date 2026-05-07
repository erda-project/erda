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
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestIssueListOutputsTableByDefault(t *testing.T) {
	restoreIssueListStubs(t)

	getIssueListOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueListProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}

	var requested *apistructs.IssuePagingRequest
	listIssueResponse = func(_ *command.Context, req *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
		copied := *req
		requested = &copied
		return &apistructs.IssuePagingResponse{
			Data: &apistructs.IssuePagingResponseData{
				Total: 1,
				List: []apistructs.Issue{
					{ID: 3001, ProjectID: 2001, IterationID: -1, Type: apistructs.IssueTypeBug, Title: "login failed", State: 1, Assignee: "1001055"},
				},
			},
			UserInfoHeader: apistructs.UserInfoHeader{
				UserInfo: map[string]apistructs.UserInfo{
					"1001055": {ID: "1001055", Nick: "ash"},
				},
			},
		}, nil
	}
	getIssueListStates = func(*command.Context, uint64, apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		return []apistructs.IssueStateRelation{
			{IssueStatus: apistructs.IssueStatus{StateID: 0, StateName: "Unknown"}},
			{IssueStatus: apistructs.IssueStatus{StateID: 1, StateName: "Open"}},
		}, nil
	}
	getIssueListUsersByIDs = func(*command.Context, []string) (map[string]string, error) { return nil, nil }
	getIssueListIterationByID = func(*command.Context, uint64, int64) (*apistructs.Iteration, error) {
		return &apistructs.Iteration{Title: "Sprint 1"}, nil
	}

	var out bytes.Buffer
	issueListStdout = &out

	if err := IssueList(&command.Context{}, "", "", "bug", false, false, false, -1, "", "1001055", "", "", 1, 20, false, false); err != nil {
		t.Fatalf("IssueList() error = %v", err)
	}
	if requested == nil {
		t.Fatal("ListIssues was not called")
	}
	if requested.ProjectID != 2001 || requested.PageNo != 1 || requested.PageSize != 20 {
		t.Fatalf("request = %+v, want project/page populated", requested)
	}
	if len(requested.Type) != 1 || requested.Type[0] != apistructs.IssueTypeBug {
		t.Fatalf("request types = %+v, want BUG", requested.Type)
	}
	if requested.IterationID != -1 {
		t.Fatalf("iterationID = %d, want -1", requested.IterationID)
	}
	if len(requested.Assignees) != 1 || requested.Assignees[0] != "1001055" {
		t.Fatalf("assignees = %+v, want 1001055", requested.Assignees)
	}

	got := out.String()
	for _, expected := range []string{"issues (total=1, page=1, pageSize=20)", "ID", "login failed", "BUG", "Open", "ash", "UNASSIGNED"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("table output missing %q:\n%s", expected, got)
		}
	}
}

func TestIssueListOutputsJSONWithFlag(t *testing.T) {
	restoreIssueListStubs(t)

	getIssueListOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueListProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueResponse = func(_ *command.Context, _ *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
		return &apistructs.IssuePagingResponse{
			Data: &apistructs.IssuePagingResponseData{
				Total: 1,
				List: []apistructs.Issue{
					{ID: 3001, ProjectID: 2001, IterationID: -1, Type: apistructs.IssueTypeBug, Title: "login failed", State: 1, Assignee: "1001055"},
				},
			},
		}, nil
	}
	getIssueListStates = func(*command.Context, uint64, apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		t.Fatal("state lookup should not happen in json mode")
		return nil, nil
	}
	getIssueListUsersByIDs = func(*command.Context, []string) (map[string]string, error) {
		t.Fatal("user lookup should not happen in json mode")
		return nil, nil
	}
	getIssueListIterationByID = func(*command.Context, uint64, int64) (*apistructs.Iteration, error) {
		t.Fatal("iteration lookup should not happen in json mode")
		return nil, nil
	}

	var out bytes.Buffer
	issueListStdout = &out

	if err := IssueList(&command.Context{}, "", "", "bug", false, false, false, -1, "", "1001055", "", "", 1, 20, false, true); err != nil {
		t.Fatalf("IssueList() error = %v", err)
	}

	var output issueListOutput
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if output.Total != 1 || len(output.List) != 1 || output.List[0].Title != "login failed" {
		t.Fatalf("output = %+v, want one issue", output)
	}
}

func TestIssueListShowsUnknownAssigneeWhenNameLookupMisses(t *testing.T) {
	restoreIssueListStubs(t)

	getIssueListOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueListProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueResponse = func(_ *command.Context, _ *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
		return &apistructs.IssuePagingResponse{
			Data: &apistructs.IssuePagingResponseData{
				Total: 1,
				List: []apistructs.Issue{
					{ID: 3001, ProjectID: 2001, IterationID: -1, Type: apistructs.IssueTypeBug, Title: "login failed", State: 1, Assignee: "1001055"},
				},
			},
		}, nil
	}
	getIssueListStates = func(*command.Context, uint64, apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		return nil, nil
	}
	getIssueListUsersByIDs = func(*command.Context, []string) (map[string]string, error) {
		return map[string]string{}, nil
	}
	getIssueListIterationByID = func(*command.Context, uint64, int64) (*apistructs.Iteration, error) {
		return nil, nil
	}

	var out bytes.Buffer
	issueListStdout = &out

	if err := IssueList(&command.Context{}, "", "", "bug", false, false, false, -1, "", "1001055", "", "", 1, 20, false, false); err != nil {
		t.Fatalf("IssueList() error = %v", err)
	}
	if !strings.Contains(out.String(), "unknown(1001055)") {
		t.Fatalf("table output should contain unknown assignee marker:\n%s", out.String())
	}
}

func TestIssueListWideShowsNameWithIDs(t *testing.T) {
	restoreIssueListStubs(t)

	getIssueListOrgID = func(*command.Context, string) (string, uint64, error) { return "erda", 1001, nil }
	getIssueListProjectID = func(*command.Context, uint64, string) (string, uint64, error) { return "erda-project", 2001, nil }
	listIssueResponse = func(_ *command.Context, _ *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
		return &apistructs.IssuePagingResponse{
			Data: &apistructs.IssuePagingResponseData{
				Total: 1,
				List:  []apistructs.Issue{{ID: 3001, ProjectID: 2001, IterationID: 3, Type: apistructs.IssueTypeBug, Title: "login failed", State: 1, Assignee: "1001055"}},
			},
		}, nil
	}
	getIssueListStates = func(*command.Context, uint64, apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		return []apistructs.IssueStateRelation{{IssueStatus: apistructs.IssueStatus{StateID: 1, StateName: "Open"}}}, nil
	}
	getIssueListUsersByIDs = func(*command.Context, []string) (map[string]string, error) {
		return map[string]string{"1001055": "ash"}, nil
	}
	getIssueListIterationByID = func(*command.Context, uint64, int64) (*apistructs.Iteration, error) {
		return &apistructs.Iteration{Title: "Sprint 3"}, nil
	}

	var out bytes.Buffer
	issueListStdout = &out
	if err := IssueList(&command.Context{}, "", "", "bug", false, false, false, 3, "", "1001055", "", "", 1, 20, true, false); err != nil {
		t.Fatalf("IssueList() error = %v", err)
	}
	got := out.String()
	for _, expected := range []string{"Open(1)", "ash(1001055)", "Sprint 3(3)"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("wide output missing %q:\n%s", expected, got)
		}
	}
}

func TestIssueListCommandShape(t *testing.T) {
	if ISSUELIST.ParentName != "ISSUE" {
		t.Fatalf("issue list parent = %q, want ISSUE", ISSUELIST.ParentName)
	}
	if ISSUELIST.Name != "list" {
		t.Fatalf("issue list name = %q, want list", ISSUELIST.Name)
	}
	for _, flag := range []string{"type", "requirement", "task", "bug", "iteration", "state", "assignee", "creator", "owner", "page", "page-size", "wide", "json"} {
		if !hasCommandFlag(ISSUELIST.Flags, flag) {
			t.Fatalf("issue list missing --%s", flag)
		}
	}
}

func TestParseIssueListTypesSupportsLegacyBooleanFilters(t *testing.T) {
	types, err := parseIssueListTypes("", true, true, false)
	if err != nil {
		t.Fatalf("parseIssueListTypes() error = %v", err)
	}
	if len(types) != 2 || types[0] != apistructs.IssueTypeRequirement || types[1] != apistructs.IssueTypeTask {
		t.Fatalf("types = %+v, want requirement/task", types)
	}
}

func restoreIssueListStubs(t *testing.T) {
	origGetOrgID := getIssueListOrgID
	origGetProjectID := getIssueListProjectID
	origListIssues := listIssueResponse
	origGetStates := getIssueListStates
	origGetUsers := getIssueListUsersByIDs
	origGetIteration := getIssueListIterationByID
	origStdout := issueListStdout
	t.Cleanup(func() {
		getIssueListOrgID = origGetOrgID
		getIssueListProjectID = origGetProjectID
		listIssueResponse = origListIssues
		getIssueListStates = origGetStates
		getIssueListUsersByIDs = origGetUsers
		getIssueListIterationByID = origGetIteration
		issueListStdout = origStdout
	})
}
