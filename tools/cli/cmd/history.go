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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
)

var historyStdout io.Writer = os.Stdout

var HISTORY = command.Command{
	ParentName: "PIPELINE",
	Name:       "history",
	ShortHelp:  "List pipeline run history",
	Example: ` $ erda-cli pipeline history
  $ erda-cli pipeline history --branch master
  $ erda-cli pipeline history --page 2 --page-size 10 --statuses Running --sources dice
  $ erda-cli pipeline history --yml-names pipeline.yml`,
	Flags: []command.Flag{
		command.StringFlag{Short: "b", Name: "branch", Doc: "branch filter: single branch, or comma-separated branches (default is current git branch)", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "page", Doc: "page number (default 1)", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "page-size", Doc: "page size (default 20)", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "sources", Doc: "pipeline sources, comma-separated (default dice)", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "statuses", Doc: "filter by pipeline status", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "yml-names", Doc: "filter by pipeline yml name(s), comma-separated (query ymlNames)", DefaultValue: ""},
	},
	Run: PipelineHistory,
}

func PipelineHistory(ctx *command.Context, branch string, page int, pageSize int, sources string, statuses string, ymlNames string) error {
	if branch == "" {
		b, err := getWorkspaceBranch(".")
		if err != nil {
			return err
		}
		branch = b
	}

	info, err := getWorkspaceInfo(".", command.Remote)
	if err != nil {
		return errors.Wrap(err, "failed to get  workspace info")
	}

	org, err := getOrgDetail(ctx, info.Org)
	if err != nil {
		return err
	}

	_, applicationID, err := resolveWorkspaceApplication(ctx, org.ID, info.Org, info.Project, info.Application)
	if err != nil {
		return errors.Wrapf(err, "orgID: %v, projectName: %s, appName: %s", org.ID, info.Project, info.Application)
	}

	data, err := listPipelinesCICD(ctx, uint64(applicationID), branch, sources, statuses, ymlNames, page, pageSize)
	if err != nil {
		return err
	}
	return writePipelineHistory(historyStdout, applicationID, data, page, pageSize)
}

func writePipelineHistory(w io.Writer, applicationID int64, data *apistructs.PipelinePageListData, page int, pageSize int) error {
	pn := page
	if pn <= 0 {
		pn = 1
	}
	ps := pageSize
	if ps <= 0 {
		ps = 20
	}
	fmt.Fprintf(w, "pipelines (appID=%d, total=%d, page=%d, pageSize=%d)\n",
		applicationID, data.Total, pn, ps)
	var rows [][]string
	for _, p := range data.Pipelines {
		br := p.FilterLabels[apistructs.LabelBranch]
		commit := p.Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		tb := ""
		if p.TimeBegin != nil {
			tb = p.TimeBegin.Format("2006-01-02 15:04:05")
		}
		rows = append(rows, []string{
			strconv.FormatUint(p.ID, 10),
			string(p.Status),
			br,
			commit,
			p.YmlName,
			tb,
		})
	}
	if err := table.NewTable(table.WithWriter(w)).Header([]string{
		"pipelineID", "status", "branch", "commit", "ymlName", "startedAt",
	}).Data(rows).Flush(); err != nil {
		return err
	}
	fmt.Fprintf(w, "view one: erda-cli pipeline status -i <pipelineID>\n")
	return nil
}
