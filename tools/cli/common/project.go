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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetProjectDetail(ctx *command.Context, project string, orgID string) (apistructs.ProjectDetailResponse, error) {
	var resp apistructs.ProjectDetailResponse
	var b bytes.Buffer
	var projectID string

	if project == "" {
		return apistructs.ProjectDetailResponse{}, fmt.Errorf(
			format.FormatErrMsg("get project detail", "missing required arg project", false))
	}

	projectID = project

	if orgID != "" {
		projectFlag := false
		projects, err := GetProjectList(ctx)
		if err != nil {
			return apistructs.ProjectDetailResponse{}, err
		}
		for i := range projects {
			porgID, err := strconv.ParseUint(orgID, 10, 64)
			if err != nil {
				return apistructs.ProjectDetailResponse{}, fmt.Errorf(
					format.FormatErrMsg("get project detail", err.Error(), false))
			}
			if projects[i].Name == project && projects[i].OrgID == porgID {
				projectID = strconv.FormatUint(projects[i].ID, 10)
				projectFlag = true
				break
			}
		}

		if !projectFlag {
			return apistructs.ProjectDetailResponse{}, fmt.Errorf(format.FormatErrMsg("get project detail",
				"failed to get project from projects list", false))
		}
	}

	response, err := ctx.Get().Path(fmt.Sprintf("/api/projects/%s", projectID)).Do().Body(&b)
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

func GetProjectList(ctx *command.Context) ([]apistructs.ProjectDTO, error) {
	var resp apistructs.ProjectListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/projects").Param("joined", "true").
		Param("orgId", strconv.FormatUint(ctx.Sessions[ctx.CurrentOpenApiHost].OrgID, 10)).Do().Body(&b)
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

	if resp.Data.Total == 0 {
		fmt.Printf(format.FormatErrMsg("list", "no projects created\n", false))
		return nil, nil
	}

	return resp.Data.List, nil
}
