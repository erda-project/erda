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
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetOrgDetail(ctx *command.Context, org string) (apistructs.OrgFetchResponse, error) {
	var resp apistructs.OrgFetchResponse
	var b bytes.Buffer

	if org == "" {
		return apistructs.OrgFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get organization detail",
			"missing required parameter organization", false))
	}

	response, err := ctx.Get().Path(fmt.Sprintf("/api/orgs/%s", org)).Do().Body(&b)
	if err != nil {
		return apistructs.OrgFetchResponse{}, fmt.Errorf(format.FormatErrMsg(
			"get organization detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.OrgFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get organization detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.OrgFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get organization detail",
			fmt.Sprintf("failed to unmarshal organization detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.OrgFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get organization detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}
