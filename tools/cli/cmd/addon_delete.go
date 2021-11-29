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
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ADDONDELETE = command.Command{
	Name:       "delete",
	ParentName: "ADDON",
	ShortHelp:  "Delete addon",
	Example:    "$ erda-cli addon delete --addon-id=<id>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "addon-id", Doc: "The id of an addon", DefaultValue: ""},
	},
	Run: DeleteAddon,
}

func DeleteAddon(ctx *command.Context, orgId uint64, org, addonId string) error {
	checkOrgParam(org, orgId)
	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	err = common.DeleteAddon(ctx, orgId, addonId)
	if err != nil {
		return err
	}

	ctx.Succ("Addon deleted.")
	return nil
}
