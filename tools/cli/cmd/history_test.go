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
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

func TestPipelineHistoryOwnsListQueryFlags(t *testing.T) {
	if hasCommandFlag(STATUS.Flags, "list") {
		t.Fatal("pipeline status should not expose --list; use pipeline history")
	}

	for _, name := range []string{"branch", "page", "page-size", "sources", "statuses", "yml-names"} {
		if !hasCommandFlag(HISTORY.Flags, name) {
			t.Fatalf("pipeline history missing --%s", name)
		}
	}
}

func TestPipelineStatusCommandName(t *testing.T) {
	if STATUS.ParentName != "PIPELINE" {
		t.Fatalf("pipeline status parent = %q, want PIPELINE", STATUS.ParentName)
	}
	if STATUS.Name != "status" {
		t.Fatalf("pipeline status command name = %q, want status", STATUS.Name)
	}
	if !bytes.Contains([]byte(STATUS.Example), []byte("erda-cli pipeline status")) {
		t.Fatalf("pipeline status example = %q, want status usage", STATUS.Example)
	}
}

func TestPipelineHistoryPassesQueryFilters(t *testing.T) {
	origGetWorkspaceBranch := getWorkspaceBranch
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origListPipelinesCICD := listPipelinesCICD
	t.Cleanup(func() {
		getWorkspaceBranch = origGetWorkspaceBranch
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		listPipelinesCICD = origListPipelinesCICD
	})

	getWorkspaceBranch = func(string) (string, error) {
		return "master", nil
	}
	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}

	var got struct {
		appID    uint64
		branch   string
		sources  string
		statuses string
		ymlNames string
		page     int
		pageSize int
	}
	listPipelinesCICD = func(_ *command.Context, appID uint64, branches, sources, statuses, ymlNames string, pageNo, pageSize int) (*apistructs.PipelinePageListData, error) {
		got.appID = appID
		got.branch = branches
		got.sources = sources
		got.statuses = statuses
		got.ymlNames = ymlNames
		got.page = pageNo
		got.pageSize = pageSize
		return &apistructs.PipelinePageListData{}, nil
	}

	err := PipelineHistory(&command.Context{}, "release/1.0", 2, 10, "dice", "Running,Failed", "pipeline.yml")
	if err != nil {
		t.Fatalf("PipelineHistory() error = %v", err)
	}
	if got.appID != 3001 {
		t.Fatalf("appID = %d, want 3001", got.appID)
	}
	if got.branch != "release/1.0" || got.sources != "dice" || got.statuses != "Running,Failed" || got.ymlNames != "pipeline.yml" {
		t.Fatalf("filters = %#v, want passed through", got)
	}
	if got.page != 2 || got.pageSize != 10 {
		t.Fatalf("pagination = page %d pageSize %d, want 2/10", got.page, got.pageSize)
	}
}

func TestWritePipelineHistoryUsesProvidedWriter(t *testing.T) {
	startedAt := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	data := &apistructs.PipelinePageListData{
		Total: 1,
		Pipelines: []apistructs.PagePipeline{
			{
				ID:      1001,
				Status:  apistructs.PipelineStatusSuccess,
				Commit:  "9335c33bf5abc6647fc8ffe5c67b6ed3426f8740",
				YmlName: "pipeline.yml",
				FilterLabels: map[string]string{
					apistructs.LabelBranch: "master",
				},
				TimeBegin: &startedAt,
			},
		},
	}

	var out bytes.Buffer
	if err := writePipelineHistory(&out, 3001, data, 1, 20); err != nil {
		t.Fatalf("writePipelineHistory() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("pipelines (appID=3001, total=1, page=1, pageSize=20)")) {
		t.Fatalf("history output = %q, want header", got)
	}
	if !bytes.Contains([]byte(got), []byte("1001")) || !bytes.Contains([]byte(got), []byte("pipeline.yml")) {
		t.Fatalf("history output = %q, want pipeline row", got)
	}
	if !bytes.Contains([]byte(got), []byte("erda-cli pipeline status -i <pipelineID>")) {
		t.Fatalf("history output = %q, want status hint", got)
	}
}

func hasCommandFlag(flags []command.Flag, name string) bool {
	for _, flag := range flags {
		switch v := flag.(type) {
		case command.IntFlag:
			if v.Name == name {
				return true
			}
		case command.Uint64Flag:
			if v.Name == name {
				return true
			}
		case command.BoolFlag:
			if v.Name == name {
				return true
			}
		case command.StringFlag:
			if v.Name == name {
				return true
			}
		}
	}
	return false
}
