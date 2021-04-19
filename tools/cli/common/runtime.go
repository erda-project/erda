// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetRuntimeDetail(ctx *command.Context, runtime string, app string) (apistructs.RuntimeInspectResponse, error) {
	var resp apistructs.RuntimeInspectResponse
	var b bytes.Buffer
	var request *httpclient.Request

	if runtime == "" {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(
			format.FormatErrMsg("releases inspect", "missing required parameter runtime", false))
	}

	if app == "" {
		request = ctx.Get().Path(fmt.Sprintf("/api/runtimes/%s", runtime))
	} else {
		request = ctx.Get().Path(fmt.Sprintf("/api/runtimes/%s", runtime)).Param("applicationId", app)
	}

	response, err := request.Do().Body(&b)
	if err != nil {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(format.FormatErrMsg(
			"get runtime detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(format.FormatErrMsg("get runtime detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(format.FormatErrMsg("get runtime detail",
			fmt.Sprintf("failed to unmarshal runtime detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(format.FormatErrMsg("get runtime detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}
