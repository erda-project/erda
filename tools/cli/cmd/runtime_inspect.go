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

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/prettyjson"
)

var RUNTIMEINSPECT = command.Command{
	Name:       "inspect",
	ParentName: "RUNTIME",
	ShortHelp:  "Inspect runtime",
	Example:    "erda-cli runtime inspect",
	Flags: []command.Flag{
		command.IntFlag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "application-id", Doc: "The id of an application", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "workspace", Doc: "The workspace of a runtime", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "runtime", Doc: "The id/name of a runtime", DefaultValue: ""},
	},
	Run: RuntimeInspect,
}

func RuntimeInspect(ctx *command.Context, orgId, applicationId int, workspace, runtime string) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	if runtime == "" {
		return errors.New("invalid runtime")
	}

	resp, err := common.GetRuntimeDetail(ctx, orgId, applicationId, workspace, runtime)
	if err != nil {
		return err
	}

	s, err := prettyjson.Marshal(resp.Data)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("runtime inspect",
			"failed to prettyjson marshal runtime data ("+err.Error()+")", false))
	}

	fmt.Println(string(s))
	return nil
}
