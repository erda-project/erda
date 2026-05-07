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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestIssueBindCreatesInclusionRelation(t *testing.T) {
	restoreIssueBindStubs(t)

	getIssueBindOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIssueBindProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}

	var parentID uint64
	var created *apistructs.IssueRelationCreateRequest
	bindIssueRelation = func(_ *command.Context, _ uint64, p uint64, req *apistructs.IssueRelationCreateRequest) error {
		parentID = p
		created = req
		return nil
	}

	var out bytes.Buffer
	issueBindStdout = &out

	if err := IssueBind(&command.Context{}, "", "", 18, "101,102", ""); err != nil {
		t.Fatalf("IssueBind() error = %v", err)
	}
	if parentID != 18 {
		t.Fatalf("parentID = %d, want 18", parentID)
	}
	if created == nil || created.IssueID != 18 || created.ProjectID != 2001 || created.Type != apistructs.IssueRelationInclusion {
		t.Fatalf("created = %+v, want inclusion relation for parent 18/project 2001", created)
	}
	if len(created.RelatedIssue) != 2 || created.RelatedIssue[0] != 101 || created.RelatedIssue[1] != 102 {
		t.Fatalf("children = %+v, want 101,102", created.RelatedIssue)
	}

	var output issueBindOutput
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if output.ParentID != 18 || len(output.Children) != 2 || output.Type != apistructs.IssueRelationInclusion {
		t.Fatalf("output = %+v, want parent/children/type", output)
	}
}

func TestIssueBindCommandShape(t *testing.T) {
	if ISSUEBIND.ParentName != "ISSUE" {
		t.Fatalf("issue bind parent = %q, want ISSUE", ISSUEBIND.ParentName)
	}
	if ISSUEBIND.Name != "bind" {
		t.Fatalf("issue bind name = %q, want bind", ISSUEBIND.Name)
	}
	for _, flag := range []string{"parent", "children", "type"} {
		if !hasCommandFlag(ISSUEBIND.Flags, flag) {
			t.Fatalf("issue bind missing --%s", flag)
		}
	}
}

func restoreIssueBindStubs(t *testing.T) {
	origGetOrgID := getIssueBindOrgID
	origGetProjectID := getIssueBindProjectID
	origBindIssueRelation := bindIssueRelation
	origStdout := issueBindStdout
	t.Cleanup(func() {
		getIssueBindOrgID = origGetOrgID
		getIssueBindProjectID = origGetProjectID
		bindIssueRelation = origBindIssueRelation
		issueBindStdout = origStdout
	})
}
