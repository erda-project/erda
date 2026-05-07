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
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestIssueSchemaOutputsProjectIssueMetadataAsJSON(t *testing.T) {
	restoreIssueSchemaStubs(t)

	getIssueSchemaOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueSchemaProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listIssueSchemaStates = func(_ *command.Context, _ uint64, req apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		if req.IssueType != apistructs.IssueTypeBug {
			return nil, nil
		}
		return []apistructs.IssueStateRelation{
			{
				IssueStatus: apistructs.IssueStatus{
					StateID:     11,
					StateName:   "Open",
					StateBelong: apistructs.IssueStateBelongOpen,
					IssueType:   apistructs.IssueTypeBug,
				},
				StateRelation: []int64{12},
			},
			{
				IssueStatus: apistructs.IssueStatus{
					StateID:     12,
					StateName:   "Resolved",
					StateBelong: apistructs.IssueStateBelongResolved,
					IssueType:   apistructs.IssueTypeBug,
				},
			},
		}, nil
	}
	listIssueSchemaLabels = func(*command.Context, uint64) ([]apistructs.ProjectLabel, error) {
		return []apistructs.ProjectLabel{
			{ID: 31, Name: "cli", Color: "green", Type: apistructs.LabelTypeIssue, ProjectID: 2001, CreatedAt: time.Unix(0, 0), UpdatedAt: time.Unix(0, 0)},
		}, nil
	}
	listIssueSchemaProperties = func(_ *command.Context, _ uint64, _ uint64, issueType apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
		if issueType != apistructs.IssueTypeBug {
			return nil, nil
		}
		return []apistructs.IssuePropertyIndex{
			{
				PropertyID:        41,
				PropertyName:      "rootCause",
				DisplayName:       "Root Cause",
				PropertyType:      apistructs.PropertyTypeSelect,
				Required:          true,
				PropertyIssueType: apistructs.PropertyIssueTypeBug,
				EnumeratedValues:  []apistructs.Enumerate{{ID: 1, Name: "code", Index: 1}},
			},
		}, nil
	}
	listIssueSchemaStages = func(_ *command.Context, _ uint64, issueType apistructs.IssueType) ([]apistructs.IssueStage, error) {
		if issueType != apistructs.IssueTypeBug {
			return nil, nil
		}
		return []apistructs.IssueStage{{Name: "代码开发", Value: "dev"}}, nil
	}

	var out bytes.Buffer
	issueSchemaStdout = &out

	if err := IssueSchema(&command.Context{}, "", "", "bug"); err != nil {
		t.Fatalf("IssueSchema() error = %v", err)
	}

	var schema issueSchemaDocument
	if err := json.Unmarshal(out.Bytes(), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v\n%s", err, out.String())
	}
	if schema.Project.ID != 2001 || schema.Project.Name != "erda-project" {
		t.Fatalf("project = %+v, want id=2001 name=erda-project", schema.Project)
	}
	if len(schema.IssueTypes) != 1 || schema.IssueTypes[0].Type != string(apistructs.IssueTypeBug) {
		t.Fatalf("issueTypes = %+v, want only BUG", schema.IssueTypes)
	}

	bugSchema := schema.IssueTypes[0]
	if !containsString(bugSchema.RequiredFields, "title") || !containsString(bugSchema.RequiredFields, "rootCause") {
		t.Fatalf("requiredFields = %+v, want title and custom rootCause", bugSchema.RequiredFields)
	}
	if len(bugSchema.States) != 2 {
		t.Fatalf("states len = %d, want 2", len(bugSchema.States))
	}
	if len(bugSchema.States[0].Transitions) != 1 || bugSchema.States[0].Transitions[0].Name != "Resolved" {
		t.Fatalf("transitions = %+v, want transition to Resolved", bugSchema.States[0].Transitions)
	}
	if len(schema.Labels) != 1 || schema.Labels[0].Name != "cli" {
		t.Fatalf("labels = %+v, want cli label", schema.Labels)
	}
	if !fieldHasOption(bugSchema.Fields, "bugStage", "dev") {
		t.Fatalf("fields = %+v, want bugStage stage option dev", bugSchema.Fields)
	}
}

func TestIssueSchemaCommandShape(t *testing.T) {
	if ISSUESCHEMA.ParentName != "ISSUE" {
		t.Fatalf("issue schema parent = %q, want ISSUE", ISSUESCHEMA.ParentName)
	}
	if ISSUESCHEMA.Name != "schema" {
		t.Fatalf("issue schema name = %q, want schema", ISSUESCHEMA.Name)
	}
	if !hasCommandFlag(ISSUESCHEMA.Flags, "type") {
		t.Fatal("issue schema missing --type")
	}
}

func restoreIssueSchemaStubs(t *testing.T) {
	origGetOrgID := getIssueSchemaOrgID
	origGetProjectID := getIssueSchemaProjectID
	origListStates := listIssueSchemaStates
	origListLabels := listIssueSchemaLabels
	origListProperties := listIssueSchemaProperties
	origListStages := listIssueSchemaStages
	origStdout := issueSchemaStdout
	t.Cleanup(func() {
		getIssueSchemaOrgID = origGetOrgID
		getIssueSchemaProjectID = origGetProjectID
		listIssueSchemaStates = origListStates
		listIssueSchemaLabels = origListLabels
		listIssueSchemaProperties = origListProperties
		listIssueSchemaStages = origListStages
		issueSchemaStdout = origStdout
	})
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func fieldHasOption(fields []issueSchemaField, fieldName string, value string) bool {
	for _, field := range fields {
		if field.Name != fieldName {
			continue
		}
		for _, option := range field.Options {
			if option.Value == value {
				return true
			}
		}
	}
	return false
}
