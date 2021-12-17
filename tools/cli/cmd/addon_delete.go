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
	"time"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var ADDONDELETE = command.Command{
	Name:       "delete",
	ParentName: "ADDON",
	ShortHelp:  "delete addon",
	Example:    "$ erda-cli addon delete --addon-id=<id>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "addon-id", Doc: "the id of an addon", DefaultValue: ""},
		command.BoolFlag{Short: "", Name: "wait-delete", Doc: "if true, wait addon to be deleted", DefaultValue: false},
		command.IntFlag{Short: "", Name: "wait-minutes", Doc: "minutes to wait addon to be deleted", DefaultValue: 3},
	},
	Run: DeleteAddon,
}

func DeleteAddon(ctx *command.Context, orgId uint64, org string, addonId string, waitDelete bool, waitMinutes int) error {
	checkOrgParam(org, orgId)
	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	err = common.DeleteAddon(ctx, orgId, addonId)
	if err != nil {
		return err
	}

	if waitDelete {
		err = dicedir.DoTaskWithTimeout(func() (bool, error) {
			resp, _, err := common.GetAddonResp(ctx, orgId, addonId)
			if err != nil {
				return false, err
			}

			if resp.StatusCode() == 404 {
				return true, nil
			}

			return false, err
		}, time.Duration(waitMinutes)*time.Minute)
	}

	ctx.Succ("Addon deleted.")
	return nil
}
