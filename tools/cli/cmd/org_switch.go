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

	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/format"
)

var ORGANIZATIONSSWITCH = command.Command{
	Name:       "switch",
	ParentName: "ORG",
	ShortHelp:  "Switch organization",
	Example: `
  $ erda-cli orgs switch 2
`,
	Args: []command.Arg{
		command.StringArg{}.Name("org"),
	},
	Run: RunOrganizationsSwitch,
}

// RunOrganizationsSwitch switches organization
func RunOrganizationsSwitch(ctx *command.Context, org string) error {
	preOrg := ctx.CurrentOrg

	orgResp, err := common.GetOrgDetail(ctx, org)
	if err != nil {
		return err
	}

	// switch org, and save to config
	ctx.CurrentOrg.ID = orgResp.Data.ID
	ctx.CurrentOrg.Name = orgResp.Data.Name
	ctx.CurrentOrg.Desc = orgResp.Data.Desc

	f, conf, err := command.GetConfig()
	if err != nil {
		return err
	}

	// TODO make sure api endpoint
	for _, p := range conf.Platforms {
		if p.Server == ctx.CurrentOpenApiHost {
			p.OrgInfo = &ctx.CurrentOrg
		}
	}

	if err := command.SetConfig(f, conf); err != nil {
		return fmt.Errorf(
			format.FormatErrMsg(
				"orgs switch", "failed to switch ("+err.Error()+")", false))
	}


	fmt.Printf("  Before: %-15s(%d)\n", preOrg.Name, preOrg.ID)
	ctx.Succ(color_str.Green(fmt.Sprintf("Current: %-15s(%d)", ctx.CurrentOrg.Name, ctx.CurrentOrg.ID)))

	return nil
}
