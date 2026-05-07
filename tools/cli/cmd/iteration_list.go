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
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ITERATIONLIST = command.Command{
	ParentName: "ITERATION",
	Name:       "list",
	ShortHelp:  "list current project iterations",
	Example:    "$ erda-cli iteration list",
	Flags: []command.Flag{
		command.StringFlag{Name: "org", Doc: "organization name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "project", Doc: "project name; defaults to current workspace context", DefaultValue: ""},
		command.StringFlag{Name: "state", Doc: "iteration state, FILED or UNFILED", DefaultValue: ""},
		command.IntFlag{Name: "page", Doc: "page number", DefaultValue: 1},
		command.IntFlag{Name: "page-size", Doc: "page size", DefaultValue: 20},
		command.BoolFlag{Name: "json", Doc: "if true, output JSON", DefaultValue: false},
	},
	Run: IterationList,
}

var (
	getIterationListOrgID     = common.GetOrgID
	getIterationListProjectID = common.GetProjectID
	listProjectIterations     = queryProjectIterations
	iterationListStdout       io.Writer = os.Stdout
)

type iterationListOutput struct {
	Total uint64                 `json:"total"`
	List  []apistructs.Iteration `json:"list"`
}

func IterationList(ctx *command.Context, org string, project string, state string, page int, pageSize int, jsonOutput bool) error {
	_, orgID, err := resolveIssueScope(ctx, org, project, getIterationListOrgID, getIterationListProjectID)
	if err != nil {
		return err
	}
	projectID := ctx.CurrentProject.ProjectID

	req, err := buildIterationListRequest(orgID, projectID, state, page, pageSize)
	if err != nil {
		return err
	}
	resp, err := listProjectIterations(ctx, req, strconv.FormatUint(orgID, 10))
	if err != nil {
		return err
	}
	if resp == nil {
		resp = &apistructs.IterationPagingResponseData{}
	}

	if !jsonOutput {
		return writeIterationListTable(iterationListStdout, resp.List, resp.Total, page, pageSize)
	}

	encoder := json.NewEncoder(iterationListStdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(iterationListOutput{Total: resp.Total, List: resp.List})
}

func writeIterationListTable(w io.Writer, list []apistructs.Iteration, total uint64, page int, pageSize int) error {
	_, _ = fmt.Fprintf(w, "iterations (total=%d, page=%d, pageSize=%d)\n", total, page, pageSize)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tTITLE\tSTATE\tSTARTED AT\tFINISHED AT")
	for _, item := range list {
		_, _ = fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Title,
			item.State,
			formatIterationTime(item.StartedAt),
			formatIterationTime(item.FinishedAt),
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if len(list) == 0 {
		_, _ = fmt.Fprintln(w, "No iterations found.")
	}
	return nil
}

func formatIterationTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func buildIterationListRequest(orgID, projectID uint64, state string, page int, pageSize int) (*apistructs.IterationPagingRequest, error) {
	if page <= 0 {
		return nil, fmt.Errorf("--page must be greater than 0")
	}
	if pageSize <= 0 {
		return nil, fmt.Errorf("--page-size must be greater than 0")
	}

	var iterationState apistructs.IterationState
	if state != "" {
		iterationState = apistructs.IterationState(state)
	}

	return &apistructs.IterationPagingRequest{
		PageNo:              uint64(page),
		PageSize:            uint64(pageSize),
		ProjectID:           projectID,
		State:               iterationState,
		WithoutIssueSummary: false,
	}, nil
}

// queryProjectIterations queries project iterations.
func queryProjectIterations(ctx *command.Context, req *apistructs.IterationPagingRequest, orgID string) (*apistructs.IterationPagingResponseData, error) {
	var listResp apistructs.IterationPagingResponse
	resp, err := ctx.Get().
		Path("/api/iterations").
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("pageNo", strconv.FormatUint(req.PageNo, 10)).
		Param("pageSize", strconv.FormatUint(req.PageSize, 10)).
		Param("withoutIssueSummary", strconv.FormatBool(req.WithoutIssueSummary)).
		Param("state", string(req.State)).
		Header("org", orgID).
		Do().JSON(&listResp)
	if err != nil {
		return nil, fmt.Errorf("failed to list iterations: %w", err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, fmt.Errorf("failed to list iterations: %s", listResp.Error.Msg)
	}

	return listResp.Data, nil
}
