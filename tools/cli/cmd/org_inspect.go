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
	ShortHelp:  "display detail information of one organization",
	Example:    "$ erda-cli org inspect --org=<name>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
	},
	Run: OrgInspect,
}

func OrgInspect(ctx *command.Context, orgId uint64, org string) error {
	checkOrgParam(org, orgId)

	orgIdorName := ""
	if org != "" {
		orgIdorName = org
	} else if orgId > 0 {
		orgIdorName = strconv.FormatUint(orgId, 10)
	} else if orgId <= 0 && ctx.CurrentOrg.ID > 0 {
		orgIdorName = strconv.FormatUint(ctx.CurrentOrg.ID, 10)
	} else {
		return fmt.Errorf(format.FormatErrMsg("org inspect", "invalid Org or OrgID", true))
	}

	o, err := common.GetOrgDetail(ctx, orgIdorName)
	if err != nil {
		return err
	}

	s, err := prettyjson.Marshal(o)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("org inspect",
			"failed to prettyjson marshal organization data ("+err.Error()+")", false))
	}

	fmt.Println(string(s))

	return nil
}
