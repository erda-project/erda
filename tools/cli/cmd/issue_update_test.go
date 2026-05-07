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
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestIssueUpdateCommandShape(t *testing.T) {
	if ISSUEUPDATE.ParentName != "ISSUE" {
		t.Fatalf("issue update parent = %q, want ISSUE", ISSUEUPDATE.ParentName)
	}
	if ISSUEUPDATE.Name != "update" {
		t.Fatalf("issue update name = %q, want update", ISSUEUPDATE.Name)
	}
	for _, flag := range []string{"state", "state-id"} {
		if !hasCommandFlag(ISSUEUPDATE.Flags, flag) {
			t.Fatalf("issue update missing --%s", flag)
		}
	}
}

func TestIssueUpdateRequiresExactlyOneStateFlag(t *testing.T) {
	restoreIssueUpdateStubs(t)
	ctx := &command.Context{}

	if err := IssueUpdate(ctx, 1001, "", "", "", 0); err == nil {
		t.Fatal("expected error when both --state and --state-id are missing")
	}
	if err := IssueUpdate(ctx, 1001, "", "", "done", 10); err == nil {
		t.Fatal("expected error when both --state and --state-id are set")
	}
}

func TestIssueUpdateByStateName(t *testing.T) {
	restoreIssueUpdateStubs(t)
	ctx := &command.Context{}
	ctx.CurrentOrg.ID = 1001
	ctx.CurrentProject.ProjectID = 2001

	getIssueUpdateIssue = func(*command.Context, uint64, uint64, uint64) (*apistructs.Issue, error) {
		return &apistructs.Issue{
			ID:          3001,
			Type:        apistructs.IssueTypeRequirement,
			State:       1,
			IterationID: 2,
		}, nil
	}
	getIssueUpdateStates = func(*command.Context, uint64, apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
		return []apistructs.IssueStateRelation{
			{IssueStatus: apistructs.IssueStatus{StateID: 1, StateName: "Open"}, StateRelation: []int64{2}},
			{IssueStatus: apistructs.IssueStatus{StateID: 2, StateName: "In Progress"}},
		}, nil
	}

	called := false
	updateIssueState = func(_ *command.Context, _ uint64, req *apistructs.IssueUpdateRequest) error {
		called = true
		if req.State == nil || *req.State != 2 {
			t.Fatalf("state = %v, want 2", req.State)
		}
		if req.IterationID != nil {
			t.Fatalf("iteration id should be nil for state-only update, got %v", *req.IterationID)
		}
		return nil
	}

	if err := IssueUpdate(ctx, 3001, "", "", "in progress", 0); err != nil {
		t.Fatalf("IssueUpdate() error = %v", err)
	}
	if !called {
		t.Fatal("UpdateIssue was not called")
	}
}

func restoreIssueUpdateStubs(t *testing.T) {
	origGetIssue := getIssueUpdateIssue
	origGetStates := getIssueUpdateStates
	origUpdate := updateIssueState
	t.Cleanup(func() {
		getIssueUpdateIssue = origGetIssue
		getIssueUpdateStates = origGetStates
		updateIssueState = origUpdate
	})
}
