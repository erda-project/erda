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

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/prettyjson"
)

var ORGINSPECT = command.Command{
	Name:       "inspect",
	ParentName: "ORG",
	ShortHelp:  "Display detailed information of one organization",
	Example: `
  $ erda-cli org inspect 1
`,
	Args: []command.Arg{
		command.StringArg{}.Name("org"),
	},
	Run: OrgInspect,
}

func OrgInspect(ctx *command.Context, org string) error {
	if org == "" {
		org = strconv.FormatUint(ctx.CurrentOrg.ID, 10)

		if org == "" {
			return fmt.Errorf(format.FormatErrMsg("orgs inspect", "should input OrgID", true))
		}
	}

	resp, err := common.GetOrgDetail(ctx, org)
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
