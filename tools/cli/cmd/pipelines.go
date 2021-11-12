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
	"path"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var PIPELINE = command.Command{
	Name:      "pipeline",
	ShortHelp: "List pipelines in .dice/pipelines directory (current repo)",
	Example:   "erda-cli pipeline",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers",
			Doc:          "When using the default or custom-column output format, don't print headers (default print headers)",
			DefaultValue: false},
	},
	Run: GetPipelines,
}

func GetPipelines(ctx *command.Context, noHeaders bool) error {
	branch, err := common.GetWorkspaceBranch()
	if err != nil {
		return err
	}

	var pipelineymls []string

	// TODO default dir as ".erda" ?
	//erdaDir, err := dicedir.FindProjectErdaDir()
	//if err != nil && err != dicedir.NotExist {
	//	return err
	//} else if err == nil {
	//	ymls, err := common.GetWorkspacePipelines(erdaDir)
	//	if err != nil {
	//		return err
	//	}
	//	for _, y := range ymls {
	//		pipelineymls = append(pipelineymls, path.Join(dicedir.ProjectErdaDir, y))
	//	}
	//}

	// compatible to dice
	diceDir, err := dicedir.FindProjectDiceDir()
	if err != nil && err != dicedir.NotExist {
		return err
	} else if err == nil {
		ymls, err := common.GetWorkspacePipelines(diceDir)
		if err != nil {
			return err
		}
		if len(ymls) > 0 {
			// TODO
			// fmt.Println(color_str.Yellow("Warning! Should rename .dice to .erda"))
		}
		for _, y := range ymls {
			pipelineymls = append(pipelineymls, path.Join(dicedir.ProjectDiceDir, y))
		}
	}

	var data [][]string
	for _, p := range pipelineymls {
		data = append(data, []string{
			branch,
			p,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"branch", "pipeline",
		})
	}
	return t.Data(data).Flush()
}
