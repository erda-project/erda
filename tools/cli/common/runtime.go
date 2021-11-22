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
	"strconv"

	"github.com/erda-project/erda/tools/cli/httputils"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetRuntimeDetail(ctx *command.Context, orgId, applicationId int, workspace, runtime string) (
	apistructs.RuntimeInspectResponse, error) {
	var resp apistructs.RuntimeInspectResponse
	var b bytes.Buffer
	var request *httpclient.Request

	if runtime == "" {
		return apistructs.RuntimeInspectResponse{}, fmt.Errorf(
			format.FormatErrMsg("releases inspect", "missing required parameter runtime", false))
	}

	request = ctx.Get().Path(fmt.Sprintf("/api/runtimes/%s", runtime)).
		Header("Org-ID", strconv.Itoa(orgId)).
		Param("applicationId", strconv.Itoa(applicationId)).
		Param("workspace", workspace)

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

func GetRuntimeList(ctx *command.Context, orgId, applicationId int, workspace, name string) (
	[]apistructs.RuntimeSummaryDTO, error) {
	var resp apistructs.RuntimeListResponse
	var b bytes.Buffer

	req := ctx.Get().Path("/api/runtimes").
		Header("Org-ID", strconv.Itoa(orgId)).
		Param("applicationId", strconv.Itoa(applicationId)).
		Param("workspace", workspace).
		Param("name", name)

	response, err := req.Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf(
			format.FormatErrMsg("list", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to unmarshal runtimes list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return nil, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	if len(resp.Data) == 0 {
		fmt.Printf(format.FormatErrMsg("list", "no runtimes created\n", false))
		return nil, nil
	}

	return resp.Data, nil
}

func DeleteRuntime(ctx *command.Context, orgID, runtimeID int) error {
	r := ctx.Delete().
		Header("Org-ID", strconv.Itoa(orgID)).
		Path(fmt.Sprintf("/api/runtimes/%d", runtimeID))
	resp, err := httputils.DoResp(r)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("remove", "failed to remove runtime, error: "+err.Error(), false))
	}
	if err := resp.ParseData(nil); err != nil {
		return fmt.Errorf(
			format.FormatErrMsg(
				"remove", "failed to parse remove runtime response, error: "+err.Error(), false))
	}

	return nil
}
