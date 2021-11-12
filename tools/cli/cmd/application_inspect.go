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

var APPLICATIONINSPECT = command.Command{
	Name:       "inspect",
	ParentName: "APPLICATION",
	ShortHelp:  "Inspect application",
	Example:    "erda-cli application inspect",
	Flags: []command.Flag{
		command.IntFlag{Short: "", Name: "org-id", Doc: "the id of an organization ", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "project-id", Doc: "the id of a project ", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "application-id", Doc: "the id of an application ", DefaultValue: 0},
	},
	Run: ApplicationInspect,
}

func ApplicationInspect(ctx *command.Context, orgId, projectId, applicationId int) error {
	if applicationId <= 0 {
		errors.New("invalid application id")
	}

	if projectId <= 0 {
		errors.New("invalid project id")
	}

	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	resp, err := common.GetApplicationDetail(ctx, orgId, projectId, applicationId)
	if err != nil {
		return err
	}

	s, err := prettyjson.Marshal(resp.Data)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("orgs inspect",
			"failed to prettyjson marshal organization data ("+err.Error()+")", false))
	}

	fmt.Println(string(s))

	return nil
}
