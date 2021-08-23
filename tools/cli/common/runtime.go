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

package common

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
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
