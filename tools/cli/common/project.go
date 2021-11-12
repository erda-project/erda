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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetProjectDetail(ctx *command.Context, orgID, projectID int) (apistructs.ProjectDetailResponse, error) {
	var resp apistructs.ProjectDetailResponse
	var b bytes.Buffer

	response, err := ctx.Get().
		Header("Org-ID", strconv.Itoa(orgID)).
		Path(fmt.Sprintf("/api/projects/%d", projectID)).
		Do().Body(&b)
	if err != nil {
		return apistructs.ProjectDetailResponse{}, fmt.Errorf(format.FormatErrMsg(
			"get project detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ProjectDetailResponse{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ProjectDetailResponse{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to unmarshal project detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ProjectDetailResponse{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}

// TODO paging
func GetProjectList(ctx *command.Context, orgId int) ([]apistructs.ProjectDTO, error) {
	var resp apistructs.ProjectListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/projects").
		Param("joined", "true").
		Param("orgId", strconv.Itoa(orgId)).Do().Body(&b)
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
			fmt.Sprintf("failed to unmarshal projects list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return nil, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	if resp.Data.Total < 0 {
		return nil, fmt.Errorf(
			format.FormatErrMsg("list", "critical: the number of projects is less than 0", false))
	}

	//if resp.Data.Total == 0 {
	//	fmt.Printf(format.FormatErrMsg("list", "no projects created\n", false))
	//	return nil, nil
	//}

	return resp.Data.List, nil
}
