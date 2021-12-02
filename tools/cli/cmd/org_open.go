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

var ORGOPEN = command.Command{
	Name:       "open",
	ParentName: "ORG",
	ShortHelp:  "Open the organization page in browser",
	Example:    "$ erda-cli org open --org=<name>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
	},
	Run: OrgOpen,
}

func OrgOpen(ctx *command.Context, orgId uint64, org string) error {
	checkOrgParam(org, orgId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	err = common.Open(ctx, common.OrgEntity, org, orgId, 0, 0)
	if err != nil {
		return err
	}

	return nil
}
