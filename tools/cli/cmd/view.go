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
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/utils"
)

var VIEW = command.Command{
	Name:      "view",
	ShortHelp: "View pipeline status",
	Example: ` $ erda-cli view
  $ erda-cli view -i <pipelineId>
  $ erda-cli view --list
  $ erda-cli view --list --page 2 --page-size 10 --statuses Running --sources dice`,
	Flags: []command.Flag{
		command.StringFlag{Short: "b", Name: "branch", Doc: "branch filter: single branch, or comma-separated branches (matches /api/cicds branches); default is current git branch", DefaultValue: ""},
		command.Uint64Flag{Short: "i", Name: "pipelineID", Doc: "specify pipeline id to show pipeline status", DefaultValue: 0},
		command.BoolFlag{Short: "w", Name: "watch", Doc: "watch the status (not with --list)", DefaultValue: false},
		command.BoolFlag{Short: "l", Name: "list", Doc: "list pipelines for the workspace app via /api/cicds", DefaultValue: false},
		command.IntFlag{Short: "", Name: "page", Doc: "for --list: page number (default 1)", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "page-size", Doc: "for --list: page size (default 20)", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "sources", Doc: "for --list: pipeline sources, comma-separated (default dice)", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "statuses", Doc: "for --list: filter by pipeline status", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "yml-names", Doc: "for --list: filter by pipeline yml name(s), comma-separated (query ymlNames)", DefaultValue: ""},
	},
	Run: PipelineView,
}

func PipelineView(ctx *command.Context, branch string, pipelineID uint64, watch bool, list bool, page int, pageSize int, sources string, statuses string, ymlNames string) (err error) {
	if _, err := os.Stat(".git"); err != nil {
		return errors.New("Not a valid git repository directory.")
	}

	if branch == "" {
		b, err := utils.GetWorkspaceBranch(".")
		if err != nil {
			return err
		}
		branch = b
	}

	info, err := utils.GetWorkspaceInfo(".", command.Remote)
	if err != nil {
		return errors.Wrap(err, "failed to get  workspace info")
	}

	org, err := common.GetOrgDetail(ctx, info.Org)
	if err != nil {
		return err
	}

	projectID, applicationID, err := common.ResolveWorkspaceApplication(ctx, org.ID, info.Project, info.Application)
	if err != nil {
		return errors.Wrapf(err, "orgID: %v, projectName: %s, appName: %s", org.ID, info.Project, info.Application)
	}

	if list {
		if watch {
			return errors.New("cannot use --watch with --list")
		}
		if pipelineID > 0 {
			ctx.Info("ignoring --pipelineID when using --list")
		}
		data, err := common.ListPipelinesCICD(ctx, uint64(applicationID), branch, sources, statuses, ymlNames, page, pageSize)
		if err != nil {
			return err
		}
		pn := page
		if pn <= 0 {
			pn = 1
		}
		ps := pageSize
		if ps <= 0 {
			ps = 20
		}
		fmt.Printf("pipelines (appID=%d, total=%d, page=%d, pageSize=%d)\n",
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
		if err = table.NewTable().Header([]string{
			"pipelineID", "status", "branch", "commit", "ymlName", "startedAt",
		}).Data(rows).Flush(); err != nil {
			return err
		}
		fmt.Printf("view one: erda-cli view -i <pipelineID>\n")
		return nil
	}

	if pipelineID <= 0 {
		pipelineID, err = common.GetLatestPipelineID(ctx, uint64(applicationID), branch)
		if err != nil {
			return err
		}
		ctx.Info("no pipeline id provided, using latest pipeline on branch %s: %d", branch, pipelineID)
	}

	if watch {
		if err = common.BuildCheckLoop(ctx, pipelineID); err != nil {
			fmt.Println("[Warn]", err)
		}
	}

	// fetch ymlName path
	var pipelineCombResp apistructs.PipelineInvokedComboResponse
	response, err := ctx.Get().Path("/api/cicds/actions/app-invoked-combos").
		Param("appID", strconv.FormatInt(applicationID, 10)).
		Do().JSON(&pipelineCombResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("status fail, status code: %d, err: %+v", response.StatusCode(), pipelineCombResp.Error)
	}
	if !pipelineCombResp.Success {
		return errors.Errorf("status fail: %+v", pipelineCombResp.Error)
	}
	var ymlName string
	for _, v := range pipelineCombResp.Data {
		if v.Branch == branch {
			for _, item := range v.PagingYmlNames {
				if item != apistructs.DefaultPipelineYmlName &&
					!strings.HasPrefix(item, utils.ProjectPipelineDir) {
					ymlName = item
				}
			}
		}
	}

	// fetch pipeline info
	pipelineInfo, err := common.GetPipeline(ctx, pipelineID)
	if err != nil {
		return err
	}

	var (
		data              [][]string
		currentStageIndex int
		total             = len(pipelineInfo.PipelineStages)
	)
	for i, stage := range pipelineInfo.PipelineStages {
		success := true
		for _, task := range stage.PipelineTasks {
			success = success && apistructs.PipelineStatus(task.Status).IsSuccessStatus()
			data = append(data, []string{
				strconv.FormatUint(pipelineID, 10),
				strconv.FormatUint(task.ID, 10),
				task.Name,
				task.Status,
				task.TimeBegin.AsTime().Format("2006-01-02 15:04:05"),
			})
		}
		if success {
			currentStageIndex = i
		}
	}
	fmt.Printf("Pipeline progress (current/total): %d/%d\n",
		currentStageIndex+1, total)
	if err = table.NewTable().Header([]string{
		"pipelineID", "taskID", "taskName", "taskStatus", "startedAt",
	}).Data(data).Flush(); err != nil {
		return err
	}
	printMetadata(pipelineInfo.PipelineStages)
	seeMore(ctx, info.Org, int(projectID), int(applicationID), branch, pipelineID, ymlName)
	return nil
}

func seeMore(ctx *command.Context, orgName string, projectID, appID int, branch string, pipelineID uint64, pipelineName string) {
	frontUrl := ctx.Get().Path(fmt.Sprintf("/%s/dop/projects/%v/apps/%v/pipeline",
		orgName, projectID, appID)).
		Param("pipelineID", strconv.FormatUint(pipelineID, 10)).
		Param("nodeId", base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{
			strconv.FormatInt(int64(projectID), 10),
			strconv.FormatInt(int64(appID), 10),
			"tree",
			branch,
			pipelineName,
		}, "/")))).GetUrl()
	frontUrl = strings.TrimPrefix(frontUrl, "https://")
	frontUrl = strings.TrimPrefix(frontUrl, "http://")
	if u, err := url.Parse(frontUrl); err == nil {
		u.Host = strings.ReplaceAll(u.Host, "openapi.", "")
		fmt.Printf("see more at %s\n", u.String())
	} else {
		fmt.Printf("failed to parse %s", frontUrl)
	}
}

func printMetadata(pipelineStages []*basepb.PipelineStageDetailDTO) {
	var meta []map[string]map[string]string
	for _, pipelineStage := range pipelineStages {
		s := make(map[string]map[string]string)
		for _, pipelineTask := range pipelineStage.PipelineTasks {
			t := make(map[string]string)
			for _, metadata := range pipelineTask.Result.Metadata {
				t[metadata.Name] = metadata.Value
			}
			s[pipelineTask.Name] = t
		}
		meta = append(meta, s)
	}
	fmt.Println("\nPipeline Metadata")
	_ = yaml.NewEncoder(os.Stdout).Encode(meta)
}
