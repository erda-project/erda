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
	"os"
	"path/filepath"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

func TestIssueCreateFromFileCreatesIssueAndCustomFields(t *testing.T) {
	restoreIssueCreateStubs(t)

	getIssueCreateOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueCreateProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueCreateProperties = func(_ *command.Context, _ uint64, _ uint64, issueType apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		if issueType != apistructs.IssueTypeBug {
			t.Fatalf("issueType = %s, want BUG", issueType)
		}
		return []apistructs.IssuePropertyIndex{
			{
				PropertyID:        41,
				ScopeID:           2001,
				ScopeType:         apistructs.ProjectScope,
				OrgID:             1001,
				PropertyName:      "rootCause",
				DisplayName:       "Root Cause",
				PropertyType:      apistructs.PropertyTypeSelect,
				Required:          true,
				PropertyIssueType: apistructs.PropertyIssueTypeBug,
				EnumeratedValues:  []apistructs.Enumerate{{ID: 7, Name: "code"}},
			},
		}, nil
	}
	listIssueCreateStages = func(*command.Context, uint64, apistructs.IssueType) ([]apistructs.IssueStage, error) {
		return []apistructs.IssueStage{{Name: "开发", Value: "dev"}}, nil
	}

	var created apistructs.IssueCreateRequest
	createIssue = func(_ *command.Context, _ uint64, req *apistructs.IssueCreateRequest) (uint64, error) {
		created = *req
		return 3001, nil
	}
	var propertyReq common.IssuePropertyInstanceCreateRequest
	createIssuePropertyInstance = func(_ *command.Context, req *common.IssuePropertyInstanceCreateRequest) error {
		propertyReq = *req
		return nil
	}
	var relationReq *apistructs.IssueRelationCreateRequest
	createIssueRelation = func(_ *command.Context, _ uint64, parentID uint64, req *apistructs.IssueRelationCreateRequest) error {
		relationReq = req
		if parentID != 18 {
			t.Fatalf("parentID = %d, want 18", parentID)
		}
		return nil
	}

	path := writeIssueCreateFile(t, `{
		"type": "BUG",
		"title": "login failed",
		"content": "reproduce steps",
		"assignee": "1001055",
		"iterationID": -1,
		"priority": "HIGH",
		"severity": "SERIOUS",
		"parentID": 18,
		"relationType": "inclusion",
		"labels": ["cli"],
		"customFields": {
			"rootCause": "code"
		}
	}`)

	var out bytes.Buffer
	issueCreateStdout = &out

	if err := IssueCreate(&command.Context{}, "", "", path, ""); err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}

	if created.ProjectID != 2001 || created.Type != apistructs.IssueTypeBug || created.Title != "login failed" {
		t.Fatalf("created request = %+v, want project/type/title populated", created)
	}
	if created.IterationID != -1 || created.Assignee != "1001055" || created.Priority != apistructs.IssuePriorityHigh {
		t.Fatalf("created request = %+v, want iteration/assignee/priority populated", created)
	}
	if propertyReq.IssueID != 3001 || len(propertyReq.Property) != 1 {
		t.Fatalf("property request = %+v, want one property for issue 3001", propertyReq)
	}
	property := propertyReq.Property[0]
	if property.PropertyID != 41 || len(property.Values) != 1 || property.Values[0] != 7 {
		t.Fatalf("property = %+v, want rootCause value id 7", property)
	}
	if relationReq == nil || relationReq.Type != apistructs.IssueRelationInclusion || len(relationReq.RelatedIssue) != 1 || relationReq.RelatedIssue[0] != 3001 {
		t.Fatalf("relation request = %+v, want parent inclusion relation to issue 3001", relationReq)
	}

	var resp issueCreateOutput
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if resp.ID != 3001 {
		t.Fatalf("output id = %d, want 3001", resp.ID)
	}
}

func TestIssueCreateFromInlineJSON(t *testing.T) {
	restoreIssueCreateStubs(t)

	getIssueCreateOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueCreateProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueCreateProperties = func(*command.Context, uint64, uint64, apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		return nil, nil
	}
	listIssueCreateStages = func(*command.Context, uint64, apistructs.IssueType) ([]apistructs.IssueStage, error) {
		return []apistructs.IssueStage{{Name: "开发", Value: "dev"}}, nil
	}

	var created apistructs.IssueCreateRequest
	createIssue = func(_ *command.Context, _ uint64, req *apistructs.IssueCreateRequest) (uint64, error) {
		created = *req
		return 3002, nil
	}

	var out bytes.Buffer
	issueCreateStdout = &out

	inlineJSON := `{"type":"bug","title":"inline issue","assignee":"1001055","iterationID":-1}`
	if err := IssueCreate(&command.Context{}, "", "", "", inlineJSON); err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}
	if created.ProjectID != 2001 || created.Type != apistructs.IssueTypeBug || created.Title != "inline issue" {
		t.Fatalf("created request = %+v, want inline JSON request", created)
	}

	var resp issueCreateOutput
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if resp.ID != 3002 {
		t.Fatalf("output id = %d, want 3002", resp.ID)
	}
}

func TestIssueCreateConvertsTaskTypeFromStageName(t *testing.T) {
	restoreIssueCreateStubs(t)

	getIssueCreateOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueCreateProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueCreateProperties = func(*command.Context, uint64, uint64, apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		return nil, nil
	}
	listIssueCreateStages = func(*command.Context, uint64, apistructs.IssueType) ([]apistructs.IssueStage, error) {
		return []apistructs.IssueStage{{Name: "开发", Value: "dev"}}, nil
	}

	var created apistructs.IssueCreateRequest
	createIssue = func(_ *command.Context, _ uint64, req *apistructs.IssueCreateRequest) (uint64, error) {
		created = *req
		return 3003, nil
	}

	inlineJSON := `{"type":"TASK","title":"task issue","assignee":"1001055","iterationID":-1,"taskType":"开发"}`
	if err := IssueCreate(&command.Context{}, "", "", "", inlineJSON); err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}
	if created.TaskType != "dev" {
		t.Fatalf("created taskType = %q, want dev", created.TaskType)
	}
}

func TestIssueCreateRejectsUnknownTaskType(t *testing.T) {
	restoreIssueCreateStubs(t)

	getIssueCreateOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueCreateProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueCreateProperties = func(*command.Context, uint64, uint64, apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		return nil, nil
	}
	listIssueCreateStages = func(*command.Context, uint64, apistructs.IssueType) ([]apistructs.IssueStage, error) {
		return []apistructs.IssueStage{{Name: "开发", Value: "dev"}}, nil
	}
	createIssue = func(*command.Context, uint64, *apistructs.IssueCreateRequest) (uint64, error) {
		t.Fatal("CreateIssue should not be called when taskType is unknown")
		return 0, nil
	}

	inlineJSON := `{"type":"TASK","title":"task issue","assignee":"1001055","iterationID":-1,"taskType":"development"}`
	if err := IssueCreate(&command.Context{}, "", "", "", inlineJSON); err == nil {
		t.Fatal("IssueCreate() error = nil, want unknown taskType error")
	}
}

func TestIssueCreateRejectsNonNumericUserIDs(t *testing.T) {
	restoreIssueCreateStubs(t)

	createIssue = func(*command.Context, uint64, *apistructs.IssueCreateRequest) (uint64, error) {
		t.Fatal("CreateIssue should not be called when assignee is not a user ID")
		return 0, nil
	}

	inlineJSON := `{"type":"TASK","title":"task issue","assignee":"ash","iterationID":-1}`
	if err := IssueCreate(&command.Context{}, "", "", "", inlineJSON); err == nil {
		t.Fatal("IssueCreate() error = nil, want non-numeric assignee error")
	}
}

func TestIssueCreateRejectsNonNumericPersonCustomField(t *testing.T) {
	property := apistructs.IssuePropertyIndex{
		PropertyID:        41,
		ScopeID:           2001,
		ScopeType:         apistructs.ProjectScope,
		OrgID:             1001,
		PropertyName:      "reviewer",
		DisplayName:       "Reviewer",
		PropertyType:      apistructs.PropertyTypePerson,
		PropertyIssueType: apistructs.PropertyIssueTypeTask,
	}

	_, err := buildIssuePropertyInstance(1001, 2001, property, json.RawMessage(`"ash"`))
	if err == nil {
		t.Fatal("buildIssuePropertyInstance() error = nil, want non-numeric person custom field error")
	}
}

func TestIssueCreateRejectsAmbiguousInput(t *testing.T) {
	path := writeIssueCreateFile(t, `{"type":"BUG","title":"x","assignee":"1001055","iterationID":-1}`)
	if err := IssueCreate(&command.Context{}, "", "", path, `{"type":"BUG"}`); err == nil {
		t.Fatal("IssueCreate() error = nil, want file/json mutual exclusion error")
	}
	if err := IssueCreate(&command.Context{}, "", "", "", ""); err == nil {
		t.Fatal("IssueCreate() error = nil, want missing input error")
	}
}

func TestIssueCreateRejectsMissingRequiredCustomField(t *testing.T) {
	restoreIssueCreateStubs(t)

	getIssueCreateOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueCreateProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueCreateProperties = func(*command.Context, uint64, uint64, apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		return []apistructs.IssuePropertyIndex{
			{PropertyName: "rootCause", DisplayName: "Root Cause", PropertyType: apistructs.PropertyTypeText, Required: true},
		}, nil
	}
	listIssueCreateStages = func(*command.Context, uint64, apistructs.IssueType) ([]apistructs.IssueStage, error) {
		return nil, nil
	}
	createIssue = func(*command.Context, uint64, *apistructs.IssueCreateRequest) (uint64, error) {
		t.Fatal("CreateIssue should not be called when required custom field is missing")
		return 0, nil
	}

	path := writeIssueCreateFile(t, `{
		"type": "BUG",
		"title": "login failed",
		"assignee": "1001055",
		"iterationID": -1
	}`)

	if err := IssueCreate(&command.Context{}, "", "", path, ""); err == nil {
		t.Fatal("IssueCreate() error = nil, want missing required custom field error")
	}
}

func TestIssueCreateCommandShape(t *testing.T) {
	if ISSUECREATE.ParentName != "ISSUE" {
		t.Fatalf("issue create parent = %q, want ISSUE", ISSUECREATE.ParentName)
	}
	if ISSUECREATE.Name != "create" {
		t.Fatalf("issue create name = %q, want create", ISSUECREATE.Name)
	}
	if !hasCommandFlag(ISSUECREATE.Flags, "file") {
		t.Fatal("issue create missing --file")
	}
	if !hasCommandFlag(ISSUECREATE.Flags, "json") {
		t.Fatal("issue create missing --json")
	}
}

func writeIssueCreateFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "issue.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func restoreIssueCreateStubs(t *testing.T) {
	origGetOrgID := getIssueCreateOrgID
	origGetProjectID := getIssueCreateProjectID
	origListProperties := listIssueCreateProperties
	origListStages := listIssueCreateStages
	origCreateIssue := createIssue
	origCreateIssuePropertyInstance := createIssuePropertyInstance
	origCreateIssueRelation := createIssueRelation
	origStdout := issueCreateStdout
	t.Cleanup(func() {
		getIssueCreateOrgID = origGetOrgID
		getIssueCreateProjectID = origGetProjectID
		listIssueCreateProperties = origListProperties
		listIssueCreateStages = origListStages
		createIssue = origCreateIssue
		createIssuePropertyInstance = origCreateIssuePropertyInstance
		createIssueRelation = origCreateIssueRelation
		issueCreateStdout = origStdout
	})
}
