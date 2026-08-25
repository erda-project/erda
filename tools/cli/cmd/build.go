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
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

var PIPELINERUN = command.Command{
	ParentName: "PIPELINE",
	Name:       "run",
	ShortHelp:  "create a pipeline and run it",
	LongHelp:   `With global -V (--verbose): prints the JSON body sent to POST /api/cicds and the full HTTP response body when the call fails or returns success=false, in addition to the HTTP client debug log. Use repeatable --param name=value to pass runtime parameters.`,
	Example:    "$ erda-cli pipeline run <path-to/pipeline.yml>\n$ erda-cli pipeline run pipeline.yml --param verificationHoldAfterPass=60m\n$ erda-cli -V pipeline run <path-to/pipeline.yml>",
	Args: []command.Arg{
		command.StringArg{}.Name("filename"),
	},
	Flags: []command.Flag{
		command.StringFlag{Short: "", Name: "branch", Doc: "branch to create pipeline, default is current branch", DefaultValue: ""},
		command.BoolFlag{Short: "w", Name: "watch", Doc: "watch the status", DefaultValue: false},
		command.StringListFlag{Short: "", Name: "param", Doc: "runtime parameter (repeatable, name=value; values are strings)", DefaultValue: nil},
	},
	ValidArgsFunction:          FilenameCompletion,
	RegisterFlagCompletionFunc: map[string]interface{}{"branch": BranchCompletion},
	Run:                        PipelineRun,
}

func FilenameCompletion(ctx *cobra.Command, args []string, toComplete string, filename, branch string, watch bool, params []string) []string {
	comps := []string{}
	if branch != "" {
		b, err := utils.GetWorkspaceBranch(".")
		if err != nil || branch != b {
			return comps
		}
	}

	p, err := getWorkspacePipelines()
	if err == nil {
		comps = p
	}
	return comps
}

func getWorkspacePipelines() ([]string, error) {
	var pipelineymls []string
	for _, d := range []string{utils.ProjectErdaDir} {
		dir := d + "/pipelines"
		ymls, err := utils.GetWorkspacePipelines(dir)
		if err != nil {
			return pipelineymls, err
		}
		for _, y := range ymls {
			pipelineymls = append(pipelineymls, path.Join(dir, y))
		}
	}

	if _, err := os.Stat("./pipeline.yml"); err == nil {
		pipelineymls = append(pipelineymls, "pipeline.yml")
	}

	return pipelineymls, nil
}

func BranchCompletion(ctx *cobra.Command, args []string, toComplete string, filename, branch string, watch bool, params []string) []string {
	return applicationBranches(".")
}

func applicationBranches(dir string) []string {
	var comps []string

	c1 := exec.Command("git", "branch")
	c1.Dir = dir
	c2 := exec.Command("cut", "-c", "3-")
	output, err := utils.PipeCmds(c1, c2)
	if err == nil {
		splites := strings.Split(output, "\n")
		for _, s := range splites {
			comps = append(comps, s)
		}
	}
	return comps
}

// Create an pipeline and run it
var pipelineStatusForRun = PipelineStatus

func PipelineRun(ctx *command.Context, filename, branch string, watch bool, rawParams []string) error {
	params, err := parsePipelineParams(rawParams)
	if err != nil {
		return err
	}
	// 1. check if current directory is inside a Git work tree
	// 2. parse current branch
	// 3. create pipeline, run it
	gitRepo, err := utils.IsWorkspaceGitRepository(".")
	if err != nil || !gitRepo {
		return errors.New("Current directory is not a local git repository")
	}

	dirty, err := utils.IsWorkspaceDirty(".")
	if err != nil {
		return err
	}
	if dirty {
		return errors.New("Changes should be committed first")
	}

	if branch == "" {
		b, err := utils.GetWorkspaceBranch(".")
		if err != nil {
			return err
		}
		branch = b
	}

	// fetch appID
	info, err := getWorkspaceInfo(".", command.Remote)
	if err != nil {
		return err
	}

	org, err := getOrgDetail(ctx, info.Org)
	if err != nil {
		return err
	}

	_, applicationID, err := resolveWorkspaceApplication(ctx, org.ID, info.Project, info.Application)
	if err != nil {
		return err
	}

	var pipelineResp apistructs.PipelineCreateResponse
	request := pipelineCreateRequest(uint64(applicationID), branch, filename, len(params) > 0)

	if ctx.Debug {
		if reqJSON, err := json.Marshal(request); err == nil {
			ctx.Info("debug: cicds request body: %s", string(reqJSON))
		}
	}

	// create pipeline
	response, err := ctx.Post().Path("/api/cicds").JSONBody(request).Do().JSON(&pipelineResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		if ctx.Debug {
			ctx.Info("debug: cicds response status=%d body=%s", response.StatusCode(), string(response.Body()))
		}
		return errors.Errorf("build fail, status code: %d, err: %+v", response.StatusCode(), pipelineResp.Error)
	}
	if !pipelineResp.Success {
		if ctx.Debug {
			ctx.Info("debug: cicds response status=%d body=%s", response.StatusCode(), string(response.Body()))
		}
		return errors.Errorf("build fail: %+v", pipelineResp.Error)
	}

	if len(params) > 0 {
		if err := runPipelineWithParams(ctx, pipelineResp.Data.ID, params); err != nil {
			return err
		}
	}

	if watch {
		err = watchCreatedPipeline(ctx, branch, pipelineResp.Data.ID, true)
		if err != nil {
			ctx.Fail("failed to watch status of pipeline %d", pipelineResp.Data.ID)
		}
	} else {
		ctx.Succ("run pipeline: %s for branch: %s, pipelineID: %d, you can view building status via `erda-cli pipeline status -i %d`",
			filename, branch, pipelineResp.Data.ID, pipelineResp.Data.ID)
	}

	return nil
}

func watchCreatedPipeline(ctx *command.Context, branch string, pipelineID uint64, watch bool) error {
	if !watch {
		return nil
	}
	return pipelineStatusForRun(ctx, branch, pipelineID, true)
}

func pipelineCreateRequest(appID uint64, branch, filename string, parameterized bool) apistructs.PipelineCreateRequest {
	return apistructs.PipelineCreateRequest{
		AppID:             appID,
		Branch:            branch,
		Source:            apistructs.PipelineSourceDice,
		PipelineYmlSource: apistructs.PipelineYmlSourceGittar,
		PipelineYmlName:   normalizePipelineYmlName(filename),
		AutoRun:           !parameterized,
	}
}

func parsePipelineParams(raw []string) ([]apistructs.PipelineRunParam, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	params := make([]apistructs.PipelineRunParam, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		name, value, ok := strings.Cut(item, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --param %q: expected name=value", item)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("invalid --param %q: parameter name cannot be empty", item)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate --param name %q", name)
		}
		seen[name] = struct{}{}
		params = append(params, apistructs.PipelineRunParam{Name: name, Value: value})
	}
	return params, nil
}

func runPipelineWithParams(ctx *command.Context, pipelineID uint64, params []apistructs.PipelineRunParam) error {
	request := struct {
		PipelineRunParams []apistructs.PipelineRunParam `json:"pipelineRunParams"`
	}{PipelineRunParams: params}
	var runResp apistructs.PipelineRunResponse
	response, err := ctx.Post().Path(fmt.Sprintf("/api/cicds/%d/actions/run", pipelineID)).JSONBody(request).Do().JSON(&runResp)
	if err != nil {
		return fmt.Errorf("run pipeline %d after create: %w", pipelineID, err)
	}
	if !response.IsOK() {
		return fmt.Errorf("run pipeline %d failed, status code: %d, err: %+v", pipelineID, response.StatusCode(), runResp.Error)
	}
	if !runResp.Success {
		return fmt.Errorf("run pipeline %d failed: %+v", pipelineID, runResp.Error)
	}
	return nil
}

func normalizePipelineYmlName(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.TrimPrefix(filename, "./")
	return path.Clean(filename)
}
