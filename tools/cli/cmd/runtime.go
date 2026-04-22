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
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RUNTIME = command.Command{
	Name:      "runtime",
	ShortHelp: "Runtime operations",
	Example:   `erda-cli runtime status`,
}

type resolvedRuntimeContext struct {
	orgName       string
	orgID         uint64
	projectID     uint64
	applicationID uint64
	application   string
	workspace     string
	runtimeID     uint64
}

var (
	getCurrentBranchWorkspace           = common.GetCurrentBranchWorkspace
	listApplicationRuntimes             = common.ListApplicationRuntimes
	inspectRuntime                      = common.InspectRuntime
	runtimeStdout             io.Writer = os.Stdout
)

func RuntimeList(ctx *command.Context, workspace string, runtimeID uint64) error {
	resolved, err := resolveRuntimeContext(ctx, workspace, runtimeID, false)
	if err != nil {
		return err
	}

	if resolved.runtimeID > 0 {
		runtime, err := inspectRuntime(ctx, resolved.orgID, resolved.runtimeID, resolved.applicationID, resolved.workspace)
		if err != nil {
			return err
		}
		return writeRuntimeList(runtimeStdout, resolved.applicationID, resolved.workspace, []apistructs.RuntimeSummaryDTO{
			{RuntimeInspectDTO: runtime},
		})
	}

	runtimes, err := listApplicationRuntimes(ctx, strconv.FormatUint(resolved.orgID, 10), resolved.applicationID)
	if err != nil {
		return err
	}
	return writeRuntimeList(runtimeStdout, resolved.applicationID, resolved.workspace, filterRuntimesByWorkspace(runtimes, resolved.workspace))
}

func RuntimeStatus(ctx *command.Context, workspace string, runtimeID uint64) error {
	resolved, err := resolveRuntimeContext(ctx, workspace, runtimeID, true)
	if err != nil {
		return err
	}

	runtime, err := inspectRuntime(ctx, resolved.orgID, resolved.runtimeID, resolved.applicationID, resolved.workspace)
	if err != nil {
		return err
	}
	return writeRuntimeStatus(runtimeStdout, runtime)
}

func resolveRuntimeContext(ctx *command.Context, workspace string, runtimeID uint64, requireRuntimeID bool) (*resolvedRuntimeContext, error) {
	info, err := getWorkspaceInfo(".", command.Remote)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workspace info")
	}

	org, err := getOrgDetail(ctx, info.Org)
	if err != nil {
		return nil, err
	}

	projectID, applicationID, err := resolveWorkspaceApplication(ctx, org.ID, info.Org, info.Project, info.Application)
	if err != nil {
		return nil, errors.Wrapf(err, "orgID: %v, projectName: %s, appName: %s", org.ID, info.Project, info.Application)
	}

	resolved := &resolvedRuntimeContext{
		orgName:       info.Org,
		orgID:         org.ID,
		projectID:     projectID,
		applicationID: uint64(applicationID),
		application:   info.Application,
		workspace:     strings.ToUpper(workspace),
		runtimeID:     runtimeID,
	}

	if resolved.workspace == "" {
		branch, err := getWorkspaceBranch(".")
		if err != nil {
			return nil, err
		}
		resolved.workspace, err = getCurrentBranchWorkspace(ctx, resolved.applicationID, branch)
		if err != nil {
			return nil, err
		}
	}

	if resolved.runtimeID > 0 || !requireRuntimeID {
		return resolved, nil
	}

	runtimes, err := listApplicationRuntimes(ctx, strconv.FormatUint(resolved.orgID, 10), resolved.applicationID)
	if err != nil {
		return nil, err
	}
	current, err := pickCurrentRuntime(filterRuntimesByWorkspace(runtimes, resolved.workspace), resolved.workspace)
	if err != nil {
		return nil, err
	}
	resolved.runtimeID = current.ID
	return resolved, nil
}

func filterRuntimesByWorkspace(runtimes []apistructs.RuntimeSummaryDTO, workspace string) []apistructs.RuntimeSummaryDTO {
	if workspace == "" {
		return append([]apistructs.RuntimeSummaryDTO(nil), runtimes...)
	}

	var filtered []apistructs.RuntimeSummaryDTO
	for _, runtime := range runtimes {
		if strings.EqualFold(runtimeWorkspace(runtime.RuntimeInspectDTO.Extra), workspace) {
			filtered = append(filtered, runtime)
		}
	}
	return filtered
}

func pickCurrentRuntime(runtimes []apistructs.RuntimeSummaryDTO, workspace string) (*apistructs.RuntimeSummaryDTO, error) {
	if len(runtimes) == 0 {
		return nil, fmt.Errorf("no runtime found for workspace %s", workspace)
	}

	current := runtimes[0]
	for _, runtime := range runtimes[1:] {
		if runtime.UpdatedAt.After(current.UpdatedAt) {
			current = runtime
			continue
		}
		if runtime.UpdatedAt.Equal(current.UpdatedAt) && runtime.ID > current.ID {
			current = runtime
		}
	}
	return &current, nil
}

func runtimeWorkspace(extra map[string]interface{}) string {
	if extra == nil {
		return ""
	}
	if value, ok := extra["workspace"]; ok {
		return fmt.Sprint(value)
	}
	if value, ok := extra["Workspace"]; ok {
		return fmt.Sprint(value)
	}
	return ""
}

func writeRuntimeList(w io.Writer, applicationID uint64, workspace string, runtimes []apistructs.RuntimeSummaryDTO) error {
	fmt.Fprintf(w, "runtimes (appID=%d, workspace=%s, total=%d)\n", applicationID, workspace, len(runtimes))

	rows := make([][]string, 0, len(runtimes))
	for _, runtime := range runtimes {
		rows = append(rows, []string{
			strconv.FormatUint(runtime.ID, 10),
			runtimeDisplayName(runtime.RuntimeInspectDTO),
			runtimeWorkspace(runtime.RuntimeInspectDTO.Extra),
			runtime.Status,
			runtime.ReleaseVersion,
			formatRuntimeTime(runtime.UpdatedAt),
		})
	}

	return table.NewTable(table.WithWriter(w)).Header([]string{
		"runtimeID", "name", "workspace", "status", "release", "updatedAt",
	}).Data(rows).Flush()
}

func writeRuntimeStatus(w io.Writer, runtime apistructs.RuntimeInspectDTO) error {
	fmt.Fprintln(w, "Runtime")
	fmt.Fprintf(w, "  runtimeID: %d\n", runtime.ID)
	fmt.Fprintf(w, "  name: %s\n", runtime.Name)
	fmt.Fprintf(w, "  application: %s\n", runtime.ApplicationName)
	fmt.Fprintf(w, "  workspace: %s\n", runtimeWorkspace(runtime.Extra))
	fmt.Fprintf(w, "  status: %s\n", runtime.Status)
	fmt.Fprintf(w, "  release: %s\n", runtime.ReleaseVersion)
	fmt.Fprintf(w, "  updatedAt: %s\n", formatRuntimeTime(runtime.UpdatedAt))

	rows := make([][]string, 0, len(runtime.Services))
	names := make([]string, 0, len(runtime.Services))
	for name := range runtime.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		service := runtime.Services[name]
		rows = append(rows, []string{
			name,
			service.Status,
			strconv.Itoa(service.Deployments.Replicas),
		})
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Services")
	return table.NewTable(table.WithWriter(w)).Header([]string{
		"service", "status", "replicas",
	}).Data(rows).Flush()
}

func formatRuntimeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func runtimeDisplayName(runtime apistructs.RuntimeInspectDTO) string {
	if strings.TrimSpace(runtime.Name) != "" {
		return runtime.Name
	}

	workspace := runtimeWorkspace(runtime.Extra)
	if runtime.ApplicationName != "" && workspace != "" {
		return fmt.Sprintf("%s/%s", runtime.ApplicationName, workspace)
	}
	if runtime.ApplicationName != "" {
		return runtime.ApplicationName
	}
	if workspace != "" {
		return workspace
	}
	return "-"
}
