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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestIterationListOutputsTableByDefault(t *testing.T) {
	restoreIterationListStubs(t)

	getIterationListOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIterationListProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}

	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	listProjectIterations = func(_ *command.Context, req *apistructs.IterationPagingRequest, orgID string) (*apistructs.IterationPagingResponseData, error) {
		if req.ProjectID != 2001 || orgID != "1001" {
			t.Fatalf("unexpected request project/org: %+v %s", req, orgID)
		}
		return &apistructs.IterationPagingResponseData{
			Total: 1,
			List: []apistructs.Iteration{
				{ID: 3001, Title: "Sprint 1", State: apistructs.IterationStateUnfiled, StartedAt: &now, FinishedAt: &now},
			},
		}, nil
	}

	var out bytes.Buffer
	iterationListStdout = &out

	if err := IterationList(&command.Context{}, "", "", "", 1, 20, false); err != nil {
		t.Fatalf("IterationList() error = %v", err)
	}

	got := out.String()
	for _, expected := range []string{"iterations (total=1, page=1, pageSize=20)", "ID", "Sprint 1", "UNFILED"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("table output missing %q:\n%s", expected, got)
		}
	}
}

func TestIterationListOutputsJSONWithFlag(t *testing.T) {
	restoreIterationListStubs(t)

	getIterationListOrgID = func(*command.Context, string) (string, uint64, error) {
		return "erda", 1001, nil
	}
	getIterationListProjectID = func(*command.Context, uint64, string) (string, uint64, error) {
		return "erda-project", 2001, nil
	}
	listProjectIterations = func(_ *command.Context, _ *apistructs.IterationPagingRequest, _ string) (*apistructs.IterationPagingResponseData, error) {
		return &apistructs.IterationPagingResponseData{
			Total: 1,
			List: []apistructs.Iteration{
				{ID: 3001, Title: "Sprint 1", State: apistructs.IterationStateUnfiled},
			},
		}, nil
	}

	var out bytes.Buffer
	iterationListStdout = &out

	if err := IterationList(&command.Context{}, "", "", "", 1, 20, true); err != nil {
		t.Fatalf("IterationList() error = %v", err)
	}

	var output iterationListOutput
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if output.Total != 1 || len(output.List) != 1 || output.List[0].Title != "Sprint 1" {
		t.Fatalf("output = %+v, want one iteration", output)
	}
}

func TestIterationListCommandShape(t *testing.T) {
	if ITERATIONLIST.ParentName != "ITERATION" {
		t.Fatalf("iteration list parent = %q, want ITERATION", ITERATIONLIST.ParentName)
	}
	if ITERATIONLIST.Name != "list" {
		t.Fatalf("iteration list name = %q, want list", ITERATIONLIST.Name)
	}
	for _, flag := range []string{"state", "page", "page-size", "json"} {
		if !hasCommandFlag(ITERATIONLIST.Flags, flag) {
			t.Fatalf("iteration list missing --%s", flag)
		}
	}
}

func restoreIterationListStubs(t *testing.T) {
	origGetOrgID := getIterationListOrgID
	origGetProjectID := getIterationListProjectID
	origList := listProjectIterations
	origStdout := iterationListStdout
	t.Cleanup(func() {
		getIterationListOrgID = origGetOrgID
		getIterationListProjectID = origGetProjectID
		listProjectIterations = origList
		iterationListStdout = origStdout
	})
}
