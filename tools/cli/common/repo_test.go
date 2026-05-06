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

package common

import (
	"errors"
	"strings"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func TestResolveWorkspaceApplicationPrefersContextIDs(t *testing.T) {
	origRepoStatsResolver := resolveWorkspaceRepoStats
	t.Cleanup(func() {
		resolveWorkspaceRepoStats = origRepoStatsResolver
	})

	resolveWorkspaceRepoStats = func(*command.Context, uint64, string, string) (apistructs.GittarStatsData, error) {
		t.Fatal("repo stats resolver should not be called when context IDs match")
		return apistructs.GittarStatsData{}, nil
	}

	ctx := &command.Context{
		CurrentProject: command.ProjectInfo{
			Project:   "demo-project",
			ProjectID: 2001,
		},
		CurrentApplication: command.ApplicationInfo{
			Application:   "demo-app",
			ApplicationID: 3001,
		},
	}

	projectID, applicationID, err := ResolveWorkspaceApplication(ctx, 1001, "demo-project", "demo-app")
	if err != nil {
		t.Fatalf("ResolveWorkspaceApplication() error = %v", err)
	}
	if projectID != 2001 || applicationID != 3001 {
		t.Fatalf("ResolveWorkspaceApplication() = (%d, %d), want (2001, 3001)", projectID, applicationID)
	}
}

func TestResolveWorkspaceApplicationFallsBackToRepoStatsWhenContextMissingIDs(t *testing.T) {
	origRepoStatsResolver := resolveWorkspaceRepoStats
	t.Cleanup(func() {
		resolveWorkspaceRepoStats = origRepoStatsResolver
	})

	resolveWorkspaceRepoStats = func(_ *command.Context, orgID uint64, projectName, applicationName string) (apistructs.GittarStatsData, error) {
		if orgID != 1001 || projectName != "demo-project" || applicationName != "demo-app" {
			t.Fatalf("resolveWorkspaceRepoStats() = (%d, %q, %q), want (1001, %q, %q)", orgID, projectName, applicationName, "demo-project", "demo-app")
		}
		return apistructs.GittarStatsData{
			ProjectID:     2001,
			ApplicationID: 3001,
		}, nil
	}

	projectID, applicationID, err := ResolveWorkspaceApplication(&command.Context{}, 1001, "demo-project", "demo-app")
	if err != nil {
		t.Fatalf("ResolveWorkspaceApplication() error = %v", err)
	}
	if projectID != 2001 || applicationID != 3001 {
		t.Fatalf("ResolveWorkspaceApplication() = (%d, %d), want (2001, 3001)", projectID, applicationID)
	}
}

func TestResolveWorkspaceApplicationFallsBackToRepoStatsWhenNamesDoNotMatch(t *testing.T) {
	origRepoStatsResolver := resolveWorkspaceRepoStats
	t.Cleanup(func() {
		resolveWorkspaceRepoStats = origRepoStatsResolver
	})

	resolveWorkspaceRepoStats = func(_ *command.Context, orgID uint64, projectName, applicationName string) (apistructs.GittarStatsData, error) {
		if orgID != 1001 || projectName != "target-project" || applicationName != "target-app" {
			t.Fatalf("resolveWorkspaceRepoStats() = (%d, %q, %q), want (1001, %q, %q)", orgID, projectName, applicationName, "target-project", "target-app")
		}
		return apistructs.GittarStatsData{
			ProjectID:     4001,
			ApplicationID: 5001,
		}, nil
	}

	ctx := &command.Context{
		CurrentProject: command.ProjectInfo{
			Project:   "other-project",
			ProjectID: 2001,
		},
		CurrentApplication: command.ApplicationInfo{
			Application:   "other-app",
			ApplicationID: 3001,
		},
	}

	projectID, applicationID, err := ResolveWorkspaceApplication(ctx, 1001, "target-project", "target-app")
	if err != nil {
		t.Fatalf("ResolveWorkspaceApplication() error = %v", err)
	}
	if projectID != 4001 || applicationID != 5001 {
		t.Fatalf("ResolveWorkspaceApplication() = (%d, %d), want (4001, 5001)", projectID, applicationID)
	}
}

func TestResolveWorkspaceApplicationReturnsExplicitErrorWhenRepoStatsFails(t *testing.T) {
	origRepoStatsResolver := resolveWorkspaceRepoStats
	t.Cleanup(func() {
		resolveWorkspaceRepoStats = origRepoStatsResolver
	})

	resolveWorkspaceRepoStats = func(*command.Context, uint64, string, string) (apistructs.GittarStatsData, error) {
		return apistructs.GittarStatsData{}, errors.New("repo not found")
	}

	_, _, err := ResolveWorkspaceApplication(&command.Context{}, 1001, "demo-project", "demo-app")
	if err == nil {
		t.Fatal("expected error when repo stats cannot resolve workspace")
	}
	if !strings.Contains(err.Error(), "local config or repo stats") {
		t.Fatalf("ResolveWorkspaceApplication() error = %v, want explicit local config/repo stats hint", err)
	}
}
