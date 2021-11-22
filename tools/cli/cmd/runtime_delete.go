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
	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RUNTIMEDELETE = command.Command{
	Name:       "delete",
	ParentName: "RUNTIME",
	ShortHelp:  "Delete runtime",
	Example:    "erda-cli runtime delete",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.IntFlag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "runtime-id", Doc: "The id of a runtime", DefaultValue: 0},
	},
	Run: GetRuntimes,
}

func GetRuntimes(ctx *command.Context, noHeaders bool, orgId, runtimeId int) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	if runtimeId <= 0 {
		return errors.New("invalid runtime id")
	}

	err := common.DeleteRuntime(ctx, orgId, runtimeId)
	if err != nil {
		return err
	}

	ctx.Succ("Runtime deleted.")
	return nil
}
