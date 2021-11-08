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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

var ORG = command.Command{
	Name: "org",
	ShortHelp: "List organizations",
	Example: "erda-cli org",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
	},
	Run: GetOrgs,
}

func GetOrgs(ctx *command.Context, noHeaders bool) error {
	var resp apistructs.OrgSearchResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/orgs").Do().Body(&b)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("orgs", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("failed to unmarshal organizations list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(format.FormatErrMsg("orgs",
			fmt.Sprintf("error code(%s), error message(%s)", resp.Error.Code, resp.Error.Msg), false))
	}

	if resp.Data.Total < 0 {
		return fmt.Errorf(
			format.FormatErrMsg("orgs",
				"the number of organizations can not be less than 0", false))
	}

	if resp.Data.Total == 0 {
		fmt.Printf("no organizations found\n")
		return nil
	}

	data := [][]string{}
	for i := range resp.Data.List {
		data = append(data, []string{
			strconv.FormatUint(resp.Data.List[i].ID, 10),
			resp.Data.List[i].Name,
			resp.Data.List[i].Status,
			resp.Data.List[i].Desc,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"OrgID", "Name", "Status", "Description",
		})
	}
	return t.Data(data).Flush()
}