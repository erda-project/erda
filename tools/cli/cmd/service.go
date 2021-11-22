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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var SERVICE = command.Command{
	Name:      "service",
	ShortHelp: "List services",
	Example:   "erda-cli service",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.IntFlag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "application-id", Doc: "The id of an application", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "workspace", Doc: "The workspace for runtime", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "runtime", Doc: "The id of an application", DefaultValue: ""},
	},
	Run: ServiceList,
}

func ServiceList(ctx *command.Context, noHeaders bool, orgId, projectId int, workspace, runtime string) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	if projectId <= 0 {
		return errors.New("invalid project id")
	}

	if workspace == "" || runtime == "" {
		return errors.New("invalid workspace or runtime")
	}

	list, err := common.GetSerivceList(ctx, orgId, projectId, workspace, runtime)
	if err != nil {
		return err
	}

	data := [][]string{}
	for n, v := range list {
		data = append(data, []string{
			n,
			v.Status,
			fmt.Sprintf("%.2f", v.Resources.CPU),
			strconv.Itoa(v.Resources.Mem),
			strconv.Itoa(v.Resources.Disk),
			strconv.Itoa(v.Deployments.Replicas),
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"Name", "Status", "CPU", "Memory(MB)", "Disk(MB)", "Replicas",
		})
	}
	return t.Data(data).Flush()
}
