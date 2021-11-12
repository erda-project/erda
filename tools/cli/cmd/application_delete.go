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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/pkg/errors"
)

var APPLICATIONDELETE = command.Command{
	Name: "delete",
	ParentName: "APPLICATION",
	ShortHelp: "Delete application",
	Example: "erda-cli application delete",
	Flags: []command.Flag{
		command.IntFlag{Short: "", Name: "application-id", Doc: "the id of an application ", DefaultValue: 0},
	},
	Run: ApplicationDelete,
}

func ApplicationDelete(ctx *command.Context, appID int) error {
	if appID <= 0 {
		return errors.New("invalid application id")
	}

	var resp apistructs.ApplicationDeleteResponse
	var b bytes.Buffer

	response, err := ctx.Delete().
		Path(fmt.Sprintf("/api/applications/%d", appID)).Do().Body(&b)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("delete", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("delete",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(format.FormatErrMsg("delete",
			fmt.Sprintf("failed to unmarshal releases remove application response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(format.FormatErrMsg("delete",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	ctx.Succ("Application deleted.")
	return nil
}