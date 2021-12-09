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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
)

var ORGANIZATIONSSWITCH = command.Command{
	Name:       "switch",
	ParentName: "ORG",
	ShortHelp:  "switch organization",
	Example:    "$ erda-cli org switch <orgID/orgName>",
	Args: []command.Arg{
		command.StringArg{}.Name("orgIdorName"),
	},
	Run: OrganizationsSwitch,
}

func OrganizationsSwitch(ctx *command.Context, org string) error {
	preOrg := ctx.CurrentOrg

	orgResp, err := common.GetOrgDetail(ctx, org)
	if err != nil {
		return err
	}

	// switch org, and save to config
	ctx.CurrentOrg.ID = orgResp.ID
	ctx.CurrentOrg.Name = orgResp.Name
	ctx.CurrentOrg.Desc = orgResp.Desc

	f, conf, err := command.GetConfig()
	if err == dicedir.NotExist {
		return errors.New("Please use 'erda-cli config-set' command to set configurations first")
	} else if err != nil {
		return err
	}

	// TODO make sure api endpoint
	switched := false
	for _, p := range conf.Platforms {
		if p.Server == ctx.CurrentOpenApiHost {
			p.OrgInfo = &ctx.CurrentOrg
			switched = true
			break
		}
	}

	if !switched {
		return errors.New("org switch failed, due to no platform in configuration file")
	}

	if err := command.SetConfig(f, conf); err != nil {
		return fmt.Errorf(
			format.FormatErrMsg(
				"org switch", "failed to switch ("+err.Error()+")", false))
	}

	fmt.Printf("  Before : %-15s(%d)\n", preOrg.Name, preOrg.ID)
	ctx.Succ(color_str.Green(fmt.Sprintf("Current: %-15s(%d)", ctx.CurrentOrg.Name, ctx.CurrentOrg.ID)))

	return nil
}
